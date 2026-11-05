# Portfolio demonstration: CRM lead pipeline snapshot

## Business scenario

A small service business exports its leads from a CRM before a weekly operations review. The operations coordinator needs a quick, repeatable check of the file structure, missing data, and numeric fields before building a report.

This example uses a fictional dataset. It contains eight leads across three sources, several pipeline stages, estimated values, response-time measurements, and assigned owners.

## Run it locally

Requirements: Go 1.23 or newer. No API key, external service, or paid model usage is required.

From the repository root:

```bash
go run ./cmd/agentforge -tool csv_analyze -input-file examples/lead_pipeline.json
```

The input file is JSON containing the CSV payload. The CLI passes it through the same registered tool runner used by the other direct-mode tools.

## Expected result

```json
{
  "rows": 8,
  "columns": 6,
  "column_names": [
    "lead_id",
    "source",
    "stage",
    "estimated_value",
    "days_to_first_response",
    "owner"
  ],
  "missing_values": 2,
  "numeric_columns": [
    {
      "column": "estimated_value",
      "non_empty_values": 7,
      "min": 4200,
      "max": 15000,
      "mean": 9100
    },
    {
      "column": "days_to_first_response",
      "non_empty_values": 7,
      "min": 1,
      "max": 5,
      "mean": 2.5714285714285716
    }
  ]
}
```

## What the result tells an operator

- The export contains **8 lead records** and **6 columns**.
- There are **2 missing cells** to investigate before relying on the full dataset.
- Seven leads have a numeric estimated value. Their mean is **9,100**, with a range from **4,200 to 15,000**.
- Seven leads have a numeric first-response measurement. Their mean is approximately **2.57 days**, with a range from **1 to 5 days**.

These are descriptive data-quality and numeric summaries, not CRM performance conclusions. The tool does not infer conversion rates, pipeline probability, revenue, or service-level compliance from column names.

## What this demonstrates technically

1. A user can provide a JSON input file instead of manually escaping a CSV string in a shell command.
2. The CLI resolves a registered capability through the typed tool runner.
3. The CSV tool enforces input, row, and column limits; validates UTF-8, headers, and row shape; and parses quoted CSV fields.
4. The result is deterministic JSON that can be inspected, tested, or passed to a model-backed workflow.
5. In agent mode, the same capability is exposed only when the application grants `tool.csv.analyze`.

## Scope and safety

- The dataset is synthetic and contains no customer information.
- The tool is read-only; it does not modify a CRM or source file.
- This local demonstration does not call a model API.
- Do not use real customer or other sensitive records in a public portfolio demo.
- AgentForge permissions are application-level controls, not operating-system sandboxing.

## Current limitations

This is a small, deterministic CSV profiling example. It is not a complete CRM analytics product, does not persist runs, and does not calculate business-specific funnel metrics. A live model-backed tool cycle remains a separate validation step requiring a locally configured API key and a model ID supported by that API account.
