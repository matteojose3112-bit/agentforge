# Guided CRM CSV Inspection

AgentForge now includes a read-only `csv_inspect` tool to help an operator understand an unfamiliar CRM export before configuring a sales report.

## What it returns

- CSV row count and column headers.
- Non-empty and missing cell counts for each column.
- Up to five distinct, non-empty sample values per column.
- No inferred business meaning, calculated sales metrics, or external side effects.

Inspection is bounded by the same CSV limits as the existing analyzer: 1 MiB, 10,000 data rows, and 100 columns. Ragged rows, duplicate/empty headers, invalid UTF-8, and malformed CSV return errors.

## Run it directly

The tool accepts a JSON payload with a `csv` string. For a small export, create a JSON input file:

```json
{
  "csv": "Lead ID,Opportunity Stage,Deal Amount,Lead Source\n001,Qualified,1200,Referral\n002,Closed Won,2500,Google Ads\n"
}
```

Then run:

```bash
go run ./cmd/agentforge -tool csv_inspect -input-file examples/evaluations/inspection_input.json
```

For a real export, place the CSV contents in the `csv` field of a local JSON payload. Keep customer-sensitive exports local and do not commit them to the repository.

## Operator-guided mapping sequence

1. Inspect the export with `csv_inspect`.
2. Propose a status/stage column using header names and sample values.
3. Propose a deal-value column using header names and numeric evidence.
4. Propose a source column if the operator wants source reporting; otherwise use an empty source column.
5. Ask the operator to confirm the proposed mappings and the exact won/lost status labels. Never silently assume that a value such as `Complete`, `Successful`, or `Inactive` means won or lost.
6. Only after confirmation, call `sales_ops_report` with the original CSV and confirmed mapping.
7. Present the deterministic report and disclose any missing fields or limitations.

The inspector provides evidence, not an automatic mapping decision. Column mapping remains an operator-confirmed step. The report tool remains read-only and calculates metrics independently of any model-generated explanation.

## Permission boundary

In agent mode, inspection requires the explicit `tool.csv.inspect` permission. Report generation separately requires `tool.sales_ops.report`. Both remain denied by default unless explicitly allow-listed.
