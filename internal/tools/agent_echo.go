package tools

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/matteojose3112-bit/agentforge/internal/agent"
)

// AgentEcho is a JSON-schema-described echo capability for model tool calling.
type AgentEcho struct{}

func (AgentEcho) Name() string { return "echo" }

func (AgentEcho) Definition() agent.ToolDefinition {
	return agent.ToolDefinition{
		Name: "echo",
		Description: "Return the supplied text unchanged.",
		Permission: "tool.echo",
		Parameters: json.RawMessage(`{"type":"object","properties":{"text":{"type":"string"}},"required":["text"],"additionalProperties":false}`),
	}
}

func (AgentEcho) Execute(_ context.Context, input string) (string, error) {
	var args struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", err
	}
	if args.Text == "" {
		return "", errors.New("text cannot be empty")
	}
	return args.Text, nil
}
