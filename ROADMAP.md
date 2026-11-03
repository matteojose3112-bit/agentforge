# Roadmap

## v0.1 — Execution foundation
- [x] Establish the Go module and CLI entry point.
- [x] Define a typed tool contract and registry-backed runner.
- [x] Add a safe local example tool.
- [x] Add unit tests and Go CI.
- [x] Document architecture and development commands.

## v0.2 — Reliable task lifecycle
- [ ] Add structured task IDs, statuses, timestamps, and results.
- [ ] Add execution events and structured logging.
- [ ] Add context-aware timeouts and bounded retry policy.
- [ ] Add tests for cancellation and tool failures.

## v0.3 — Model and tool integration
- [ ] Define an LLM provider interface independent of any vendor.
- [ ] Add one provider implementation with configuration via environment variables.
- [ ] Add schema-validated tool input and output.
- [ ] Add explicit tool permission boundaries.

## v0.4 — Persistence and human control
- [ ] Define a storage interface for task state and memory.
- [ ] Add a human-approval checkpoint for sensitive actions.
- [ ] Evaluate PostgreSQL only when durable multi-run state is needed.

## Later / evidence-driven
- [ ] Add workflow composition and scheduling.
- [ ] Add evaluation fixtures and repeatable agent benchmarks.
- [ ] Add background workers only when asynchronous workloads justify them.

## Non-goals for the initial milestone
- No web dashboard.
- No multi-agent framework before single-agent execution is reliable.
- No mandatory database, queue, or hosted model dependency.
