---
sidebar_position: 5
---

# Health Checks

OSAPI exposes health endpoints for load balancers, monitoring systems, and
operational tooling. These endpoints report whether the controller is alive,
ready to serve traffic, and the status of its dependencies. All three runtime
components (controller, agent, NATS server) participate in a shared component
registry so operators can see the health of the entire deployment from a single
endpoint.

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

Checks connectivity to NATS and KV stores. Returns `200 OK` when the controller
can process requests, or `503 Service Unavailable` when dependencies are down.
Use this for load balancer health checks to avoid routing traffic to an unready
instance.

### Status (`/health/status`)

Returns per-component health with system metrics (uptime, goroutine count,
memory usage). Requires authentication with the `health:read` permission. Use
this for dashboards and monitoring.

## Component Registry

All three runtime components heartbeat into a shared registry KV bucket on a
regular interval. Each heartbeat writes a JSON record keyed by component type
and hostname (e.g., `agents.web-01`, `controller.api-server`,
`nats.nats-server`). The records include process metrics collected at heartbeat
time:

| Metric      | Description                              |
| ----------- | ---------------------------------------- |
| CPU percent | Process CPU utilisation at sample time   |
| RSS bytes   | Resident set size (physical memory used) |
| Goroutines  | Number of active goroutines              |

The `/health/status` response includes a `components` table that aggregates
these heartbeat records. A component whose registry key has expired (TTL elapsed
without a fresh heartbeat) is reported as unreachable.

Example registry output:

```
TYPE        HOSTNAME      STATUS  CONDITIONS  AGE     CPU    MEM
agent       web-01        Ready   -           7h 6m   1.2%   96 MB
agent       web-02        Ready   -           3h 2m   0.8%   82 MB
controller  api-server    Ready   -           7h 6m   2.1%   128 MB
nats        nats-server   Ready   -           7h 6m   0.3%   64 MB
```

## Sub-Component Health

Each component publishes the status of its internal services alongside its
heartbeat registration. The `/health/status` endpoint aggregates these
sub-components from the registry so operators can see the health of every
internal service across all hosts — even in multi-node deployments.

Sub-components use a `{type}.{name}` naming convention. Each component reports
only its own sub-components:

| Component  | Sub-Components                                                                                              |
| ---------- | ----------------------------------------------------------------------------------------------------------- |
| controller | `controller.api`, `controller.heartbeat`, `controller.metrics`, `controller.notifier`, `controller.tracing` |
| agent      | `agent.heartbeat`, `agent.metrics`                                                                          |
| nats       | `nats.server`, `nats.heartbeat`, `nats.metrics`                                                             |

Sub-components report a status (`ok`, `disabled`, or `error`) and an optional
network address. The controller also performs live connectivity checks against
NATS and KV, which appear as `controller.nats (connectivity)` and
`controller.kv (connectivity)` in the response.

## Metrics Server Health Probes

In addition to the controller's REST health endpoints, each component's metrics
server also exposes lightweight probes on its own port. These are separate from
the REST API health checks and require no authentication.

| Endpoint        | Port (default)     | Description                              |
| --------------- | ------------------ | ---------------------------------------- |
| `/health`       | 9090 / 9091 / 9092 | Liveness — always returns 200            |
| `/health/ready` | 9090 / 9091 / 9092 | Readiness — 200 when ready, 503 when not |

These probes are useful for monitoring individual components independently of
the controller API — for example, probing the agent or NATS metrics server
directly from a Kubernetes pod spec or an external load balancer.

See [Metrics](metrics.md) for probe details and a Kubernetes example.

## Condition Notifications

When conditions fire, the notification system dispatches alerts. See
[Notifications](notifications.md) for the notification backends, configuration,
and re-notification settings.

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
