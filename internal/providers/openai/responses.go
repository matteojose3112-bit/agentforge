package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/matteojose3112-bit/agentforge/internal/agent"
)

const DefaultEndpoint = "https://api.openai.com/v1/responses"

// Provider implements the AgentForge provider interface using OpenAI's Responses API.
type Provider struct {
	APIKey   string
	Endpoint string
	Client   *http.Client
}

func New(apiKey string) (*Provider, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("OpenAI API key is required")
	}
	return &Provider{APIKey: apiKey, Endpoint: DefaultEndpoint, Client: &http.Client{Timeout: 90 * time.Second}}, nil
}

type requestBody struct {
	Model string          `json:"model"`
	Input []json.RawMessage `json:"input"`
	Tools []functionTool   `json:"tools,omitempty"`
	ToolChoice string      `json:"tool_choice,omitempty"`
}

type functionTool struct {
	Type        string          `json:"type"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters"`
	Strict      bool            `json:"strict"`
}

type responseBody struct {
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
	Output []struct {
		Type      string `json:"type"`
		Role      string `json:"role"`
		Content   []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		ID        string `json:"id"`
		CallID    string `json:"call_id"`
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"output"`
}

func (p *Provider) Generate(ctx context.Context, request agent.ModelRequest) (agent.ModelResponse, error) {
	if p == nil || strings.TrimSpace(p.APIKey) == "" {
		return agent.ModelResponse{}, errors.New("OpenAI API key is required")
	}
	endpoint := p.Endpoint
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}
	input := make([]json.RawMessage, 0, len(request.Messages))
	for _, message := range request.Messages {
		if message.Role == "assistant" && len(message.ToolCalls) > 0 {
			for _, call := range message.ToolCalls {
				item, err := json.Marshal(map[string]any{
					"type": "function_call", "call_id": call.ID, "name": call.Name, "arguments": string(call.Arguments),
				})
				if err != nil {
					return agent.ModelResponse{}, err
				}
				input = append(input, item)
			}
			continue
		}
		if message.Role == "tool" {
			item, err := json.Marshal(map[string]any{
				"type": "function_call_output", "call_id": message.ToolCallID, "output": message.Content,
			})
			if err != nil {
				return agent.ModelResponse{}, err
			}
			input = append(input, item)
			continue
		}
		item, err := json.Marshal(map[string]any{"role": message.Role, "content": message.Content})
		if err != nil {
			return agent.ModelResponse{}, err
		}
		input = append(input, item)
	}
	body := requestBody{Model: request.Model, Input: input}
	if len(request.Tools) > 0 {
		body.ToolChoice = "auto"
		for _, spec := range request.Tools {
			body.Tools = append(body.Tools, functionTool{
				Type: "function", Name: spec.Name, Description: spec.Description, Parameters: spec.Parameters, Strict: true,
			})
		}
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return agent.ModelResponse{}, fmt.Errorf("encode OpenAI request: %w", err)
	}
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return agent.ModelResponse{}, err
	}
	httpRequest.Header.Set("Authorization", "Bearer "+p.APIKey)
	httpRequest.Header.Set("Content-Type", "application/json")
	client := p.Client
	if client == nil {
		client = http.DefaultClient
	}
	httpResponse, err := client.Do(httpRequest)
	if err != nil {
		return agent.ModelResponse{}, fmt.Errorf("OpenAI request failed: %w", err)
	}
	defer httpResponse.Body.Close()
	responseBytes, err := io.ReadAll(io.LimitReader(httpResponse.Body, 4<<20))
	if err != nil {
		return agent.ModelResponse{}, fmt.Errorf("read OpenAI response: %w", err)
	}
	var decoded responseBody
	if err := json.Unmarshal(responseBytes, &decoded); err != nil {
		return agent.ModelResponse{}, fmt.Errorf("decode OpenAI response (HTTP %d): %w", httpResponse.StatusCode, err)
	}
	if httpResponse.StatusCode < 200 || httpResponse.StatusCode >= 300 {
		if decoded.Error != nil {
			return agent.ModelResponse{}, fmt.Errorf("OpenAI API error (HTTP %d): %s", httpResponse.StatusCode, decoded.Error.Message)
		}
		return agent.ModelResponse{}, fmt.Errorf("OpenAI API returned HTTP %d", httpResponse.StatusCode)
	}
	result := agent.ModelResponse{}
	for _, item := range decoded.Output {
		switch item.Type {
		case "function_call":
			result.ToolCalls = append(result.ToolCalls, agent.ToolCall{
				ID: item.CallID, Name: item.Name, Arguments: json.RawMessage(item.Arguments),
			})
		case "message":
			for _, content := range item.Content {
				if content.Type == "output_text" {
					result.Text += content.Text
				}
			}
		}
	}
	return result, nil
}
