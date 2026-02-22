---
title: Support audit log export to alternative backends
status: backlog
created: 2026-02-21
updated: 2026-02-21
---

## Objective

Add support for exporting audit logs to backends beyond NATS KV, starting
with file-based export. This enables long-term retention and integration
with external log aggregation systems.

## Context

Audit logs are currently stored only in a NATS KV bucket with a
configurable TTL (default 30 days). For compliance and operational needs,
organizations may want to export audit entries to durable storage that
outlives the KV TTL.

## Potential Backends

- **File** — append-only JSON lines file, rotated by size or date
- **S3** — object storage for cloud deployments
- **Syslog** — integration with existing log infrastructure

## Notes

- File export should be the first backend implemented
- The audit `Store` interface may need an `Export` or `Sink` abstraction
- Consider a fan-out pattern: write to KV for real-time queries, export
  to file/S3 for long-term retention
- Configuration would go under `nats.audit.export` or a new
  `audit.export` top-level section
