package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/matteojose3112-bit/agentforge/internal/agent"
	"github.com/matteojose3112-bit/agentforge/internal/tools"
)

const version = "0.2.0"

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
	toolName := flags.String("tool", "echo", "registered tool to execute")
	input := flags.String("input", "", "task input")
	timeout := flags.Duration("timeout", 0, "execution timeout (for example 30s)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *showVersion {
		_, err := fmt.Fprintln(out, version)
		return err
	}
	if *input == "" {
		return fmt.Errorf("missing -input; try: agentforge -tool echo -input 'hello'")
	}
	runner, err := agent.NewRunnerWithOptions(agent.Options{Timeout: *timeout}, tools.Echo{})
	if err != nil {
		return err
	}
	result, err := runner.Run(context.Background(), *toolName, *input)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, result.Output)
	return err
}

