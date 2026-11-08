package tools

import (
	"context"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type evaluationFixtureInput struct {
	CSV          string   `json:"csv"`
	StatusColumn string   `json:"status_column"`
	ValueColumn  string   `json:"value_column"`
	SourceColumn string   `json:"source_column"`
	WonStatuses  []string `json:"won_statuses"`
	LostStatuses []string `json:"lost_statuses"`
}

func (in evaluationFixtureInput) payload() string {
	encoded, _ := json.Marshal(in)
	return string(encoded)
}

func TestSalesOpsEvaluationFixtures(t *testing.T) {
	tests := []struct {
		name            string
		total           int
		won             int
		lost            int
		open            int
		closed          int
		winRate         float64
		openValue       float64
		missingStatuses int
		missingValues   int
		missingSources  int
		wantAction      string
	}{
		{name: "healthy_pipeline", total: 6, won: 2, lost: 1, open: 3, closed: 3, winRate: 200.0 / 3, openValue: 40000},
		{name: "weak_pipeline", total: 8, won: 1, lost: 3, open: 4, closed: 4, winRate: 25, openValue: 47000},
		{name: "incomplete_pipeline", total: 7, won: 1, lost: 1, open: 4, closed: 2, winRate: 50, openValue: 17000, missingStatuses: 1, missingValues: 2, missingSources: 2, wantAction: "Complete missing lead stages"},
		{name: "edge_case_pipeline", total: 0, openValue: 0, wantAction: "Accumulate and classify closed outcomes"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join("..", "..", "examples", "evaluations", tc.name+".json")
			contents, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			var input evaluationFixtureInput
			if err := json.Unmarshal(contents, &input); err != nil {
				t.Fatalf("decode fixture: %v", err)
			}

			output, err := (SalesOpsReport{}).Execute(context.Background(), input.payload())
			if err != nil {
				t.Fatalf("execute report: %v", err)
			}
			var got salesOpsReportResult
			if err := json.Unmarshal([]byte(output), &got); err != nil {
				t.Fatalf("decode report: %v", err)
			}

			if got.TotalLeads != tc.total || got.ClosedLeads != tc.closed || got.OpenLeads != tc.open ||
				math.Abs(got.WinRatePercent-tc.winRate) > 1e-9 || got.OpenPipelineValue != tc.openValue ||
				got.MissingStatuses != tc.missingStatuses || got.MissingValues != tc.missingValues ||
				got.MissingSources != tc.missingSources {
				t.Fatalf("unexpected report metrics: %+v", got)
			}

			// Won/lost counts are also asserted through the deterministic analyzer;
			// the report intentionally exposes only aggregate closed-lead metrics.
			metricsJSON, err := (LeadPipelineAnalyze{}).Execute(context.Background(), input.payload())
			if err != nil {
				t.Fatalf("execute pipeline analysis: %v", err)
			}
			var metrics leadPipelineResult
			if err := json.Unmarshal([]byte(metricsJSON), &metrics); err != nil {
				t.Fatalf("decode pipeline metrics: %v", err)
			}
			if metrics.WonLeads != tc.won || metrics.LostLeads != tc.lost {
				t.Fatalf("won/lost = %d/%d, want %d/%d", metrics.WonLeads, metrics.LostLeads, tc.won, tc.lost)
			}

			if tc.wantAction != "" && !strings.Contains(strings.Join(got.RecommendedActions, " "), tc.wantAction) {
				t.Errorf("expected grounded action containing %q; got %v", tc.wantAction, got.RecommendedActions)
			}
		})
	}
}
