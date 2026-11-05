package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/matteojose3112-bit/agentforge/internal/agent"
	"github.com/matteojose3112-bit/agentforge/internal/providers/openai"
	"github.com/matteojose3112-bit/agentforge/internal/tools"
)

const version = "0.3.0"

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "agentforge:", err)
		os.Exit(1)
	}
}

func run(args []string, out *os.File) error {
	flags := flag.NewFlagSet("agentforge", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	showVersion := flags.Bool("version", false, "print version")
	agentMode := flags.Bool("agent", false, "run a model-backed agent using OpenAI")
	toolName := flags.String("tool", "echo", "registered tool to execute in direct mode")
	input := flags.String("input", "", "direct tool input as a JSON string")
	inputFile := flags.String("input-file", "", "read direct tool JSON input from a file")
	prompt := flags.String("prompt", "", "user prompt for agent mode")
	model := flags.String("model", "", "model ID supported by your provider account (required in agent mode)")
	timeout := flags.Duration("timeout", 0, "execution timeout (for example 30s)")
	permissions := flags.String("allow-permission", "", "comma-separated tool permissions to grant (default: deny all)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *showVersion {
		_, err := fmt.Fprintln(out, version)
		return err
	}
	registeredTools := []agent.Tool{tools.Echo{}, tools.TextAnalyze{}, tools.CSVAnalyze{}, tools.LeadPipelineAnalyze{}, tools.SalesOpsReport{}}
	if *agentMode {
		if strings.TrimSpace(*prompt) == "" {
			return fmt.Errorf("missing -prompt for agent mode")
		}
		if strings.TrimSpace(*model) == "" {
			return fmt.Errorf("missing -model for agent mode; pass a model ID supported by your API account")
		}
		provider, err := openai.New(os.Getenv("OPENAI_API_KEY"))
		if err != nil {
			return err
		}
		allow := agent.AllowListPolicy{}
		for _, permission := range strings.Split(*permissions, ",") {
			if permission = strings.TrimSpace(permission); permission != "" {
				allow[permission] = true
			}
		}
		runner, err := agent.NewAgent(provider, *model, allow, 5, registeredTools...)
		if err != nil {
			return err
		}
		ctx := context.Background()
		if *timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, *timeout)
			defer cancel()
		}
		result, err := runner.Run(ctx, *prompt)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(out, result)
		return err
	}
	if *input != "" && *inputFile != "" {
		return fmt.Errorf("use either -input or -input-file, not both")
	}
	toolInput := *input
	if *inputFile != "" {
		contents, readErr := os.ReadFile(*inputFile)
		if readErr != nil {
			return fmt.Errorf("read input file: %w", readErr)
		}
		toolInput = string(contents)
	}
	if toolInput == "" {
		return fmt.Errorf("missing -input or -input-file")
	}
	runner, err := agent.NewRunnerWithOptions(agent.Options{Timeout: *timeout}, registeredTools...)
	if err != nil {
		return err
	}
	result, err := runner.Run(context.Background(), *toolName, toolInput)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, result.Output)
	return err
}
