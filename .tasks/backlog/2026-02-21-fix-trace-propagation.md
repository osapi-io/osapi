---
title: Fix trace context propagation from API server to worker via NATS
status: backlog
created: 2026-02-21
updated: 2026-02-21
---

## Objective

Trace IDs differ between the API server and worker for the same job,
meaning OpenTelemetry trace context is not propagating through NATS
JetStream message headers. The worker creates a new root span instead
of continuing the API server's trace.

**Observed**: API server logs `trace_id=0b3e...` for job `64564a8e`,
worker logs `trace_id=7e68...` for the same job.

**Expected**: Both should share the same trace ID.

## Analysis

The code path looks correct on paper:

1. `otelecho.Middleware` creates span → context has trace_id
2. Handler passes `ctx` through to `publishAndWait` →
   `natsClient.Publish(ctx, ...)`
3. nats-client `Publish` injects via
   `otel.GetTextMapPropagator().Inject(ctx, HeaderCarrier(msg.Header))`
4. JetStream `PublishMsg` sends message with headers
5. Worker receives message via `consumer.Fetch` → `msg.Headers()`
6. Worker calls `ExtractTraceContextFromHeader(context.Background(),
   http.Header(msg.Header))`

OTel versions match (`v1.40.0`) in both main module and nats-client.
Propagator is set before server starts. JetStream supports headers
since NATS 2.2.

The extraction silently falls back to `context.Background()` when no
valid trace context is found, causing the worker to create a new root
span.

## Debugging Steps

1. Add temporary debug logging in nats-client `Publish` to dump
   `msg.Header` after `Inject` — verify `Traceparent` key is present
2. Add temporary debug logging in worker `handleJobMessage` to dump
   `msg.Header` before `ExtractTraceContextFromHeader` — verify
   `traceparent` key arrives
3. This narrows the break to: injection failure, transit loss, or
   extraction failure

## Possible Causes

- JetStream header preservation issue (unlikely but check NATS server
  version)
- `nats.Header` vs `http.Header` type conversion subtlety in the
  `Inject` call
- The `Inject` call is a no-op because the span context in `ctx` is
  somehow invalid despite appearing in logs
- Header key casing lost in transit (extraction normalizes, but verify)

## Fix Options

1. **Debug and fix the header propagation** — add logging, find the
   break, fix it
2. **Inject trace context into KV job data** — as a fallback, store
   `traceparent` in the job data written to KV (which the worker
   already reads), not just in NATS message headers. This is more
   reliable since KV data is definitely preserved.

## Files Involved

| File | Role |
|------|------|
| `nats-client/pkg/client/jetstream.go:137-165` | Publish with header injection |
| `internal/job/worker/handler.go:96` | Trace extraction from NATS headers |
| `internal/telemetry/propagation.go:105-115` | Header normalization + extraction |
| `internal/job/worker/consumer.go:226-246` | JetStream → nats.Msg adapter |
| `internal/job/client/client.go:137` | publishAndWait Publish call |

## Notes

- The refactor in #164 (NATS config ownership) did not change any
  trace propagation code — this bug likely predates that PR.
- Architecture docs claim trace propagation works; update docs if this
  turns out to be a known limitation.
