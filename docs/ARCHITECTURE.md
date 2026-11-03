# Architecture

## Runtime shape

AgentForge is a single Go process with explicitly registered capabilities.

- **CLI** (`cmd/agentforge`): parses direct tool or model-backed agent requests.
- **Agent runtime** (`internal/agent`): owns provider-neutral conversation types, tool loop, schema validation, permission checks, and reliable execution.
- **Provider adapter** (`internal/providers/openai`): translates the provider-neutral contract to the OpenAI Responses API.
- **Tools** (`internal/tools`): concrete capabilities and their JSON Schema descriptions.

## Model provider boundary

The `agent.Provider` interface accepts a provider-neutral model request and returns text or tool calls. OpenAI-specific HTTP payloads and response parsing remain inside the OpenAI adapter. A second provider can implement the same interface without changing the agent loop.

The first adapter uses the Responses API over the standard library HTTP client. API credentials are passed through `OPENAI_API_KEY`, never source-controlled. Provider tests use a local HTTP test server; CI does not call a paid API. Requests set `store:false` and do not rely on server-side conversation persistence.

## Tool contracts and validation

A model tool implements the runtime tool contract and exposes a `ToolDefinition` containing a stable name, description, JSON Schema parameter contract, and permission string. Arguments are validated before the tool is executed, and tool outputs are validated before being sent back to the model.

The validator intentionally supports a bounded JSON Schema subset: object, string, number, integer, boolean, array, enum, required, properties, and `additionalProperties:false`. Unsupported schema features are not silently interpreted as supported.

## Permission boundaries

Every tool definition names a permission. The runtime asks a `PermissionPolicy` before executing a model-requested tool. The included allow-list policy grants only explicitly configured permission strings. A nil policy denies all tool calls. Model output cannot add tools or grant permissions.

This is an execution boundary, not a complete sandbox: an allowed tool still runs with the process's operating-system privileges. Tool implementations must minimize authority and validate their own business rules.

## Bounded orchestration

The model/tool loop is limited to a configured number of turns. Context cancellation propagates to provider requests and tools. The runtime rejects unknown tool names, invalid JSON arguments, and denied permissions.

## Reliability

The lower-level runner retains task IDs, lifecycle events, timeouts, cancellation, and bounded opt-in retries. Tool retries remain disabled by default and require an explicit tool opt-in.

## Design constraints

1. Keep the first releases as a single deployable Go binary.
2. Keep provider-specific code behind a vendor-neutral interface.
3. Make side effects explicit and permissioned.
4. Validate model-supplied arguments before tool execution.
5. Do not retry side-effecting tools unless repeating the operation is safe.
6. Add persistence, approval gates, and distributed infrastructure only when requirements justify them.
