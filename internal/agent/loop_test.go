package agent_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/matteojose3112-bit/agentforge/internal/agent"
	"github.com/matteojose3112-bit/agentforge/internal/tools"
)

type fakeProvider struct {
	responses []agent.ModelResponse
	requests  []agent.ModelRequest
	err       error
}

func (f *fakeProvider) Generate(_ context.Context, request agent.ModelRequest) (agent.ModelResponse, error) {
	f.requests = append(f.requests, request)
	if f.err != nil {
		return agent.ModelResponse{}, f.err
	}
	response := f.responses[0]
	f.responses = f.responses[1:]
	return response, nil
}

func TestAgentExecutesPermittedSchemaValidToolCall(t *testing.T) {
	provider := &fakeProvider{responses: []agent.ModelResponse{
		{ToolCalls: []agent.ToolCall{{ID: "call-1", Name: "echo", Arguments: json.RawMessage(`{"text":"hello"}`)}}},
		{Text: "The tool returned hello."},
	}}
	runner, err := agent.NewAgent(provider, "test-model", agent.AllowListPolicy{"tool.echo": true}, 3, tools.AgentEcho{})
	if err != nil {
		t.Fatal(err)
	}
	result, err := runner.Run(context.Background(), "repeat hello")
	if err != nil || result != "The tool returned hello." {
		t.Fatalf("Run() = %q, %v", result, err)
	}
	if len(provider.requests) != 2 || len(provider.requests[0].Tools) != 1 {
		t.Fatalf("provider requests = %#v", provider.requests)
	}
}

func TestAgentDeniesToolWithoutPermission(t *testing.T) {
	provider := &fakeProvider{responses: []agent.ModelResponse{
		{ToolCalls: []agent.ToolCall{{ID: "call-1", Name: "echo", Arguments: json.RawMessage(`{"text":"hello"}`)}}},
	}}
	runner, err := agent.NewAgent(provider, "test-model", agent.AllowListPolicy{}, 2, tools.AgentEcho{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runner.Run(context.Background(), "repeat hello"); err == nil {
		t.Fatal("expected permission denial")
	}
}

func TestAgentRejectsInvalidToolArguments(t *testing.T) {
	provider := &fakeProvider{responses: []agent.ModelResponse{
		{ToolCalls: []agent.ToolCall{{ID: "call-1", Name: "echo", Arguments: json.RawMessage(`{"text":"hello","admin":true}`)}}},
	}}
	runner, err := agent.NewAgent(provider, "test-model", agent.AllowListPolicy{"tool.echo": true}, 2, tools.AgentEcho{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runner.Run(context.Background(), "repeat hello"); err == nil {
		t.Fatal("expected schema validation failure")
	}
}

func TestAgentRejectsUnknownTool(t *testing.T) {
	provider := &fakeProvider{responses: []agent.ModelResponse{
		{ToolCalls: []agent.ToolCall{{ID: "call-1", Name: "shell", Arguments: json.RawMessage(`{}`)}}},
	}}
	runner, err := agent.NewAgent(provider, "test-model", agent.AllowListPolicy{"tool.shell": true}, 2)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runner.Run(context.Background(), "run shell"); err == nil {
		t.Fatal("expected unknown tool failure")
	}
}

type invalidOutputTool struct{}

func (invalidOutputTool) Name() string { return "invalid-output" }
func (invalidOutputTool) Definition() agent.ToolDefinition {
	return agent.ToolDefinition{
		Name: "invalid-output",
		Permission: "tool.invalid-output",
		Parameters: json.RawMessage(`{"type":"object","properties":{},"additionalProperties":false}`),
		OutputSchema: json.RawMessage(`{"type":"object","properties":{"ok":{"type":"boolean"}},"required":["ok"],"additionalProperties":false}`),
	}
}
func (invalidOutputTool) Execute(context.Context, string) (string, error) { return `{"unexpected":true}`, nil }

func TestAgentRejectsInvalidToolOutput(t *testing.T) {
	provider := &fakeProvider{responses: []agent.ModelResponse{
		{ToolCalls: []agent.ToolCall{{ID: "call-1", Name: "invalid-output", Arguments: json.RawMessage(`{}`)}}},
	}}
	runner, err := agent.NewAgent(provider, "test-model", agent.AllowListPolicy{"tool.invalid-output": true}, 2, invalidOutputTool{})
	if err != nil { t.Fatal(err) }
	if _, err := runner.Run(context.Background(), "run it"); err == nil {
		t.Fatal("expected output schema validation failure")
	}
}

func TestAgentPropagatesProviderErrors(t *testing.T) {
	provider := &fakeProvider{err: errors.New("provider offline")}
	runner, err := agent.NewAgent(provider, "test-model", agent.AllowListPolicy{}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runner.Run(context.Background(), "hello"); err == nil {
		t.Fatal("expected provider error")
	}
}


func TestAgentExecutesTextAnalysisTool(t *testing.T) {
	provider := &fakeProvider{responses: []agent.ModelResponse{
		{ToolCalls: []agent.ToolCall{{ID: "call-text-1", Name: "text_analyze", Arguments: json.RawMessage("{\"text\":\"one two three\"}")}}},
		{Text: "The text contains three words."},
	}}
	runner, err := agent.NewAgent(provider, "test-model", agent.AllowListPolicy{"tool.text.analyze": true}, 2, tools.TextAnalyze{})
	if err != nil { t.Fatal(err) }
	result, err := runner.Run(context.Background(), "Analyze this text.")
	if err != nil || result != "The text contains three words." { t.Fatalf("Run() = %q, %v", result, err) }
	if len(provider.requests) != 2 { t.Fatalf("provider requests = %d, want 2", len(provider.requests)) }
	var toolResult struct { Words int `json:"words"` }
	if err := json.Unmarshal([]byte(provider.requests[1].Messages[2].Content), &toolResult); err != nil { t.Fatal(err) }
	if toolResult.Words != 3 { t.Fatalf("tool word count = %d, want 3", toolResult.Words) }
}
