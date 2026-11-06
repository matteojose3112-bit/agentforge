# Changelog

All notable changes to AgentForge are documented here.

## [Unreleased]

### Configuration and credential safety
- Added preflight validation for missing OpenAI credentials, untouched template placeholders, accidental `Bearer ` prefixes, and whitespace; errors explain how to correct the local configuration without exposing the secret.
- Added provider-constructor and CLI tests for credential validation cases. This validation does not claim to verify remote key validity or permissions.
- Added a placeholder-only `.env.example` for `OPENAI_API_KEY`; the real `.env` remains ignored by Git.
- Documented shell-specific ways to load the environment variable and clarified that AgentForge does not automatically parse dotenv files.

### CLI usability
- Replaced the minimal flag listing with structured CLI help covering direct and agent modes, usage examples, required environment configuration, mutually exclusive input options, and default-deny permissions.
- Improved individual flag descriptions and made `-h`/`-help` return successfully after printing help.
- Added a CLI help regression test and synchronized README, demo guide, and roadmap documentation.

### CLI and build metadata
- Expanded `-version` output to include the application version, Git commit, and UTC build timestamp.
- Added linker-variable defaults so local builds remain usable while release builds can embed build metadata.
- Documented Linux/macOS and PowerShell release-style build commands in the README.

### Documentation and project direction — 2026-11-05
- Reconciled the README with the implemented v0.3.0 repository baseline and explicitly separated automated/local verification from the still-outstanding live provider smoke test.
- Reworked the roadmap to define v0.4 as a focused, read-only AI Sales Operations Assistant built on the existing CRM pipeline analysis capabilities.
- Added phased acceptance criteria for synthetic evaluation datasets, grounded recommendations, end-to-end tests, and a reproducible portfolio demonstration.
- Clarified that persistence, human approval, scheduling, background workers, and UI work remain evidence-driven future scope rather than current capabilities.

### v0.4 initial implementation slice — 2026-11-05
- Added the read-only `sales_ops_report` tool, composing existing CRM pipeline metrics into a concise operations snapshot with observations, data-quality signals, and rule-based follow-up actions.
- Registered the report in direct CLI and model-backed agent modes behind the dedicated default-deny permission `tool.sales_ops.report`.
- Added deterministic tests for report schema, metrics, recommendation triggers, missing closed outcomes, and invalid input.
- Added a synthetic sales operations snapshot and documented its CLI usage, output scope, and limitations.

### Previously implemented, unreleased changes
- Published the AgentForge project landing page on GitHub Pages at https://matteojose3112-bit.github.io/agentforge/.
- Added the live project-site link near the top of the README and updated the roadmap to record the successful Pages deployment.
- Added a CRM lead-pipeline analysis tool for configurable won/lost stages, open and closed lead counts, win rate, estimated pipeline values, missing fields, and stage/source breakdowns.
- Registered lead-pipeline analysis for direct CLI and model-backed agent execution with the dedicated `tool.leads.analyze` permission.
- Added deterministic tests for stage classification, custom/default status mappings, missing data, numeric validation, and the synthetic CRM case study.
- Added JSON file input for direct CLI tool execution, avoiding shell-escaping long CSV payloads.
- Added a synthetic CRM lead-pipeline example and portfolio case study with reproducible expected output.
- Require an explicit provider-supported model ID in agent mode instead of shipping an unverified default model name.
- Added a read-only text analysis tool with word, Unicode character, line, and estimated reading-time metrics.
- Added a bounded CSV analysis tool reporting row/column counts, missing values, and min/max/mean for fully numeric columns.
- Registered CSV analysis in direct CLI and model-backed agent modes with the dedicated `tool.csv.analyze` permission.
- Added CSV parser, size, row-count, header, numeric-summary, and agent-loop tests; documented a no-key CSV demo.
- Registered text analysis in direct CLI and model-backed agent modes, with a dedicated permission string and a 256 KiB input limit.
- Added unit tests and a no-API-key demo command for the new capability.

## [0.3.0] — Model and tool integration
- Added vendor-neutral model, message, tool-call, and provider contracts.
- Added an OpenAI Responses API adapter using the standard library HTTP client.
- Added a bounded model/tool loop with strict tool registration.
- Added JSON Schema validation for tool inputs and outputs, plus default-deny permission allow-listing.
- Added a JSON-schema-described local echo tool and provider/agent tests.
- Added CLI agent mode with environment-based API key configuration.
- Added a hands-on demo guide and contribution guidelines.
- Expanded the README with architecture flow, quick start, safety boundaries, and current limitations.

## [0.2.0] — Reliable task execution
- Added task IDs, lifecycle statuses, UTC timestamps, structured results, duration, and attempt counts.
- Added lifecycle and attempt events with structured `log/slog` output and an optional event callback.
- Added context-aware execution timeouts and cancellation handling.
- Added bounded retries for tools that explicitly opt in as retryable; retries default to disabled.
- Expanded tests for task metadata, event sequences, failures, cancellation, timeouts, and retry safety.

## [0.1.0] — Initial foundation
- Created the Go module and command-line entry point.
- Added a typed tool contract and minimal execution runner.
- Added a local echo tool and runner unit tests.
- Added Makefile commands, GitHub Actions CI, architecture notes, and roadmap.

## [0.3.1] — Documentation synchronization — 2026-11-03
- Refreshed README and roadmap to match the implemented v0.3.0 Go runtime, OpenAI Responses adapter, schema validation, and default-deny permissions.
- Reaffirmed the project's early-stage status, explicit tool permissions, and non-production limitations.
- Clarified the live API smoke test as the immediate next validation step; persistence and human approval remain future work.
