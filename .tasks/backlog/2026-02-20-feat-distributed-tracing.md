---
title: Distributed tracing with OpenTelemetry
status: backlog
created: 2026-02-20
updated: 2026-02-20
---

## Objective

Add end-to-end distributed tracing using OpenTelemetry (OTel) so that
a single trace ID follows a request from CLI or API origin, through
NATS messaging, to worker execution and back. After this work,
filtering logs by trace ID should show the full flow of any request
across all components.

## Prerequisite

Job IDs are already surfaced in API responses and CLI output (done).
This task adds trace context propagation so that the trace ID
correlates all log lines and spans across components automatically.

## Approach

Use the OpenTelemetry Go SDK (CNCF standard):

- **Trace ID as correlation ID** — W3C `traceparent` header carries a
  globally unique trace ID across HTTP and NATS boundaries
- **Automatic context propagation** — no custom headers needed
- **Span-based instrumentation** — each operation (HTTP request, KV
  read, job processing) becomes a span with timing and attributes
- **slog integration** — `otelslog` bridge attaches `trace_id` and
  `span_id` to all structured log lines automatically
- **Echo middleware** — `otelecho` instruments all HTTP handlers with
  zero per-handler code

### Key Go Packages

- `go.opentelemetry.io/otel` — core API
- `go.opentelemetry.io/otel/sdk/trace` — trace SDK
- `go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho`
  — Echo middleware
- `go.opentelemetry.io/otel/exporters/stdout/stdouttrace` — stdout
  exporter for debug mode
- `go.opentelemetry.io/otel/exporters/otlp/otlptrace` — OTLP exporter
  for production backends (Jaeger, Tempo, etc.)
- `go.opentelemetry.io/otel/bridge/otelslog` — slog bridge

## Scope

### Tracer Provider Setup

- Initialize OTel tracer provider at startup in API server, worker,
  and CLI
- Use stdout exporter when `--debug` is enabled
- Use OTLP exporter when configured (optional, for production)
- Use noop tracer when tracing is not enabled (zero overhead)

### HTTP Layer (Echo API Server)

- Add `otelecho` middleware to instrument all HTTP handlers
- Trace context propagates automatically via W3C `traceparent` header

### HTTP Client (CLI → API)

- Inject OTel propagator into HTTP client so outgoing requests carry
  `traceparent`
- CLI creates root span, API server continues the trace

### NATS Propagation

- Inject trace context into NATS message headers when publishing jobs
- Extract trace context from NATS message headers in the worker
- This links API-side spans to worker-side spans under the same trace

### Worker Spans

- Create spans for job processing, provider execution, result writes
- Attach job ID, operation type, and target as span attributes

### slog Integration

- Use `otelslog` bridge so that `trace_id` and `span_id` appear in
  all structured log lines automatically
- Enables `grep trace_id=<hex>` to see full end-to-end flow in logs

### Configuration

Add optional config section to `osapi.yaml`:

```yaml
telemetry:
  tracing:
    # Enable tracing (default: false, --debug also enables stdout)
    enabled: false
    # OTLP endpoint for production tracing backends
    # otlp_endpoint: "localhost:4317"
```

### Components to Update

- `cmd/api_server_start.go` — init tracer provider, add `otelecho`
- `cmd/job_worker_start.go` — init tracer provider for worker
- `cmd/` (CLI commands) — init tracer provider, create root spans
- `internal/client/` — inject OTel propagator into HTTP client
- `internal/job/client/` — inject trace context into NATS headers
- `internal/job/worker/` — extract trace context from NATS headers,
  create processing spans
- `internal/config/` — add `telemetry.tracing` config section

## Notes

- NATS does not have an official OTel instrumentation library, so
  trace context injection/extraction will be manual (set/read
  `traceparent` header on messages)
- Keep backwards compatible — missing trace context should not cause
  errors, just start a new trace
- The stdout exporter is sufficient for `--debug` — no Jaeger needed
  during development

## Outcome

_To be filled in when done._
