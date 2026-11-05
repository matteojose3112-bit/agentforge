package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/matteojose3112-bit/agentforge/internal/agent"
)

// SalesOpsReport turns deterministic CRM pipeline metrics into a concise,
// read-only operations report. Recommendations are rule-based and tied to
// observed fields; the tool does not forecast or modify external systems.
type SalesOpsReport struct{}

type salesOpsReportResult struct {
	ReportTitle       string   `json:"report_title"`
	TotalLeads        int      `json:"total_leads"`
	ClosedLeads       int      `json:"closed_leads"`
	OpenLeads         int      `json:"open_leads"`
	WinRatePercent    float64  `json:"win_rate_percent"`
	OpenPipelineValue float64  `json:"open_pipeline_value"`
	MissingStatuses   int      `json:"missing_statuses"`
	MissingValues     int      `json:"missing_values"`
	MissingSources    int      `json:"missing_sources"`
	Observations      []string `json:"observations"`
	RecommendedActions []string `json:"recommended_actions"`
	Limitations       []string `json:"limitations"`
}

func (SalesOpsReport) Name() string { return "sales_ops_report" }

func (SalesOpsReport) Definition() agent.ToolDefinition {
	return agent.ToolDefinition{
		Name:        "sales_ops_report",
		Description: "Create a concise, read-only sales operations report from a CRM lead CSV. It calculates pipeline metrics deterministically and returns observations and rule-based follow-up actions grounded in the supplied fields. It does not forecast, benchmark, prioritize individual leads, or modify a CRM.",
		Permission:  "tool.sales_ops.report",
		Parameters: json.RawMessage(`{"type":"object","properties":{"csv":{"type":"string"},"status_column":{"type":"string"},"value_column":{"type":"string"},"source_column":{"type":"string"},"won_statuses":{"type":"array","items":{"type":"string"}},"lost_statuses":{"type":"array","items":{"type":"string"}}},"required":["csv","status_column","value_column","source_column","won_statuses","lost_statuses"],"additionalProperties":false}`),
		OutputSchema: json.RawMessage(`{"type":"object","properties":{"report_title":{"type":"string"},"total_leads":{"type":"integer"},"closed_leads":{"type":"integer"},"open_leads":{"type":"integer"},"win_rate_percent":{"type":"number"},"open_pipeline_value":{"type":"number"},"missing_statuses":{"type":"integer"},"missing_values":{"type":"integer"},"missing_sources":{"type":"integer"},"observations":{"type":"array","items":{"type":"string"}},"recommended_actions":{"type":"array","items":{"type":"string"}},"limitations":{"type":"array","items":{"type":"string"}}},"required":["report_title","total_leads","closed_leads","open_leads","win_rate_percent","open_pipeline_value","missing_statuses","missing_values","missing_sources","observations","recommended_actions","limitations"],"additionalProperties":false}`),
	}
}

func (SalesOpsReport) Execute(ctx context.Context, input string) (string, error) {
	metricsJSON, err := (LeadPipelineAnalyze{}).Execute(ctx, input)
	if err != nil {
		return "", fmt.Errorf("analyze CRM pipeline: %w", err)
	}
	var metrics leadPipelineResult
	if err := json.Unmarshal([]byte(metricsJSON), &metrics); err != nil {
		return "", fmt.Errorf("decode CRM pipeline metrics: %w", err)
	}

	result := salesOpsReportResult{
		ReportTitle:       "Sales Operations Snapshot",
		TotalLeads:        metrics.TotalLeads,
		ClosedLeads:       metrics.ClosedLeads,
		OpenLeads:         metrics.OpenLeads,
		WinRatePercent:    metrics.WinRatePercent,
		OpenPipelineValue: metrics.OpenPipelineValue,
		MissingStatuses:   metrics.MissingStatuses,
		MissingValues:     metrics.MissingValues,
		MissingSources:    metrics.MissingSources,
		Observations:      make([]string, 0, 5),
		RecommendedActions: make([]string, 0, 5),
		Limitations: []string{
			"This report describes only the supplied CRM snapshot; it does not establish trends over time.",
			"Win rate is calculated from closed leads only (won divided by won plus lost).",
			"Recommendations are general review prompts, not forecasts or a ranking of individual leads.",
			"The report is read-only and does not update a CRM or contact customers.",
		},
	}

	result.Observations = append(result.Observations,
		fmt.Sprintf("The snapshot contains %d leads: %d open and %d closed.", metrics.TotalLeads, metrics.OpenLeads, metrics.ClosedLeads))
	if metrics.ClosedLeads > 0 {
		result.Observations = append(result.Observations,
			fmt.Sprintf("The observed win rate is %.1f%% across %d closed leads.", metrics.WinRatePercent, metrics.ClosedLeads))
	} else {
		result.Observations = append(result.Observations,
			"No leads are marked won or lost, so a closed-lead win rate cannot yet be meaningfully assessed.")
	}
	result.Observations = append(result.Observations,
		fmt.Sprintf("The recorded value of open leads is %.2f.", metrics.OpenPipelineValue))
	if metrics.MissingStatuses+metrics.MissingValues+metrics.MissingSources == 0 {
		result.Observations = append(result.Observations, "No missing status, estimated-value, or source fields were detected in the configured columns.")
	} else {
		result.Observations = append(result.Observations,
			fmt.Sprintf("Data gaps detected: %d missing statuses, %d missing estimated values, and %d missing sources.", metrics.MissingStatuses, metrics.MissingValues, metrics.MissingSources))
	}

	if metrics.OpenLeads > 0 {
		result.RecommendedActions = append(result.RecommendedActions,
			"Review open leads in the CRM and confirm each has a clear next step and an accountable owner.")
	}
	if metrics.MissingStatuses > 0 {
		result.RecommendedActions = append(result.RecommendedActions,
			"Complete missing lead stages before using this snapshot for funnel or win-rate decisions.")
	}
	if metrics.MissingValues > 0 {
		result.RecommendedActions = append(result.RecommendedActions,
			"Verify missing estimated values with the relevant lead owners before relying on pipeline-value totals.")
	}
	if metrics.MissingSources > 0 {
		result.RecommendedActions = append(result.RecommendedActions,
			"Standardize lead-source capture so source reporting is more complete.")
	}
	if metrics.ClosedLeads > 0 {
		result.RecommendedActions = append(result.RecommendedActions,
			"Review closed-won and closed-lost records, including outcome reasons if available, to identify patterns for the next sales review.")
	} else {
		result.RecommendedActions = append(result.RecommendedActions,
			"Accumulate and classify closed outcomes before drawing conclusions about conversion performance.")
	}
	if len(result.RecommendedActions) == 0 {
		result.RecommendedActions = append(result.RecommendedActions, "Continue reviewing pipeline changes on a consistent reporting cadence.")
	}

	// Normalize whitespace in generated strings defensively.
	for i := range result.Observations {
		result.Observations[i] = strings.TrimSpace(result.Observations[i])
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("encode sales operations report: %w", err)
	}
	return string(encoded), nil
}
