---
title: HTTP metrics with Prometheus endpoint
status: done
created: 2026-02-20
updated: 2026-02-20
---

## Objective

Expose an opt-in `/metrics` endpoint on the API server that serves
Prometheus-compatible HTTP metrics using OpenTelemetry's metrics SDK.
The `otelecho` middleware (already wired) automatically records
`http.server.request.duration` and `http.server.active_requests` when
a global `MeterProvider` is set.

## Scope

**HTTP metrics only** (automatic via otelecho middleware):
- `http_server_request_duration_seconds` — request latency histogram
- `http_server_active_requests` — in-flight request gauge

Custom job/worker metrics are a separate follow-up task — see
`.tasks/backlog/2026-02-20-feat-custom-job-metrics.md`.

## Approach

1. Add `go.opentelemetry.io/otel/sdk/metric` and
   `go.opentelemetry.io/otel/exporters/prometheus` as direct deps
2. Add `MetricsConfig` to `internal/config/types.go`
3. Create `internal/telemetry/metrics.go` with `InitMeter()` that
   creates a Prometheus exporter and sets it as global MeterProvider
4. Wire into `cmd/api_server_start.go` and register `/metrics` route
   on Echo via `WithMetricsHandler` option
5. Auto-enable in debug mode, default path `/metrics`

## Notes

- `/metrics` is unauthenticated (Prometheus convention)
- Registered directly on Echo, not through OpenAPI handler pipeline
- Debug mode auto-enables metrics (same pattern as tracing)

## Outcome

Implemented Prometheus `/metrics` endpoint and added missing integration tests.

### Prometheus metrics
- Added `MetricsConfig` to `internal/config/types.go`
- Created `internal/telemetry/metrics.go` with `InitMeter()` — creates
  Prometheus exporter, sets global MeterProvider, returns promhttp handler
- Created `internal/telemetry/metrics_test.go` — covers default path,
  custom path, and exporter error path
- Created `internal/api/handler_metrics.go` — `GetMetricsHandler()` registers
  unauthenticated `/metrics` route directly on Echo
- Wired into `cmd/api_server_start.go` with shutdown in cleanup closure
- Added `TestGetMetricsHandler` to `internal/api/handler_public_test.go`

### Missing integration tests
- `internal/api/health/health_get_integration_test.go`
- `internal/api/health/health_ready_get_integration_test.go`
- `internal/api/health/health_status_get_integration_test.go`
- `internal/api/job/job_status_integration_test.go`
- `internal/api/job/job_workers_get_integration_test.go`
