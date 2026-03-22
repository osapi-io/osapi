# Health Probes and Application Metrics Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development
> (if subagents available) or superpowers:executing-plans to implement this plan.
> Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `/health` and `/health/ready` probes to the per-component metrics
server, register application-specific Prometheus metrics for each component, and
bridge health into metrics with an `osapi_component_up` gauge.

**Architecture:** Health handlers live in a new `health.go` file under
`internal/telemetry/metrics/`. The existing metrics `Server` gains a
`SetReadinessFunc` method and registers `/health` + `/health/ready` routes in
`New()`. Each component wires its readiness check and application metrics in
`cmd/`. The `otelecho` middleware (already imported) gets the metrics server's
`MeterProvider` for controller HTTP metrics. Agent job metrics are instrumented
via OTEL in the handler layer.

**Tech Stack:** Go, Prometheus client_golang, OpenTelemetry (otel/metric),
otelecho middleware, testify/suite

---

## Chunk 1: Health probes on the metrics server

### Task 1: Add health handlers

**Files:**
- Create: `internal/telemetry/metrics/health.go`
- Create: `internal/telemetry/metrics/health_public_test.go`
- Modify: `internal/telemetry/metrics/types.go`
- Modify: `internal/telemetry/metrics/server.go`

- [ ] **Step 1: Add `readinessFunc` field to `Server` struct**

In `internal/telemetry/metrics/types.go`, add the field:

```go
type Server struct {
	httpServer    *http.Server
	logger        *slog.Logger
	registry      *prometheus.Registry
	meterProvider *sdkmetric.MeterProvider
	readinessFunc func() error
}
```

- [ ] **Step 2: Create `health.go` with liveness and readiness handlers**

Create `internal/telemetry/metrics/health.go`:

```go
package metrics

import (
	"encoding/json"
	"net/http"
)

// handleHealth returns a liveness probe response.
// Always returns 200 OK if the process is running.
func (s *Server) handleHealth(
	w http.ResponseWriter,
	_ *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

// handleReady returns a readiness probe response.
// Returns 200 when the component is ready, 503 when not.
func (s *Server) handleReady(
	w http.ResponseWriter,
	_ *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	if s.readinessFunc == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "not_ready",
			"error":  "readiness check not configured",
		})
		return
	}

	if err := s.readinessFunc(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "not_ready",
			"error":  err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ready",
	})
}
```

- [ ] **Step 3: Refactor `New()` to support health routes and add `SetReadinessFunc`**

Refactor `internal/telemetry/metrics/server.go` so the `Server` is created
before the mux routes are registered (health handlers are methods on
`*Server` and need a receiver). The full refactored `New()`:

```go
func New(
	host string,
	port int,
	logger *slog.Logger,
) *Server {
	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewProcessCollector(
		collectors.ProcessCollectorOpts{},
	))

	exporter, err := prometheusNewFn(
		prometheusExporter.WithRegisterer(reg),
	)
	if err != nil {
		logger.Error("failed to create prometheus exporter", "error", err)
		return nil
	}

	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))

	srv := &Server{
		logger:        logger,
		registry:      reg,
		meterProvider: mp,
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{Registry: reg},
	))
	mux.HandleFunc("/health", srv.handleHealth)
	mux.HandleFunc("/health/ready", srv.handleReady)

	srv.httpServer = &http.Server{
		Addr:              fmt.Sprintf("%s:%d", host, port),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	return srv
}
```

Add the setter method:
```go
// SetReadinessFunc sets the function called by /health/ready and the
// osapi_component_up gauge. The function should return nil when the
// component is ready, or an error describing why it is not.
func (s *Server) SetReadinessFunc(
	fn func() error,
) {
	s.readinessFunc = fn
}
```

- [ ] **Step 4: Write tests for health handlers**

Create `internal/telemetry/metrics/health_public_test.go` with a
`HealthPublicTestSuite`. Test cases:

1. `/health` returns 200 with `{"status":"ok"}`
2. `/health/ready` returns 503 when no readiness func set
3. `/health/ready` returns 503 when readiness func returns error
4. `/health/ready` returns 200 when readiness func returns nil
5. Both endpoints return `Content-Type: application/json`

Use the same `getFreePort()` + `Start()`/`Stop()` pattern as
`server_public_test.go`.

