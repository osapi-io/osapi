---
title: Metrics endpoint with OpenTelemetry
status: backlog
created: 2026-02-20
updated: 2026-02-20
---

## Objective

Expose a `/metrics` endpoint on the API server that serves Prometheus-
compatible metrics using OpenTelemetry's metrics SDK. This gives
operators visibility into request rates, latencies, job throughput,
error rates, and system health without requiring a separate metrics
agent.

## Approach

Use the OpenTelemetry Go metrics SDK with a Prometheus exporter:

- **OTel metrics API** — define counters, histograms, and gauges using
  the standard OTel metrics API
- **Prometheus exporter** — serves metrics in Prometheus exposition
  format at `/metrics`
- **Echo middleware** — automatically record HTTP request metrics
  (duration, status codes, method, path)

### Key Go Packages

- `go.opentelemetry.io/otel/metric` — metrics API
- `go.opentelemetry.io/otel/sdk/metric` — metrics SDK
- `go.opentelemetry.io/otel/exporters/prometheus` — Prometheus
  exporter
- `go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho`
  — Echo middleware (shared with tracing task)

## Scope

### Metrics to Expose

**HTTP metrics** (automatic via otelecho middleware):
- `http_server_request_duration_seconds` — request latency histogram
- `http_server_active_requests` — in-flight request gauge
- `http_server_request_total` — request counter by method/status

**Job metrics** (custom instrumentation):
- `osapi_jobs_created_total` — counter of jobs created
- `osapi_jobs_completed_total` — counter by status (completed/failed)
- `osapi_job_duration_seconds` — histogram of job processing time
- `osapi_jobs_active` — gauge of currently processing jobs

**Worker metrics**:
- `osapi_workers_connected` — gauge of connected workers
- `osapi_worker_jobs_processed_total` — counter per worker

### Endpoint

- `GET /metrics` — Prometheus exposition format, unauthenticated
  (standard for metrics scraping)
- Add to health domain or as standalone route on the Echo instance

### Configuration

Extend the `telemetry` config section:

```yaml
telemetry:
  metrics:
    # Enable metrics endpoint (default: false)
    enabled: false
    # Path for the metrics endpoint
    path: "/metrics"
```

### Components to Update

- `cmd/api_server_start.go` — init meter provider, register
  Prometheus handler
- `cmd/job_worker_start.go` — init meter provider for worker metrics
- `internal/api/server.go` — register `/metrics` route
- `internal/job/client/` — instrument job creation/completion
- `internal/job/worker/` — instrument job processing
- `internal/config/` — add `telemetry.metrics` config section

## Notes

- The Prometheus exporter is pull-based (scrape) which works with
  standard Prometheus/Grafana setups
- If the distributed tracing task is done first, the OTel SDK init
  can be shared between tracing and metrics
- `/metrics` should be unauthenticated (Prometheus convention) but
  could be on a separate port if security is a concern
- Keep it opt-in via config so it adds zero overhead when disabled

## Outcome

_To be filled in when done._
