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
4. The result is deterministic JSON that can be inspected, tested, or passed to a model-backed workflow. The CRM-specific tool can also calculate funnel counts, win rate, and value totals when the stage and value columns are mapped explicitly.
5. In agent mode, the same capability is exposed only when the application grants `tool.csv.analyze`.

## Scope and safety

- The dataset is synthetic and contains no customer information.
- The tool is read-only; it does not modify a CRM or source file.
- This local demonstration does not call a model API.
- Do not use real customer or other sensitive records in a public portfolio demo.
- AgentForge permissions are application-level controls, not operating-system sandboxing.

## Current limitations

These are small, deterministic data-profiling and pipeline-summary examples. They are not a complete CRM analytics product, do not persist runs, and do not forecast future revenue or infer stage probabilities. A live model-backed tool cycle remains a separate validation step requiring a locally configured API key and a model ID supported by that API account.


## Business-specific pipeline metrics

The same sample export can be passed to the CRM-specific analysis tool. Its input explicitly maps the CRM stage and estimated-value columns and declares which stages count as won or lost:

```bash
go run ./cmd/agentforge -tool lead_pipeline_analyze -input-file examples/lead_pipeline_analysis.json
```

Expected key metrics:

```json
{
  "total_leads": 8,
  "won_leads": 1,
  "lost_leads": 1,
  "open_leads": 6,
  "closed_leads": 2,
  "win_rate_percent": 50,
  "missing_statuses": 0,
  "missing_values": 1,
  "missing_sources": 0,
  "total_estimated_value": 63700,
  "open_pipeline_value": 44500,
  "won_value": 4200,
  "lost_value": 15000
}
```

The win rate is calculated as won leads divided by closed leads (won + lost), not won leads divided by all leads. Open pipeline value includes only records with a non-empty status that is not mapped as won or lost. Values with missing statuses contribute to total estimated value but are not assigned to an outcome bucket.

The tool defaults to common status labels (`won`, `closed won`, `lost`, and `closed lost`) and accepts explicit `won_statuses` and `lost_statuses` arrays for a CRM's own terminology. Status matching is case-insensitive. If no source column is supplied, source breakdown is empty.

This adds a real operational calculation to the generic CSV profile: an operator can inspect both data completeness and a basic funnel snapshot before a weekly review. It remains a deterministic summary, not a forecast or a substitute for validating the CRM's stage definitions and financial data.