- [ ] **Step 5: Run tests**

```bash
go test -count=1 ./internal/telemetry/metrics/... -v
```

- [ ] **Step 6: Verify 100% coverage**

```bash
go test -coverprofile=/tmp/cover.out ./internal/telemetry/metrics/ && \
  go tool cover -func=/tmp/cover.out | grep metrics
```

- [ ] **Step 7: Commit**

```bash
git add internal/telemetry/metrics/
git commit -m "feat: add /health and /health/ready probes to metrics server"
```

---

### Task 2: Add `osapi_component_up` gauge

**Files:**
- Modify: `internal/telemetry/metrics/server.go`
- Modify: `internal/telemetry/metrics/server_public_test.go`

- [ ] **Step 1: Register `osapi_component_up` GaugeFunc in `New()`**

After creating the registry and before creating the mux, register the gauge:

```go
reg.MustRegister(prometheus.NewGaugeFunc(
	prometheus.GaugeOpts{
		Name: "osapi_component_up",
		Help: "Whether the component is ready (1) or not (0).",
	},
	func() float64 {
		if srv.readinessFunc == nil {
			return 0
		}
		if srv.readinessFunc() != nil {
			return 0
		}
		return 1
	},
))
```

Note: the GaugeFunc closure captures `srv` (the `*Server` pointer). Since
`readinessFunc` is set post-construction via `SetReadinessFunc`, the closure
sees the updated value at scrape time.

- [ ] **Step 2: Add test for `osapi_component_up` in metrics output**

Add test cases to `TestStartAndStop` in `server_public_test.go`:

1. When no readiness func set, `/metrics` contains
   `osapi_component_up 0`
2. After `SetReadinessFunc(func() error { return nil })`, `/metrics`
   contains `osapi_component_up 1`
3. After `SetReadinessFunc(func() error { return errors.New("...") })`,
   `/metrics` contains `osapi_component_up 0`

- [ ] **Step 3: Run tests and verify coverage**

```bash
go test -count=1 -coverprofile=/tmp/cover.out \
  ./internal/telemetry/metrics/... && \
  go tool cover -func=/tmp/cover.out | grep metrics
```

- [ ] **Step 4: Commit**

```bash
git add internal/telemetry/metrics/
git commit -m "feat: add osapi_component_up gauge to metrics server"
```

---

### Task 3: Wire readiness checks for all three components

**Files:**
- Modify: `cmd/controller_start.go`
- Modify: `cmd/agent_start.go`
- Modify: `cmd/agent_setup.go`
- Modify: `cmd/nats_server_start.go`
- Modify: `cmd/start.go`
- Modify: `internal/agent/agent.go` (add `IsReady()` method)
- Modify: `internal/agent/agent_public_test.go` (test `IsReady()`)

- [ ] **Step 1: Add `IsReady()` method to agent**

In `internal/agent/agent.go`:

```go
// IsReady returns nil when the agent is ready to process jobs,
// or an error describing why it is not.
func (a *Agent) IsReady() error {
	if a.ctx == nil || a.ctx.Err() != nil {
		return fmt.Errorf("agent not started")
	}
	return nil
}
```

Test in `agent_public_test.go`: call before `Start()` → error, after
`Start()` → nil.

- [ ] **Step 2: Wire controller readiness**

The `NATSChecker` is created inside `setupController` as a local variable
and not stored on the bundle. Add a `checker` field to `natsBundle` (or the
equivalent controller bundle struct in `cmd/controller_setup.go`) and store
the `NATSChecker` there so it can be accessed from the startup command.

Then in `cmd/controller_start.go`, after creating the metrics server:

```go
if metricsServer != nil {
	metricsServer.SetReadinessFunc(func() error {
		return b.checker.CheckHealth(context.Background())
	})
}
```

- [ ] **Step 3: Wire agent readiness**

Note: `setupAgent` returns `(cli.Lifecycle, *natsBundle)`, not
`*agent.Agent`. The `cli.Lifecycle` interface only has `Start()` and
`Stop(ctx)`. To call `IsReady()`, `SetMeterProvider()`, and
`LastHeartbeatTime()`, change `setupAgent` to return `*agent.Agent`
directly (it satisfies `cli.Lifecycle` since it has `Start()` and
`Stop(ctx)`). Update the call sites in `agent_start.go` and `start.go`.

