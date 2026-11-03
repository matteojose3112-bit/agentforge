package agent

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// Tool is a capability that an agent can invoke to complete a task.
type Tool interface {
	Name() string
	Execute(ctx context.Context, input string) (string, error)
}

// Result is the outcome of one tool-backed task execution.
type Result struct {
	Tool   string
	Output string
}

// Runner executes a task using an explicitly registered tool.
type Runner struct {
	tools map[string]Tool
}

// NewRunner creates a runner and registers the supplied tools.
func NewRunner(tools ...Tool) (*Runner, error) {
	runner := &Runner{tools: make(map[string]Tool, len(tools))}
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

// Run executes input with the named tool.
func (r *Runner) Run(ctx context.Context, toolName, input string) (Result, error) {
	if r == nil {
		return Result{}, errors.New("runner is nil")
	}
	tool, ok := r.tools[toolName]
	if !ok {
		return Result{}, fmt.Errorf("tool %q is not registered", toolName)
	}
	if strings.TrimSpace(input) == "" {
		return Result{}, errors.New("task input cannot be empty")
	}

	output, err := tool.Execute(ctx, input)
	if err != nil {
		return Result{}, fmt.Errorf("execute tool %q: %w", toolName, err)
	}
	return Result{Tool: toolName, Output: output}, nil
}
