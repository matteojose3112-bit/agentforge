package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestCSVAnalyzeSummarizesRowsMissingValuesAndNumericColumns(t *testing.T) {
	input := map[string]string{"csv": "name,amount,score\nA,10,2\nB,,4\nC,30,6\n"}
	encoded, _ := json.Marshal(input)
	output, err := (CSVAnalyze{}).Execute(context.Background(), string(encoded))
	if err != nil {
		t.Fatal(err)
	}
	var got csvAnalysisResult
	if err := json.Unmarshal([]byte(output), &got); err != nil {
		t.Fatal(err)
	}
	if got.Rows != 3 || got.Columns != 3 || got.MissingValues != 1 {
		t.Fatalf("unexpected dimensions or missing count: %+v", got)
	}
	if len(got.ColumnNames) != 3 || got.ColumnNames[0] != "name" {
		t.Fatalf("unexpected headers: %#v", got.ColumnNames)
	}
	if len(got.NumericColumns) != 2 {
		t.Fatalf("numeric columns = %d, want 2: %+v", len(got.NumericColumns), got.NumericColumns)
	}
	amount := got.NumericColumns[0]
	if amount.Column != "amount" || amount.NonEmptyValues != 2 || amount.Min != 10 || amount.Max != 30 || amount.Mean != 20 {
		t.Fatalf("unexpected amount summary: %+v", amount)
	}
	score := got.NumericColumns[1]
	if score.Column != "score" || score.NonEmptyValues != 3 || score.Mean != 4 {
		t.Fatalf("unexpected score summary: %+v", score)
	}
}

func TestCSVAnalyzeSupportsQuotedFieldsAndUTF8BOM(t *testing.T) {
	input, _ := json.Marshal(map[string]string{"csv": "\uFEFFname,note\nA,\"likes commas, and quotes\"\n"})
	output, err := (CSVAnalyze{}).Execute(context.Background(), string(input))
	if err != nil {
		t.Fatal(err)
	}
	var got csvAnalysisResult
	if err := json.Unmarshal([]byte(output), &got); err != nil {
		t.Fatal(err)
	}
	if got.Rows != 1 || got.Columns != 2 || got.ColumnNames[0] != "name" || got.MissingValues != 0 {
		t.Fatalf("unexpected analysis: %+v", got)
	}
}

func TestCSVAnalyzeRejectsInvalidDocuments(t *testing.T) {
	cases := []struct {
		name string
		csv string
	}{
		{name: "empty"},
		{name: "duplicate headers", csv: "name,name\na,b"},
		{name: "empty header", csv: "name,\na,b"},
		{name: "ragged row", csv: "a,b\n1"},
		{name: "invalid quoting", csv: "a,b\n1,\"unterminated"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			encoded, _ := json.Marshal(map[string]string{"csv": tc.csv})
			if _, err := (CSVAnalyze{}).Execute(context.Background(), string(encoded)); err == nil {
				t.Fatal("expected invalid CSV error")
			}
		})
	}
}

func TestCSVAnalyzeEnforcesInputAndRowLimits(t *testing.T) {
	large, _ := json.Marshal(map[string]string{"csv": strings.Repeat("x", maxCSVBytes+1)})
	if _, err := (CSVAnalyze{}).Execute(context.Background(), string(large)); err == nil {
		t.Fatal("expected input size limit error")
	}

	var rows strings.Builder
	rows.WriteString("value\n")
	for i := 0; i <= maxCSVRows; i++ {
		rows.WriteString("1\n")
	}
	encoded, _ := json.Marshal(map[string]string{"csv": rows.String()})
	if _, err := (CSVAnalyze{}).Execute(context.Background(), string(encoded)); err == nil {
		t.Fatal("expected data row limit error")
	}
}
