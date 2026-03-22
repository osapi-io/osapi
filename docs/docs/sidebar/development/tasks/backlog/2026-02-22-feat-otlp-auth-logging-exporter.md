---
title: Improve tracing exporter options (OTLP auth, logger exporter)
status: backlog
created: 2026-02-22
updated: 2026-02-22
---

## Objective

Improve the tracing exporter configuration with two changes:

1. **OTLP authentication** -- the current OTLP exporter uses
   `otlptracegrpc.WithInsecure()` with no auth. This only works for
   local/internal backends (e.g., Jaeger on localhost). Cloud backends like
   Grafana Cloud, Honeycomb, and Datadog require TLS and/or API key headers. Add
   config options for TLS and auth headers.

2. **Logger-based exporter** -- the `stdout` exporter dumps raw OpenTelemetry
   span JSON directly to stdout, bypassing the structured logger. This is noisy
   and doesn't go wherever logs are routed. Add a `log` exporter option that
   writes span data through the structured `slog` logger so spans follow the
   same output path as all other log lines.

## Context

Current code in `internal/telemetry/telemetry.go`:

```go
case "otlp":
    exp, err := otlptraceNewFn(
        ctx,
        otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
        otlptracegrpc.WithInsecure(), // no TLS, no auth
    )
```

The `stdout` exporter uses `stdouttrace.New(stdouttrace.WithPrettyPrint())`
which writes directly to `os.Stdout`.

## Proposed Config

```yaml
telemetry:
  tracing:
    enabled: true
    exporter: otlp # "stdout", "log", "otlp", or unset
    otlp_endpoint: localhost:4317
    otlp_insecure: true # default false (use TLS)
    otlp_headers: # optional auth headers
      Authorization: 'Bearer <token>'
```

## Notes

- The `log` exporter could use the `stdouttrace` exporter writing to an
  `io.Writer` backed by the slog handler, or implement a custom `SpanExporter`
  that logs span summaries as structured log entries.
- For OTLP TLS: replace `WithInsecure()` with
  `WithTLSCredentials(credentials.NewTLS(...))` when `otlp_insecure` is false.
- For OTLP headers: use `otlptracegrpc.WithHeaders(cfg.OTLPHeaders)`.
- Update feature docs and configuration reference when implemented.
