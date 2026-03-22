---
title: Add custom OTel metrics for jobs, agents, and NATS
status: backlog
created: 2026-02-20
updated: 2026-02-20
---

## Objective

The `/metrics` endpoint and Prometheus exporter are in place (see
`.tasks/done/2026-02-20-feat-metrics-endpoint.md`), and `otelecho` already
provides HTTP request metrics automatically. This task adds custom
application-level metrics using the OTel metrics API so operators get visibility
into job throughput, agent health, and NATS connectivity.

## Metrics to Add

### Job metrics (instrument in `internal/job/client/` and agent)

- `osapi_jobs_created_total` — counter of jobs created
- `osapi_jobs_completed_total` — counter by status (completed/failed)
- `osapi_job_duration_seconds` — histogram of job processing time
- `osapi_jobs_active` — gauge of currently processing jobs

### Agent metrics (instrument in `internal/agent/`)

- `osapi_agents_connected` — gauge of connected agents
- `osapi_agent_jobs_processed_total` — counter per agent

### NATS metrics

- `osapi_nats_connected` — gauge (1/0) for connection status
- `osapi_nats_reconnects_total` — counter of reconnect events

## Approach

Use the OTel metrics API (`go.opentelemetry.io/otel/metric`) to define counters,
histograms, and gauges. The global `MeterProvider` is already set by
`InitMeter()` in `internal/telemetry/metrics.go`, so instruments created via
`otel.Meter("osapi")` will automatically be scraped by the existing Prometheus
exporter at `/metrics`.

### Key packages

- `go.opentelemetry.io/otel/metric` — create instruments
- `go.opentelemetry.io/otel` — get global meter

### Components to update

- `internal/agent/processor.go` — record job duration, active jobs, completion
  status
- `internal/agent/consumer.go` — record agent connection status
- `internal/job/client/` — record job creation counter
- `cmd/node_agent_start.go` — init agent-side meter provider

## Notes

- Agent runs as a separate process — it needs its own `InitMeter()` call and
  `/metrics` endpoint (or push-based exporter)
- All metric names should use `osapi_` prefix to avoid collisions
- Use `metric.WithDescription()` and `metric.WithUnit()` for each instrument so
  Prometheus exposition includes HELP and UNIT lines
- Consider whether agent metrics should be exposed via a separate HTTP port or
  pushed to an OTel collector

## Outcome

_To be filled in when done._
