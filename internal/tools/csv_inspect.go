package tools

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/matteojose3112-bit/agentforge/internal/agent"
)

const csvInspectSampleLimit = 5

// CSVInspect returns bounded structural evidence to help an operator map CRM exports.
type CSVInspect struct{}

type csvInspectColumn struct {
	Name          string   `json:"name"`
	NonEmptyCount int      `json:"non_empty_count"`
	MissingCount  int      `json:"missing_count"`
	SampleValues  []string `json:"sample_values"`
}

type csvInspectResult struct {
	Rows    int               `json:"rows"`
	Columns []csvInspectColumn `json:"columns"`
}

func (CSVInspect) Name() string { return "csv_inspect" }

func (CSVInspect) Definition() agent.ToolDefinition {
	return agent.ToolDefinition{
		Name: "csv_inspect",
		Description: "Inspect a bounded CSV export without calculating business metrics or changing data. Return header names, non-empty and missing counts, and up to five distinct sample values per column. Use this evidence to propose CRM column mappings, but ask the operator to confirm ambiguous mappings before running sales_ops_report.",
		Permission: "tool.csv.inspect",
		Parameters: json.RawMessage(`{"type":"object","properties":{"csv":{"type":"string"}},"required":["csv"],"additionalProperties":false}`),
		OutputSchema: json.RawMessage(`{"type":"object","properties":{"rows":{"type":"integer"},"columns":{"type":"array","items":{"type":"object","properties":{"name":{"type":"string"},"non_empty_count":{"type":"integer"},"missing_count":{"type":"integer"},"sample_values":{"type":"array","items":{"type":"string"}}},"required":["name","non_empty_count","missing_count","sample_values"],"additionalProperties":false}}},"required":["rows","columns"],"additionalProperties":false}`),
	}
}

func (CSVInspect) Execute(_ context.Context, input string) (string, error) {
	var args struct { CSV string `json:"csv"` }
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("decode CSV inspection input: %w", err)
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

	reader := csv.NewReader(strings.NewReader(strings.TrimPrefix(args.CSV, "\uFEFF")))
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true
	header, err := reader.Read()
	if err != nil {
		if errors.Is(err, io.EOF) { return "", errors.New("CSV must contain a header row") }
		return "", fmt.Errorf("read CSV header: %w", err)
	}
	if len(header) == 0 || len(header) > maxCSVColumns {
		return "", fmt.Errorf("CSV must contain between 1 and %d columns", maxCSVColumns)
	}

	result := csvInspectResult{Columns: make([]csvInspectColumn, len(header))}
	seen := make(map[string]struct{}, len(header))
	for i, name := range header {
		name = strings.TrimSpace(name)
		if name == "" { return "", fmt.Errorf("CSV header column %d is empty", i+1) }
		key := strings.ToLower(name)
		if _, exists := seen[key]; exists { return "", fmt.Errorf("CSV header contains duplicate column %q", name) }
		seen[key] = struct{}{}
		result.Columns[i] = csvInspectColumn{Name: name, SampleValues: make([]string, 0, csvInspectSampleLimit)}
	}

	samples := make([]map[string]struct{}, len(header))
	for i := range samples { samples[i] = make(map[string]struct{}) }
	for {
		record, readErr := reader.Read()
		if errors.Is(readErr, io.EOF) { break }
		if readErr != nil { return "", fmt.Errorf("read CSV row %d: %w", result.Rows+2, readErr) }
		if len(record) != len(header) { return "", fmt.Errorf("CSV row %d has %d fields; expected %d", result.Rows+2, len(record), len(header)) }
		result.Rows++
		if result.Rows > maxCSVRows { return "", fmt.Errorf("CSV exceeds the %d data-row limit", maxCSVRows) }
		for i, raw := range record {
			value := strings.TrimSpace(raw)
			column := &result.Columns[i]
			if value == "" { column.MissingCount++; continue }
			column.NonEmptyCount++
			if len(column.SampleValues) >= csvInspectSampleLimit { continue }
			if _, exists := samples[i][value]; exists { continue }
			samples[i][value] = struct{}{}
			column.SampleValues = append(column.SampleValues, value)
		}
	}

	encoded, err := json.Marshal(result)
	if err != nil { return "", fmt.Errorf("encode CSV inspection: %w", err) }
	return string(encoded), nil
}
