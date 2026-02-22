---
title: Rename health detailed to health status and enrich with system metrics
status: done
created: 2026-02-19
updated: 2026-02-19
---

## Objective

Rename `/health/detailed` to `/health/status` and make it the single operator
debugging view for the entire OSAPI system. Currently it only reports "nats: ok,
kv: ok" — it should show everything an operator needs to understand the state of
the system.

## Outcome

Implemented the rename and enriched the endpoint with system metrics:

- **Rename**: `/health/detailed` → `/health/status` across OpenAPI spec,
  handler, client, CLI, tests, and docs
- **MetricsProvider**: New interface with `ClosureMetricsProvider` for NATS
  info, stream stats, KV bucket stats, and job queue counts
- **Graceful degradation**: Each metric call is independent — failures are
  logged and skipped
- **CLI output**: Compact inline `printKV` header + tables for multi-row data
  (components, streams, KV buckets)
- **CLI UX overhaul**: Replaced `printStyledMap` with `printKV` across all
  client commands for consistent, column-aligned output
- **Docs**: Updated all 11 CLI doc pages and architecture docs

### Out of scope (deferred)

- Workers list (needs consumer enumeration or heartbeat tracking)
- Object Store stats (blob store not implemented yet)
- job-responses KV bucket stats (not wired in API server)
- Per-consumer lag metrics
