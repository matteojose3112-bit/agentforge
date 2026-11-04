package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestTextAnalyzeCountsUnicodeWordsAndLines(t *testing.T) {
	tool := TextAnalyze{}
	output, err := tool.Execute(context.Background(), `{"text":"Hola, mundo!\nこんにちは 世界"}`)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Words                   int `json:"words"`
		Characters              int `json:"characters"`
		Lines                   int `json:"lines"`
		EstimatedReadingMinutes int `json:"estimated_reading_minutes"`
	}
	if err := json.Unmarshal([]byte(output), &got); err != nil {
		t.Fatal(err)
	}
	if got.Words != 4 || got.Characters != 21 || got.Lines != 2 || got.EstimatedReadingMinutes != 1 {
		t.Fatalf("unexpected analysis: %+v", got)
	}
}

func TestTextAnalyzeRoundsReadingTimeUp(t *testing.T) {
	text := strings.Repeat("word ", 201)
	input, _ := json.Marshal(map[string]string{"text": text})
	output, err := (TextAnalyze{}).Execute(context.Background(), string(input))
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		EstimatedReadingMinutes int `json:"estimated_reading_minutes"`
	}
	if err := json.Unmarshal([]byte(output), &got); err != nil {
		t.Fatal(err)
	}
	if got.EstimatedReadingMinutes != 2 {
		t.Fatalf("reading time = %d, want 2", got.EstimatedReadingMinutes)
	}
}

func TestTextAnalyzeRejectsEmptyAndOversizedInput(t *testing.T) {
	tool := TextAnalyze{}
	if _, err := tool.Execute(context.Background(), `{"text":"  \n "}`); err == nil {
		t.Fatal("expected empty text error")
	}
	large, _ := json.Marshal(map[string]string{"text": strings.Repeat("x", maxAnalyzedTextBytes+1)})
	if _, err := tool.Execute(context.Background(), string(large)); err == nil {
		t.Fatal("expected size limit error")
	}
}
