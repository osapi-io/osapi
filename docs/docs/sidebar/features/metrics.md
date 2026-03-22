---
sidebar_position: 8
---

# Metrics

Each OSAPI component exposes a Prometheus-compatible `/metrics` endpoint on a
dedicated port. Metrics are collected using OpenTelemetry with an isolated
Prometheus registry per component, so each endpoint shows only that component's
data.

## Endpoints

Each component's metrics server exposes three endpoints:

| Endpoint        | Description                                       |
| --------------- | ------------------------------------------------- |
| `/metrics`      | Prometheus metrics in text exposition format      |
| `/health`       | Liveness probe — always returns `{"status":"ok"}` |
| `/health/ready` | Readiness probe — 200 when ready, 503 when not    |

The health probes are unauthenticated and always available when the metrics
server is enabled.

| Component  | Default Port | Config Key                 |
| ---------- | ------------ | -------------------------- |
| Controller | 9090         | `controller.metrics.port`  |
| Agent      | 9091         | `agent.metrics.port`       |
| NATS       | 9092         | `nats.server.metrics.port` |

## Application Metrics Reference

Go runtime (goroutines, memory, GC) and process metrics (CPU, memory, file
descriptors) are included on every component.

| Metric                        | Type      | Labels | Component | Description                        |
| ----------------------------- | --------- | ------ | --------- | ---------------------------------- |
| `osapi_component_up`          | gauge     |        | all       | 1 when ready, 0 when not           |
| `osapi_jobs_processed_total`  | counter   | status | agent     | Jobs completed or failed           |
| `osapi_jobs_active`           | gauge     |        | agent     | Currently executing jobs           |
| `osapi_job_duration_seconds`  | histogram |        | agent     | Job execution duration             |
| `osapi_heartbeat_age_seconds` | gauge     |        | agent     | Seconds since last heartbeat write |

The controller also exposes HTTP request metrics from the OTEL middleware using
standard `http.server.*` names (`http.server.request.duration`,
`http.server.active_requests`, etc.).

The NATS component exposes `osapi_component_up` only — NATS has its own native
monitoring.

## Health Probes

Each metrics server also serves lightweight health probes on the same port.
These are always unauthenticated.

### Liveness (`/health`)

Always returns `200 OK` with `{"status":"ok"}` when the metrics server is
running. Use this to detect hung or crashed processes.

### Readiness (`/health/ready`)

Returns `200 OK` when the component is ready, or `503 Service Unavailable` when
it is not. Use this to gate traffic until the component has fully started.

### Kubernetes Example

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 9090
  initialDelaySeconds: 5
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /health/ready
    port: 9090
  initialDelaySeconds: 5
  periodSeconds: 10
```

Adjust the port to match the component (`9090` for controller, `9091` for agent,
`9092` for NATS).

## Integration

Point your Prometheus instance at each component:

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'osapi-controller'
    static_configs:
      - targets: ['localhost:9090']
  - job_name: 'osapi-agent'
    static_configs:
      - targets: ['localhost:9091']
  - job_name: 'osapi-nats'
    static_configs:
      - targets: ['localhost:9092']
```

## Configuration

Each component's metrics server can be enabled or disabled independently:

```yaml
controller:
  metrics:
    enabled: true
    port: 9090

agent:
  metrics:
    enabled: true
    port: 9091

nats:
  server:
    metrics:
      enabled: true
      port: 9092
```

Set `enabled: false` to disable the metrics endpoint for a component. See the
[Configuration](../usage/configuration.md) reference for the full list of
settings and environment variable overrides.

## Related

- [Health Checks](health-checks.md) -- liveness, readiness, and status probes
- [Distributed Tracing](distributed-tracing.md) -- OpenTelemetry tracing
- [Configuration](../usage/configuration.md) -- full configuration reference
