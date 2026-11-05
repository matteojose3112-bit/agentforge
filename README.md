# AgentForge

**A lightweight Go runtime for model-driven tool execution and practical workflow automation.**

> **Live project site:** [AgentForge on GitHub Pages](https://matteojose3112-bit.github.io/agentforge/) · **Repository:** [GitHub](https://github.com/matteojose3112-bit/agentforge)

> **Release status:** v0.3.0 foundation implemented. Early-stage developer project; not production-ready. The live model/tool workflow has not yet been verified with a real API key.

AgentForge keeps orchestration in Go and puts model providers behind a small interface. Models can request capabilities, but only tools explicitly registered by the application can execute, and permissions are denied unless explicitly granted.

## Current project status

The v0.3.0 runtime and local business-analysis examples are implemented. The first v0.4 slice now adds a read-only sales operations report that combines existing CRM pipeline metrics with observations and rule-based review actions. Next priorities are broader evaluation datasets, a guided operator workflow, and full model-loop tests before treating this as a complete assistant.

**What is verified in the repository:** local tool execution paths, deterministic analysis tools, automated tests, provider adapter tests using a local HTTP test server, and CI checks that do not require paid model API calls.

**What is not yet verified or implemented:** a successful live model/tool cycle against an actual provider account; persisted conversations or task history; approval checkpoints; scheduled/background workflows; and production deployment or security readiness. See [ROADMAP.md](ROADMAP.md) for milestone status.

## Why AgentForge

- **Small by design:** one Go process, standard-library HTTP integration, no mandatory database or queue.
- **Provider boundary:** model orchestration is independent of the first OpenAI adapter.
- **Controlled tool use:** registered tools, schema-validated inputs and outputs, explicit permission allow-list.
- **Bounded execution:** limited model/tool turns, context cancellation, execution timeouts, and opt-in retries in the lower-level runner.
- **Testable:** provider behavior is tested with a local HTTP server; CI runs tests, vet, and a CLI build without paid API calls.

## What works today

- Go CLI with direct tool mode, JSON input-file support, and model-backed agent mode.
- Typed tool registry and task execution lifecycle.
- Vendor-independent model provider interface.
- OpenAI Responses API adapter using Go's standard `net/http` client.
- JSON Schema subset validation for model-supplied tool arguments and tool outputs.
- Default-deny tool permissions.
- Local echo tool for a complete first tool-call cycle.
- Read-only text analysis tool for word, Unicode character, line, and estimated reading-time metrics.
- Bounded CSV analysis for row/column counts, missing cells, and numeric-column summaries.
- CRM lead-pipeline analysis for won/lost/open counts, closed-lead win rate, pipeline value, missing fields, and stage/source breakdowns.
- Read-only sales operations report combining deterministic pipeline metrics with rule-based observations and follow-up recommendations.
- Unit tests, Go vet, and GitHub Actions CI.

## Quick start

Requirements: Go 1.23 or newer.

```bash
git clone https://github.com/matteojose3112-bit/agentforge.git
cd agentforge

go test ./...
go vet ./...
go run ./cmd/agentforge -tool echo -input "hello AgentForge"
go run ./cmd/agentforge -tool csv_analyze -input-file examples/lead_pipeline.json
go run ./cmd/agentforge -tool lead_pipeline_analyze -input-file examples/lead_pipeline_analysis.json
go run ./cmd/agentforge -tool sales_ops_report -input-file examples/sales_ops_snapshot.json
```

Expected output:

```text
hello AgentForge
```

## Run a model-backed agent

The first provider uses the OpenAI Responses API. Set an API key in your local environment; never commit it to the repository.

Linux/macOS:

```bash
export OPENAI_API_KEY="your-api-key"
go run ./cmd/agentforge -agent \
  -model your-supported-model-id \
  -prompt "Use the echo tool to repeat hello AgentForge" \
  -allow-permission tool.echo
```

PowerShell:

```powershell
$env:OPENAI_API_KEY = "your-api-key"
go run ./cmd/agentforge -agent `
  -model your-supported-model-id `
  -prompt "Use the echo tool to repeat hello AgentForge" `
  -allow-permission tool.echo
```

Agent mode requires an explicit `-model` value supported by your OpenAI API account; no unverified default model is assumed. Model API usage may incur charges. Tool permissions are denied by default; grant only the exact permission strings required. Agent mode exposes local echo, text analysis, CSV profiling, CRM lead-pipeline analysis, and the sales operations report tools.

See the [step-by-step demo guide](docs/DEMO.md) for a no-key local check, model-backed walkthrough, CLI flags, and permission-boundary test. The [CRM lead pipeline case study](docs/PORTFOLIO.md) is a reproducible business-facing example using synthetic data.

## How it works

```text
User prompt
    |
    v
AgentForge runtime -----> Model provider (OpenAI adapter)
    |                              |
    |<------ text / tool call -----|
    |
    v
Registered tool lookup
    |
    v
Permission check (default deny)
    |
    v
Input schema validation
    |
    v
Tool execution
    |
    v
Output schema validation
    |
    v
Tool result returned to model
    |
    v
Final response
```

## Architecture

- `cmd/agentforge`: CLI entry point and flags.
- `internal/agent`: provider-neutral contracts, model/tool loop, schema validation, permissions, and reliable execution.
- `internal/providers/openai`: OpenAI Responses API adapter.
- `internal/tools`: concrete capabilities and their JSON Schema descriptions.

Go owns orchestration and tool execution. Provider-specific request/response translation stays behind `agent.Provider`. See [architecture notes](docs/ARCHITECTURE.md).

## Safety boundaries and limitations

- A nil permission policy denies all model-requested tool calls.
- Tool arguments are validated before execution and tool outputs before being returned to the model.
- The supported JSON Schema subset includes object, string, number, integer, boolean, array, enum, required, properties, and `additionalProperties:false`; this is not full JSON Schema.
- The model cannot register tools or grant itself permissions.
- API keys are read from environment variables and must not be committed.
- Model/tool turns are bounded and provider calls inherit context cancellation.
- Requests use `store:false`; this does not replace reviewing the provider's applicable data-processing and retention terms.
- Only OpenAI is implemented today; the provider interface is vendor-neutral.
- Agent mode includes local echo, text analysis, CSV profiling, CRM lead-pipeline analysis, and sales operations reporting tools. Text analysis accepts at most 256 KiB per call; CSV tools accept at most 1 MiB, 10,000 data rows, and 100 columns.
- Persistent conversation history, durable task storage, human approval workflows, background workers, and production security review are not implemented.
- CI does not call a paid live model API. The live smoke test remains a local release check.

## Development

Run the local checks and build:

```bash
go test ./...
go vet ./...
go build ./cmd/agentforge
```

The CLI's `-version` flag prints the version, Git commit, and build timestamp. Local builds use the version with `unknown` commit/build metadata unless values are supplied through linker flags.

Example release-style build (Linux/macOS):

```bash
VERSION=0.3.0
COMMIT=$(git rev-parse --short HEAD)
BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)
go build -ldflags "-X main.version=$VERSION -X main.commit=$COMMIT -X main.buildDate=$BUILD_DATE" -o agentforge ./cmd/agentforge
./agentforge -version
```

PowerShell:

```powershell
$version = "0.3.0"
$commit = git rev-parse --short HEAD
$buildDate = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
go build -ldflags "-X main.version=$version -X main.commit=$commit -X main.buildDate=$buildDate" -o agentforge.exe ./cmd/agentforge
.\agentforge.exe -version
```

See [ROADMAP.md](ROADMAP.md), [CHANGELOG.md](CHANGELOG.md), and [CONTRIBUTING.md](CONTRIBUTING.md).

## Contributing

Keep changes small, typed, testable, and explicit about side effects. See [contribution guidelines](CONTRIBUTING.md).

## Documentation review — 2026-11-05

The repository is at the v0.3.0 implementation baseline. The local CLI, provider abstraction, OpenAI adapter, bounded model/tool loop, schema validation, default-deny permissions, text and CSV analysis, CRM lead-pipeline metrics, synthetic examples, and GitHub Pages project site are present. Automated tests and CI cover local behavior without paid model calls. A real provider-backed tool cycle has not yet been confirmed, so the project is not presented as a verified live AI assistant or production system.

The active v0.4 work is a focused AI Sales Operations Assistant built on the existing CRM lead-pipeline analysis capabilities. Its first read-only report slice is implemented; next priorities are broader synthetic evaluation cases, a guided workflow, and full agent-loop testing before considering persistence, scheduling, dashboards, or more infrastructure.
