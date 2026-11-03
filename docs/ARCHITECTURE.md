# Architecture

## Runtime shape

AgentForge is a single Go process with explicitly registered capabilities.

- **CLI** (`cmd/agentforge`): parses user input and starts a task.
- **Agent runner** (`internal/agent`): validates the requested capability, owns task lifecycle and applies reliability policy.
- **Tools** (`internal/tools`): concrete capabilities that implement the `Tool` contract.

The initial echo tool proves the execution path without requiring credentials, network access, or an LLM provider.

## Task lifecycle

Every runner invocation creates a task with a generated ID and UTC creation timestamp. Valid executions transition from `pending` to `running`, then to `succeeded`, `failed`, or `cancelled`. The returned result includes timestamps, elapsed duration, attempt count, output, and error text.

The runner emits task and attempt events through an optional callback and structured `log/slog` records. Events are currently in-process only; durable execution history requires a later storage implementation.

## Reliability

The caller's context is propagated to the tool. An optional timeout bounds the entire run, including retry delays. Tools must observe context cancellation; the runtime cannot forcibly terminate code that ignores its context.

Retries default to one total attempt. A tool must explicitly implement `RetryableTool` and return true to receive the configured retry allowance. The maximum is five total attempts, with optional capped exponential backoff. Tool authors are responsible for declaring retry safety; non-idempotent tools should not opt in.

## Design constraints

1. Keep the first releases as a single deployable Go binary.
2. Keep tool contracts small, typed, and testable.
3. Make side effects explicit; tools should not receive unrestricted ambient authority.
4. Do not retry side-effecting tools unless their operation is safe to repeat.
5. Add persistence, model providers, and approval gates only with clear requirements and tests.
6. Keep provider-specific code behind interfaces when model integration is introduced.

## Planned evolution

The next increment is a vendor-independent model provider interface, schema-validated tool contracts, and explicit permission boundaries. Durable task storage and human approval follow when real workflows require them. Avoid distributed infrastructure until a real workload justifies it.
