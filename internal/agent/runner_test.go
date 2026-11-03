package agent

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type testTool struct {
	name      string
	run       func(context.Context, string) (string, error)
	retryable bool
}

func (t testTool) Name() string { return t.name }
func (t testTool) Execute(ctx context.Context, input string) (string, error) {
	if t.run == nil {
		return "", nil
	}
	return t.run(ctx, input)
}
func (t testTool) Retryable() bool { return t.retryable }

func quietOptions() Options {
	return Options{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
}

func TestRunnerExecutesRegisteredToolAndReturnsTask(t *testing.T) {
	runner, err := NewRunnerWithOptions(quietOptions(), testTool{
		name: "echo",
		run: func(_ context.Context, input string) (string, error) { return input, nil },
	})
	if err != nil {
		t.Fatalf("NewRunner() error = %v", err)
	}
	result, err := runner.Run(context.Background(), "echo", "hello")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Tool != "echo" || result.Output != "hello" || result.Status != StatusSucceeded {
		t.Fatalf("Run() result = %#v, want successful echo task", result)
	}
	if result.ID == "" || result.Attempts != 1 || result.CreatedAt.IsZero() || result.StartedAt.IsZero() || result.CompletedAt.IsZero() {
		t.Fatalf("task metadata incomplete: %#v", result.Task)
	}
	if result.Duration < 0 {
		t.Fatalf("Duration = %v, want non-negative", result.Duration)
	}
}

func TestRunnerRecordsValidationFailure(t *testing.T) {
	runner, err := NewRunnerWithOptions(quietOptions())
	if err != nil {
		t.Fatal(err)
	}
	result, err := runner.Run(context.Background(), "missing", "hello")
	if err == nil || result.Status != StatusFailed || result.ID == "" || result.CompletedAt.IsZero() {
		t.Fatalf("Run() = %#v, %v; want a structured failed task", result, err)
	}
}

func TestRunnerRejectsEmptyInput(t *testing.T) {
	runner, err := NewRunnerWithOptions(quietOptions(), testTool{name: "echo"})
	if err != nil {
		t.Fatal(err)
	}
	result, err := runner.Run(context.Background(), "echo", "  ")
	if err == nil || result.Status != StatusFailed {
		t.Fatalf("Run() = %#v, %v; want failed task for empty input", result, err)
	}
}

func TestRunnerWrapsToolError(t *testing.T) {
	wantErr := errors.New("tool failed")
	runner, err := NewRunnerWithOptions(quietOptions(), testTool{
		name: "broken",
		run: func(context.Context, string) (string, error) { return "", wantErr },
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := runner.Run(context.Background(), "broken", "work")
	if !errors.Is(err, wantErr) || result.Status != StatusFailed || !strings.Contains(result.Error, wantErr.Error()) {
		t.Fatalf("Run() = %#v, %v; want wrapped tool error", result, err)
	}
}

func TestRunnerRecordsEvents(t *testing.T) {
	var events []Event
	options := quietOptions()
	options.OnEvent = func(event Event) { events = append(events, event) }
	runner, err := NewRunnerWithOptions(options, testTool{name: "echo"})
	if err != nil {
		t.Fatal(err)
	}
	result, err := runner.Run(context.Background(), "echo", "hello")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 5 {
		t.Fatalf("got %d events, want 5: %#v", len(events), events)
	}
	if events[0].Type != "task.created" || events[len(events)-1].Type != "task.succeeded" {
		t.Fatalf("unexpected event sequence: %#v", events)
	}
	for _, event := range events {
		if event.TaskID != result.ID || event.Timestamp.IsZero() {
			t.Fatalf("event missing task metadata: %#v", event)
		}
	}
}

func TestRunnerRetriesOnlyOptedInTools(t *testing.T) {
	var calls atomic.Int32
	options := quietOptions()
	options.Retry = RetryPolicy{MaxAttempts: 3}
	runner, err := NewRunnerWithOptions(options, testTool{
		name: "flaky",
		retryable: true,
		run: func(context.Context, string) (string, error) {
			if calls.Add(1) < 3 {
				return "", errors.New("temporary")
			}
			return "recovered", nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := runner.Run(context.Background(), "flaky", "work")
	if err != nil || result.Output != "recovered" || result.Attempts != 3 || calls.Load() != 3 {
		t.Fatalf("Run() = %#v, %v; calls=%d", result, err, calls.Load())
	}
}

func TestRunnerDoesNotRetryNonRetryableTool(t *testing.T) {
	var calls atomic.Int32
	options := quietOptions()
	options.Retry = RetryPolicy{MaxAttempts: 4}
	runner, err := NewRunnerWithOptions(options, testTool{
		name: "side-effect",
		run: func(context.Context, string) (string, error) {
			calls.Add(1)
			return "", errors.New("failed")
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := runner.Run(context.Background(), "side-effect", "work")
	if err == nil || result.Attempts != 1 || calls.Load() != 1 {
		t.Fatalf("Run() = %#v, %v; calls=%d; want exactly one attempt", result, err, calls.Load())
	}
}

func TestRunnerStopsWhenContextCancelled(t *testing.T) {
	started := make(chan struct{})
	options := quietOptions()
	runner, err := NewRunnerWithOptions(options, testTool{
		name: "blocking",
		run: func(ctx context.Context, _ string) (string, error) {
			close(started)
			<-ctx.Done()
			return "", ctx.Err()
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan Result, 1)
	go func() {
		result, _ := runner.Run(ctx, "blocking", "work")
		done <- result
	}()
	<-started
	cancel()
	select {
	case result := <-done:
		if result.Status != StatusCancelled {
			t.Fatalf("status = %q, want cancelled", result.Status)
		}
	case <-time.After(time.Second):
		t.Fatal("Run() did not stop after cancellation")
	}
}

func TestRunnerEnforcesTimeout(t *testing.T) {
	options := quietOptions()
	options.Timeout = 20 * time.Millisecond
	runner, err := NewRunnerWithOptions(options, testTool{
		name: "blocking",
		run: func(ctx context.Context, _ string) (string, error) {
			<-ctx.Done()
			return "", ctx.Err()
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := runner.Run(context.Background(), "blocking", "work")
	if !errors.Is(err, context.DeadlineExceeded) || result.Status != StatusCancelled {
		t.Fatalf("Run() = %#v, %v; want timed-out cancelled task", result, err)
	}
}

func TestRunnerValidatesRetryPolicy(t *testing.T) {
	options := quietOptions()
	options.Retry = RetryPolicy{MaxAttempts: 6}
	if _, err := NewRunnerWithOptions(options); err == nil {
		t.Fatal("expected retry attempt limit validation error")
	}
}
