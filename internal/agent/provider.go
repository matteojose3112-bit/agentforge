package agent

import (
	"context"
	"encoding/json"
)

// Message is a provider-neutral conversation item.
type Message struct {
	Role       string          `json:"role"`
	Content    string          `json:"content,omitempty"`
	ToolCallID string          `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall       `json:"tool_calls,omitempty"`
	Name       string          `json:"name,omitempty"`
}

// ToolCall is a provider-neutral request to execute a named tool.
type ToolCall struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// ToolDefinition describes a tool using a JSON Schema parameter contract.
type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
	Permission  string          `json:"permission"`
}

// ModelRequest is the provider-neutral input to a model.
type ModelRequest struct {
	Model    string
	Messages []Message
	Tools    []ToolDefinition
}

// ModelResponse is the provider-neutral model output.
type ModelResponse struct {
	Text      string
	ToolCalls []ToolCall
}

// Provider is the vendor-independent model interface.
type Provider interface {
	Generate(ctx context.Context, request ModelRequest) (ModelResponse, error)
}

// PermissionPolicy decides whether a tool permission is granted.
type PermissionPolicy interface {
	Allows(ctx context.Context, permission string) bool
}

// AllowListPolicy grants only explicitly listed permission strings.
type AllowListPolicy map[string]bool

func (p AllowListPolicy) Allows(_ context.Context, permission string) bool {
	return p[permission]
}
