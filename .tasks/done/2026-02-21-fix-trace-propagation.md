---
title: Fix trace context propagation from API server to worker via NATS
status: done
created: 2026-02-21
updated: 2026-02-21
---

## Objective

Trace IDs differ between the API server and worker for the same job,
meaning OpenTelemetry trace context is not propagating through NATS
JetStream message headers. The worker creates a new root span instead
of continuing the API server's trace.

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

**Root cause**: The nats-client version pinned in go.mod
(`v0.0.0-20260215190839-25dd316864d1`) used `ExtJS.Publish()` which
does not pass headers. The old `Publish(ctx, subject, data)` method
had no way to attach headers to the message. The Feb 20 commit
`80144e1` in nats-client switched to `ExtJS.PublishMsg(ctx, msg)`
with OTel trace context injected into NATS message headers via
`otel.GetTextMapPropagator().Inject()`.

## Files Involved

| File | Role |
|------|------|
| `nats-client/pkg/client/jetstream.go:137-165` | Publish with header injection |
| `internal/job/worker/handler.go:96` | Trace extraction from NATS headers |
| `internal/telemetry/propagation.go:105-115` | Header normalization + extraction |
| `internal/job/worker/consumer.go:226-246` | JetStream → nats.Msg adapter |
| `internal/job/client/client.go:137` | publishAndWait Publish call |

## Outcome

**Fix**: Updated nats-client dependency to
`v0.0.0-20260221001231-80144e1c7d21` via `go get
github.com/osapi-io/nats-client@latest`. No code changes needed in
osapi — the trace injection/extraction code paths were already correct.

**Verification**: `go build ./...` compiles, `just go::unit` all tests
pass.
