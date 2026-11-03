# Contributing

Thanks for taking an interest in AgentForge. The project prioritizes a small, understandable runtime with explicit side effects and practical value.

## Development setup

- Go 1.23 or newer
- Git

Run the checks before submitting a change:

```bash
go test ./...
go vet ./...
go build ./cmd/agentforge
```

## Design principles

1. Keep orchestration and policy in the Go runtime.
2. Keep provider-specific request/response handling behind `agent.Provider`.
3. Register tools explicitly; never let model output add capabilities.
4. Give every tool a narrow permission and input/output schema.
5. Validate inputs before execution and outputs before returning them to a model.
6. Treat side effects carefully; retries must be opt-in and safe to repeat.
7. Propagate context cancellation and bound loops, timeouts, and resource use.
8. Avoid adding databases, queues, services, or dependencies without a concrete requirement.

## Pull requests

- Explain the problem and the smallest useful solution.
- Add or update tests for behavior changes.
- Update README, architecture notes, roadmap, or changelog when behavior or user-facing workflows change.
- Never commit API keys, credentials, personal data, or generated binaries.
- Do not add tests that call paid external APIs; use local test servers and deterministic fixtures.

## Scope

AgentForge is an early development foundation. Contributions should not imply production readiness or complete sandboxing. Permission checks are application-level execution controls, not an operating-system security boundary.
