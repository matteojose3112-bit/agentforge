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

## v0.4 — AI Sales Operations Assistant (next milestone)

**Goal:** turn the existing CRM pipeline analyzer into a credible, repeatable business workflow that produces useful operational insight—not simply add another generic tool.

### Phase 1 — Define the workflow and evidence
- [ ] Define the target user, input format, business questions, and expected report structure.
- [ ] Create representative synthetic CRM datasets covering healthy, weak, incomplete, and edge-case pipelines.
- [ ] Define deterministic acceptance criteria for every metric and recommendation.
- [ ] Document the boundary between calculated facts, model-generated interpretation, and suggested actions.

### Phase 2 — Build the assistant workflow
- [x] Compose the existing CRM lead-pipeline analyzer into a registered sales operations report tool.
- [x] Generate a concise report with pipeline snapshot metrics, data-quality signals, observations, and follow-up priorities.
- [x] Keep recommendations rule-based and tied to observed counts/fields; avoid unsupported causes or forecasts.
- [x] Keep the report read-only; no CRM writes, emails, or external side effects.
- [x] Register the capability behind its own default-deny permission (`tool.sales_ops.report`) and retain bounded agent execution.
- [ ] Add a guided prompt/workflow that helps an operator map their CRM columns and interpret the report.

### Phase 3 — Evaluate and demonstrate
- [ ] Add repeatable evaluation fixtures for multiple pipeline shapes, expected metrics, grounded recommendations, malformed input, and permission denial.
- [x] Add deterministic unit tests for report metrics, recommendation triggers, unavailable win-rate context, schema validity, and invalid input.
- [ ] Add tests for the full model-backed workflow using a fake/local provider so CI remains API-cost-free.
- [ ] Publish a reproducible demo with sample input, command, output, and operator interpretation.
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

Last reviewed: 2026-11-05

## Current direction

**Implemented baseline:** v0.3.0 local runtime and CRM analysis capabilities.

**Outstanding validation:** live provider-backed model/tool smoke test and release-candidate evidence.

**Active development target:** v0.4 AI Sales Operations Assistant, starting with workflow definition, synthetic evaluation datasets, and deterministic acceptance criteria. The immediate objective is a demonstrable, read-only business workflow built on the existing code—not premature persistence or infrastructure.