Then in `cmd/agent_start.go`, after creating the metrics server:

```go
if metricsServer != nil {
	metricsServer.SetReadinessFunc(func() error {
		return a.IsReady()
	})
}
```

`IsReady()` will return error until `a.Start()` is called.

- [ ] **Step 4: Wire NATS readiness**

The `natsembedded.Server` does not expose `JetStreamEnabled()` on its
public interface. Since `setupNATSServer` only returns successfully after
JetStream infrastructure is fully configured, the NATS server is always
ready if the process is running. Use a simple always-ready check:

```go
if metricsServer != nil {
	metricsServer.SetReadinessFunc(func() error {
		return nil
	})
}
```

If a more meaningful readiness check is needed later (e.g., checking that
the NATS server is still accepting connections), the `natsembedded` package
can be extended with a health method. For now, the `osapi_component_up`
gauge will reflect 1 as long as the process is running, and the heartbeat
TTL expiry handles the "server is down" case in the registry.

- [ ] **Step 5: Wire readiness in combined start mode**

In `cmd/start.go`, wire each metrics server's readiness after creation,
using the same patterns as the standalone commands.

- [ ] **Step 6: Run full test suite**

```bash
go build ./... && go test -count=1 ./...
```

- [ ] **Step 7: Commit**

```bash
git add cmd/ internal/agent/
git commit -m "feat: wire readiness checks for all components"
```

---

## Chunk 2: Application metrics

### Task 4: Controller HTTP metrics via otelecho

**Files:**
- Modify: `internal/controller/api/server.go`
- Modify: `internal/controller/api/types.go`
- Modify: `cmd/controller_start.go`
- Modify: `cmd/controller_setup.go`
- Modify: `cmd/start.go`

The `otelecho` middleware is already wired in `server.go:54`:
```go
e.Use(otelecho.Middleware("osapi-api"))
```

Currently it uses the global OTEL provider (from tracing). To route metrics
to the metrics server's isolated `MeterProvider`, we need to pass the
`MeterProvider` to the middleware via options.

- [ ] **Step 1: Add `MeterProvider` option to `Server`**

In `internal/controller/api/types.go`, add a field and option:

```go
import sdkmetric "go.opentelemetry.io/otel/sdk/metric"

// In Server struct:
meterProvider *sdkmetric.MeterProvider

// Option:
func WithMeterProvider(mp *sdkmetric.MeterProvider) Option {
	return func(s *Server) {
		s.meterProvider = mp
	}
}
```

- [ ] **Step 2: Move otelecho middleware after option application**

In `server.go`, the current `e.Use(otelecho.Middleware("osapi-api"))` at
line 54 runs before the options loop at lines 77-79, so `s.meterProvider`
is always nil when the middleware is registered.

Fix: move all `e.Use(...)` calls to after the options loop. The order
should be:

1. Create echo instance
2. Apply options (which sets `s.meterProvider`, `s.auditStore`, etc.)
3. Register middleware (otelecho with optional MeterProvider, slogecho,
   recover, requestID, CORS, audit)

```go
// After opts loop:
otelOpts := []otelecho.Option{
	otelecho.WithTracerProvider(otel.GetTracerProvider()),
}
if s.meterProvider != nil {
	otelOpts = append(otelOpts,
		otelecho.WithMeterProvider(s.meterProvider))
}
e.Use(otelecho.Middleware("osapi-api", otelOpts...))
e.Use(slogecho.New(logger))
e.Use(middleware.Recover())
e.Use(middleware.RequestID())
e.Use(middleware.CORSWithConfig(corsConfig))
```

Remove the duplicate `e.Use(middleware.Recover())` that currently exists
at line 59.

- [ ] **Step 3: Pass `MeterProvider` from cmd**

In `cmd/controller_start.go` (and the controller setup in `start.go`),
pass the metrics server's `MeterProvider` to `api.New()`:

```go
if metricsServer != nil {
	opts = append(opts, api.WithMeterProvider(metricsServer.MeterProvider()))
}
```

Check how `setupController` creates the API server — the options are built
in `controller_setup.go`. Pass the `MeterProvider` through.

- [ ] **Step 4: Add `osapi_jobs_created_total` counter**

