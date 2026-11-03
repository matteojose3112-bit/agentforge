# AgentForge

A Go-based foundation for reliable AI-agent execution, tool calling, and workflow automation.

> **Status:** v0.1.0 — execution foundation. This is an early development scaffold, not yet an LLM-powered autonomous agent.

## What works today

- A small Go CLI.
- A typed tool contract and explicit tool registration.
- A runner that rejects unknown tools and empty input.
- Context propagation and wrapped tool errors.
- A local echo tool that exercises the full execution path.
- Unit tests, Go vet, and GitHub Actions CI.

## Quick start

Requirements: Go 1.23 or newer.

```bash
go test ./...
go vet ./...
go run ./cmd/agentforge -tool echo -input "hello AgentForge"
```

Expected output:

```text
hello AgentForge
```

Print the current scaffold version:

```bash
go run ./cmd/agentforge -version
```

Build a binary:

```bash
make build
./bin/agentforge -tool echo -input "hello"
```

## Project structure

```text
agentforge/
├── cmd/agentforge/       # CLI entry point
├── internal/
│   ├── agent/            # Tool contract and execution runner
│   └── tools/            # Built-in example tools
├── docs/                 # Architecture and design notes
├── .github/workflows/    # Continuous integration
├── go.mod
├── Makefile
├── ROADMAP.md
└── CHANGELOG.md
```

## Design direction

Go owns the runtime and control flow: task execution, tool registration, lifecycle management, reliability, and orchestration. Provider-specific AI integrations can be added behind interfaces when the execution foundation is ready. Python remains an option for specialized AI/data workloads, not a parallel orchestration runtime.

The project is intentionally starting as one process and one binary. A database, queue, multi-agent layer, and dashboard are not prerequisites for the first useful release.

## Current limitations

- The only built-in tool is a local echo example.
- No LLM provider, planning loop, persistent memory, or workflow scheduler exists yet.
- The CLI is a thin development interface; it is not a production service.
- Retry policy, approval gates, and durable task history are future work.

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md), [ROADMAP.md](ROADMAP.md), and [CHANGELOG.md](CHANGELOG.md).

## Contributing

Keep changes small, typed, testable, and explicit about side effects. Run `go test ./...` and `go vet ./...` before opening a pull request.
