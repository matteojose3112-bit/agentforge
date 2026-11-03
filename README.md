# AgentForge

A Go-based foundation for reliable AI-agent execution, tool calling, and workflow automation.

> **Status:** v0.3.0 — model and tool integration. This is an early agent runtime, not a production-ready autonomous system.

## What works today

- A small Go CLI and an explicit typed tool registry.
- Structured task records, lifecycle events, cancellation, timeouts, and bounded opt-in retries.
- A vendor-independent model provider interface.
- An OpenAI Responses API provider using Go's standard HTTP client.
- JSON Schema validation for model-supplied tool arguments.
- Explicit default-deny tool permissions.
- A local echo tool, unit tests, Go vet, and GitHub Actions CI.

## Quick start

Requirements: Go 1.23 or newer.

```bash
go test ./...
go vet ./...
go run ./cmd/agentforge -tool echo -input "hello AgentForge"
```

## Run a model-backed agent

The OpenAI integration uses the Responses API. Create an API key in your OpenAI developer account and set it locally; never commit it to the repository.

Linux/macOS:

```bash
export OPENAI_API_KEY="your-api-key"
go run ./cmd/agentforge -agent -model gpt-6-luna -prompt "Use the echo tool to repeat hello" -allow-permission tool.echo
```

PowerShell:

```powershell
$env:OPENAI_API_KEY = "your-api-key"
go run ./cmd/agentforge -agent -model gpt-6-luna -prompt "Use the echo tool to repeat hello" -allow-permission tool.echo
```

Tool permissions are denied by default. Grant only the exact permission strings required by the task using `-allow-permission`. The current model mode exposes only the local echo tool. Model API usage may incur charges on the API account.

Direct tool mode remains available:

```bash
go run ./cmd/agentforge -tool echo -input "hello AgentForge"
```

## Architecture

- `internal/agent`: provider-neutral interfaces, tool loop, schema validation, permissions, and reliable execution.
- `internal/providers/openai`: OpenAI Responses API adapter.
- `internal/tools`: local capabilities and JSON-schema-described model tools.
- `cmd/agentforge`: CLI entry point.

Go owns orchestration and tool execution. Provider-specific code is isolated behind `agent.Provider`. The first implementation uses the standard library HTTP client to keep the runtime small.

## Safety boundaries

- Tool permissions use an explicit allow-list; a nil policy denies all calls.
- Tool arguments are validated before execution against a supported JSON Schema subset.
- Unknown tools and invalid arguments are rejected.
- The model can request tools, but it cannot register new tools or grant itself permissions.
- API keys are read from environment variables and must not be committed.
- The current schema validator supports object, string, number, integer, boolean, array, enum, required, properties, and `additionalProperties:false`. It is not a complete JSON Schema implementation.
- The model loop has a bounded number of turns. Provider calls inherit context cancellation.

## Current limitations

- OpenAI is the only implemented provider; the provider interface is vendor-neutral.
- Only the local echo tool is included in agent mode.
- No persistent conversation history, durable task store, human approval workflow, or background worker exists yet.
- No live API call is part of CI; provider behavior is tested with an HTTP test server.
- This is an early development foundation and has not undergone a production security review.

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md), [ROADMAP.md](ROADMAP.md), and [CHANGELOG.md](CHANGELOG.md).

## Contributing

Keep changes small, typed, testable, and explicit about side effects. Run `go test ./...` and `go vet ./...` before opening a pull request.