Job creation happens in `internal/controller/api/job/` handlers that call
`jc.PublishJob()`. The `jobclient.JobClient` is passed into the job handler
via `sm.GetJobHandler(jc)`. To instrument job creation:

1. Add a `jobsCreated metric.Int64Counter` field to the job handler struct
   in `internal/controller/api/job/types.go`
2. Create the counter in `cmd/controller_setup.go` using the metrics
   server's `MeterProvider` and pass it to the job handler via a new option
   (e.g., `job.WithJobsCreatedCounter(counter)`)
3. In each job handler that calls `PublishJob`, increment the counter after
   a successful publish

If the metrics server is nil (disabled), skip counter creation and the
handler nil-checks the counter before incrementing (same pattern as agent
metrics).

- [ ] **Step 5: Test controller metrics appear in `/metrics` output**

Write an integration-style test or verify manually that after wiring,
hitting the API server and then scraping `/metrics` on the metrics port
shows `osapi_api_request_duration_seconds` and `osapi_api_requests_total`
(these come from otelecho automatically).

- [ ] **Step 6: Commit**

```bash
git add internal/controller/api/ cmd/
git commit -m "feat: add controller HTTP and job creation metrics"
```

---

### Task 5: Agent job metrics

**Files:**
- Modify: `internal/agent/types.go` (add meter fields)
- Modify: `internal/agent/agent.go` (add `SetMeterProvider()`)
- Modify: `internal/agent/handler.go` (instrument job processing)
- Modify: `internal/agent/heartbeat.go` (track lastHeartbeatTime)
- Modify: `cmd/agent_start.go` (wire MeterProvider + heartbeat gauge)
- Modify: `cmd/agent_setup.go` (wire MeterProvider)
- Modify: `cmd/start.go` (wire in combined mode)

- [ ] **Step 1: Add OTEL meter fields to Agent**

In `internal/agent/types.go`, add fields for the OTEL instruments:

```go
import (
	"go.opentelemetry.io/otel/metric"
)

// In Agent struct:
jobsProcessed metric.Int64Counter
jobsActive    metric.Int64UpDownCounter
jobDuration   metric.Float64Histogram
```

- [ ] **Step 2: Add `SetMeterProvider()` to Agent**

In `internal/agent/agent.go`:

```go
// SetMeterProvider creates OTEL instruments for job metrics.
func (a *Agent) SetMeterProvider(
	mp *sdkmetric.MeterProvider,
) {
	meter := mp.Meter("osapi-agent")

	a.jobsProcessed, _ = meter.Int64Counter(
		"osapi_jobs_processed_total",
		metric.WithDescription("Total jobs processed"),
	)
	a.jobsActive, _ = meter.Int64UpDownCounter(
		"osapi_jobs_active",
		metric.WithDescription("Currently executing jobs"),
	)
	a.jobDuration, _ = meter.Float64Histogram(
		"osapi_job_duration_seconds",
		metric.WithDescription("Job execution duration in seconds"),
	)
}
```

- [ ] **Step 3: Instrument `handleJobMessage`**

In `internal/agent/handler.go`, around the job processing:

At the start of job processing (after "Write started event"):
```go
if a.jobsActive != nil {
	a.jobsActive.Add(ctx, 1)
}
```

After processing completes (both success and failure paths):
```go
if a.jobsActive != nil {
	a.jobsActive.Add(ctx, -1)
}
if a.jobDuration != nil {
	a.jobDuration.Record(ctx, time.Since(startTime).Seconds())
}
if a.jobsProcessed != nil {
	status := "completed"
	if response.Status == job.StatusFailed {
		status = "failed"
	}
	a.jobsProcessed.Add(ctx, 1,
		metric.WithAttributes(attribute.String("status", status)))
}
```

- [ ] **Step 4: Track `lastHeartbeatTime`**

In `internal/agent/types.go`, add:
```go
lastHeartbeatTime atomic.Value // stores time.Time
```

In `internal/agent/heartbeat.go`, after successful KV put (line 189):
```go
a.lastHeartbeatTime.Store(time.Now())
```

In `internal/agent/agent.go`, add accessor:
```go
// LastHeartbeatTime returns the timestamp of the last successful
// heartbeat write. Returns zero time if no heartbeat has been written.
func (a *Agent) LastHeartbeatTime() time.Time {
	if t, ok := a.lastHeartbeatTime.Load().(time.Time); ok {
		return t
	}
	return time.Time{}
}
```

