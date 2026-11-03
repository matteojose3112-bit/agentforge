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
	if err := run(os.Args[1:], os.Stdout); err != nil { fmt.Fprintln(os.Stderr, "agentforge:", err); os.Exit(1) }
}

func run(args []string, out *os.File) error {
	flags := flag.NewFlagSet("agentforge", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	showVersion := flags.Bool("version", false, "print version")
	agentMode := flags.Bool("agent", false, "run a model-backed agent using OpenAI")
	toolName := flags.String("tool", "echo", "registered tool to execute in direct mode")
	input := flags.String("input", "", "direct tool input")
	prompt := flags.String("prompt", "", "user prompt for agent mode")
	model := flags.String("model", "gpt-6-luna", "model name for agent mode")
	timeout := flags.Duration("timeout", 0, "execution timeout (for example 30s)")
	permissions := flags.String("allow-permission", "", "comma-separated tool permissions to grant (default: deny all)")
	if err := flags.Parse(args); err != nil { return err }
	if *showVersion { _, err := fmt.Fprintln(out, version); return err }
	if *agentMode {
		if strings.TrimSpace(*prompt) == "" { return fmt.Errorf("missing -prompt for agent mode") }
		provider, err := openai.New(os.Getenv("OPENAI_API_KEY")); if err != nil { return err }
		allow := agent.AllowListPolicy{}
		for _, permission := range strings.Split(*permissions, ",") {
			if permission = strings.TrimSpace(permission); permission != "" { allow[permission] = true }
		}
		runner, err := agent.NewAgent(provider, *model, allow, 5, tools.AgentEcho{}); if err != nil { return err }
		ctx := context.Background()
		if *timeout > 0 { var cancel context.CancelFunc; ctx, cancel = context.WithTimeout(ctx, *timeout); defer cancel() }
		result, err := runner.Run(ctx, *prompt); if err != nil { return err }
		_, err = fmt.Fprintln(out, result); return err
	}
	if *input == "" { return fmt.Errorf("missing -input; try: agentforge -tool echo -input 'hello'") }
	runner, err := agent.NewRunnerWithOptions(agent.Options{Timeout:*timeout}, tools.Echo{}); if err != nil { return err }
	result, err := runner.Run(context.Background(), *toolName, *input); if err != nil { return err }
	_, err = fmt.Fprintln(out, result.Output); return err
}
