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
- [ ] Run a live API smoke test after a local API key is configured.

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