- [ ] **Step 5: Wire agent metrics in cmd**

In `cmd/agent_start.go` (and `cmd/start.go` for combined mode), after
creating the metrics server:

```go
if metricsServer != nil {
	a.SetMeterProvider(metricsServer.MeterProvider())

	// Register heartbeat age gauge
	metricsServer.Registry().MustRegister(
		prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "osapi_heartbeat_age_seconds",
				Help: "Seconds since last successful heartbeat write.",
			},
			func() float64 {
				t := a.LastHeartbeatTime()
				if t.IsZero() {
					return 0
				}
				return time.Since(t).Seconds()
			},
		),
	)
}
```

- [ ] **Step 6: Test agent metrics**

Add test for `SetMeterProvider` in `agent_public_test.go` — verify it
doesn't panic and instruments are created.

Add test for `LastHeartbeatTime` — returns zero before heartbeat, non-zero
after.

- [ ] **Step 7: Run tests and verify coverage**

```bash
go test -count=1 ./internal/agent/... ./cmd/...
```

- [ ] **Step 8: Commit**

```bash
git add internal/agent/ cmd/
git commit -m "feat: add agent job and heartbeat metrics"
```

---

## Chunk 3: Documentation

### Task 6: Update documentation

**Files:**
- Modify: `docs/docs/sidebar/features/metrics.md`
- Modify: `docs/docs/sidebar/features/health-checks.md`
- Modify: `docs/docs/sidebar/usage/configuration.md`

- [ ] **Step 1: Update metrics.md**

Rewrite the "What It Exposes" section and add new sections:

**Health Probes section** (new, after Endpoints):

Document that each metrics server port also serves `/health` (liveness,
always 200) and `/health/ready` (readiness, 200 or 503). No authentication
required. Useful for Kubernetes probes:

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 9091
readinessProbe:
  httpGet:
    path: /health/ready
    port: 9091
```

**Application Metrics Reference section** (new, replace "What It Exposes"):

Table of all custom metrics with columns: Metric, Type, Labels, Component,
Description. Include all metrics from the spec:

- `osapi_component_up` (gauge, all)
- `osapi_api_requests_total` (counter, controller)
- `osapi_api_request_duration_seconds` (histogram, controller)
- `osapi_jobs_created_total` (counter, controller)
- `osapi_jobs_processed_total` (counter, agent)
- `osapi_jobs_active` (gauge, agent)
- `osapi_job_duration_seconds` (histogram, agent)
- `osapi_heartbeat_age_seconds` (gauge, agent)

Plus note that Go runtime and process metrics are always included.

- [ ] **Step 2: Update health-checks.md**

Add a note in the "Endpoints" section or a new "Metrics Server Health
Probes" section explaining that `/health` and `/health/ready` are also
available on each component's metrics port (9090, 9091, 9092) without
authentication. Cross-reference the metrics page.

- [ ] **Step 3: Update configuration.md**

In `docs/docs/sidebar/usage/configuration.md`, add a note to each metrics
server section (`controller.metrics`, `agent.metrics`, `nats.server.metrics`)
that the metrics port also serves `/health` and `/health/ready` endpoints
for liveness and readiness probes.

- [ ] **Step 4: Run prettier**

```bash
npx prettier docs/docs/sidebar/features/metrics.md --write \
  --config docs/prettier.config.js
npx prettier docs/docs/sidebar/features/health-checks.md --write \
  --config docs/prettier.config.js
npx prettier docs/docs/sidebar/usage/configuration.md --write \
  --config docs/prettier.config.js
```

- [ ] **Step 5: Commit**

```bash
git add docs/
git commit -m "docs: add health probes and application metrics reference"
```

---

### Task 7: Verify

- [ ] **Step 1: Full build and test**

```bash
go build ./... && go test -count=1 ./...
```

- [ ] **Step 2: Lint**

```bash
just go::vet
```

- [ ] **Step 3: Coverage gaps check**

```bash
just go::unit-cov-gaps
```

Expect only the pre-existing defense-in-depth gaps (docker, file, job).

- [ ] **Step 4: Format check**

```bash
just go::fmt-check && just docs::fmt-check
```
