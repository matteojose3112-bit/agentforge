package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/matteojose3112-bit/agentforge/internal/agent"
)

func TestProviderSendsResponsesRequestAndParsesToolCall(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/responses" { t.Fatalf("request = %s %s", r.Method, r.URL.Path) }
		if r.Header.Get("Authorization") != "Bearer test-key" { t.Fatal("missing bearer authorization") }
		var body struct {
			Model string `json:"model"`
			Tools []struct { Type string `json:"type"`; Strict bool `json:"strict"` } `json:"tools"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil { t.Fatal(err) }
		if body.Model != "test-model" || len(body.Tools) != 1 || body.Tools[0].Type != "function" || !body.Tools[0].Strict { t.Fatalf("request body = %#v", body) }
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"output":[{"type":"function_call","call_id":"call-1","name":"echo","arguments":"{\"text\":\"hi\"}"}]}`))
	}))
	defer server.Close()
	provider := &Provider{APIKey:"test-key", Endpoint:server.URL+"/v1/responses", Client:server.Client()}
	response, err := provider.Generate(context.Background(), agent.ModelRequest{
		Model:"test-model", Messages:[]agent.Message{{Role:"user", Content:"say hi"}},
		Tools:[]agent.ToolDefinition{{Name:"echo", Parameters:json.RawMessage(`{"type":"object","properties":{"text":{"type":"string"}},"required":["text"],"additionalProperties":false}`)}},
	})
	if err != nil { t.Fatal(err) }
	if len(response.ToolCalls) != 1 || response.ToolCalls[0].Name != "echo" || string(response.ToolCalls[0].Arguments) != `{"text":"hi"}` { t.Fatalf("response = %#v", response) }
}

func TestProviderReturnsAPIErrorMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"message":"invalid API key"}}`))
	}))
	defer server.Close()
	provider := &Provider{APIKey:"bad", Endpoint:server.URL, Client:server.Client()}
	_, err := provider.Generate(context.Background(), agent.ModelRequest{Model:"test", Messages:[]agent.Message{{Role:"user", Content:"hello"}}})
	if err == nil || !strings.Contains(err.Error(), "invalid API key") { t.Fatalf("Generate() error = %v", err) }
}

func TestProviderRejectsMissingAPIKey(t *testing.T) {
	provider := &Provider{}
	if _, err := provider.Generate(context.Background(), agent.ModelRequest{}); err == nil { t.Fatal("expected missing key error") }
}


func TestNewRejectsInvalidAPIKeyConfiguration(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want string
	}{
		{name: "missing", key: "", want: "OPENAI_API_KEY is not set"},
		{name: "template placeholder", key: "replace-with-your-openai-api-key", want: "template placeholder"},
		{name: "bearer prefix", key: "Bearer sk-example", want: "without the 'Bearer ' prefix"},
		{name: "surrounding whitespace", key: " sk-example ", want: "contains whitespace"},
		{name: "embedded whitespace", key: "sk-example key", want: "contains whitespace"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := New(test.key)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("New() error = %v, want it to contain %q", err, test.want)
			}
		})
	}
}

func TestNewAcceptsNonPlaceholderAPIKey(t *testing.T) {
	provider, err := New("sk-test-example")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if provider.APIKey != "sk-test-example" {
		t.Fatal("provider did not retain the configured key")
	}
}
