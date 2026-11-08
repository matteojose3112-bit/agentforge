package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestCSVInspectReturnsHeadersCountsAndDistinctSamples(t *testing.T) {
	input := `{"csv":"Lead ID,Stage,Amount,Source\n1,Qualified,1200,Referral\n2,Won,2500,Google Ads\n3,Qualified,,Referral\n4,Lost,800,Website\n"}`
	output, err := (CSVInspect{}).Execute(context.Background(), input)
	if err != nil { t.Fatalf("Execute() error = %v", err) }
	var got csvInspectResult
	if err := json.Unmarshal([]byte(output), &got); err != nil { t.Fatalf("decode output: %v", err) }
	if got.Rows != 4 || len(got.Columns) != 4 { t.Fatalf("unexpected dimensions: %+v", got) }
	stage := got.Columns[1]
	if stage.Name != "Stage" || stage.NonEmptyCount != 4 || stage.MissingCount != 0 || len(stage.SampleValues) != 3 {
		t.Fatalf("unexpected stage inspection: %+v", stage)
	}
	amount := got.Columns[2]
	if amount.NonEmptyCount != 3 || amount.MissingCount != 1 { t.Fatalf("unexpected amount counts: %+v", amount) }
	source := got.Columns[3]
	if len(source.SampleValues) != 3 || source.SampleValues[0] != "Referral" || source.SampleValues[1] != "Google Ads" {
		t.Fatalf("unexpected source samples: %+v", source)
	}
}

func TestCSVInspectHeaderOnlyAndValidation(t *testing.T) {
	output, err := (CSVInspect{}).Execute(context.Background(), `{"csv":"id,stage\n"}`)
	if err != nil { t.Fatalf("header-only CSV should be valid: %v", err) }
	var got csvInspectResult
	if err := json.Unmarshal([]byte(output), &got); err != nil { t.Fatalf("decode output: %v", err) }
	if got.Rows != 0 || len(got.Columns) != 2 { t.Fatalf("unexpected empty result: %+v", got) }

	for _, input := range []string{
		`{"csv":""}`,
		`{"csv":"id,stage\n1,Won,extra\n"}`,
		`{"csv":"id,stage,Stage\n1,Won,Lost\n"}`,
	} {
		if _, err := (CSVInspect{}).Execute(context.Background(), input); err == nil {
			t.Errorf("Execute(%s) expected an error", input)
		}
	}
}

func TestCSVInspectSampleValuesAreBounded(t *testing.T) {
	var csv strings.Builder
	csv.WriteString("stage\n")
	for i := 0; i < 20; i++ { csv.WriteString("stage-" + string(rune('A'+i)) + "\n") }
	input, _ := json.Marshal(map[string]string{"csv": csv.String()})
	output, err := (CSVInspect{}).Execute(context.Background(), string(input))
	if err != nil { t.Fatalf("Execute() error = %v", err) }
	var got csvInspectResult
	if err := json.Unmarshal([]byte(output), &got); err != nil { t.Fatalf("decode output: %v", err) }
	if len(got.Columns[0].SampleValues) != csvInspectSampleLimit { t.Fatalf("sample count = %d, want %d", len(got.Columns[0].SampleValues), csvInspectSampleLimit) }
}
