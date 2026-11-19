---
sidebar_position: 5
---

# Health checks

osapi has health endpoints for load balancers, monitoring systems and operations
tools. These endpoints report if the controller runs, if it is ready for
traffic, and the status of its dependencies. The three components (controller,
agent, NATS server) write to one shared registry. Thus an operator can see the
health of all of the deployment from one endpoint.

## Endpoints

| Endpoint         | Auth Required | Description                                   |
| ---------------- | ------------- | --------------------------------------------- |
| `/health`        | No            | Liveness probe: does the process run?         |
| `/health/ready`  | No            | Readiness probe: is it ready for traffic?     |
| `/health/status` | Yes           | Detailed status with component health metrics |

### Liveness (`/health`)

Always returns `200 OK` if the process runs. Use it in container orchestrators,
for example in a Kubernetes liveness probe, to find processes that do not
respond.

### Readiness (`/health/ready`)

Checks connectivity to NATS and KV stores. Returns `200 OK` when the controller
can process requests, or `503 Service Unavailable` when dependencies are down.
Use this for load balancer health checks to avoid routing traffic to an unready
instance.

### Status (`/health/status`)

Returns the health of each component, with system metrics (uptime, goroutine
count, memory use). It needs authentication with the `health:read` permission.
Use it for dashboards and monitoring. It answers 200 when a component is not
healthy, with the overall status `degraded`: the request worked. Use
`/health/ready` to route traffic.

## Component registry

The three components write a heartbeat to a shared registry KV bucket at a
regular interval. Each heartbeat writes a JSON record. The key is the component
type and the hostname, for example `agents.web-01`, `controller.api-server` or
`nats.nats-server`. The records have process metrics from the time of the
heartbeat:

| Metric      | Description                              |
| ----------- | ---------------------------------------- |
| CPU percent | Process CPU utilisation at sample time   |
| RSS bytes   | Resident set size (physical memory used) |
| Goroutines  | Number of active goroutines              |

The `/health/status` response has a `components` table that puts these heartbeat
records together. If the registry key of a component expired (the TTL ended
without a new heartbeat), osapi reports the component as unreachable.

Each row has the build version of the component's binary, so a mixed fleet shows
which hosts run an older build. Example registry output:

```
TYPE        HOSTNAME      STATUS  CONDITIONS  AGE     CPU    MEM     VERSION
agent       web-01        Ready   -           7h 6m   1.2%   96 MB   v0.2.1-0.20261119003517-f6bc5af55ceb
agent       web-02        Ready   -           3h 2m   0.8%   82 MB   v0.2.1-0.20261119003517-f6bc5af55ceb
controller  api-server    Ready   -           7h 6m   2.1%   128 MB  v0.2.1-0.20261119003517-f6bc5af55ceb
nats        nats-server   Ready   -           7h 6m   0.3%   64 MB   v0.2.1-0.20261119003517-f6bc5af55ceb
```

## Sub-Component health

Each component publishes the status of its internal services with its heartbeat.
The `/health/status` endpoint puts these parts together from the registry. Thus
an operator can see the health of each internal service on all hosts, also in a
deployment with many nodes.

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

## Metrics server health probes

In addition to the controller's REST health endpoints, each component's metrics
server also exposes lightweight probes on its own port. These are separate from
the REST API health checks and require no authentication.

| Endpoint        | Port (default)     | Description                             |
| --------------- | ------------------ | --------------------------------------- |
| `/health`       | 9090 / 9091 / 9092 | Liveness. Always returns 200            |
| `/health/ready` | 9090 / 9091 / 9092 | Readiness. 200 when ready, 503 when not |

Use these probes to monitor each component without the controller API. For
example, a Kubernetes pod spec or an external load balancer can probe the
metrics server of the agent or of NATS directly.

Refer to [Metrics](metrics.md) for probe details and a Kubernetes example.

## Condition notifications

When conditions fire, the notification system dispatches alerts. Refer to
[Notifications](notifications.md) for the notification backends, configuration,
and re-notification settings.

## Configuration

The health check endpoints (`/health` and `/health/ready`) need no
authentication. They must work before clients have tokens. The `/health/status`
endpoint needs a valid JWT with the `health:read` permission.

Health checks need only the standard server and authentication settings. Refer
to [Configuration](../usage/configuration.md) for the full reference,
[CLI Reference](../usage/cli/client/health/health.mdx) for usage and examples,
or the [API Reference](/gen/api/health-check-api-health) for the REST endpoints.

## Permissions

| Endpoint         | Permission    |
| ---------------- | ------------- |
| `/health`        | None          |
| `/health/ready`  | None          |
| `/health/status` | `health:read` |

All built-in roles (`admin`, `write`, `read`) include `health:read`.

## Related

- [CLI Reference](../usage/cli/client/health/health.mdx): health commands
- [API Reference](/gen/api/health-check-api-health): REST API documentation
- [Architecture](../architecture/architecture.md): system design overview
