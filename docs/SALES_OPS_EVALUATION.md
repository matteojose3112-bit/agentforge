# Sales Operations Evaluation Set

This evaluation set defines repeatable, synthetic CRM snapshots for the v0.4 sales operations workflow. The fixtures contain no customer data and are intended for deterministic local tests and portfolio demonstrations.

## Run a fixture

From the repository root:

```bash
go run ./cmd/agentforge -tool sales_ops_report -input-file examples/evaluations/healthy_pipeline.json
go run ./cmd/agentforge -tool sales_ops_report -input-file examples/evaluations/weak_pipeline.json
go run ./cmd/agentforge -tool sales_ops_report -input-file examples/evaluations/incomplete_pipeline.json
go run ./cmd/agentforge -tool sales_ops_report -input-file examples/evaluations/edge_case_pipeline.json
```

No model API key or external service is required for these deterministic report runs.

## Scenario inventory

| Fixture | Purpose | Expected pipeline facts |
|---|---|---|
| `healthy_pipeline.json` | Balanced, complete pipeline with both closed outcomes and active opportunities | 6 leads; 2 won; 1 lost; 3 open; 3 closed; 66.7% win rate; 40,000 open pipeline value; no missing configured fields |
| `weak_pipeline.json` | Low observed win rate with a larger open pipeline | 8 leads; 1 won; 3 lost; 4 open; 4 closed; 25.0% win rate; 47,000 open pipeline value; no missing configured fields |
| `incomplete_pipeline.json` | Missing stage, estimated-value, and source fields | 7 leads; 1 won; 1 lost; 4 open; 2 closed; 50.0% win rate; 17,000 open pipeline value; 1 missing status; 2 missing values; 2 missing sources |
| `edge_case_pipeline.json` | Header-only CSV, custom outcome labels, and zero records | 0 leads; 0 closed; 0 open; 0% numeric win-rate field with explanatory unavailable-context observation; zero missing fields |

## Deterministic acceptance criteria

### Metrics

- Total leads equals the number of CSV data rows, excluding the header.
- Won and lost counts use the explicitly configured status labels and case-insensitive matching.
- Closed leads equals won plus lost.
- Open leads counts non-empty statuses not mapped to won or lost. Rows with missing status are not classified as open.
- Win rate is won divided by closed leads, multiplied by 100. When there are no closed leads, the numeric field is zero and the report must explain that a meaningful rate is unavailable.
- Open pipeline value includes numeric values only for rows with a non-empty status that is neither won nor lost.
- Missing status, value, and source counts are computed only for their configured columns.
- A missing value is not silently treated as a known zero for data-quality reporting.
- Empty/header-only datasets return a valid zero-count report rather than fabricated activity.

### Observations and recommendations

- Every numeric statement in the report must match the deterministic metrics returned by the pipeline analyzer.
- The report may describe counts, totals, missing fields, and the presence or absence of closed outcomes.
- Open leads may trigger a general review prompt to confirm next steps and ownership.
- Missing statuses, values, or sources may trigger only the corresponding data-completion recommendation.
- When closed leads exist, the report may recommend reviewing recorded outcomes; when none exist, it must recommend collecting classified closed outcomes.
- Recommendations are review prompts, not assertions that a process is broken.
- No recommendation may invent a cause, forecast revenue, infer conversion trends, benchmark performance, or rank individual leads.

### Input and safety behavior

- Valid fixtures must execute through the registered read-only `sales_ops_report` capability.
- Malformed JSON, malformed CSV, missing configured columns, ragged rows, and non-numeric non-empty values must return errors rather than partial or invented reports.
- The capability remains behind the explicit `tool.sales_ops.report` permission in agent mode; no permission is granted by the fixture or model.
- Evaluation runs must not write to a CRM, send messages, or trigger external side effects.
- All example data must remain clearly identified as synthetic.

## Evidence boundary

The CSV analyzer calculates facts. The sales operations report adds deterministic, rule-based observations and follow-up prompts. A model may later help an operator map unfamiliar CRM columns or explain the report, but it must not alter the calculated metrics or present unsupported interpretation as fact.

These fixtures establish the first repeatable evaluation baseline. They do not prove live provider integration, production readiness, or real-world business impact.
