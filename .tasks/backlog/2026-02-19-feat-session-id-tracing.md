---
title: Distributed tracing with OpenTelemetry
status: backlog
created: 2026-02-19
updated: 2026-02-19
---

## Objective

Add end-to-end distributed tracing using OpenTelemetry (OTel) so that
a single trace ID follows a request from CLI or API origin, through
NATS messaging, to worker execution and back. Currently, Echo generates
`X-Request-ID` but it is never correlated with job IDs or propagated to
NATS/worker logs. After this work, filtering logs by trace ID should
show the full flow of any request across all components.

## Approach: OpenTelemetry

Use the OpenTelemetry Go SDK instead of a custom session ID scheme.
OTel is the CNCF standard for distributed tracing and gives us:

- **Trace ID as session ID** — the W3C `traceparent` header carries a
  globally unique trace ID that serves the same purpose as a custom
  session ID, but uses an industry standard
- **Automatic context propagation** — OTel propagates trace context
  through HTTP headers (W3C Trace Context) without custom headers
- **Span-based instrumentation** — each operation (HTTP request, KV
  read, job processing) becomes a span with timing, status, and
  attributes
- **slog integration** — `otelslog` bridge attaches trace/span IDs to
  structured log lines automatically
- **Echo middleware** — `otelecho` instruments all HTTP handlers with
  zero per-handler code
- **Exporter flexibility** — export to stdout (for `--debug`), Jaeger,
  Grafana Tempo, or OTLP-compatible backends

### Key Go Packages

- `go.opentelemetry.io/otel` — core API
- `go.opentelemetry.io/otel/sdk/trace` — trace SDK
- `go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho`
  — Echo middleware
- `go.opentelemetry.io/otel/exporters/stdout/stdouttrace` — stdout
  exporter for debug mode
- `go.opentelemetry.io/otel/exporters/otlp/otlptrace` — OTLP exporter
  for production backends (Jaeger, Tempo, etc.)
- `go.opentelemetry.io/otel/bridge/otelslog` — slog bridge for
  attaching trace IDs to log lines (Go 1.21+, `log/slog`)

## Scope

### Tracer Provider Setup

- Initialize OTel tracer provider at startup in API server, worker,
  and CLI
- Use stdout exporter when `--debug` is enabled (human-readable trace
  output alongside slog)
- Use OTLP exporter when configured (optional, for production)
- Use noop tracer when tracing is not enabled (zero overhead)

### HTTP Layer (Echo API Server)

- Add `otelecho` middleware to instrument all HTTP handlers
- Trace context propagates automatically via W3C `traceparent` header
- Replaces need for custom `X-Session-ID` header

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
  even without a tracing backend

### Components to Update

- `cmd/api_server_start.go` — init tracer provider, add `otelecho`
  middleware
- `cmd/job_worker_start.go` — init tracer provider for worker
- `cmd/` (CLI commands) — init tracer provider, create root spans
- `internal/client/` — inject OTel propagator into HTTP client
- `internal/job/client/` — inject trace context into NATS headers
- `internal/job/worker/` — extract trace context from NATS headers,
  create processing spans
- `internal/config/` — add optional `telemetry` config section

### Configuration

Add optional config section to `osapi.yaml`:

```yaml
telemetry:
  # Enable tracing (default: false, --debug also enables with stdout)
  enabled: false
  # OTLP endpoint for production tracing backends
  # otlp_endpoint: "localhost:4317"
```

## Prerequisite (in progress)

`.tasks/in-progress/2026-02-19-quick-wins-job-id-tracing.md` surfaces
job IDs in API responses and CLI output. This is purely about giving
users a way to run `osapi client job get --job-id <id>` without
hunting through logs. It does NOT add any slog/logging correlation —
that is what THIS task (OTel) solves properly.

## Key design principle

Do NOT sprinkle `slog.String("job_id", ...)` in handlers to simulate
tracing. OTel handles correlation automatically via `otelslog` bridge:
every log line in a traced context gets `trace_id` and `span_id`
attached without per-call changes. Manual slog additions are noise
that OTel makes unnecessary.

## Notes

- OTel adds dependencies but they are well-maintained CNCF projects
  with stable APIs (1.x)
- The stdout exporter is sufficient for `--debug` use cases — no need
  to run Jaeger locally during development
- Jaeger and Grafana Tempo both accept OTLP natively, so the OTLP
  exporter covers most production backends
- NATS does not have an official OTel instrumentation library, so
  trace context injection/extraction into NATS headers will be manual
  (straightforward — just set/read `traceparent` header on messages)
- Keep backwards compatible — missing trace context should not cause
  errors, just start a new trace
- This complements the consistent debug logging task — OTel adds
  trace correlation, debug logging adds the log lines themselves

## Outcome

_To be filled in when done._
