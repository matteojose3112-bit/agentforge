package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/matteojose3112-bit/agentforge/internal/agent"
	"github.com/matteojose3112-bit/agentforge/internal/providers/openai"
	"github.com/matteojose3112-bit/agentforge/internal/tools"
)

var (
	version   = "0.3.0"
	commit    = "unknown"
	buildDate = "unknown"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "agentforge:", err)
		os.Exit(1)
	}
}

func run(args []string, out *os.File) error {
	flags := flag.NewFlagSet("agentforge", flag.ContinueOnError)
	flags.SetOutput(out)
	flags.Usage = func() {
		fmt.Fprintln(out, "AgentForge - controlled tool execution and AI-assisted workflows")
		fmt.Fprintln(out, "\nUsage:")
		fmt.Fprintln(out, "  agentforge [flags]")
		fmt.Fprintln(out, "\nModes:")
		fmt.Fprintln(out, "  Direct tool mode (default): execute one registered tool with its input payload.")
		fmt.Fprintln(out, "  Agent mode (-agent): send a prompt to the configured model; tool access is denied unless explicitly allowed.")
		fmt.Fprintln(out, "\nFlags:")
		flags.PrintDefaults()
		fmt.Fprintln(out, "\nExamples:")
		fmt.Fprintln(out, "  agentforge -tool echo -input \"hello AgentForge\"")
		fmt.Fprintln(out, "  agentforge -tool sales_ops_report -input-file examples/sales_ops_snapshot.json")
		fmt.Fprintln(out, "  agentforge -agent -model YOUR_MODEL -prompt \"Analyze this pipeline\" -allow-permission tool.sales_ops.report")
		fmt.Fprintln(out, "  agentforge -version")
		fmt.Fprintln(out, "\nNotes:")
		fmt.Fprintln(out, "  -input and -input-file are mutually exclusive.")
		fmt.Fprintln(out, "  Structured tools expect JSON; the echo tool accepts plain text.")
		fmt.Fprintln(out, "  Agent mode requires OPENAI_API_KEY and an explicit -model value.")
		fmt.Fprintln(out, "  -allow-permission is a comma-separated allow-list; empty means deny all.")
	}
	showVersion := flags.Bool("version", false, "print version, Git commit, and build timestamp")
	agentMode := flags.Bool("agent", false, "run a model-backed agent (requires -model and OPENAI_API_KEY)")
	toolName := flags.String("tool", "echo", "registered direct-mode tool (default: echo)")
	input := flags.String("input", "", "input payload for direct tool mode (JSON for structured tools; plain text for echo); mutually exclusive with -input-file")
	inputFile := flags.String("input-file", "", "read direct tool input payload from a file; mutually exclusive with -input")
	prompt := flags.String("prompt", "", "user prompt supplied to the model in agent mode")
	model := flags.String("model", "", "provider-supported model ID (required in agent mode)")
	timeout := flags.Duration("timeout", 0, "optional execution timeout (for example 30s); 0 means no CLI timeout")
	permissions := flags.String("allow-permission", "", "comma-separated tool permissions to grant in agent mode (default: deny all)")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if *showVersion {
		_, err := fmt.Fprintf(out, "AgentForge %s\nCommit: %s\nBuilt: %s\n", version, commit, buildDate)
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
