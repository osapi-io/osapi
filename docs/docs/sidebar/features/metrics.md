---
sidebar_position: 8
---

# Metrics

Each osapi component has a `/metrics` endpoint for Prometheus, on its own port.
osapi collects metrics with OpenTelemetry. Each component has its own Prometheus
registry. Thus each endpoint shows only the data of its component.

## Endpoints

Each component's metrics server exposes three endpoints:

| Endpoint        | Description                                      |
| --------------- | ------------------------------------------------ |
| `/metrics`      | Prometheus metrics in text exposition format     |
| `/health`       | Liveness probe. Always returns `{"status":"ok"}` |
| `/health/ready` | Readiness probe. 200 when ready, 503 when not    |

The health probes need no authentication. They are always available when the
metrics server is enabled.

<!-- osapi:config controller.metrics.port agent.metrics.port nats.server.metrics.port -->

| Key                        | Type | Default                              | Environment                      | Description                         |
| -------------------------- | ---- | ------------------------------------ | -------------------------------- | ----------------------------------- |
| `controller.metrics.port`  | int  | `9090`                               | `OSAPI_CONTROLLER_METRICS_PORT`  | Port the metrics server listens on. |
| `agent.metrics.port`       | int  | none; `9091` in `configs/osapi.yaml` | `OSAPI_AGENT_METRICS_PORT`       | Port the metrics server listens on. |
| `nats.server.metrics.port` | int  | none; `9092` in `configs/osapi.yaml` | `OSAPI_NATS_SERVER_METRICS_PORT` | Port the metrics server listens on. |

<!-- /osapi:config -->

## Application metrics reference

osapi records all application metrics through OpenTelemetry, and exports them
through the OTEL Prometheus exporter with the `osapi` namespace. The exporter
adds the `osapi_` prefix automatically. The output has the OTEL scope labels,
for example `otel_scope_name`, so that you can find the source of a metric.

Each component also has the Go runtime metrics (goroutines, memory, GC) and the
process metrics (CPU, memory, file descriptors).

| Metric                                         | Type      | Labels                                      | Component  | Description                        |
| ---------------------------------------------- | --------- | ------------------------------------------- | ---------- | ---------------------------------- |
| `osapi_component_up`                           | gauge     |                                             | all        | 1 when ready, 0 when not           |
| `osapi_subsystem_up`                           | gauge     | subsystem                                   | all        | 1 when subsystem is ok, 0 when not |
| `osapi_jobs_created_total`                     | counter   |                                             | controller | Jobs submitted via API             |
| `osapi_jobs_processed_total`                   | counter   | status                                      | agent      | Jobs completed or failed           |
| `osapi_jobs_active`                            | gauge     |                                             | agent      | Currently executing jobs           |
| `osapi_job_duration_seconds`                   | histogram |                                             | agent      | Job execution duration             |
| `osapi_heartbeat_age_seconds`                  | gauge     |                                             | agent      | Seconds since last heartbeat write |
| `osapi_nats_client_operation_duration_seconds` | histogram | db_operation_name, db_namespace, error_type | all        | The duration of each call to NATS  |

The `osapi_subsystem_up` gauge has a `subsystem` label that names its internal
service, for example `api`, `heartbeat`, `metrics`, `notifier`, `tracing` or
`facts`. A value of 1 means that the status of the subsystem is `ok`. A value of
0 means `disabled` or `error`.

The controller also has HTTP request metrics from the OTEL middleware, with the
standard `http.server.*` names, for example `http.server.request.duration` and
`http.server.active_requests`.

Each component records `osapi_nats_client_operation_duration_seconds` for each
call that it makes to NATS. `db_operation_name` is the call, for example
`kv.get`, `kv.put` or `nats.publish`. `db_namespace` is the bucket or the
stream. A failed call also has `error_type`. A key, an object name and a subject
are never labels, because they have no limit. A key that is not there is the
answer of a read, not a failure, so it has no `error_type`.

The NATS component also has `osapi_component_up` and `osapi_subsystem_up`. NATS
has its own monitoring.

## Health probes

Each metrics server also serves lightweight health probes on the same port.
These are always unauthenticated.

### Liveness (`/health`)

Always returns `200 OK` with `{"status":"ok"}` when the metrics server is
running. Use this to detect hung or crashed processes.

### Readiness (`/health/ready`)

Returns `200 OK` when the component is ready, or `503 Service Unavailable` when
it is not. Use this to gate traffic until the component has fully started.

### Kubernetes example

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

You can enable or disable the metrics server of each component separately:

<!-- osapi:config controller.metrics agent.metrics nats.server.metrics -->

| Key                           | Type   | Default                                 | Environment                         | Description                                                                         |
| ----------------------------- | ------ | --------------------------------------- | ----------------------------------- | ----------------------------------------------------------------------------------- |
| `controller.metrics.enabled`  | bool   | none; `true` in `configs/osapi.yaml`    | `OSAPI_CONTROLLER_METRICS_ENABLED`  | Enable the metrics endpoint (default: true).                                        |
| `controller.metrics.host`     | string | none; `0.0.0.0` in `configs/osapi.yaml` | `OSAPI_CONTROLLER_METRICS_HOST`     | Host the metrics server binds to.                                                   |
| `controller.metrics.port`     | int    | `9090`                                  | `OSAPI_CONTROLLER_METRICS_PORT`     | Port the metrics server listens on.                                                 |
| `controller.metrics.url`      | string | none                                    | `OSAPI_CONTROLLER_METRICS_URL`      | Where a person opens the metrics endpoint, shown by health beside the bind address. |
| `agent.metrics.enabled`       | bool   | none; `true` in `configs/osapi.yaml`    | `OSAPI_AGENT_METRICS_ENABLED`       | Enable the metrics endpoint (default: true).                                        |
| `agent.metrics.host`          | string | none; `0.0.0.0` in `configs/osapi.yaml` | `OSAPI_AGENT_METRICS_HOST`          | Host the metrics server binds to.                                                   |
| `agent.metrics.port`          | int    | none; `9091` in `configs/osapi.yaml`    | `OSAPI_AGENT_METRICS_PORT`          | Port the metrics server listens on.                                                 |
| `agent.metrics.url`           | string | none                                    | `OSAPI_AGENT_METRICS_URL`           | Where a person opens the metrics endpoint, shown by health beside the bind address. |
| `nats.server.metrics.enabled` | bool   | none; `true` in `configs/osapi.yaml`    | `OSAPI_NATS_SERVER_METRICS_ENABLED` | Enable the metrics endpoint (default: true).                                        |
| `nats.server.metrics.host`    | string | none; `0.0.0.0` in `configs/osapi.yaml` | `OSAPI_NATS_SERVER_METRICS_HOST`    | Host the metrics server binds to.                                                   |
| `nats.server.metrics.port`    | int    | none; `9092` in `configs/osapi.yaml`    | `OSAPI_NATS_SERVER_METRICS_PORT`    | Port the metrics server listens on.                                                 |
| `nats.server.metrics.url`     | string | none                                    | `OSAPI_NATS_SERVER_METRICS_URL`     | Where a person opens the metrics endpoint, shown by health beside the bind address. |

<!-- /osapi:config -->

Set `enabled: false` to disable the metrics endpoint for a component. See the
[Configuration](../usage/configuration.md) reference for the full list of
settings and environment variable overrides.

## Related

- [Health Checks](health-checks.md): liveness, readiness, and status probes
- [Distributed Tracing](distributed-tracing.md): OpenTelemetry tracing
- [Configuration](../usage/configuration.md): full configuration reference
