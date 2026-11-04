package tools

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/matteojose3112-bit/agentforge/internal/agent"
)

const (
	maxCSVBytes = 1024 * 1024
	maxCSVRows  = 10000
	maxCSVColumns = 100
)

// CSVAnalyze summarizes a bounded CSV document without modifying it.
type CSVAnalyze struct{}

type numericColumnSummary struct {
	Column         string  `json:"column"`
	NonEmptyValues int     `json:"non_empty_values"`
	Min            float64 `json:"min"`
	Max            float64 `json:"max"`
	Mean           float64 `json:"mean"`
}

type csvAnalysisResult struct {
	Rows          int                    `json:"rows"`
	Columns       int                    `json:"columns"`
	ColumnNames   []string               `json:"column_names"`
	MissingValues int                    `json:"missing_values"`
	NumericColumns []numericColumnSummary `json:"numeric_columns"`
}

func (CSVAnalyze) Name() string { return "csv_analyze" }

func (CSVAnalyze) Definition() agent.ToolDefinition {
	return agent.ToolDefinition{
		Name:        "csv_analyze",
		Description: "Analyze a CSV string with a header row. Report data row and column counts, missing cells, and min/max/mean for columns containing only numeric non-empty values.",
		Permission:  "tool.csv.analyze",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"csv":{"type":"string"}},"required":["csv"],"additionalProperties":false}`),
		OutputSchema: json.RawMessage(`{"type":"object","properties":{"rows":{"type":"integer"},"columns":{"type":"integer"},"column_names":{"type":"array","items":{"type":"string"}},"missing_values":{"type":"integer"},"numeric_columns":{"type":"array","items":{"type":"object","properties":{"column":{"type":"string"},"non_empty_values":{"type":"integer"},"min":{"type":"number"},"max":{"type":"number"},"mean":{"type":"number"}},"required":["column","non_empty_values","min","max","mean"],"additionalProperties":false}}},"required":["rows","columns","column_names","missing_values","numeric_columns"],"additionalProperties":false}`),
	}
}

func (CSVAnalyze) Execute(_ context.Context, input string) (string, error) {
	var args struct {
		CSV string `json:"csv"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("decode CSV tool input: %w", err)
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

	seen := make(map[string]struct{}, len(header))
	for i := range header {
		header[i] = strings.TrimSpace(header[i])
		if header[i] == "" {
			return "", fmt.Errorf("CSV header column %d is empty", i+1)
		}
		if _, exists := seen[header[i]]; exists {
			return "", fmt.Errorf("CSV header contains duplicate column %q", header[i])
		}
		seen[header[i]] = struct{}{}
	}

	type accumulator struct {
		count int
		min float64
		max float64
		sum float64
		numeric bool
	}
	accumulators := make([]accumulator, len(header))
	for i := range accumulators {
		accumulators[i].numeric = true
	}

	result := csvAnalysisResult{
		Columns: len(header),
		ColumnNames: append([]string(nil), header...),
		NumericColumns: make([]numericColumnSummary, 0, len(header)),
	}

	for {
		record, readErr := reader.Read()
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return "", fmt.Errorf("read CSV row %d: %w", result.Rows+2, readErr)
		}
		if len(record) != len(header) {
			return "", fmt.Errorf("CSV row %d has %d fields; expected %d", result.Rows+2, len(record), len(header))
		}
		result.Rows++
		if result.Rows > maxCSVRows {
			return "", fmt.Errorf("CSV exceeds the %d data-row limit", maxCSVRows)
		}

		for i, raw := range record {
			value := strings.TrimSpace(raw)
			if value == "" {
				result.MissingValues++
				continue
			}
			number, parseErr := strconv.ParseFloat(value, 64)
			if parseErr != nil || math.IsNaN(number) || math.IsInf(number, 0) {
				accumulators[i].numeric = false
				continue
			}
			acc := &accumulators[i]
			if acc.count == 0 {
				acc.min, acc.max = number, number
			} else {
				if number < acc.min { acc.min = number }
				if number > acc.max { acc.max = number }
			}
			nextSum := acc.sum + number
			if math.IsNaN(nextSum) || math.IsInf(nextSum, 0) {
				acc.numeric = false
				continue
			}
			acc.count++
			acc.sum = nextSum
		}
	}

	for i, acc := range accumulators {
		if !acc.numeric || acc.count == 0 {
			continue
		}
		result.NumericColumns = append(result.NumericColumns, numericColumnSummary{
			Column: header[i],
			NonEmptyValues: acc.count,
			Min: acc.min,
			Max: acc.max,
			Mean: acc.sum / float64(acc.count),
		})
	}

	encoded, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("encode CSV analysis: %w", err)
	}
	return string(encoded), nil
}
