package agent_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
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


func TestAgentExecutesPermittedCSVAnalysisTool(t *testing.T) {
	csvInput := "name,amount\nAlpha,10\nBeta,20\n"
	arguments, err := json.Marshal(map[string]string{"csv": csvInput})
	if err != nil {
		t.Fatal(err)
	}
	provider := &fakeProvider{responses: []agent.ModelResponse{
		{ToolCalls: []agent.ToolCall{{ID: "call-csv-1", Name: "csv_analyze", Arguments: arguments}}},
		{Text: "The CSV contains two rows and one numeric column."},
	}}
	runner, err := agent.NewAgent(provider, "test-model", agent.AllowListPolicy{"tool.csv.analyze": true}, 2, tools.CSVAnalyze{})
	if err != nil {
		t.Fatal(err)
	}
	result, err := runner.Run(context.Background(), "Analyze this CSV.")
	if err != nil || result != "The CSV contains two rows and one numeric column." {
		t.Fatalf("Run() = %q, %v", result, err)
	}
	if len(provider.requests) != 2 {
		t.Fatalf("provider requests = %d, want 2", len(provider.requests))
	}
	var toolResult struct {
		Rows int `json:"rows"`
		Columns int `json:"columns"`
	}
	if err := json.Unmarshal([]byte(provider.requests[1].Messages[2].Content), &toolResult); err != nil {
		t.Fatal(err)
	}
	if toolResult.Rows != 2 || toolResult.Columns != 2 {
		t.Fatalf("unexpected CSV tool result: %+v", toolResult)
	}
}

func TestAgentExecutesPermittedLeadPipelineAnalysisTool(t *testing.T) {
	csvInput := "source,stage,value\nGoogle Ads,Won,100\nReferral,Open,50\nGoogle Ads,Lost,25\n"
	arguments, err := json.Marshal(map[string]any{
		"csv": csvInput,
		"status_column": "stage",
		"value_column": "value",
		"source_column": "source",
		"won_statuses": []string{"Won"},
		"lost_statuses": []string{"Lost"},
	})
	if err != nil {
		t.Fatal(err)
	}
	provider := &fakeProvider{responses: []agent.ModelResponse{
		{ToolCalls: []agent.ToolCall{{ID: "call-leads-1", Name: "lead_pipeline_analyze", Arguments: arguments}}},
		{Text: "The pipeline has three leads and a 50% closed-lead win rate."},
	}}
	runner, err := agent.NewAgent(provider, "test-model", agent.AllowListPolicy{"tool.leads.analyze": true}, 2, tools.LeadPipelineAnalyze{})
	if err != nil {
		t.Fatal(err)
	}
	result, err := runner.Run(context.Background(), "Analyze this CRM pipeline.")
	if err != nil || result != "The pipeline has three leads and a 50% closed-lead win rate." {
		t.Fatalf("Run() = %q, %v", result, err)
	}
	if len(provider.requests) != 2 || len(provider.requests[0].Tools) != 1 || provider.requests[0].Tools[0].Name != "lead_pipeline_analyze" {
		t.Fatalf("provider requests = %#v", provider.requests)
	}
	var toolResult struct {
		TotalLeads int `json:"total_leads"`
		WonLeads int `json:"won_leads"`
		LostLeads int `json:"lost_leads"`
		OpenLeads int `json:"open_leads"`
		WinRatePercent float64 `json:"win_rate_percent"`
		OpenPipelineValue float64 `json:"open_pipeline_value"`
	}
	if err := json.Unmarshal([]byte(provider.requests[1].Messages[2].Content), &toolResult); err != nil {
		t.Fatal(err)
	}
	if toolResult.TotalLeads != 3 || toolResult.WonLeads != 1 || toolResult.LostLeads != 1 || toolResult.OpenLeads != 1 || toolResult.WinRatePercent != 50 || toolResult.OpenPipelineValue != 50 {
		t.Fatalf("unexpected lead pipeline tool result: %+v", toolResult)
	}
}

func TestAgentDeniesLeadPipelineAnalysisWithoutPermission(t *testing.T) {
	arguments, err := json.Marshal(map[string]any{
		"csv": "stage,value\nWon,100\n",
		"status_column": "stage",
		"value_column": "value",
		"source_column": "",
		"won_statuses": []string{},
		"lost_statuses": []string{},
	})
	if err != nil {
		t.Fatal(err)
	}
	provider := &fakeProvider{responses: []agent.ModelResponse{
		{ToolCalls: []agent.ToolCall{{ID: "call-leads-denied", Name: "lead_pipeline_analyze", Arguments: arguments}}},
	}}
	runner, err := agent.NewAgent(provider, "test-model", agent.AllowListPolicy{}, 2, tools.LeadPipelineAnalyze{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runner.Run(context.Background(), "Analyze this CRM pipeline."); err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("Run() error = %v, want permission denial", err)
	}
}

func TestAgentRequiresNewUserTurnBetweenCSVInspectionAndSalesReport(t *testing.T) {
	csv := "stage,value,source\\nQualified,1200,Referral\\nWon,2500,Google Ads\\n"
	inspectArgs, err := json.Marshal(map[string]string{"csv": csv})
	if err != nil { t.Fatal(err) }
	reportArgs, err := json.Marshal(map[string]any{
		"csv": csv, "status_column": "stage", "value_column": "value", "source_column": "source",
		"won_statuses": []string{"Won"}, "lost_statuses": []string{"Lost"},
	})
	if err != nil { t.Fatal(err) }

	t.Run("same response is blocked before either tool runs", func(t *testing.T) {
		provider := &fakeProvider{responses: []agent.ModelResponse{{
			ToolCalls: []agent.ToolCall{
				{ID: "inspect-1", Name: "csv_inspect", Arguments: inspectArgs},
				{ID: "report-1", Name: "sales_ops_report", Arguments: reportArgs},
			},
		}}}
		runner, err := agent.NewAgent(provider, "test-model", agent.AllowListPolicy{
			"tool.csv.inspect": true, "tool.sales_ops.report": true,
		}, 3, tools.CSVInspect{}, tools.SalesOpsReport{})
		if err != nil { t.Fatal(err) }
		_, err = runner.Run(context.Background(), "Inspect this CRM export and prepare a report.")
		if err == nil || !strings.Contains(err.Error(), "requires a new user turn") {
			t.Fatalf("Run() error = %v, want confirmation turn boundary", err)
		}
		if len(provider.requests) != 1 { t.Fatalf("provider requests = %d, want 1", len(provider.requests)) }
	})

	t.Run("report is available in a separate confirmed user turn", func(t *testing.T) {
		provider := &fakeProvider{responses: []agent.ModelResponse{
			{ToolCalls: []agent.ToolCall{{ID: "report-2", Name: "sales_ops_report", Arguments: reportArgs}}},
			{Text: "Confirmed mapping report is ready."},
		}}
		runner, err := agent.NewAgent(provider, "test-model", agent.AllowListPolicy{
			"tool.csv.inspect": true, "tool.sales_ops.report": true,
		}, 3, tools.CSVInspect{}, tools.SalesOpsReport{})
		if err != nil { t.Fatal(err) }
		result, err := runner.Run(context.Background(), "I confirm stage, value, and source mappings and the Won/Lost labels. Generate the report using this CSV: " + csv)
		if err != nil || result != "Confirmed mapping report is ready." {
			t.Fatalf("Run() = %q, %v", result, err)
		}
	})
}
