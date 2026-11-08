# AgentForge Workspace API

The workspace API exposes the existing deterministic Go tools over HTTP. It does not call a model, persist uploaded CSVs, write to a CRM, send messages, or trigger external actions.

## Run locally

```bash
go test ./...
AGENTFORGE_ALLOWED_ORIGINS=http://localhost:5500,http://127.0.0.1:5500 AGENTFORGE_API_TOKEN=replace-with-a-long-random-secret go run ./cmd/agentforge-server
```

The service listens on `:8080` by default. Set `PORT` for platforms that inject a port, or `AGENTFORGE_ADDR` to override the listen address. Configure `AGENTFORGE_ALLOWED_ORIGINS` as a comma-separated exact-origin allow-list for browser clients.

## Endpoints

- `GET /healthz` — service health.
- `POST /api/inspect` — execute the Go `csv_inspect` tool.
- `POST /api/report` — execute the Go `sales_ops_report` tool with operator-confirmed mappings.

Both POST endpoints accept and return JSON. When `AGENTFORGE_API_TOKEN` is configured, they require `Authorization: Bearer <token>`. Set a long random token as a deployment secret; never commit it or expose it in public source. If unset, endpoints are unauthenticated for local development only. CSV input remains subject to the tool's 1 MiB, 10,000-row, and 100-column bounds. Requests are limited to 2 MiB and unknown JSON fields are rejected.

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

CORS is disabled for browser origins unless an exact origin is listed in `AGENTFORGE_ALLOWED_ORIGINS`. Keep `AGENTFORGE_API_TOKEN` enabled for any deployed service. The public GitHub Pages origin is `https://matteojose3112-bit.github.io`; add it only when the API is deployed and intentionally serving that site.

The API is read-only and has no credential or account layer. Do not send confidential customer exports or deploy it as a general-purpose public data-processing service without adding authentication, abuse controls, and an appropriate privacy review. Use synthetic or anonymized data for the portfolio demo.

## Workspace integration

The workspace includes an explicit API URL and token connection panel. When connected, inspection and report generation call these Go endpoints. Without a configured API URL, it remains in local demo mode.
