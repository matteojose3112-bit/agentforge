# AgentForge v0.3 — Demo guide

This guide takes AgentForge from a clean checkout to a local tool execution. It includes a zero-cost local check and an optional model-backed demonstration.

## 1. Requirements

- Go 1.23 or newer
- Git
- An OpenAI API key only for the model-backed section

Clone the repository:

```bash
git clone https://github.com/matteojose3112-bit/agentforge.git
cd agentforge
```

## 2. Verify the local runtime (no API key, no model usage)

Run the test suite, static checks, build, and direct echo command:

```bash
go test ./...
go vet ./...
go build -o agentforge ./cmd/agentforge
./agentforge -version
./agentforge -tool echo -input "hello AgentForge"
```

Expected direct-tool output:

```text
hello AgentForge
```

Try the new read-only text analysis tool without an API key:

```bash
go run ./cmd/agentforge -tool text_analyze -input '{"text":"A small business can automate repetitive work."}'
```

It returns JSON containing word count, Unicode character count, line count, and an estimated reading time. The tool accepts up to 256 KiB of text per call.

Analyze a synthetic CRM lead-pipeline export without an API key:

```bash
go run ./cmd/agentforge -tool csv_analyze -input-file examples/lead_pipeline.json
```

The JSON input file contains the CSV payload, so you do not need to escape newlines or pass a long string through the shell. See [PORTFOLIO.md](PORTFOLIO.md) for the full scenario, expected output, and operator interpretation.

Run the business-specific lead pipeline analysis on the same fictional CRM export:

```bash
go run ./cmd/agentforge -tool lead_pipeline_analyze -input-file examples/lead_pipeline_analysis.json
```

The tool reports won, lost, and open leads; win rate among closed leads; estimated values by outcome; missing fields; and optional stage/source breakdowns. Won/lost stage names can be configured in the input file.

Generate a read-only sales operations snapshot from a synthetic CRM export:

```bash
go run ./cmd/agentforge -tool sales_ops_report -input-file examples/sales_ops_snapshot.json
```

This composes the CRM pipeline metrics into a concise report with observations, data-quality signals, and rule-based follow-up actions. Recommendations are review prompts grounded in the supplied snapshot; they are not forecasts, lead rankings, or automated CRM actions.

CSV analysis reports data rows, columns, header names, missing cells, and min/max/mean for columns whose non-empty values are all numeric. It accepts CSV input up to 1 MiB, 10,000 data rows, and 100 columns; the first record must be a unique, non-empty header row.

On PowerShell, the built executable is `agentforge.exe`:

```powershell
go test ./...
go vet ./...
go build -o agentforge.exe ./cmd/agentforge
.\agentforge.exe -version
.\agentforge.exe -tool echo -input "hello AgentForge"
```

## 3. Run the model-backed agent (API usage may be billed)

Start with the tracked [`.env.example`](../.env.example): copy it to `.env` and replace the placeholder with your key. The real `.env` file is ignored by Git. AgentForge reads the key from the process environment and does not automatically load dotenv files. Before making a model request, startup validation checks for a missing key, the untouched template placeholder, a `Bearer ` prefix, and whitespace in the value. This catches common setup errors locally; it cannot verify whether a key is active or authorized without contacting the API.

Load it into your current shell. Do not paste it into source code, command history, screenshots, or GitHub.

Linux/macOS:

```bash
set -a
source ./.env
set +a
go run ./cmd/agentforge -agent \
  -model your-supported-model-id \
  -prompt "Use the echo tool to repeat hello AgentForge" \
  -allow-permission tool.echo
```

PowerShell:

```powershell
$env:OPENAI_API_KEY = (Get-Content .env | Where-Object { $_ -match '^OPENAI_API_KEY=' } | Select-Object -First 1) -replace '^OPENAI_API_KEY=', ''
go run ./cmd/agentforge -agent `
  -model your-supported-model-id `
  -prompt "Use the echo tool to repeat hello AgentForge" `
  -allow-permission tool.echo
```

The model can request the registered `echo`, `text_analyze`, `csv_analyze`, `lead_pipeline_analyze`, or `sales_ops_report` function. AgentForge validates the arguments and the matching permission before execution, returns the tool output to the model, and prints the model's final response. Exact wording is model-generated and will vary.

Agent mode requires an explicit `-model` value. Pass a model ID supported by your OpenAI API account; AgentForge does not assume a default model ID.

## 4. Check the permission boundary

The runtime denies tool calls by default. The model-backed command above explicitly grants only `tool.echo`. To let the agent analyze text, grant `tool.text.analyze`; to analyze CSV data, grant `tool.csv.analyze`; to analyze CRM pipeline data, grant `tool.leads.analyze`; to generate a sales operations report, grant `tool.sales_ops.report` as well.

You can verify the default-deny behavior by omitting `-allow-permission tool.echo`. This still makes a model API request and may incur a small charge, but AgentForge should reject the requested tool call rather than execute it.

## 5. CLI help and reference

Print the built-in help at any time without running a tool or making an API request:

```bash
go run ./cmd/agentforge -h
```

For a compiled binary, use `./agentforge -h` (PowerShell: `./agentforge.exe -h`). The help output groups the direct and agent modes, lists flags and defaults, explains the default-deny permission behavior, and includes copyable examples. Both `-h` and `-help` are supported and exit successfully.


| Flag | Purpose |
| --- | --- |
| `-version` | Print the CLI version, Git commit, and build timestamp |
| `-tool` | Select a registered tool in direct mode (default: `echo`) |
| `-input` | JSON input for direct tool mode; mutually exclusive with `-input-file` |
| `-input-file` | Read direct tool JSON input from a file; mutually exclusive with `-input` |
| `-agent` | Enable model-backed agent mode |
| `-prompt` | User prompt for agent mode |
| `-model` | API model ID |
| `-timeout` | Optional execution timeout, e.g. `30s` |
| `-allow-permission` | Comma-separated permission allow-list for agent mode; empty means deny all |

## 6. What this demo proves — and does not prove

It demonstrates one model provider, five local tools, a bounded model/tool loop, schema validation, and explicit permissions. The text- and CSV-analysis tools can also be exercised locally without a model API key. It does not demonstrate persistent conversations, durable task storage, human approval workflows, background jobs, multi-agent coordination, or production-grade sandboxing.

The automated test suite uses a local HTTP test server; CI does not make paid model calls. A successful local test run is not a substitute for recording a successful live API smoke test.
