# Roadmap

AgentForge is being developed in small, verifiable milestones. Checkboxes describe repository implementation status, not aspirations or unverified live behavior.

## v0.1 — Execution foundation (complete)
- [x] Establish the Go module and CLI entry point.
- [x] Define a typed tool contract and registry-backed runner.
- [x] Add a safe local example tool.
- [x] Add unit tests and Go CI.
- [x] Document architecture and development commands.

## v0.2 — Reliable task lifecycle (complete)
- [x] Add structured task IDs, statuses, timestamps, and results.
- [x] Add execution events and structured logging.
- [x] Add context-aware timeouts and bounded, opt-in retry policy.
- [x] Add tests for failures, cancellation, timeout, and retry behavior.

## v0.3 — Model and tool integration (implemented; live validation outstanding)
- [x] Improve CLI help with mode descriptions, flag guidance, safety defaults, and runnable examples.
- [x] Add a version/build-metadata check to the quick-start commands.
- [x] Clarify direct-mode input formats: structured tools accept JSON payloads; `echo` accepts plain text.
- [x] Add a safe `.env.example` template and document how to load credentials without committing the real `.env` file.
- [x] Validate OpenAI credential configuration before agent startup and report actionable local setup errors.
- [x] Add automated coverage for help output and successful help exit behavior.
- [x] Define a vendor-independent model provider interface.
- [x] Add an OpenAI Responses API provider with environment-based credentials.
- [x] Add JSON Schema validation for tool inputs and outputs.
- [x] Add explicit default-deny tool permission boundaries.
- [x] Add bounded model/tool turns and provider adapter tests.
- [x] Publish a step-by-step local and model-backed demo guide.
- [x] Document contribution and development expectations.
- [x] Add deterministic, read-only text analysis with no model/API dependency.
- [x] Add bounded CSV analysis for row/column counts, missing values, and numeric summaries.
- [x] Add JSON input-file support and a reproducible synthetic CRM lead-pipeline case study.
- [x] Add deterministic CRM pipeline metrics with configurable won/lost stages, closed-lead win rate, pipeline values, and stage/source breakdowns.
- [x] Publish the static GitHub Pages project site from the repository's `/docs` directory.
- [ ] Run and record a live provider-backed model/tool smoke test using a locally configured API key.

### v0.3 validation / release gate
- [x] Local no-key execution paths and commands documented.
- [x] Model-backed happy path documented.
- [x] Default-deny permission behavior documented and covered by tests.
- [x] CI workflow runs tests, vet, and CLI build without paid model calls.
- [x] Public GitHub Pages site deployed.
- [ ] Confirm a real model response that requests a registered tool, passes validation and permission checks, executes, and returns a final response.
- [ ] Record provider/model/date and outcome of the live smoke test; do not publish credentials or sensitive output.
- [ ] Tag a release candidate only after the live validation gate is complete.

## v0.4 — AI Sales Operations Assistant (implemented baseline; live report validation outstanding)

**Goal:** turn the existing CRM pipeline analyzer into a credible, repeatable business workflow that produces useful operational insight—not simply add another generic tool.

### Phase 1 — Define the workflow and evidence
- [x] Define the target user, input format, business questions, and expected report structure.
- [x] Create representative synthetic CRM datasets covering healthy, weak, incomplete, and edge-case pipelines.
- [x] Define deterministic acceptance criteria for every metric and recommendation.
- [x] Document the boundary between calculated facts, model-generated interpretation, and suggested actions.

### Phase 2 — Build the assistant workflow
- [x] Compose the existing CRM lead-pipeline analyzer into a registered sales operations report tool.
- [x] Generate a concise report with pipeline snapshot metrics, data-quality signals, observations, and follow-up priorities.
- [x] Keep recommendations rule-based and tied to observed counts/fields; avoid unsupported causes or forecasts.
- [x] Keep the report read-only; no CRM writes, emails, or external side effects.
- [x] Register the capability behind its own default-deny permission (`tool.sales_ops.report`) and retain bounded agent execution.
- [x] Add bounded CSV inspection that returns headers, missing/non-empty counts, and up to five distinct sample values per column.
- [x] Document the operator-confirmed mapping sequence and separate inspection/report permissions.
- [x] Enforce a new user-turn checkpoint between CSV inspection and sales report generation, with tests for same-response blocking.
- [x] Document that the operator must confirm mappings and re-supply the CSV in the next CLI run.
- [ ] Improve mapping suggestions with deterministic header/value heuristics and explicit confidence levels.

