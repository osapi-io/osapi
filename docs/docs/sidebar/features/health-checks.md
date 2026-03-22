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
and hostname (e.g., `agents.web-01`, `api.api-server`, `nats.nats-server`). The
records include process metrics collected at heartbeat time:

| Metric      | Description                              |
| ----------- | ---------------------------------------- |
| CPU percent | Process CPU utilisation at sample time   |
| RSS bytes   | Resident set size (physical memory used) |
| Goroutines  | Number of active goroutines              |

The `/health/status` response includes a `components` table that aggregates
these heartbeat records. A component whose registry key has expired (TTL elapsed
without a fresh heartbeat) is reported as unreachable.

Example `components` output:

```
COMPONENT    HOSTNAME      STATUS    CPU    RSS       GOROUTINES
api          api-server    ready     0.4%   42.3 MiB  87
agent        web-01        ready     0.1%   18.1 MiB  23
agent        web-02        ready     0.2%   17.8 MiB  22
nats         nats-server   ready     0.3%   31.6 MiB  41
```

## Condition Notifications

The component registry enables an optional condition notification system. When
`controller.notifications.enabled` is true, a watcher monitors the registry KV
bucket for condition transitions on any component and dispatches events via the
configured notifier backend.

### Conditions Reference

| Condition               | Components       | Description                                               |
| ----------------------- | ---------------- | --------------------------------------------------------- |
| `MemoryPressure`        | agent            | Host memory usage exceeds threshold (default 90%)         |
| `HighLoad`              | agent            | Load average exceeds CPU count × multiplier (default 2.0) |
| `DiskPressure`          | agent            | Any disk usage exceeds threshold (default 90%)            |
| `ProcessMemoryPressure` | agent, api, nats | Process RSS exceeds threshold                             |
| `ProcessHighCPU`        | agent, api, nats | Process CPU usage exceeds threshold                       |
| `ComponentUnreachable`  | agent, api, nats | Heartbeat expired (TTL timeout)                           |

Host-level conditions (`MemoryPressure`, `HighLoad`, `DiskPressure`) are
evaluated on agents only. Process-level conditions (`ProcessMemoryPressure`,
`ProcessHighCPU`) are evaluated on all components. `ComponentUnreachable` is
emitted by the notification watcher when a heartbeat TTL expires — it does not
appear on the component's registration (the component is already gone).

Conditions fire when a threshold is crossed and resolve automatically when it
drops back below the threshold. Thresholds are configurable per component in
`osapi.yaml`.

The default notifier (`log`) writes condition events to the structured log.
Fired conditions log at WARN level, resolved conditions at INFO:

```
WRN condition fired   component=agent hostname=web-01 condition=MemoryPressure active=true
INF condition resolved component=agent hostname=web-01 condition=MemoryPressure active=false
WRN condition fired   component=agent hostname=web-02 condition=ComponentUnreachable active=true reason="heartbeat expired"
```

See [Configuration](../usage/configuration.md) for how to enable notifications
and select a notifier.

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
