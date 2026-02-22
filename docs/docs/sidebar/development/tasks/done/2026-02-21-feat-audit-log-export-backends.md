---
title: Support audit log export to alternative backends
status: done
created: 2026-02-21
updated: 2026-02-21
---

## Objective

Add support for exporting audit logs to backends beyond NATS KV, starting with
file-based export. This enables long-term retention and integration with
external log aggregation systems.

## Context

Audit logs are currently stored only in a NATS KV bucket with a configurable TTL
(default 30 days). For compliance and operational needs, organizations may want
to export audit entries to durable storage that outlives the KV TTL.

## Potential Backends

- **File** — append-only JSON lines file, rotated by size or date
- **S3** — object storage for cloud deployments
- **Syslog** — integration with existing log infrastructure

## Outcome

Implemented file-based audit log export with a pluggable `Exporter` interface
that supports future backends (S3, syslog).

### New files

- `internal/audit/export/types.go` — Exporter interface, Fetcher type, Result
  struct
- `internal/audit/export/file.go` — FileExporter (JSONL format)
- `internal/audit/export/export.go` — Run orchestrator with pagination and
  progress callbacks
- `internal/audit/export/file_public_test.go` — FileExporter tests
- `internal/audit/export/export_public_test.go` — Orchestrator tests
- `cmd/client_audit_export.go` — CLI command
- `docs/docs/sidebar/usage/cli/client/audit/export.md` — Documentation

### CLI usage

```bash
osapi client audit export --output audit.jsonl [--type file] [--batch-size 100]
```
