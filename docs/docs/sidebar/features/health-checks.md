---
sidebar_position: 5
---

# Health Checks

OSAPI exposes health endpoints for load balancers, monitoring systems, and
operational tooling. These endpoints report whether the API server is alive,
ready to serve traffic, and the status of its dependencies.

## Endpoints

| Endpoint         | Auth Required | Description                                   |
| ---------------- | ------------- | --------------------------------------------- |
| `/health`        | No            | Liveness probe -- is the process alive?       |
| `/health/ready`  | No            | Readiness probe -- can it serve traffic?      |
| `/health/status` | Yes           | Detailed status with component health metrics |

### Liveness (`/health`)

Always returns `200 OK` if the process is running. Use this for container
orchestrators (e.g., Kubernetes liveness probes) to detect hung processes.

### Readiness (`/health/ready`)

Checks connectivity to NATS and KV stores. Returns `200 OK` when the API server
can process requests, or `503 Service Unavailable` when dependencies are down.
Use this for load balancer health checks to avoid routing traffic to an unready
instance.

### Status (`/health/status`)

Returns per-component health with system metrics (uptime, goroutine count,
memory usage). Requires authentication with the `health:read` permission. Use
this for dashboards and monitoring.

## Configuration

Health check endpoints (`/health` and `/health/ready`) are unauthenticated by
design -- they need to work before clients have tokens. The `/health/status`
endpoint requires a valid JWT with the `health:read` permission.

No specific configuration is needed for health checks beyond the standard server
and authentication settings. See [Configuration](../usage/configuration.md) for
the full reference, [CLI Reference](../usage/cli/client/health/health.mdx) for
usage and examples, or the [API Reference](/gen/api/health-check-api-health) for
the REST endpoints.

## Permissions

| Endpoint         | Permission    |
| ---------------- | ------------- |
| `/health`        | None          |
| `/health/ready`  | None          |
| `/health/status` | `health:read` |

All built-in roles (`admin`, `write`, `read`) include `health:read`.

## Related

- [CLI Reference](../usage/cli/client/health/health.mdx) -- health commands
- [API Reference](/gen/api/health-check-api-health) -- REST API documentation
- [Architecture](../architecture/architecture.md) -- system design overview
