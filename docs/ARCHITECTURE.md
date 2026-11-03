# Architecture

## Current foundation

AgentForge starts as a single Go process with explicit, typed capabilities.

- **CLI** (cmd/agentforge): parses user input and starts a task.
- **Agent runner** (internal/agent): validates the requested capability and executes it.
- **Tools** (internal/tools): concrete capabilities that implement the agent Tool contract.

The initial echo tool is intentionally simple. It proves the registration and execution path without requiring credentials, network access, or an LLM provider.

## Execution contract

A tool has a stable name and an Execute(context.Context, string) method. The runner rejects empty task input and unknown tools, propagates cancellation through the context, and wraps tool errors with the tool name.

## Design constraints

1. Keep the first release a single deployable Go binary.
2. Keep tool contracts small, typed, and testable.
3. Make side effects explicit; tools should not receive unrestricted ambient authority.
4. Add persistence, model providers, retries, and approval gates only with clear requirements and tests.
5. Keep provider-specific code behind interfaces when model integration is introduced.

## Planned evolution

The next architectural increments are structured task/result types, execution IDs and lifecycle events, configurable tool permissions, a provider interface, and durable state. Avoid introducing distributed infrastructure until a real workload requires it.
