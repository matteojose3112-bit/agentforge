package agent

import (
	"context"
	"errors"
	"testing"
)

type testTool struct {
	name string
	run  func(context.Context, string) (string, error)
}

func (t testTool) Name() string { return t.name }

func (t testTool) Execute(ctx context.Context, input string) (string, error) {
	return t.run(ctx, input)
}

func TestRunnerExecutesRegisteredTool(t *testing.T) {
	runner, err := NewRunner(testTool{
		name: "echo",
		run: func(_ context.Context, input string) (string, error) {
			return input, nil
		},
	})
	if err != nil {
		t.Fatalf("NewRunner() error = %v", err)
	}
	result, err := runner.Run(context.Background(), "echo", "hello")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Tool != "echo" || result.Output != "hello" {
		t.Fatalf("Run() result = %#v, want echo/hello", result)
	}
}

func TestRunnerRejectsUnknownTool(t *testing.T) {
	runner, err := NewRunner()
	if err != nil {
		t.Fatalf("NewRunner() error = %v", err)
	}
	if _, err := runner.Run(context.Background(), "missing", "hello"); err == nil {
		t.Fatal("Run() expected an error for an unknown tool")
	}
}

func TestRunnerRejectsEmptyInput(t *testing.T) {
	runner, err := NewRunner(testTool{name: "echo"})
	if err != nil {
		t.Fatalf("NewRunner() error = %v", err)
	}
	if _, err := runner.Run(context.Background(), "echo", "  "); err == nil {
		t.Fatal("Run() expected an error for empty input")
	}
}

func TestRunnerWrapsToolError(t *testing.T) {
	wantErr := errors.New("tool failed")
	runner, err := NewRunner(testTool{
		name: "broken",
		run: func(context.Context, string) (string, error) {
			return "", wantErr
		},
	})
	if err != nil {
		t.Fatalf("NewRunner() error = %v", err)
	}
	if _, err := runner.Run(context.Background(), "broken", "work"); !errors.Is(err, wantErr) {
		t.Fatalf("Run() error = %v, want wrapped tool error", err)
	}
}
