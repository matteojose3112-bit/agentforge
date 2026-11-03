# Changelog

All notable changes to AgentForge are documented here.

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
