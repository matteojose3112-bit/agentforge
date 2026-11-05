# Changelog

All notable changes to AgentForge are documented here.

## [Unreleased]
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

### Documentation
- Refreshed README and roadmap to match the implemented v0.3.0 Go runtime, OpenAI Responses adapter, schema validation, and default-deny permissions.
- Reaffirmed the project's early-stage status, explicit tool permissions, and non-production limitations.
- Clarified the live API smoke test as the immediate next validation step; persistence and human approval remain future work.
