package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/matteojose3112-bit/agentforge/internal/agent"
)

func TestSalesOpsReportProducesGroundedSnapshot(t *testing.T) {
	input := leadPipelineInput(
		"source,stage,value\nGoogle Ads,Won,100\nReferral,Lost,50\nAngi,Qualified,25\n,New,\n",
		"stage", "value", "source", []string{"Won"}, []string{"Lost"},
	)
	tool := SalesOpsReport{}
	if err := agent.ValidateJSON(tool.Definition().Parameters, []byte(input)); err != nil {
		t.Fatalf("input does not match schema: %v", err)
	}
	output, err := tool.Execute(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if err := agent.ValidateJSON(tool.Definition().OutputSchema, []byte(output)); err != nil {
		t.Fatalf("output does not match schema: %v", err)
	}
	var result salesOpsReportResult
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatal(err)
	}
	if result.TotalLeads != 4 || result.ClosedLeads != 2 || result.OpenLeads != 2 || result.WinRatePercent != 50 || result.OpenPipelineValue != 25 {
		t.Fatalf("unexpected report metrics: %+v", result)
	}
	if result.MissingValues != 1 || result.MissingSources != 1 {
		t.Fatalf("missing-field counts not carried into report: %+v", result)
	}
	joined := strings.Join(result.RecommendedActions, " ")
	for _, expected := range []string{"next step", "estimated values", "lead-source", "closed-won"} {
		if !strings.Contains(joined, expected) {
			t.Errorf("recommended actions do not include %q: %s", expected, joined)
		}
	}
	if strings.Contains(strings.ToLower(joined), "forecast") {
		t.Errorf("report should not make forecasting claims: %s", joined)
	}
}

func TestSalesOpsReportDoesNotInventWinRateWithoutClosedLeads(t *testing.T) {
	input := leadPipelineInput("stage,value\nQualified,20\nIn Progress,30\n", "stage", "value", "", nil, nil)
	output, err := (SalesOpsReport{}).Execute(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	var result salesOpsReportResult
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatal(err)
	}
	if result.ClosedLeads != 0 || result.WinRatePercent != 0 {
		t.Fatalf("unexpected closed-lead metrics: %+v", result)
	}
	if !strings.Contains(strings.Join(result.Observations, " "), "cannot yet be meaningfully assessed") {
		t.Fatalf("report should explain unavailable win-rate context: %+v", result.Observations)
	}
	if !strings.Contains(strings.Join(result.RecommendedActions, " "), "Accumulate and classify closed outcomes") {
		t.Fatalf("report should recommend collecting closed outcomes: %+v", result.RecommendedActions)
	}
}

func TestSalesOpsReportRejectsInvalidPipelineInput(t *testing.T) {
	_, err := (SalesOpsReport{}).Execute(context.Background(), "{")
	if err == nil || !strings.Contains(err.Error(), "analyze CRM pipeline") {
		t.Fatalf("Execute() error = %v, want wrapped pipeline analysis error", err)
	}
}
