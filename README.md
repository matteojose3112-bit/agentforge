# AgentForge

A Go-based foundation for reliable AI-agent execution, tool calling, and workflow automation.

> **Status:** v0.2.0 — reliable task execution. This is an execution runtime foundation, not yet an LLM-powered autonomous agent.

## What works today

- A small Go CLI and an explicit typed tool registry.
- Structured task records with unique IDs, lifecycle status, UTC timestamps, attempts, output, and errors.
- Execution events and structured `log/slog` logging.
- Context cancellation and optional execution timeouts.
- Bounded retries (maximum five attempts), only for tools that explicitly opt in as retryable.
- A local echo tool, unit tests, Go vet, and GitHub Actions CI.

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

Set an execution timeout:

```bash
go run ./cmd/agentforge -timeout 30s -tool echo -input "hello"
```

Print the current version:

```bash
go run ./cmd/agentforge -version
```

Build a binary:

```bash
make build
./bin/agentforge -tool echo -input "hello"
```

## Task lifecycle and reliability

Each call to `Runner.Run` creates a task record. A task moves through `pending`, `running`, and one terminal state: `succeeded`, `failed`, or `cancelled`. The returned `Result` includes the task ID, timestamps, duration, attempt count, output, and error text.

The runner emits lifecycle and attempt events through an optional callback and writes structured records with Go's `log/slog`. The runner itself does not persist task history; durable storage is a later milestone.

Timeouts are optional and apply to the whole execution, including retry waits. Cancellation is propagated through the tool context. Tools must cooperate with context cancellation; Go cannot forcibly stop a tool implementation that ignores its context.

Retries default to one attempt. A tool must implement `RetryableTool` and return `true` before the configured retry policy is used. The maximum is five total attempts. This is a safety boundary, not a guarantee that a tool is idempotent: tool authors are responsible for opting in only when repeating the operation is safe.

## Project structure

```text
agentforge/
├── cmd/agentforge/       # CLI entry point
├── internal/
│   ├── agent/            # Task lifecycle, runner, events, reliability
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

The project intentionally starts as one process and one binary. A database, queue, multi-agent layer, and dashboard are not prerequisites for the first useful release.

## Current limitations

- The only built-in tool is a local echo example.
- No LLM provider, planning loop, persistent memory, or workflow scheduler exists yet.
- Execution events are observable but not durably stored.
- A tool that ignores its context cannot be forcibly interrupted.
- Human approval gates and durable task history remain future work.

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md), [ROADMAP.md](ROADMAP.md), and [CHANGELOG.md](CHANGELOG.md).

## Contributing

Keep changes small, typed, testable, and explicit about side effects. Run `go test ./...` and `go vet ./...` before opening a pull request.
