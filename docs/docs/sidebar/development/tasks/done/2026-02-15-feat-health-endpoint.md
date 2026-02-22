---
title: Health check and readiness endpoints
status: done
created: 2026-02-15
updated: 2026-02-19
---

## Objective

Add standard health check endpoints for load balancers, orchestrators, and
monitoring systems. A production appliance needs to report its own health.

## API Endpoints

```
GET    /health               - Basic health check (200 = alive)
GET    /health/ready         - Readiness check (NATS connected, etc.)
GET    /health/detailed      - Detailed component health
```

## Implementation

- `/health` — lightweight, no auth required, returns 200 with `{"status": "ok"}`
- `/health/ready` — checks NATS connection, KV bucket access, worker
  availability
- `/health/detailed` — authenticated, returns status of each subsystem:
  - API server: up
  - NATS connection: connected/disconnected
  - KV bucket: accessible/inaccessible
  - Worker count: N active workers
  - Disk space: ok/warning/critical
  - Memory: ok/warning/critical

## Notes

- `/health` should NOT require auth (used by load balancers)
- `/health/ready` and `/health/detailed` may require auth
- Follow Kubernetes liveness/readiness probe conventions
- Consider custom thresholds for disk/memory warnings
- No scope needed for basic health; `health:read` for detailed
