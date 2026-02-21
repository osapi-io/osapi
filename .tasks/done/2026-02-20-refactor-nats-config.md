---
title: Refactor NATS config so each consumer owns its connection settings
status: done
created: 2026-02-20
updated: 2026-02-21
---

## Objective

The current config puts the API server's NATS connection under `job.client`,
which is confusing — it looks like a job-specific setting when it's really
the API server's NATS connection. Each component that connects to NATS should
declare its own `nats.host` / `nats.port` in its own config section.

## Outcome

Completed in PR #164 (`refactor(config): Move NATS connection settings to
owning components`). NATS connection settings now live under
`api.server.nats` and `job.worker.nats` in their respective config sections.
The `job.client` section was removed. Config docs and environment variable
mapping were updated accordingly.
