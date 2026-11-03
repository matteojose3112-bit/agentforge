package agent

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// Tool is a capability that an agent can invoke to complete a task.
type Tool interface {
	Name() string
	Execute(ctx context.Context, input string) (string, error)
}

// RetryableTool marks a tool as safe to retry after an execution error.
// Tools that perform non-idempotent side effects should not implement it.
type RetryableTool interface {
	Retryable() bool
}

// TaskStatus describes the lifecycle state of one execution.
type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusRunning   TaskStatus = "running"
	StatusSucceeded TaskStatus = "succeeded"
	StatusFailed    TaskStatus = "failed"
	StatusCancelled TaskStatus = "cancelled"
)

// Task is the structured record returned for each execution.
type Task struct {
	ID          string
	Tool        string
	Status      TaskStatus
	CreatedAt   time.Time
	StartedAt   time.Time
	CompletedAt time.Time
	Attempts    int
	Output      string
	Error       string
}

// Result is the structured outcome of one tool-backed task execution.
type Result struct {
	Task
	Duration time.Duration
}

// Event records a meaningful transition or execution attempt.
type Event struct {
	TaskID    string
	Type      string
	Tool      string
	Attempt   int
	Timestamp time.Time
	Error     string
}

// RetryPolicy bounds retries for tools that explicitly opt in.
type RetryPolicy struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
}

// Options configures task execution. Zero values preserve safe defaults.
type Options struct {
	Timeout     time.Duration
	Retry       RetryPolicy
	Logger      *slog.Logger
	OnEvent     func(Event)
}

// Runner executes tasks using explicitly registered tools.
type Runner struct {
	tools   map[string]Tool
	options Options
}

// NewRunner creates a runner with safe defaults and registers the supplied tools.
func NewRunner(tools ...Tool) (*Runner, error) {
	return NewRunnerWithOptions(Options{}, tools...)
}

// NewRunnerWithOptions creates a configured runner and registers the supplied tools.
func NewRunnerWithOptions(options Options, tools ...Tool) (*Runner, error) {
	if options.Timeout < 0 {
		return nil, errors.New("timeout cannot be negative")
	}
	if options.Retry.MaxAttempts == 0 {
		options.Retry.MaxAttempts = 1
	}
	if options.Retry.MaxAttempts < 1 || options.Retry.MaxAttempts > 5 {
		return nil, errors.New("retry max attempts must be between 1 and 5")
	}
	if options.Retry.InitialDelay < 0 || options.Retry.MaxDelay < 0 {
		return nil, errors.New("retry delays cannot be negative")
	}
	if options.Retry.MaxDelay > 0 && options.Retry.InitialDelay > options.Retry.MaxDelay {
		return nil, errors.New("retry initial delay cannot exceed max delay")
	}
	if options.Logger == nil {
		options.Logger = slog.Default()
	}

	runner := &Runner{tools: make(map[string]Tool, len(tools)), options: options}
	for _, tool := range tools {
		if tool == nil {
			return nil, errors.New("cannot register a nil tool")
		}
		name := strings.TrimSpace(tool.Name())
		if name == "" {
			return nil, errors.New("tool name cannot be empty")
		}
		if _, exists := runner.tools[name]; exists {
			return nil, fmt.Errorf("tool %q is already registered", name)
		}
		runner.tools[name] = tool
	}
	return runner, nil
}

