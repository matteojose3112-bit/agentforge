package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunReadsDirectToolInputFromFile(t *testing.T) {
	dir := t.TempDir()
	inputPath := dir + "/input.json"
	if err := os.WriteFile(inputPath, []byte(`{"csv":"name,amount\nAlpha,10\nBeta,20\n"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	output, err := os.CreateTemp(dir, "output")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()

	if err := run([]string{"-tool", "csv_analyze", "-input-file", inputPath}, output); err != nil {
		t.Fatal(err)
	}
	if _, err := output.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	result, err := io.ReadAll(output)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(result), `"rows":2`) || !strings.Contains(string(result), `"columns":2`) {
		t.Fatalf("unexpected output: %s", result)
	}
}

func TestRunRejectsConflictingDirectInputs(t *testing.T) {
	output, err := os.CreateTemp(t.TempDir(), "output")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	err = run([]string{"-input", "{}", "-input-file", "unused.json"}, output)
	if err == nil || !strings.Contains(err.Error(), "not both") {
		t.Fatalf("run() error = %v, want conflicting-input error", err)
	}
}

func TestRunRequiresExplicitModelIDInAgentMode(t *testing.T) {
	output, err := os.CreateTemp(t.TempDir(), "output")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	err = run([]string{"-agent", "-prompt", "analyze this"}, output)
	if err == nil || !strings.Contains(err.Error(), "missing -model") {
		t.Fatalf("run() error = %v, want missing-model error", err)
	}
}

func TestPortfolioLeadPipelineFixture(t *testing.T) {
	inputPath := filepath.Join("..", "..", "examples", "lead_pipeline.json")
	output, err := os.CreateTemp(t.TempDir(), "output")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()

	if err := run([]string{"-tool", "csv_analyze", "-input-file", inputPath}, output); err != nil {
		t.Fatal(err)
	}
	if _, err := output.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	result, err := io.ReadAll(output)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{`"rows":8`, `"missing_values":2`, `"mean":9100`} {
		if !strings.Contains(string(result), expected) {
			t.Errorf("portfolio fixture output missing %s: %s", expected, result)
		}
	}
}

func TestLeadPipelineAnalysisPortfolioFixture(t *testing.T) {
	inputPath := filepath.Join("..", "..", "examples", "lead_pipeline_analysis.json")
	output, err := os.CreateTemp(t.TempDir(), "output")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()

	if err := run([]string{"-tool", "lead_pipeline_analyze", "-input-file", inputPath}, output); err != nil {
		t.Fatal(err)
	}
	if _, err := output.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	result, err := io.ReadAll(output)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		`"total_leads":8`,
		`"won_leads":1`,
		`"lost_leads":1`,
		`"open_leads":6`,
		`"win_rate_percent":50`,
		`"total_estimated_value":63700`,
		`"open_pipeline_value":44500`,
	} {
		if !strings.Contains(string(result), expected) {
			t.Errorf("lead pipeline fixture output missing %s: %s", expected, result)
		}
	}
}
