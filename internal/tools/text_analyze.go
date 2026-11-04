package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/matteojose3112-bit/agentforge/internal/agent"
)

const maxAnalyzedTextBytes = 256 * 1024

// TextAnalyze provides deterministic, read-only metrics for a supplied text.
type TextAnalyze struct{}

func (TextAnalyze) Name() string { return "text_analyze" }

func (TextAnalyze) Definition() agent.ToolDefinition {
	return agent.ToolDefinition{
		Name:        "text_analyze",
		Description: "Count words, Unicode characters, and lines in supplied text, and estimate reading time at 200 words per minute.",
		Permission:  "tool.text.analyze",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"text":{"type":"string"}},"required":["text"],"additionalProperties":false}`),
		OutputSchema: json.RawMessage(`{"type":"object","properties":{"words":{"type":"integer"},"characters":{"type":"integer"},"lines":{"type":"integer"},"estimated_reading_minutes":{"type":"integer"}},"required":["words","characters","lines","estimated_reading_minutes"],"additionalProperties":false}`),
	}
}

func (TextAnalyze) Execute(_ context.Context, input string) (string, error) {
	var args struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", err
	}
	if strings.TrimSpace(args.Text) == "" {
		return "", errors.New("text cannot be empty")
	}
	if len(args.Text) > maxAnalyzedTextBytes {
		return "", errors.New("text exceeds the 256 KiB analysis limit")
	}

	words := len(strings.Fields(args.Text))
	readingMinutes := (words + 199) / 200
	if readingMinutes < 1 {
		readingMinutes = 1
	}
	result := struct {
		Words                   int `json:"words"`
		Characters              int `json:"characters"`
		Lines                   int `json:"lines"`
		EstimatedReadingMinutes int `json:"estimated_reading_minutes"`
	}{
		Words:                   words,
		Characters:              utf8.RuneCountInString(args.Text),
		Lines:                   strings.Count(args.Text, "\n") + 1,
		EstimatedReadingMinutes: readingMinutes,
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}
