package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/matteojose3112-bit/agentforge/internal/agent"
)

func leadPipelineInput(csv, statusColumn, valueColumn, sourceColumn string, won, lost []string) string {
	input := map[string]any{
		"csv": csv,
		"status_column": statusColumn,
		"value_column": valueColumn,
	}
	if sourceColumn != "" {
		input["source_column"] = sourceColumn
	}
	if won != nil {
		input["won_statuses"] = won
	}
	if lost != nil {
		input["lost_statuses"] = lost
	}
	encoded, _ := json.Marshal(input)
	return string(encoded)
}

func TestLeadPipelineAnalyzeSummarizesCRMExport(t *testing.T) {
	input := leadPipelineInput(
		"source,stage,value\nGoogle Ads,Won,100\nReferral,Lost,50\nAngi,Qualified,25\n,New,\n",
		"stage", "value", "source", []string{"Won"}, []string{"Lost"},
	)
	output, err := (LeadPipelineAnalyze{}).Execute(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	var result leadPipelineResult
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatal(err)
	}
	if result.TotalLeads != 4 || result.WonLeads != 1 || result.LostLeads != 1 || result.OpenLeads != 2 || result.ClosedLeads != 2 {
		t.Fatalf("unexpected lead counts: %+v", result)
	}
	if result.WinRatePercent != 50 {
		t.Fatalf("win rate = %v, want 50", result.WinRatePercent)
	}
	if result.MissingStatuses != 0 || result.MissingValues != 1 || result.MissingSources != 1 {
		t.Fatalf("unexpected missing-field counts: %+v", result)
	}
	if result.TotalEstimatedValue != 175 || result.OpenPipelineValue != 25 || result.WonValue != 100 || result.LostValue != 50 {
		t.Fatalf("unexpected value summaries: %+v", result)
	}
	if len(result.StatusBreakdown) != 4 || len(result.SourceBreakdown) != 3 {
		t.Fatalf("unexpected breakdowns: %+v", result)
	}
	if err := agent.ValidateJSON((LeadPipelineAnalyze{}).Definition().OutputSchema, []byte(output)); err != nil {
		t.Fatalf("output does not match declared schema: %v", err)
	}
}

func TestLeadPipelineAnalyzeUsesDefaultStatusesCaseInsensitively(t *testing.T) {
	input := leadPipelineInput(
		"stage,value\nCLOSED WON,40\nLost,10\nIn Progress,20\n",
		"stage", "value", "", nil, nil,
	)
	output, err := (LeadPipelineAnalyze{}).Execute(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	var result leadPipelineResult
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatal(err)
	}
	if result.WonLeads != 1 || result.LostLeads != 1 || result.OpenLeads != 1 || result.WinRatePercent != 50 {
		t.Fatalf("unexpected default status classification: %+v", result)
	}
	if len(result.SourceBreakdown) != 0 {
		t.Fatalf("source breakdown should be empty when no source column is configured: %+v", result.SourceBreakdown)
	}
}

func TestLeadPipelineAnalyzeHandlesMissingStatusAndZeroClosedLeads(t *testing.T) {
	input := leadPipelineInput("stage,value\n,10\nQualified,20\n", "stage", "value", "", nil, nil)
	output, err := (LeadPipelineAnalyze{}).Execute(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	var result leadPipelineResult
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatal(err)
	}
	if result.MissingStatuses != 1 || result.OpenLeads != 1 || result.ClosedLeads != 0 || result.WinRatePercent != 0 {
		t.Fatalf("unexpected missing-status result: %+v", result)
	}
	if result.TotalEstimatedValue != 30 || result.OpenPipelineValue != 20 {
		t.Fatalf("unexpected value totals: %+v", result)
	}
}

func TestLeadPipelineAnalyzeRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "invalid JSON", input: "{", want: "decode lead pipeline input"},
		{name: "empty CSV", input: leadPipelineInput(" ", "stage", "value", "", nil, nil), want: "cannot be empty"},
		{name: "missing status column", input: leadPipelineInput("source,value\nA,10\n", "stage", "value", "", nil, nil), want: "status column"},
		{name: "missing value column", input: leadPipelineInput("stage\nWon\n", "stage", "value", "", nil, nil), want: "value column"},
		{name: "non-numeric value", input: leadPipelineInput("stage,value\nWon,many\n", "stage", "value", "", nil, nil), want: "non-numeric value"},
		{name: "ragged row", input: leadPipelineInput("stage,value\nWon,10,extra\n", "stage", "value", "", nil, nil), want: "expected 2"},
		{name: "missing source column", input: leadPipelineInput("stage,value\nWon,10\n", "stage", "value", "source", nil, nil), want: "source column"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := (LeadPipelineAnalyze{}).Execute(context.Background(), test.input)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Execute() error = %v, want substring %q", err, test.want)
			}
		})
	}
}

func TestLeadPipelineAnalyzeRejectsOverlappingStatusMappings(t *testing.T) {
	input := leadPipelineInput("stage,value\nWon,10\n", "stage", "value", "", []string{"Won"}, []string{"won"})
	_, err := (LeadPipelineAnalyze{}).Execute(context.Background(), input)
	if err == nil || !strings.Contains(err.Error(), "cannot be both won and lost") {
		t.Fatalf("Execute() error = %v, want overlapping status error", err)
	}
}
