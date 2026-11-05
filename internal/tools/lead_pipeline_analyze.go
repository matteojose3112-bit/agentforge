package tools

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/matteojose3112-bit/agentforge/internal/agent"
)

// LeadPipelineAnalyze summarizes a CRM lead export using explicitly named columns.
type LeadPipelineAnalyze struct{}

type leadPipelineBreakdown struct {
	Name  string `json:"name"`
	Leads int    `json:"leads"`
}

type leadPipelineResult struct {
	TotalLeads          int                    `json:"total_leads"`
	WonLeads            int                    `json:"won_leads"`
	LostLeads           int                    `json:"lost_leads"`
	OpenLeads           int                    `json:"open_leads"`
	ClosedLeads         int                    `json:"closed_leads"`
	WinRatePercent      float64                `json:"win_rate_percent"`
	MissingStatuses     int                    `json:"missing_statuses"`
	MissingValues       int                    `json:"missing_values"`
	MissingSources      int                    `json:"missing_sources"`
	TotalEstimatedValue float64                `json:"total_estimated_value"`
	OpenPipelineValue   float64                `json:"open_pipeline_value"`
	WonValue            float64                `json:"won_value"`
	LostValue           float64                `json:"lost_value"`
	StatusBreakdown     []leadPipelineBreakdown `json:"status_breakdown"`
	SourceBreakdown     []leadPipelineBreakdown `json:"source_breakdown"`
}

func (LeadPipelineAnalyze) Name() string { return "lead_pipeline_analyze" }

func (LeadPipelineAnalyze) Definition() agent.ToolDefinition {
	return agent.ToolDefinition{
		Name:        "lead_pipeline_analyze",
		Description: "Analyze a CRM lead CSV using explicit status and estimated-value columns. Summarize won, lost, open, closed, win rate, estimated values, missing fields, and optional source breakdown. Status matching is case-insensitive. Use an empty source_column to skip source breakdown and empty won_statuses/lost_statuses arrays to use common defaults.",
		Permission:  "tool.leads.analyze",
		Parameters: json.RawMessage(`{"type":"object","properties":{"csv":{"type":"string"},"status_column":{"type":"string"},"value_column":{"type":"string"},"source_column":{"type":"string"},"won_statuses":{"type":"array","items":{"type":"string"}},"lost_statuses":{"type":"array","items":{"type":"string"}}},"required":["csv","status_column","value_column","source_column","won_statuses","lost_statuses"],"additionalProperties":false}`),
		OutputSchema: json.RawMessage(`{"type":"object","properties":{"total_leads":{"type":"integer"},"won_leads":{"type":"integer"},"lost_leads":{"type":"integer"},"open_leads":{"type":"integer"},"closed_leads":{"type":"integer"},"win_rate_percent":{"type":"number"},"missing_statuses":{"type":"integer"},"missing_values":{"type":"integer"},"missing_sources":{"type":"integer"},"total_estimated_value":{"type":"number"},"open_pipeline_value":{"type":"number"},"won_value":{"type":"number"},"lost_value":{"type":"number"},"status_breakdown":{"type":"array","items":{"type":"object","properties":{"name":{"type":"string"},"leads":{"type":"integer"}},"required":["name","leads"],"additionalProperties":false}},"source_breakdown":{"type":"array","items":{"type":"object","properties":{"name":{"type":"string"},"leads":{"type":"integer"}},"required":["name","leads"],"additionalProperties":false}}},"required":["total_leads","won_leads","lost_leads","open_leads","closed_leads","win_rate_percent","missing_statuses","missing_values","missing_sources","total_estimated_value","open_pipeline_value","won_value","lost_value","status_breakdown","source_breakdown"],"additionalProperties":false}`),
	}
}

