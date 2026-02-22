---
sidebar_position: 8
---

# Metrics

OSAPI exposes a Prometheus-compatible metrics endpoint for monitoring and
alerting. Metrics are collected automatically by the API server and available
for scraping without additional configuration.

## Endpoint

The metrics endpoint is available at `/metrics` on the API server. It returns
metrics in the standard Prometheus text exposition format. See
[CLI Reference](../usage/cli/client/metrics/metrics.mdx) for usage, or the
[API Reference](/category/api) for the REST endpoints.

## What It Exposes

The `/metrics` endpoint exposes standard Go runtime metrics and HTTP request
metrics collected by the Echo framework, including:

- Go runtime (goroutines, memory, GC)
- HTTP request counts, durations, and response sizes by route
- NATS connection metrics

## Integration

Point your Prometheus instance at the OSAPI server:

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'osapi'
    static_configs:
      - targets: ['localhost:8080']
```

## Configuration

The metrics endpoint is always enabled and requires no configuration. It is
unauthenticated by design so that Prometheus can scrape it without a token.

## Related

- [CLI Reference](../usage/cli/client/metrics/metrics.mdx) -- metrics CLI
  command
- [API Reference](/category/api) -- REST API documentation
- [Health Checks](health-checks.md) -- liveness and readiness probes
