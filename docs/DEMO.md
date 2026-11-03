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

On PowerShell, the built executable is `agentforge.exe`:

```powershell
go test ./...
go vet ./...
go build -o agentforge.exe ./cmd/agentforge
.\agentforge.exe -version
.\agentforge.exe -tool echo -input "hello AgentForge"
```

## 3. Run the model-backed agent (API usage may be billed)

Set the key in your current shell. Do not paste it into source code, command history, screenshots, or GitHub.

Linux/macOS:

```bash
export OPENAI_API_KEY="your-api-key"
go run ./cmd/agentforge -agent \
  -model gpt-6-luna \
  -prompt "Use the echo tool to repeat hello AgentForge" \
  -allow-permission tool.echo
```

PowerShell:

```powershell
$env:OPENAI_API_KEY = "your-api-key"
go run ./cmd/agentforge -agent `
  -model gpt-6-luna `
  -prompt "Use the echo tool to repeat hello AgentForge" `
  -allow-permission tool.echo
```

The model should request the registered `echo` function, AgentForge should validate its arguments and permission, execute it, return the tool output to the model, and print the model's final response. Exact wording is model-generated and will vary.

To choose a different API model, pass its exact supported model ID with `-model`. The default is `gpt-6-luna`.

## 4. Check the permission boundary

The runtime denies tool calls by default. The model-backed command above explicitly grants only `tool.echo`.

You can verify the default-deny behavior by omitting `-allow-permission tool.echo`. This still makes a model API request and may incur a small charge, but AgentForge should reject the requested tool call rather than execute it.

## 5. CLI reference

| Flag | Purpose |
| --- | --- |
| `-version` | Print the CLI version |
| `-tool` | Select a registered tool in direct mode |
| `-input` | Input for direct tool mode |
| `-agent` | Enable model-backed agent mode |
| `-prompt` | User prompt for agent mode |
| `-model` | API model ID |
| `-timeout` | Optional execution timeout, e.g. `30s` |
| `-allow-permission` | Comma-separated permission allow-list; empty means deny all |

## 6. What this demo proves — and does not prove

It demonstrates one model provider, one local tool, a bounded model/tool loop, schema validation, and explicit permissions. It does not demonstrate persistent conversations, durable task storage, human approval workflows, background jobs, multi-agent coordination, or production-grade sandboxing.

The automated test suite uses a local HTTP test server; CI does not make paid model calls. A successful local test run is not a substitute for recording a successful live API smoke test.
