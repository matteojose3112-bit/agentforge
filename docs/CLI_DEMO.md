# AgentForge CLI demo — v0.4 Sales Operations

This is the repeatable, no-key command-line demonstration for the read-only Sales Operations workflow. It uses checked-in synthetic CRM fixtures and the real Go CLI; it does not contact a model provider or modify a CRM.

## Requirements

- Go 1.23 or newer
- Git

## 1. Clone and verify

```bash
git clone https://github.com/matteojose3112-bit/agentforge.git
cd agentforge
go test ./...
go vet ./...
go build ./cmd/agentforge
```

## 2. Generate a report

Run the healthy synthetic pipeline through the registered report tool:

```bash
go run ./cmd/agentforge -tool sales_ops_report -input-file examples/evaluations/healthy_pipeline.json
```

The command prints one JSON report. The key deterministic values are:

| Metric | Expected |
| --- | ---: |
| Total leads | 6 |
| Closed leads | 3 |
| Open leads | 3 |
| Win rate (won / closed) | 66.7% |
| Open pipeline value | 40,000 |
| Missing statuses / values / sources | 0 / 0 / 0 |

The report also includes observations, rule-based review prompts, and limitations. In particular, it describes only this snapshot; it does not forecast results or update a CRM.

## 3. Compare different pipeline conditions

Run the other fixtures to see how the same tool handles different synthetic scenarios:

```bash
go run ./cmd/agentforge -tool sales_ops_report -input-file examples/evaluations/weak_pipeline.json
go run ./cmd/agentforge -tool sales_ops_report -input-file examples/evaluations/incomplete_pipeline.json
go run ./cmd/agentforge -tool sales_ops_report -input-file examples/evaluations/edge_case_pipeline.json
```

Expected scenario summaries are documented in [SALES_OPS_EVALUATION.md](SALES_OPS_EVALUATION.md). The fixtures are synthetic and should not be presented as client data.

## 4. PowerShell commands

The same commands work in PowerShell:

```powershell
go test ./...
go vet ./...
go build ./cmd/agentforge
go run ./cmd/agentforge -tool sales_ops_report -input-file examples/evaluations/healthy_pipeline.json
go run ./cmd/agentforge -tool sales_ops_report -input-file examples/evaluations/weak_pipeline.json
go run ./cmd/agentforge -tool sales_ops_report -input-file examples/evaluations/incomplete_pipeline.json
go run ./cmd/agentforge -tool sales_ops_report -input-file examples/evaluations/edge_case_pipeline.json
```

## 5. Optional CSV inspection

To inspect a CRM CSV before mapping it, use the separate inspection tool:

```bash
go run ./cmd/agentforge -tool csv_inspect -input-file examples/evaluations/inspection_input.json
```

Inspection returns column names, row counts, missing/non-empty counts, and sample values. The operator must review and confirm mappings before a separate report run. The CLI is stateless: supply the CSV again in the report input and confirm the mappings in that later run. See [CSV_INSPECTION.md](CSV_INSPECTION.md).

## What this demo proves

- The Go CLI can execute the actual registered sales report tool locally.
- The report metrics are deterministic across checked-in scenarios.
- The workflow is read-only and can be demonstrated without a model key or API charges.

It does **not** prove a live model/provider tool cycle, persistent conversations, durable approvals, production readiness, or real-world business impact. The live provider-backed smoke test remains a separate release check; see [ROADMAP.md](../ROADMAP.md).