func (LeadPipelineAnalyze) Execute(_ context.Context, input string) (string, error) {
	var args struct {
		CSV          string   `json:"csv"`
		StatusColumn string   `json:"status_column"`
		ValueColumn  string   `json:"value_column"`
		SourceColumn string   `json:"source_column"`
		WonStatuses  []string `json:"won_statuses"`
		LostStatuses []string `json:"lost_statuses"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("decode lead pipeline input: %w", err)
	}
	if strings.TrimSpace(args.CSV) == "" {
		return "", errors.New("CSV input cannot be empty")
	}
	if len(args.CSV) > maxCSVBytes {
		return "", errors.New("CSV input exceeds the 1 MiB limit")
	}
	if !utf8.ValidString(args.CSV) {
		return "", errors.New("CSV input must be valid UTF-8")
	}
	args.StatusColumn = strings.TrimSpace(args.StatusColumn)
	args.ValueColumn = strings.TrimSpace(args.ValueColumn)
	args.SourceColumn = strings.TrimSpace(args.SourceColumn)
	if args.StatusColumn == "" || args.ValueColumn == "" {
		return "", errors.New("status_column and value_column are required")
	}
	if args.StatusColumn == args.ValueColumn || (args.SourceColumn != "" && (args.SourceColumn == args.StatusColumn || args.SourceColumn == args.ValueColumn)) {
		return "", errors.New("status, value, and source columns must be distinct")
	}

	wonStatuses, err := normalizeLeadStatuses(args.WonStatuses, []string{"won", "closed won"})
	if err != nil {
		return "", fmt.Errorf("won_statuses: %w", err)
	}
	lostStatuses, err := normalizeLeadStatuses(args.LostStatuses, []string{"lost", "closed lost"})
	if err != nil {
		return "", fmt.Errorf("lost_statuses: %w", err)
	}
	for status := range wonStatuses {
		if _, overlaps := lostStatuses[status]; overlaps {
			return "", fmt.Errorf("status %q cannot be both won and lost", status)
		}
	}

	document := strings.TrimPrefix(args.CSV, "\uFEFF")
	reader := csv.NewReader(strings.NewReader(document))
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return "", errors.New("CSV must contain a header row")
		}
		return "", fmt.Errorf("read CSV header: %w", err)
	}
	if len(header) == 0 || len(header) > maxCSVColumns {
		return "", fmt.Errorf("CSV must contain between 1 and %d columns", maxCSVColumns)
	}
	columns := make(map[string]int, len(header))
	for index, name := range header {
		name = strings.TrimSpace(name)
		if name == "" {
			return "", fmt.Errorf("CSV header column %d is empty", index+1)
		}
		key := strings.ToLower(name)
		if _, exists := columns[key]; exists {
			return "", fmt.Errorf("CSV header contains duplicate column %q", name)
		}
		header[index] = name
		columns[key] = index
	}
	statusIndex, ok := columns[strings.ToLower(args.StatusColumn)]
	if !ok {
		return "", fmt.Errorf("status column %q was not found", args.StatusColumn)
	}
	valueIndex, ok := columns[strings.ToLower(args.ValueColumn)]
	if !ok {
		return "", fmt.Errorf("value column %q was not found", args.ValueColumn)
	}
	sourceIndex := -1
	if args.SourceColumn != "" {
		sourceIndex, ok = columns[strings.ToLower(args.SourceColumn)]
		if !ok {
			return "", fmt.Errorf("source column %q was not found", args.SourceColumn)
		}
	}
	if statusIndex == valueIndex || (sourceIndex >= 0 && (sourceIndex == statusIndex || sourceIndex == valueIndex)) {
		return "", errors.New("status, value, and source columns must be distinct")
	}

	result := leadPipelineResult{
		StatusBreakdown: make([]leadPipelineBreakdown, 0),
		SourceBreakdown: make([]leadPipelineBreakdown, 0),
	}
	statusCounts := make(map[string]*leadPipelineBreakdown)
	sourceCounts := make(map[string]*leadPipelineBreakdown)
	for {
		record, readErr := reader.Read()
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return "", fmt.Errorf("read CSV row %d: %w", result.TotalLeads+2, readErr)
		}
		if len(record) != len(header) {
			return "", fmt.Errorf("CSV row %d has %d fields; expected %d", result.TotalLeads+2, len(record), len(header))
		}
		result.TotalLeads++
		if result.TotalLeads > maxCSVRows {
			return "", fmt.Errorf("CSV exceeds the %d data-row limit", maxCSVRows)
		}

		status := strings.TrimSpace(record[statusIndex])
		normalizedStatus := strings.ToLower(status)
		if status == "" {
			result.MissingStatuses++
		} else {
			incrementLeadBreakdown(statusCounts, normalizedStatus, status)
			switch {
			case wonStatuses[normalizedStatus]:
				result.WonLeads++
			case lostStatuses[normalizedStatus]:
				result.LostLeads++
			default:
				result.OpenLeads++
			}
		}

		rawValue := strings.TrimSpace(record[valueIndex])
		value := 0.0
		if rawValue == "" {
			result.MissingValues++
		} else {
			value, err = strconv.ParseFloat(rawValue, 64)
			if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
				return "", fmt.Errorf("CSV row %d has a non-numeric value in column %q", result.TotalLeads+1, args.ValueColumn)
			}
			if err := addLeadValue(&result.TotalEstimatedValue, value); err != nil {
				return "", fmt.Errorf("total estimated value: %w", err)
			}
			switch {
			case status != "" && wonStatuses[normalizedStatus]:
				if err := addLeadValue(&result.WonValue, value); err != nil {
					return "", fmt.Errorf("won value: %w", err)
				}
			case status != "" && lostStatuses[normalizedStatus]:
				if err := addLeadValue(&result.LostValue, value); err != nil {
					return "", fmt.Errorf("lost value: %w", err)
				}
			case status != "":
				if err := addLeadValue(&result.OpenPipelineValue, value); err != nil {
					return "", fmt.Errorf("open pipeline value: %w", err)
				}
			}
		}

		if sourceIndex >= 0 {
			source := strings.TrimSpace(record[sourceIndex])
			if source == "" {
				result.MissingSources++
			} else {
				incrementLeadBreakdown(sourceCounts, strings.ToLower(source), source)
			}
		}
	}
	result.ClosedLeads = result.WonLeads + result.LostLeads
	if result.ClosedLeads > 0 {
		result.WinRatePercent = float64(result.WonLeads) / float64(result.ClosedLeads) * 100
	}
	result.StatusBreakdown = sortedLeadBreakdown(statusCounts)
	result.SourceBreakdown = sortedLeadBreakdown(sourceCounts)

	encoded, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("encode lead pipeline analysis: %w", err)
	}
	return string(encoded), nil
}

func normalizeLeadStatuses(configured, defaults []string) (map[string]bool, error) {
	if len(configured) == 0 {
		configured = defaults
	}
	result := make(map[string]bool, len(configured))
	for _, status := range configured {
		normalized := strings.ToLower(strings.TrimSpace(status))
		if normalized == "" {
			return nil, errors.New("status values cannot be empty")
		}
		result[normalized] = true
	}
	return result, nil
}

func incrementLeadBreakdown(counts map[string]*leadPipelineBreakdown, key, label string) {
	if current, exists := counts[key]; exists {
		current.Leads++
		return
	}
	counts[key] = &leadPipelineBreakdown{Name: label, Leads: 1}
}

func sortedLeadBreakdown(counts map[string]*leadPipelineBreakdown) []leadPipelineBreakdown {
	result := make([]leadPipelineBreakdown, 0, len(counts))
	for _, item := range counts {
		result = append(result, *item)
	}
	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
	})
	return result
}

func addLeadValue(total *float64, value float64) error {
	next := *total + value
	if math.IsNaN(next) || math.IsInf(next, 0) {
		return errors.New("numeric total exceeds the supported range")
	}
	*total = next
	return nil
}
