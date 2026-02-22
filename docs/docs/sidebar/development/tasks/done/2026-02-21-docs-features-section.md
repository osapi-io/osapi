---
title: Add Features documentation section
status: done
created: 2026-02-21
updated: 2026-02-22
---

## Objective

Add a "Features" section to the Docusaurus docs that documents what each managed
resource (System, Network, etc.) actually does. Currently the docs cover _how to
use_ the API and CLI but not _what each resource manages_ at a feature level. As
more providers are added (NTP, users, RPM, services), this becomes essential.

## Outcome

Implemented as a "Features" navbar dropdown (renamed from "Resources") with 8
feature pages under `docs/docs/sidebar/features/`:

- `system-management.md` -- hostname, disk, memory, load
- `network-management.md` -- DNS get/update, ping
- `job-system.md` -- KV-first architecture, routing, lifecycle (with mermaid
  state diagram and sequence diagram)
- `audit-logging.md` -- middleware audit trail, export, retention (with mermaid
  sequence diagram)
- `health-checks.md` -- liveness, readiness, status endpoints
- `authentication.md` -- JWT HS256, RBAC, roles, permissions (with mermaid
  permission resolution flowchart)
- `distributed-tracing.md` -- OpenTelemetry, trace propagation (with mermaid
  sequence diagram)
- `metrics.md` -- Prometheus endpoint

Each page follows a consistent structure: what it does, how it works, links to
CLI and API docs (domain-specific), configuration, permissions, and related
links. No `osapi` CLI command duplication -- feature pages describe concepts and
link to CLI docs for actual usage.

### Additional changes

- Reorganized sidebar: Development/ (contributing, testing, roadmap, tasks),
  Architecture/ (api-guidelines, principles, configuration moved to Usage/)
- Sidebar order: Home, Usage, Features, Architecture, Development
- Added Usage to navbar
- Renamed 16 DocCardList files to `.mdx` for `format:detect` compatibility
- Moved `.tasks/` into docs site under Development/Tasks
- Slimmed `architecture.md` by extracting feature content to feature pages
- Added quickstart to Home page
- Updated CLAUDE.md with new paths and feature doc instructions
