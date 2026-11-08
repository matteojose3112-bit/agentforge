# AgentForge Roadmap

AgentForge is developed in small, verifiable milestones. Checkboxes describe implemented repository behavior or explicitly verified live behavior; a deployed endpoint is not automatically a production-readiness claim.

## v0.1 — Execution foundation (complete)
- [x] Establish the Go module and CLI entry point.
- [x] Define a typed tool contract and registry-backed runner.
- [x] Add a safe local example tool.
- [x] Add unit tests, Go CI, and development documentation.

## v0.2 — Reliable task lifecycle (complete)
- [x] Add structured task IDs, statuses, timestamps, and results.
- [x] Add execution events and structured logging.
- [x] Add context-aware timeouts and bounded, opt-in retry policy.
- [x] Test failures, cancellation, timeout, and retry behavior.

## v0.3 — Model and tool runtime (implemented; live provider validation outstanding)
- [x] Define a provider-independent model interface and OpenAI Responses API adapter.
- [x] Implement bounded model/tool turns, tool registration, schema validation, and default-deny permissions.
- [x] Add direct CLI mode, JSON input-file support, structured help, and version/build metadata.
- [x] Add local echo, text analysis, bounded CSV analysis, and CRM lead-pipeline analysis tools.
- [x] Add deterministic tests and fake-provider agent-loop coverage for registered tools.
- [x] Document credential setup, safety boundaries, architecture, and local/model-backed demos.
- [x] Publish the GitHub Pages project site.
- [ ] Run and record a real provider-backed model/tool smoke test using an explicitly configured supported model.

## v0.4 — Read-only Sales Operations workflow

**Status: Implemented baseline; live report and model validation remain.**

### Workflow and evidence
- [x] Define the business workflow, input shape, report structure, and read-only boundary.
- [x] Add synthetic healthy, weak, incomplete, and zero-record CRM evaluation fixtures.
- [x] Document expected metrics, grounded recommendation rules, malformed-input behavior, and acceptance criteria.
- [x] Add fixture-backed regression tests for all four scenarios.

### Runtime and operator controls
- [x] Add the deterministic CRM lead-pipeline analyzer with configurable won/lost stage mappings.
- [x] Add the read-only Sales Operations report with calculated pipeline metrics, data-quality signals, observations, and rule-based follow-up priorities.
- [x] Register tools behind separate default-deny permissions.
- [x] Add bounded CSV inspection with header/missing-value counts and limited sample values.
- [x] Enforce a new-user-turn checkpoint between inspection and report generation; require operator-confirmed mappings.
- [x] Expose inspection and reporting through bounded, read-only HTTP API endpoints with strict JSON decoding, configurable CORS, and bearer-token protection.
- [x] Connect the guided browser workspace to configurable API endpoints while retaining local demo mode.
- [x] Deploy the Go API to Render Free and the workspace to GitHub Pages.
- [x] Verify the live `/healthz` endpoint returns HTTP 200.
- [x] Verify unauthenticated `/api/inspect` and `/api/report` requests return HTTP 401.
- [x] Verify authenticated sample CSV inspection from the deployed GitHub Pages workspace against the live Go API.
- [x] Verify Go CI and GitHub Pages deployment workflows pass on 2026-11-08 (commit `b951b78`).

### Remaining validation
- [ ] Verify authenticated `/api/report` with representative CSV data through the deployed workspace.
- [ ] Run a live provider-backed model/tool cycle and record provider/model/date/outcome without exposing credentials or sensitive output.
- [ ] Add fake-provider end-to-end coverage for the complete Sales Operations report workflow if the current deterministic tool tests do not cover the intended orchestration path.
- [ ] Improve deterministic mapping suggestions with explicit confidence levels only if operator testing demonstrates a need.
- [ ] Capture genuine screenshots or a short recording of the live workflow for the portfolio.

## Later — evidence-driven, not current commitments
- [ ] Add durable conversation/task state only if continuity is a demonstrated requirement.
- [ ] Add workflow scheduling or background workers only when a concrete asynchronous use case justifies them.
- [ ] Consider additional model providers behind the existing provider interface.
- [ ] Design explicit human approval and action permissions before introducing any external side effects.

## Release and quality principles
- Keep business calculations deterministic and testable; distinguish calculated facts from model interpretation and suggested actions.
- Keep synthetic/demo data clearly labeled.
- Keep tools read-only; no CRM writes, outbound messages, or other external actions are implemented.
- Do not describe AgentForge as production-ready or fully autonomous.
- Do not claim a live provider-backed model/tool cycle until it has been exercised and recorded.
- CI runs tests, vet, and builds without requiring paid model API calls.

## Current direction — 2026-11-08

**v0.4 read-only Sales Operations workflow implemented; guided workspace connected to the deployed API.** Authenticated CSV inspection has been verified end to end from the deployed browser workspace. Authenticated report generation and a real provider-backed model/tool cycle are the next validation gates. The current API has no persistent conversation history, database-backed task state, scheduling, or background workers.

The latest Go CI and GitHub Pages deployment workflows both passed on 2026-11-08 (commit `b951b78`).
