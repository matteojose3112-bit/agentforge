# AgentForge Workspace API

The workspace API exposes the existing deterministic Go tools over HTTP. It does not call a model, persist uploaded CSVs, write to a CRM, send messages, or trigger external actions.

## Run locally

```bash
go test ./...
AGENTFORGE_ALLOWED_ORIGINS=http://localhost:5500,http://127.0.0.1:5500 go run ./cmd/agentforge-server
```

The service listens on `:8080` by default. Set `PORT` for platforms that inject a port, or `AGENTFORGE_ADDR` to override the listen address. Configure `AGENTFORGE_ALLOWED_ORIGINS` as a comma-separated exact-origin allow-list for browser clients.

## Endpoints

- `GET /healthz` — service health.
- `POST /api/inspect` — execute the Go `csv_inspect` tool.
- `POST /api/report` — execute the Go `sales_ops_report` tool with operator-confirmed mappings.

Both POST endpoints accept and return JSON. CSV input remains subject to the tool's 1 MiB, 10,000-row, and 100-column bounds. Requests are limited to 2 MiB and unknown JSON fields are rejected.

### Inspect

```json
{"csv":"Stage,Value\nClosed Won,100\nOpen,50\n"}
```

### Report

```json
{
  "csv": "Stage,Value,Source\nClosed Won,100,Referral\nOpen,50,Web\n",
  "status_column": "Stage",
  "value_column": "Value",
  "source_column": "Source",
  "won_statuses": ["Closed Won"],
  "lost_statuses": ["Closed Lost"]
}
```

## Browser and deployment notes

CORS is disabled for browser origins unless an exact origin is listed in `AGENTFORGE_ALLOWED_ORIGINS`. The public GitHub Pages origin is `https://matteojose3112-bit.github.io`; add it only when the API is deployed and intentionally serving that site.

The API is read-only and has no credential or account layer. Do not send confidential customer exports or deploy it as a general-purpose public data-processing service without adding authentication, abuse controls, and an appropriate privacy review. Use synthetic or anonymized data for the portfolio demo.

## Workspace integration

The current GitHub Pages workspace remains a local-only prototype until configured with a deployed API URL. The API is independently testable locally; the next frontend change will add an explicit API connection setting and call these endpoints instead of computing the inspection and report in browser JavaScript.
