package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

const DefaultMaxModelTurns = 5

// Agent coordinates a provider-neutral model with explicitly registered tools.
type Agent struct {
	provider Provider
	tools    map[string]Tool
	specs    map[string]ToolDefinition
	policy   PermissionPolicy
	model    string
	maxTurns int
}

// NewAgent creates an agent. A nil policy denies every tool call.
func NewAgent(provider Provider, model string, policy PermissionPolicy, maxTurns int, tools ...Tool) (*Agent, error) {
	if provider == nil {
		return nil, errors.New("model provider is required")
	}
	if strings.TrimSpace(model) == "" {
		return nil, errors.New("model name is required")
	}
	if maxTurns == 0 {
		maxTurns = DefaultMaxModelTurns
	}
	if maxTurns < 1 || maxTurns > 10 {
		return nil, errors.New("max model turns must be between 1 and 10")
	}
	agent := &Agent{provider: provider, model: model, policy: policy, maxTurns: maxTurns,
		tools: make(map[string]Tool), specs: make(map[string]ToolDefinition)}
	for _, tool := range tools {
		described, ok := tool.(interface{ Definition() ToolDefinition })
		if !ok || tool == nil {
			return nil, errors.New("agent tools must provide a Definition")
		}
		spec := described.Definition()
		if strings.TrimSpace(spec.Name) == "" || spec.Permission == "" {
			return nil, errors.New("tool name and permission are required")
		}
		if _, exists := agent.tools[spec.Name]; exists {
			return nil, fmt.Errorf("tool %q is already registered", spec.Name)
		}
		if err := validateSchemaDefinition(spec.Parameters); err != nil {
			return nil, fmt.Errorf("tool %q input schema: %w", spec.Name, err)
		}
		if err := validateSchemaDefinition(spec.OutputSchema); err != nil {
			return nil, fmt.Errorf("tool %q output schema: %w", spec.Name, err)
		}
		agent.tools[spec.Name] = tool
		agent.specs[spec.Name] = spec
	}
	return agent, nil
}

// Run sends a prompt to the model and executes only validated, permitted tool calls.
func (a *Agent) Run(ctx context.Context, prompt string) (string, error) {
	if a == nil {
		return "", errors.New("agent is nil")
	}
	if ctx == nil {
		return "", errors.New("context cannot be nil")
	}
	if strings.TrimSpace(prompt) == "" {
		return "", errors.New("prompt cannot be empty")
	}
	messages := []Message{{Role: "user", Content: prompt}}
	definitions := make([]ToolDefinition, 0, len(a.specs))
	for _, spec := range a.specs {
		definitions = append(definitions, spec)
	}
	sort.Slice(definitions, func(i, j int) bool { return definitions[i].Name < definitions[j].Name })
	executedTools := make(map[string]bool)
	for turn := 0; turn < a.maxTurns; turn++ {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		response, err := a.provider.Generate(ctx, ModelRequest{Model: a.model, Messages: messages, Tools: definitions})
		if err != nil {
			return "", fmt.Errorf("generate model response: %w", err)
		}
		if len(response.ToolCalls) == 0 {
			return response.Text, nil
		}
		if turn == a.maxTurns-1 {
			return "", fmt.Errorf("model exceeded maximum of %d tool turns", a.maxTurns)
		}
		// Preflight the complete batch so a dependent tool cannot run alongside
		// its prerequisite in the same model response.
		requestedInBatch := make(map[string]bool, len(response.ToolCalls))
		for _, call := range response.ToolCalls { requestedInBatch[call.Name] = true }
		for _, call := range response.ToolCalls {
			spec, exists := a.specs[call.Name]
			if !exists { return "", fmt.Errorf("model requested unregistered tool %q", call.Name) }
			for _, prerequisite := range spec.RequiresNewUserTurnAfter {
				if executedTools[prerequisite] || requestedInBatch[prerequisite] {
					return "", fmt.Errorf("tool %q requires a new user turn after %q; ask the operator to review and confirm the prerequisite result before continuing", call.Name, prerequisite)
				}
			}
		}
		messages = append(messages, Message{Role: "assistant", ToolCalls: response.ToolCalls})
		for _, call := range response.ToolCalls {
			tool, exists := a.tools[call.Name]
			if !exists {
				return "", fmt.Errorf("model requested unregistered tool %q", call.Name)
			}
			spec := a.specs[call.Name]
			if a.policy == nil || !a.policy.Allows(ctx, spec.Permission) {
				return "", fmt.Errorf("permission denied for tool %q (%s)", call.Name, spec.Permission)
			}
			if err := ValidateJSON(spec.Parameters, call.Arguments); err != nil {
				return "", fmt.Errorf("invalid arguments for tool %q: %w", call.Name, err)
			}
			if err := ctx.Err(); err != nil {
				return "", err
			}
			output, err := tool.Execute(ctx, string(call.Arguments))
			if err != nil {
				return "", fmt.Errorf("execute tool %q: %w", call.Name, err)
			}
			if err := ValidateJSON(spec.OutputSchema, []byte(output)); err != nil {
				return "", fmt.Errorf("tool %q returned invalid output: %w", call.Name, err)
			}
			messages = append(messages, Message{Role: "tool", ToolCallID: call.ID, Name: call.Name, Content: output})
			executedTools[call.Name] = true
		}
	}
	return "", fmt.Errorf("model exceeded maximum of %d tool turns", a.maxTurns)
}

func validateSchemaDefinition(schema json.RawMessage) error {
	var parsed map[string]any
	if len(schema) == 0 || json.Unmarshal(schema, &parsed) != nil {
		return errors.New("parameters must be a valid JSON schema object")
	}
	if parsed["type"] != "object" {
		return errors.New("tool parameters schema must have type object")
	}
	if _, ok := parsed["properties"].(map[string]any); !ok {
		return errors.New("tool parameters schema must define properties")
	}
	return nil
}
