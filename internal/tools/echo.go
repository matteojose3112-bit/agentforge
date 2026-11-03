package tools

import "context"

// Echo returns its input unchanged. It is a safe example tool for local runs.
type Echo struct{}

// Name returns the tool's stable registry name.
func (Echo) Name() string { return "echo" }

// Execute returns the provided input.
func (Echo) Execute(_ context.Context, input string) (string, error) {
	return input, nil
}
