---
title: Evaluate synadia-io/orbit.go for NATS integration
status: backlog
created: 2026-02-17
updated: 2026-02-17
---

## Objective

Evaluate whether https://github.com/synadia-io/orbit.go would be a good
fit for OSAPI's NATS JetStream integration, potentially replacing or
complementing the current nats-client sibling repo.

## Tasks

- [ ] Review orbit.go capabilities (KV, object store, services, etc.)
- [ ] Compare with current nats-client abstraction layer
- [ ] Identify overlap with existing OSAPI patterns (KV-first
      architecture, stream notifications, consumer management)
- [ ] Assess maturity, maintenance status, and API stability
- [ ] Determine if adoption would simplify or complicate the codebase
- [ ] Write recommendation: adopt, partial adopt, or skip

## Notes

- OSAPI currently uses a custom `nats-client` sibling repo linked via
  `go.mod replace` for NATS abstraction
- The job system relies heavily on KV watches, stream consumers, and
  subject-based routing
- Any replacement needs to support the append-only status event
  architecture and broadcast response collection
