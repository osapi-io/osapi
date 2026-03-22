---
sidebar_position: 8
---

# Metrics

Each OSAPI component exposes a Prometheus-compatible `/metrics` endpoint on a
dedicated port. Metrics are collected using OpenTelemetry with an isolated
Prometheus registry per component, so each endpoint shows only that component's
data.

## Endpoints

| Component  | Default Port | Config Key                 |
| ---------- | ------------ | -------------------------- |
| Controller | 9090         | `controller.metrics.port`  |
| Agent      | 9091         | `agent.metrics.port`       |
| NATS       | 9092         | `nats.server.metrics.port` |

Each endpoint returns metrics in the standard Prometheus text exposition format
at `/metrics`.

## What It Exposes

Each component's `/metrics` endpoint exposes:

- Go runtime (goroutines, memory, GC)
- Process metrics (CPU, memory, file descriptors)

The controller additionally exposes HTTP request metrics collected by the Echo
framework (counts, durations, response sizes by route).

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
