# Health Probes and Application Metrics Design

## Goal

Add `/health` (liveness) and `/health/ready` (readiness) probes to the
per-component metrics server, and register application-specific Prometheus
metrics for each component. Bridge health into metrics with an
`osapi_component_up` gauge.

## Background

Each OSAPI component (controller, agent, NATS server) runs a per-component
metrics server (`internal/telemetry/metrics/server.go`) with an isolated
Prometheus registry and OTEL MeterProvider. Today the metrics server only
exposes Go runtime and process collector metrics at `/metrics`. There are no
health probes on the metrics server — the controller has complex
`/health`, `/health/ready`, `/health/status` endpoints on its API server
(with auth, OpenAPI types, metrics aggregation), but agent and NATS have
nothing.

Operators need:

- Liveness and readiness probes on every component for container orchestrators
- Application metrics (request counts, job throughput, latency) for dashboards
- A Prometheus gauge that reflects component readiness for alerting

## Design

### Health probes on the metrics server

Add `/health` and `/health/ready` routes to the existing metrics server HTTP
mux. Implementation lives in a new file `internal/telemetry/metrics/health.go`.

**`/health` (liveness):**

Always returns `{"status":"ok"}` with HTTP 200 if the process is running.
No dependencies, no checks. Used by container orchestrators to detect hung
processes.

**`/health/ready` (readiness):**

Calls an injected readiness function. Returns `{"status":"ready"}` with
HTTP 200 when the component can do its job, or
`{"status":"not_ready","error":"..."}` with HTTP 503 when it cannot.

Readiness semantics per component:

| Component  | Ready when                                |
| ---------- | ----------------------------------------- |
| Controller | NATS connected, KV accessible             |
| Agent      | NATS connected, consumers started         |
| NATS       | JetStream available                       |

Both routes are registered unconditionally in `New()`. The readiness
function is injected post-construction via
`SetReadinessFunc(fn func() error)`, following the same pattern as
`agent.SetSubComponents()`. If no readiness function is set, `/health/ready`
returns 503 with `"readiness check not configured"`. Each component wires its
own check in `cmd/`.

Both health endpoints return `Content-Type: application/json`.

The readiness function is called on every `/health/ready` request and on
every Prometheus scrape (via `GaugeFunc`). Implementations should be fast
and non-blocking. If a readiness check needs to contact an external service
(e.g., NATS ping), it should use a short timeout (5 seconds) to avoid
blocking the HTTP response.

The controller's existing API health endpoints (`/health`, `/health/ready`,
`/health/status`) remain unchanged — they serve a different purpose (API
clients, auth-gated status). The metrics server health probes are for
infrastructure tooling (Kubernetes, load balancers) and don't require auth.

### `osapi_component_up` gauge

A Prometheus gauge registered on each component's metrics server registry.
Value is `1` when ready, `0` when not. Evaluated lazily on each Prometheus
scrape using a `prometheus.GaugeFunc` that calls the same readiness function
as `/health/ready`. `GaugeFunc` is used because OTEL does not support
lazy-evaluated gauges — this is the one metric that uses the native
Prometheus client directly rather than OTEL.

Each component's gauge is on its own isolated registry (different port), so
there is no label collision. When federating into a single Prometheus
instance, operators distinguish components using the `job` or `instance`
label from their scrape config.

This bridges health into metrics — operators can alert on
`osapi_component_up == 0` without polling the health endpoint separately.

### Application metrics

Each component registers application-specific metrics via the metrics
server's `MeterProvider()` or `Registry()` at startup.

**Controller metrics:**

| Metric                               | Type      | Labels               | Description           |
| ------------------------------------ | --------- | -------------------- | --------------------- |
| `osapi_api_requests_total`           | counter   | method, path, status | HTTP request count    |
| `osapi_api_request_duration_seconds` | histogram | method, path         | Request latency       |
| `osapi_jobs_created_total`           | counter   |                      | Jobs submitted        |
| `osapi_component_up`                 | gauge     |                      | 1 = ready, 0 = not   |

**Agent metrics:**

| Metric                          | Type      | Labels | Description              |
| ------------------------------- | --------- | ------ | ------------------------ |
| `osapi_jobs_processed_total`    | counter   | status | Jobs completed/failed    |
| `osapi_jobs_active`             | gauge     |        | Currently executing jobs |
| `osapi_job_duration_seconds`    | histogram |        | Job execution time       |
| `osapi_heartbeat_age_seconds`   | gauge     |        | Time since last write    |
| `osapi_component_up`            | gauge     |        | 1 = ready, 0 = not      |

**NATS server metrics:**

| Metric              | Type  | Description         |
| ------------------- | ----- | ------------------- |
| `osapi_component_up`| gauge | 1 = ready, 0 = not |

The embedded NATS server exposes its own monitoring metrics natively.
`osapi_component_up` is the only custom metric needed.

### Metrics registration approach

Controller HTTP metrics (`osapi_api_requests_total`,
`osapi_api_request_duration_seconds`) are collected via the `otelecho`
middleware on the Echo server, using the metrics server's `MeterProvider()`.
This keeps the API server unaware of Prometheus — it instruments via OTEL,
and the Prometheus exporter on the metrics server handles the translation.
The `path` label uses Echo's route template (e.g., `/api/v1/node/:hostname`)
not the literal request path, to avoid unbounded cardinality.

Agent job metrics (`osapi_jobs_processed_total`, `osapi_jobs_active`,
`osapi_job_duration_seconds`) are instrumented in the agent's handler/
processor layer using the metrics server's `MeterProvider()`. The
`MeterProvider` must be passed to the agent (or set post-construction)
so the agent package doesn't import the metrics package directly.

`osapi_heartbeat_age_seconds` is a `GaugeFunc` registered on the
Prometheus registry. The agent exposes a `LastHeartbeatTime() time.Time`
method, and the `cmd/` wiring layer creates a closure over it for the
`GaugeFunc`.

### File structure

```
internal/telemetry/metrics/
  server.go          # Existing — add SetReadinessFunc, wire health routes
  health.go          # New — /health and /health/ready handlers
  types.go           # Existing — add ReadinessFunc field
  server_test.go     # Update for new functionality
  server_public_test.go  # Update for new functionality
  health_test.go     # New — health handler tests
  health_public_test.go  # New — health handler public tests
```

### Configuration

No config changes. Health probes are always available when the metrics
server is enabled. There's no reason to want metrics without health or
vice versa.

### Documentation

Update `docs/docs/sidebar/features/metrics.md`:

- Add "Health Probes" section documenting `/health` and `/health/ready`
- Add "Application Metrics Reference" section with tables of all custom
  metrics, their types, labels, and which component reports them
- Update the "What It Exposes" section

Update `docs/docs/sidebar/features/health-checks.md`:

- Cross-reference the metrics server health probes
- Note that `/health` and `/health/ready` are available on each
  component's metrics port without authentication

Update `docs/docs/sidebar/usage/configuration.md`:

- Note the health probe endpoints in the metrics server sections
