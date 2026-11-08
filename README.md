# AgentForge

**A lightweight Go runtime for model-driven tool execution and practical workflow automation.**

> **Live project site:** [AgentForge on GitHub Pages](https://matteojose3112-bit.github.io/agentforge/) · **Repository:** [GitHub](https://github.com/matteojose3112-bit/agentforge)

> **Project status:** v0.4 read-only Sales Operations workflow implemented and API deployed on Render Free. Early-stage developer project; not production-ready. The authenticated CSV inspection flow is verified; authenticated report generation and the live model/tool workflow still require end-to-end validation.

AgentForge keeps orchestration in Go and puts model providers behind a small interface. Models can request capabilities, but only tools explicitly registered by the application can execute, and permissions are denied unless explicitly granted.

## Current project status

The v0.3.0 runtime and v0.4 read-only Sales Operations workflow are implemented. The guided workspace is connected to a deployed Go HTTP API on Render Free. The authenticated CSV inspection flow has now been exercised successfully from the GitHub Pages workspace against the live Go runtime. The authenticated sales report request and a live model/tool cycle remain validation items.

**What is verified:** local tool execution paths, deterministic analysis tools, automated tests, provider adapter tests using a local HTTP test server, and CI checks that do not require paid model API calls. The Render deployment is live; `/healthz` returns HTTP 200, while unauthenticated requests to `/api/inspect` and `/api/report` return HTTP 401.

**Not yet verified or implemented:** successful authenticated `/api/report` requests against the deployed service; a live model/tool cycle against an actual provider account; persisted conversations or task history; durable approval workflows; scheduled/background workflows; and production security readiness. See [ROADMAP.md](ROADMAP.md) for milestone status.

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
- Read-only CRM CSV inspection with sample values and an enforced new-user-turn checkpoint before sales report generation.
- CRM lead-pipeline analysis for won/lost/open counts, closed-lead win rate, pipeline value, missing fields, and stage/source breakdowns.
- Read-only sales operations report combining deterministic pipeline metrics with rule-based observations and follow-up recommendations.
- Unit tests, Go vet, and GitHub Actions CI.
- Read-only HTTP API exposing the existing CSV inspection and sales operations report tools to the guided workspace; deployed on Render Free with a generated bearer-token secret and GitHub Pages CORS allow-list.

## Quick start

Requirements: Go 1.23 or newer.

```bash
git clone https://github.com/matteojose3112-bit/agentforge.git
cd agentforge

go test ./...
go vet ./...
go run ./cmd/agentforge -h
go run ./cmd/agentforge -version
AGENTFORGE_ALLOWED_ORIGINS=http://localhost:5500 go run ./cmd/agentforge-server
go run ./cmd/agentforge -tool echo -input "hello AgentForge"
go run ./cmd/agentforge -tool csv_analyze -input-file examples/lead_pipeline.json
go run ./cmd/agentforge -tool lead_pipeline_analyze -input-file examples/lead_pipeline_analysis.json
go run ./cmd/agentforge -tool sales_ops_report -input-file examples/sales_ops_snapshot.json
```

The `-version` command prints the application version, Git commit, and build timestamp. A local `go run` uses the default version and reports `unknown` for commit/build metadata unless those values are embedded in a release build.

Use `go run ./cmd/agentforge -h` to view the CLI modes, all flags, safety defaults, and copyable examples. The compiled binary supports the same help output with `./agentforge -h` (PowerShell: `./agentforge.exe -h`).

Expected output:

```text
hello AgentForge
```

## Run a model-backed agent

The first provider uses the OpenAI Responses API. You can start from the tracked [`.env.example`](.env.example) template. Copy it to `.env`, replace the placeholder, and load the variable into your shell. The `.env` file is ignored by Git. AgentForge reads `OPENAI_API_KEY` from the process environment; it does **not** automatically load dotenv files. Before starting agent mode, it checks that the variable is present and rejects obvious configuration mistakes such as the untouched template value, a `Bearer ` prefix, or whitespace. This is a local sanity check—not a remote API-key validity check; revoked or unauthorized keys are reported by the API.

Linux/macOS (after copying `.env.example` to `.env` and editing it):

```bash
set -a
source ./.env
set +a
go run ./cmd/agentforge -agent \\
  -model your-supported-model-id \\
  -prompt "Use the echo tool to repeat hello AgentForge" \\
  -allow-permission tool.echo
```

PowerShell (after editing .env):

```powershell
$env:OPENAI_API_KEY = (Get-Content .env | Where-Object { $_ -match '^OPENAI_API_KEY=' } | Select-Object -First 1) -replace '^OPENAI_API_KEY=', ''
go run ./cmd/agentforge -agent `
  -model your-supported-model-id `
  -prompt "Use the echo tool to repeat hello AgentForge" `
  -allow-permission tool.echo
```

Agent mode requires an explicit `-model` value supported by your OpenAI API account; no unverified default model is assumed. Model API usage may incur charges. Tool permissions are denied by default; grant only the exact permission strings required. Agent mode exposes local echo, text analysis, CSV profiling and inspection, CRM lead-pipeline analysis, and the sales operations report tools. CSV inspection and report generation require separate permissions; after inspection, the runtime blocks report generation until a new user turn.

Live API: [AgentForge API on Render](https://agentforge-api-s4zx.onrender.com) · Health check: [ `/healthz` ](https://agentforge-api-s4zx.onrender.com/healthz). The health route is public; `/api/inspect` and `/api/report` require the configured bearer token. See the [workspace API guide](docs/API.md) for endpoints, local startup, CORS configuration, and deployment boundaries. The [step-by-step demo guide](docs/DEMO.md) for a no-key local check, model-backed walkthrough, CLI flags, and permission-boundary test. The [CRM lead pipeline case study](docs/PORTFOLIO.md) is a reproducible business-facing example using synthetic data. The [Sales Operations evaluation set](docs/SALES_OPS_EVALUATION.md) and [guided CSV inspection workflow](docs/CSV_INSPECTION.md) document four synthetic scenarios, deterministic acceptance criteria, and the operator-confirmation boundary.

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
- API keys are read from environment variables and must not be committed. The repository includes a placeholder `.env.example`; the real `.env` file is ignored by Git, and the CLI does not load dotenv files automatically.
- Model/tool turns are bounded and provider calls inherit context cancellation.
- Requests use `store:false`; this does not replace reviewing the provider's applicable data-processing and retention terms.
- Only OpenAI is implemented today; the provider interface is vendor-neutral.
- Agent mode includes local echo, text analysis, CSV profiling, CRM lead-pipeline analysis, and sales operations reporting tools. Text analysis accepts at most 256 KiB per call; CSV tools accept at most 1 MiB, 10,000 data rows, and 100 columns.
- Persistent conversation history, durable task storage, human approval workflows, background workers, and production security review are not implemented.
- CI does not call a paid live model API. The live smoke test remains a local release check.

## CLI help

The CLI provides contextual help for direct tool execution and model-backed agent mode. Direct-mode `-input` and `-input-file` pass the selected tool its input payload: structured analysis tools expect JSON, while the `echo` tool accepts plain text.

```bash
go run ./cmd/agentforge -h
```

Help includes flag descriptions, input-file versus inline-input behavior, the agent mode requirements, the default-deny permission model, and common command examples. `-h` and `-help` display help and exit successfully.

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
.\\agentforge.exe -version
```

See [ROADMAP.md](ROADMAP.md), [CHANGELOG.md](CHANGELOG.md), and [CONTRIBUTING.md](CONTRIBUTING.md).

## Contributing

Keep changes small, typed, testable, and explicit about side effects. See [contribution guidelines](CONTRIBUTING.md).

## Documentation review — 2026-11-08

The v0.3 runtime foundation and v0.4 read-only Sales Operations workflow are implemented. The project includes deterministic CRM analysis, synthetic evaluation fixtures, a guided CSV workspace, a Go HTTP API, unit tests, and CI. The API is deployed at https://agentforge-api-s4zx.onrender.com on Render Free. Live verification confirms `/healthz` returns HTTP 200 and both protected data-processing routes reject unauthenticated requests with HTTP 401. Authenticated `/api/inspect` has been successfully exercised from the deployed workspace using the sample CRM CSV. The authenticated `/api/report` flow and a real provider-backed model/tool cycle remain outstanding. The latest Go CI and GitHub Pages deployment workflows both passed on 2026-11-08 (commit `b951b78`).

The workspace is designed as an operator-guided AI workspace, not another dashboard. Its live authenticated CSV inspection path has been verified end to end; the report-generation path is the next integration check. It can use the deployed API after the operator supplies the bearer token in the connection panel. The API remains read-only: it inspects CSVs and generates deterministic reports without CRM writes, outbound messages, or other external side effects. Persistence, scheduling, background workers, and broader production security work remain out of scope until justified by usage.