// Run executes input with the named tool and returns a complete task record.
// Retries occur only when the tool implements RetryableTool and returns true.
func (r *Runner) Run(ctx context.Context, toolName, input string) (Result, error) {
	if r == nil {
		return Result{}, errors.New("runner is nil")
	}
	if ctx == nil {
		return Result{}, errors.New("context cannot be nil")
	}

	now := time.Now().UTC()
	task := Task{ID: newTaskID(), Tool: toolName, Status: StatusPending, CreatedAt: now}
	r.emit(Event{TaskID: task.ID, Type: "task.created", Tool: toolName, Timestamp: now})

	tool, ok := r.tools[toolName]
	if !ok {
		return r.finish(task, StatusFailed, fmt.Errorf("tool %q is not registered", toolName))
	}
	if strings.TrimSpace(input) == "" {
		return r.finish(task, StatusFailed, errors.New("task input cannot be empty"))
	}

	execCtx := ctx
	cancel := func() {}
	if r.options.Timeout > 0 {
		execCtx, cancel = context.WithTimeout(ctx, r.options.Timeout)
	}
	defer cancel()

	task.Status = StatusRunning
	task.StartedAt = time.Now().UTC()
	r.emit(Event{TaskID: task.ID, Type: "task.started", Tool: toolName, Timestamp: task.StartedAt})

	maxAttempts := 1
	if retryable, ok := tool.(RetryableTool); ok && retryable.Retryable() {
		maxAttempts = r.options.Retry.MaxAttempts
	}

	var output string
	var runErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := execCtx.Err(); err != nil {
			runErr = err
			break
		}
		task.Attempts = attempt
		r.emit(Event{TaskID: task.ID, Type: "attempt.started", Tool: toolName, Attempt: attempt, Timestamp: time.Now().UTC()})
		output, runErr = tool.Execute(execCtx, input)
		if runErr == nil {
			break
		}
		r.emit(Event{TaskID: task.ID, Type: "attempt.failed", Tool: toolName, Attempt: attempt, Timestamp: time.Now().UTC(), Error: runErr.Error()})
		if attempt == maxAttempts || execCtx.Err() != nil {
			break
		}
		delay := retryDelay(r.options.Retry, attempt)
		if delay > 0 {
			r.emit(Event{TaskID: task.ID, Type: "retry.scheduled", Tool: toolName, Attempt: attempt, Timestamp: time.Now().UTC()})
			timer := time.NewTimer(delay)
			select {
			case <-execCtx.Done():
				if !timer.Stop() {
					<-timer.C
				}
				runErr = execCtx.Err()
			case <-timer.C:
			}
			if execCtx.Err() != nil {
				runErr = execCtx.Err()
				break
			}
		}
	}

	if runErr != nil {
		if execCtx.Err() != nil {
			runErr = execCtx.Err()
		}
		status := StatusFailed
		if errors.Is(runErr, context.Canceled) || errors.Is(runErr, context.DeadlineExceeded) {
			status = StatusCancelled
		}
		return r.finish(task, status, fmt.Errorf("execute tool %q: %w", toolName, runErr))
	}
	task.Output = output
	return r.finish(task, StatusSucceeded, nil)
}

func (r *Runner) finish(task Task, status TaskStatus, err error) (Result, error) {
	task.Status = status
	task.CompletedAt = time.Now().UTC()
	if err != nil {
		task.Error = err.Error()
	}
	result := Result{Task: task, Duration: task.CompletedAt.Sub(task.CreatedAt)}
	eventType := "task." + string(status)
	event := Event{TaskID: task.ID, Type: eventType, Tool: task.Tool, Attempt: task.Attempts, Timestamp: task.CompletedAt}
	if err != nil {
		event.Error = err.Error()
	}
	r.emit(event)
	return result, err
}

func (r *Runner) emit(event Event) {
	r.options.Logger.Info("agent execution event",
		"event", event.Type,
		"task_id", event.TaskID,
		"tool", event.Tool,
		"attempt", event.Attempt,
		"timestamp", event.Timestamp,
		"error", event.Error,
	)
	if r.options.OnEvent != nil {
		r.options.OnEvent(event)
	}
}

func retryDelay(policy RetryPolicy, attempt int) time.Duration {
	delay := policy.InitialDelay
	if delay <= 0 {
		return 0
	}
	for i := 1; i < attempt; i++ {
		if policy.MaxDelay > 0 && delay >= policy.MaxDelay/2 {
			return policy.MaxDelay
		}
		delay *= 2
	}
	if policy.MaxDelay > 0 && delay > policy.MaxDelay {
		return policy.MaxDelay
	}
	return delay
}

func newTaskID() string {
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		// Keep task creation available even if the system random source fails.
		return fmt.Sprintf("task-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(id[:])
}