### Phase 3 — Evaluate and demonstrate
- [x] Add synthetic baseline fixtures for healthy, weak, incomplete, and zero-record pipelines, with expected metric summaries documented in `docs/SALES_OPS_EVALUATION.md`.
- [ ] Add repeatable evaluation fixtures for multiple pipeline shapes, expected metrics, grounded recommendations, malformed input, and permission denial.
- [x] Add deterministic unit tests for report metrics, recommendation triggers, unavailable win-rate context, schema validity, and invalid input.
- [x] Add fixture-backed regression tests that execute all four synthetic scenarios and assert pipeline/report metrics plus selected recommendation triggers.
- [ ] Add tests for the full model-backed workflow using a fake/local provider so CI remains API-cost-free.
- [ ] Publish a reproducible CLI demo with sample input, command, output, and operator interpretation.\n- [x] Add a browser-based guided workspace prototype with local CSV inspection, an explicit mapping approval checkpoint, and a deterministic read-only report using synthetic data; clearly label it as a front-end demo, not a live Go runtime integration.
- [x] Expose the existing Go CSV inspection and sales report tools through bounded, read-only HTTP endpoints with strict JSON decoding, configurable CORS origins, and optional bearer-token protection.
- [x] Connect the browser workspace to configurable Go API endpoints while retaining an explicit local-only demo mode.
- [x] Deploy the API on Render Free with a generated bearer-token secret and GitHub Pages CORS allow-list.
- [x] Verify the live `/healthz` endpoint returns HTTP 200 and unauthenticated `/api/inspect` and `/api/report` requests are rejected with HTTP 401.
- [x] Verify authenticated `/api/inspect` requests with the sample CRM CSV through the deployed GitHub Pages workspace and live Render Go API.
- [ ] Verify authenticated `/api/report` requests with representative CSV payloads through the deployed workspace.
- [ ] Capture genuine screenshots or a short recording of the working demo for the portfolio.
- [ ] Run the live provider smoke test when an API key and a suitable low-cost model are available.

### Phase 4 — Human control and state (only after the workflow proves useful)
- [ ] Design an approval interface/checkpoint before introducing any action with external side effects.
- [ ] Define a storage interface only when conversation or task continuity is a demonstrated requirement.
- [ ] Evaluate PostgreSQL only when durable multi-run state is needed.

## Later — evidence-driven
- [ ] Add workflow composition and scheduling when a concrete use case requires them.
- [ ] Add background workers only when asynchronous workloads justify them.
- [ ] Consider additional model providers behind the existing provider interface.
- [ ] Consider a user interface only after the CLI workflow is reliable and useful.

## Release and quality principles
- Do not describe the project as production-ready or fully autonomous.
- Do not claim a live model/tool cycle has passed until its result is recorded.
- Keep business calculations deterministic and testable; use the model for explanation and grounded recommendations.
- Keep synthetic/demo data clearly labeled; never imply it is real client data.
- Keep tools read-only until a human approval design and explicit action permissions exist.
- Avoid adding a database, queue, dashboard, or multi-agent architecture without evidence that the workflow needs them.

## Non-goals for the current milestone
- No web dashboard.
- No multi-agent framework.
- No autonomous CRM updates, outbound messages, or other external actions.
- No mandatory database, queue, or hosted model dependency.

---

Last reviewed: 2026-11-08

## Current direction

**Implemented baseline:** v0.3.0 local runtime and CRM analysis capabilities.

**Outstanding validation:** authenticated live API data-processing requests, live provider-backed model/tool smoke test, and release-candidate evidence.

**Recently completed usability polish:** the CLI now has structured help output (`-h`/`-help`), clearer flag descriptions, safety notes, direct/agent-mode examples, explicit guidance that structured tools expect JSON while `echo` accepts plain text, and a quick-start `-version` command that displays version/build metadata. A placeholder-only `.env.example` documents the OpenAI credential variable; the CLI still reads credentials from the process environment and does not auto-load dotenv files. Agent startup now rejects missing keys, untouched placeholders, `Bearer ` prefixes, and whitespace before attempting a provider request; remote key validity remains an API-side check.

**Active development target:** v0.4 AI Sales Operations Assistant. The HTTP API is deployed on Render Free, tracks `main`, and the workspace is connected to its configurable API boundary. Live checks confirm the service health endpoint responds successfully, both data-processing endpoints enforce bearer authentication, and authenticated sample CSV inspection succeeds from the browser workspace. Authenticated report generation remains the next live validation step. The workflow has synthetic evaluation datasets, deterministic acceptance criteria, a runtime-enforced new-turn checkpoint, and a guided workspace—not a generic dashboard. Next priorities are verifying authenticated report generation, adding full model-loop tests with a fake provider, and publishing a reproducible CLI demo.
