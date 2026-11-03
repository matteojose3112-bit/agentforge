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
