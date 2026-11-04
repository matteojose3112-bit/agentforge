# Roadmap

## v0.1 — Execution foundation
- [x] Establish the Go module and CLI entry point.
- [x] Define a typed tool contract and registry-backed runner.
- [x] Add a safe local example tool.
- [x] Add unit tests and Go CI.
- [x] Document architecture and development commands.

## v0.2 — Reliable task lifecycle
- [x] Add structured task IDs, statuses, timestamps, and results.
- [x] Add execution events and structured logging.
- [x] Add context-aware timeouts and bounded, opt-in retry policy.
- [x] Add tests for failures, cancellation, timeout, and retry behavior.

## v0.3 — Model and tool integration
- [x] Define a vendor-independent model provider interface.
- [x] Add an OpenAI Responses API provider with environment-based credentials.
- [x] Add JSON Schema validation for tool inputs and outputs.
- [x] Add explicit default-deny tool permission boundaries.
- [x] Add bounded model/tool turns and provider adapter tests.
- [x] Publish a step-by-step local and model-backed demo guide.
- [x] Document contribution and development expectations.
- [ ] Run a live API smoke test after a local API key is configured.

## Release acceptance checklist
- [x] Local no-key path documented.
- [x] Model-backed happy path documented.
- [x] Default-deny permission behavior documented.
- [x] CI workflow runs tests, vet, and CLI build.
- [ ] Confirm the live model/tool cycle with a locally configured API key.
- [ ] Record the live smoke-test result and tag the first release candidate.

## v0.4 — Persistence and human control
- [ ] Define a storage interface for task state and memory.
- [ ] Add a human-approval checkpoint for sensitive actions.
- [ ] Evaluate PostgreSQL only when durable multi-run state is needed.

## Later / evidence-driven
- [ ] Add workflow composition and scheduling.
- [ ] Add evaluation fixtures and repeatable agent benchmarks.
- [ ] Add background workers only when asynchronous workloads justify them.

## Non-goals for the initial milestones
- No web dashboard.
- No multi-agent framework before single-agent execution is reliable.
- No mandatory database, queue, or hosted model dependency.

---

Last reviewed: 2026-11-03

## Current direction — 2026-11-03

**v0.3.0 foundation implemented · Early-stage, not production-ready.** First complete the local live-model smoke test with a configured API key and verify the default-deny boundary under the documented demo. Record the result before tagging a release candidate. Keep persistence, human approval, workflow composition, scheduling, and benchmarks evidence-driven; do not add a database, queue, dashboard, or multi-agent layer before the single-agent runtime has demonstrated a need.
