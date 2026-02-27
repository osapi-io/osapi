---
title: Enhance health status with workers and consumers
status: backlog
created: 2026-02-26
updated: 2026-02-26
---

## Objective

Enrich the `client health status` output with worker registry and consumer
information so operators get a complete operational overview from a single
command.

## Changes

### Show registered workers

Read the `worker-registry` KV bucket and display a Workers section:

```
  Workers: 3 registered
  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━┓
  ┃ HOSTNAME                  ┃ LABELS          ┃ REGISTERED ┃
  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━╋━━━━━━━━━━━━━━━━━╋━━━━━━━━━━━━┫
  ┃ web-01                    ┃ group=web.prod  ┃ 15s ago    ┃
  ┃ web-02                    ┃ group=web.prod  ┃ 8s ago     ┃
  ┃ db-01                     ┃ group=db        ┃ 2m ago     ┃
  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━┻━━━━━━━━━━━━━━━━━┻━━━━━━━━━━━━┛
```

The registry KV bucket uses a 30s TTL, so only live workers appear. The
`REGISTERED` column shows the age of `registered_at` from the
`WorkerRegistration` entry.

### Show registry bucket in KV Buckets table

The `worker-registry` bucket should appear alongside `job-queue` in the existing
KV Buckets table. Investigate why it currently does not — the health endpoint
may be filtering to specific bucket names rather than listing all buckets.

### Show consumers (stretch)

Add a Consumers section showing JetStream consumer details per stream. This
would help operators see which workers have active subscriptions and whether
consumers are lagging:

```
  Consumers (osapi-JOBS):
  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━┳━━━━━━━━━┓
  ┃ NAME                      ┃ PENDING  ┃ ACKED   ┃
  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━╋━━━━━━━━━━╋━━━━━━━━━┫
  ┃ query_any_web_01          ┃ 0        ┃ 142     ┃
  ┃ query_direct_web_01       ┃ 0        ┃ 38      ┃
  ┃ modify_any_web_01         ┃ 1        ┃ 15      ┃
  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━┻━━━━━━━━━━┻━━━━━━━━━┛
```

This is lower priority — the worker list covers the main operational need.
Consumer details are useful for debugging but may be noisy with many workers.

## Notes

- The health status API response schema will need new fields for workers and
  optionally consumers
- The SDK and CLI will need updates to render the new sections
- Consider adding a `--verbose` flag to show consumers only when requested
