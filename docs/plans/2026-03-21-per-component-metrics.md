# Per-Component Metrics and Sub-Component Health Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development
> (if subagents available) or superpowers:executing-plans to implement this
> plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add per-component `/metrics` endpoints on dedicated ports for the
controller, agent, and NATS server with isolated Prometheus registries and OTEL
MeterProviders. Add sub-component health to the controller's `/health/status`.

**Architecture:** New `internal/ops/` package provides a lightweight HTTP server
with its own Prometheus registry. Each component creates one if
`metrics.enabled` is true. Remove `/metrics` from the controller's API port. Add
`notifier` and `heartbeat` to the `/health/status` components map.

**Tech Stack:** Go 1.25, OTEL SDK, Prometheus client, Echo HTTP

---

## Chunk 1: Config and ops server

### Task 1: Add OpsServer config type

**Files:**

- Modify: `internal/config/types.go`

- [ ] **Step 1: Add OpsServer struct**

Add after the existing `MetricsConfig` struct:

```go
// OpsServer configures the per-component metrics HTTP server.
type OpsServer struct {
	// Enabled activates the metrics server (default: true).
	Enabled *bool `mapstructure:"enabled"`
	// Port the metrics server listens on.
	Port int `mapstructure:"port"`
}

// IsEnabled returns true if the ops server is enabled.
// Defaults to true when Enabled is nil.
func (o OpsServer) IsEnabled() bool {
	if o.Enabled == nil {
		return true
	}
	return *o.Enabled
}
```

Use `*bool` so we can distinguish "not set" (default true) from "explicitly set
to false".

- [ ] **Step 2: Add to Controller, AgentConfig, NATSServer**

Add `Metrics OpsServer` field to each:

```go
type Controller struct {
	Client  Client         `mapstructure:"client"`
	API     APIServer      `mapstructure:"api"          mask:"struct"`
	NATS    NATSConnection `mapstructure:"nats"`
	Metrics OpsServer      `mapstructure:"metrics"`
}
```

```go
type AgentConfig struct {
	// ... existing fields ...
	Metrics OpsServer `mapstructure:"metrics"`
}
```

```go
type NATSServer struct {
	// ... existing fields ...
	Metrics OpsServer `mapstructure:"metrics"`
}
```

- [ ] **Step 3: Verify config package compiles**

Run: `go build ./internal/config/...`

- [ ] **Step 4: Commit**

```
feat(config): add OpsServer metrics config to all components
```

---

### Task 2: Update YAML config files

**Files:**

- Modify: `configs/osapi.yaml`
- Modify: `test/integration/osapi.yaml`

- [ ] **Step 1: Add metrics sections to configs/osapi.yaml**

Under `controller:`:

```yaml
controller:
  metrics:
    enabled: true
    port: 9090
```

Under `agent:`:

```yaml
agent:
  metrics:
    enabled: true
    port: 9091
```

Under `nats.server:`:

```yaml
nats:
  server:
    metrics:
      enabled: true
      port: 9092
```

- [ ] **Step 2: Add metrics to test/integration/osapi.yaml**

Use `enabled: false` for integration tests (avoid port conflicts):

```yaml
controller:
  metrics:
    enabled: false

agent:
  metrics:
    enabled: false
```

No NATS metrics in integration config (it's already minimal).

- [ ] **Step 3: Commit**

```
feat(config): add metrics sections to YAML configs
```

---

### Task 3: Create internal/ops package

**Files:**

- Create: `internal/ops/server.go`
- Create: `internal/ops/server_test.go`

- [ ] **Step 1: Write the test**

```go
package ops_test

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/ops"
)

type ServerPublicTestSuite struct {
	suite.Suite
}

func (s *ServerPublicTestSuite) TestStartAndStop() {
	tests := []struct {
		name         string
		port         int
		validateFunc func()
	}{
		{
			name: "serves metrics endpoint",
			port: 19090,
			validateFunc: func() {
				resp, err := http.Get("http://127.0.0.1:19090/metrics")
				s.Require().NoError(err)
				defer resp.Body.Close()
				s.Equal(200, resp.StatusCode)

				body, err := io.ReadAll(resp.Body)
				s.Require().NoError(err)
				s.Contains(string(body), "go_goroutines")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			logger := slog.Default()
			srv := ops.New(tc.port, logger)
			srv.Start()

			// Give server time to bind.
			time.Sleep(100 * time.Millisecond)

			tc.validateFunc()

			ctx, cancel := context.WithTimeout(
				context.Background(),
				5*time.Second,
			)
			defer cancel()
			srv.Stop(ctx)
		})
	}
}

func (s *ServerPublicTestSuite) TestMeterProvider() {
	logger := slog.Default()
	srv := ops.New(19091, logger)
	s.NotNil(srv.MeterProvider())
}

func TestServerPublicTestSuite(t *testing.T) {
	suite.Run(t, new(ServerPublicTestSuite))
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/ops/... -count=1 -v` Expected: FAIL (package doesn't
exist)

- [ ] **Step 3: Write the implementation**

```go
// Package ops provides a lightweight HTTP server for per-component
// Prometheus metrics.
package ops

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	prometheusExporter "go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// Server is a lightweight HTTP server that serves /metrics with an
// isolated Prometheus registry and OTEL MeterProvider.
type Server struct {
	httpServer    *http.Server
	logger        *slog.Logger
	registry      *prometheus.Registry
	meterProvider *sdkmetric.MeterProvider
}

// New creates a new ops server on the given port.
func New(
	port int,
	logger *slog.Logger,
) *Server {
	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewProcessCollector(
		collectors.ProcessCollectorOpts{},
	))

	exporter, err := prometheusExporter.New(
		prometheusExporter.WithRegisterer(reg),
	)
	if err != nil {
		logger.Error("failed to create prometheus exporter", "error", err)
		return nil
	}

	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{Registry: reg},
	))

	return &Server{
		httpServer: &http.Server{
			Addr:              fmt.Sprintf(":%d", port),
			Handler:           mux,
			ReadHeaderTimeout: 10 * time.Second,
		},
		logger:        logger,
		registry:      reg,
		meterProvider: mp,
	}
}

// MeterProvider returns the isolated OTEL MeterProvider for this server.
// Components use this to create instruments that appear on this server's
// /metrics endpoint.
func (s *Server) MeterProvider() *sdkmetric.MeterProvider {
	return s.meterProvider
}

// Registry returns the isolated Prometheus registry for this server.
func (s *Server) Registry() *prometheus.Registry {
	return s.registry
}

// Start starts the HTTP server in a background goroutine.
func (s *Server) Start() {
	go func() {
		s.logger.Info("ops server started", "addr", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			s.logger.Error("ops server error", "error", err)
		}
	}()
}

// Stop gracefully shuts down the HTTP server and meter provider.
func (s *Server) Stop(ctx context.Context) {
	if err := s.meterProvider.Shutdown(ctx); err != nil {
		s.logger.Error("meter provider shutdown error", "error", err)
	}

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("ops server shutdown error", "error", err)
	}

	s.logger.Info("ops server stopped")
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./internal/ops/... -count=1 -v` Expected: PASS

- [ ] **Step 5: Commit**

```
feat: add internal/ops package for per-component metrics
```

---

## Chunk 2: Wire ops server into components

### Task 4: Wire into controller

**Files:**

- Modify: `cmd/controller_start.go`
- Modify: `cmd/controller_setup.go`

- [ ] **Step 1: Update controller_start.go**

After `telemetry.InitMeter()`, add ops server creation:

```go
var opsServer *ops.Server
if appConfig.Controller.Metrics.IsEnabled() {
    opsServer = ops.New(
        appConfig.Controller.Metrics.Port,
        log.With("component", "controller-ops"),
    )
}
```

Start it before `sm.Start()`:

```go
if opsServer != nil {
    opsServer.Start()
}
```

Add to shutdown:

```go
cli.RunServer(ctx, sm, func() {
    if opsServer != nil {
        opsServer.Stop(context.Background())
    }
    _ = shutdownMeter(context.Background())
    _ = shutdownTracer(context.Background())
    cli.CloseNATSClient(b.nc)
})
```

Add import: `"github.com/retr0h/osapi/internal/ops"`

- [ ] **Step 2: Remove metrics from API port**

In `cmd/controller_setup.go`, remove the `metricsHandler` and `metricsPath`
parameters from `setupController()`. Remove `sm.GetMetricsHandler()` call from
`registerControllerHandlers()`.

Update `setupController` signature:

```go
func setupController(
    ctx context.Context,
    log *slog.Logger,
    natsConfig config.NATSConnection,
) (*api.Server, *natsBundle) {
```

Remove `metricsHandler` and `metricsPath` from `controller_start.go` call.

- [ ] **Step 3: Delete metrics handler and domain package**

```bash
rm internal/controller/api/handler_metrics.go
rm -rf internal/controller/api/metrics/
```

Remove `GetMetricsHandler` method from `internal/controller/api/types.go` if it
exists, and remove it from `registerControllerHandlers()`.

- [ ] **Step 4: Verify build and tests**

```bash
go build ./...
go test ./... -count=1
```

- [ ] **Step 5: Commit**

```
feat: wire ops server into controller, remove /metrics from API port
```

---

### Task 5: Wire into agent

**Files:**

- Modify: `cmd/agent_start.go`

- [ ] **Step 1: Add ops server to agent startup**

After agent setup, create and start the ops server:

```go
var opsServer *ops.Server
if appConfig.Agent.Metrics.IsEnabled() {
    opsServer = ops.New(
        appConfig.Agent.Metrics.Port,
        logger.With("component", "agent-ops"),
    )
    opsServer.Start()
}
```

Add to shutdown:

```go
cli.RunServer(ctx, agentServer, func() {
    if opsServer != nil {
        opsServer.Stop(context.Background())
    }
    _ = shutdownTracer(context.Background())
    cli.CloseNATSClient(b.nc)
})
```

Add import: `"github.com/retr0h/osapi/internal/ops"`

- [ ] **Step 2: Verify build**

```bash
go build ./...
```

- [ ] **Step 3: Commit**

```
feat: wire ops server into agent
```

---

### Task 6: Wire into NATS server

**Files:**

- Modify: `cmd/nats_server_start.go`

- [ ] **Step 1: Add ops server to NATS server startup**

After NATS server setup:

```go
var opsServer *ops.Server
if appConfig.NATS.Server.Metrics.IsEnabled() {
    opsServer = ops.New(
        appConfig.NATS.Server.Metrics.Port,
        logger.With("component", "nats-ops"),
    )
    opsServer.Start()
}
```

Add to shutdown/cleanup.

- [ ] **Step 2: Verify build**

```bash
go build ./...
```

- [ ] **Step 3: Commit**

```
feat: wire ops server into NATS server
```

---

### Task 7: Wire into combined start

**Files:**

- Modify: `cmd/start.go`

- [ ] **Step 1: Create ops servers for all three components**

After component setup, before `composite.Start()`:

```go
var controllerOps, agentOps, natsOps *ops.Server

if appConfig.Controller.Metrics.IsEnabled() {
    controllerOps = ops.New(
        appConfig.Controller.Metrics.Port,
        logger.With("component", "controller-ops"),
    )
}
if appConfig.Agent.Metrics.IsEnabled() {
    agentOps = ops.New(
        appConfig.Agent.Metrics.Port,
        logger.With("component", "agent-ops"),
    )
}
if appConfig.NATS.Server.Metrics.IsEnabled() {
    natsOps = ops.New(
        appConfig.NATS.Server.Metrics.Port,
        logger.With("component", "nats-ops"),
    )
}
```

Start them before `composite.Start()`:

```go
for _, o := range []*ops.Server{controllerOps, agentOps, natsOps} {
    if o != nil {
        o.Start()
    }
}
```

Stop them in the shutdown closure:

```go
cli.RunServer(ctx, composite, func() {
    for _, o := range []*ops.Server{controllerOps, agentOps, natsOps} {
        if o != nil {
            o.Stop(context.Background())
        }
    }
    _ = shutdownMeter(context.Background())
    _ = shutdownTracer(context.Background())
    cli.CloseNATSClient(agentBundle.nc)
    cli.CloseNATSClient(controllerBundle.nc)
})
```

Also remove `metricsHandler`/`metricsPath` from `setupController()` call.

- [ ] **Step 2: Verify build and tests**

```bash
go build ./...
go test ./... -count=1
```

- [ ] **Step 3: Commit**

```
feat: wire ops servers into combined start
```

---

## Chunk 3: Sub-component health

### Task 8: Add notifier and heartbeat to /health/status components

**Files:**

- Modify: `cmd/controller_setup.go`
- Modify: `internal/controller/api/health/health_status_get.go`
- Modify: `internal/controller/api/health/types.go`

- [ ] **Step 1: Add component status fields to health handler**

In `internal/controller/api/health/types.go`, add to the `Health` struct or
`MetricsProvider` a way to report sub-component status. The simplest approach is
to pass them as static values when creating the health handler.

Add to the `Health` struct in `internal/controller/api/health/health.go`:

```go
type Health struct {
    // ... existing fields ...
    SubComponents map[string]string
}
```

- [ ] **Step 2: Populate in health_status_get.go**

In `GetHealthStatus`, add sub-components to the response `Components` map:

```go
for k, v := range h.SubComponents {
    components[k] = gen.ComponentHealth{
        Status: v,
    }
}
```

- [ ] **Step 3: Wire in controller_setup.go**

When creating the health handler, pass sub-component status:

```go
subComponents := map[string]string{
    "heartbeat": "ok",
}
if appConfig.Notifications.Enabled {
    subComponents["notifier"] = "ok"
} else {
    subComponents["notifier"] = "disabled"
}
```

Pass to health handler constructor.

- [ ] **Step 4: Update tests**

Add test case for sub-components in
`internal/controller/api/health/health_status_get_public_test.go`.

- [ ] **Step 5: Verify tests pass**

```bash
go test ./internal/controller/api/health/... -count=1 -v
```

- [ ] **Step 6: Commit**

```
feat: add notifier and heartbeat to /health/status components
```

---

## Chunk 4: Cleanup and docs

### Task 9: Remove old telemetry.metrics.path config

**Files:**

- Modify: `internal/config/types.go`

- [ ] **Step 1: Remove Path from MetricsConfig**

The `MetricsConfig.Path` field is no longer used — the ops server always serves
on `/metrics`. Remove the field or leave it for backwards compat.

Since nothing reads it anymore, remove it:

```go
// MetricsConfig is retained for future telemetry configuration.
type MetricsConfig struct{}
```

Or remove `MetricsConfig` entirely and simplify `Telemetry`:

```go
type Telemetry struct {
    Tracing TracingConfig `mapstructure:"tracing,omitempty"`
}
```

- [ ] **Step 2: Remove metricsPath references from controller_start.go**

Remove the `metricsHandler, metricsPath, shutdownMeter` variables if `InitMeter`
is no longer called (since ops server handles it).

Check if `InitMeter` is still needed for OTEL initialization. If not, remove the
call.

- [ ] **Step 3: Verify build and tests**

```bash
go build ./...
go test ./... -count=1
```

- [ ] **Step 4: Commit**

```
refactor: remove unused telemetry.metrics.path config
```

---

### Task 10: Update handler_public_test.go

**Files:**

- Modify: `internal/controller/api/handler_public_test.go`

- [ ] **Step 1: Remove GetMetricsHandler test**

The `TestGetMetricsHandler` test case tests the removed handler. Delete it.

- [ ] **Step 2: Verify tests pass**

```bash
go test ./internal/controller/api/... -count=1
```

- [ ] **Step 3: Commit**

```
test: remove GetMetricsHandler test
```

---

### Task 11: Update docs

**Files:**

- Modify: `CLAUDE.md`
- Modify: `docs/docs/sidebar/usage/configuration.md`
- Modify: `docs/docs/sidebar/features/metrics.md`

- [ ] **Step 1: Update CLAUDE.md**

Add `internal/ops/` to architecture section.

- [ ] **Step 2: Update configuration.md**

Add `controller.metrics`, `agent.metrics`, `nats.server.metrics` sections with
the `enabled` and `port` fields. Add env var mappings. Remove
`telemetry.metrics.path` if removed.

- [ ] **Step 3: Update metrics.md**

Update to describe per-component metrics:

- Controller on port 9090
- Agent on port 9091
- NATS on port 9092
- Each has isolated registry
- Configurable via `metrics.enabled` and `metrics.port`

- [ ] **Step 4: Commit**

```
docs: update docs for per-component metrics
```

---

## Chunk 5: Verification

### Task 12: Full verification

- [ ] **Step 1: Build**

```bash
go build ./...
```

- [ ] **Step 2: Unit tests**

```bash
go test ./... -count=1
```

- [ ] **Step 3: Lint**

```bash
just go::vet
```

- [ ] **Step 4: Manual verification**

Start osapi and verify:

```bash
go run main.go start -f configs/osapi.yaml
```

In another terminal:

```bash
# Controller metrics
curl http://localhost:9090/metrics | head -5

# Agent metrics
curl http://localhost:9091/metrics | head -5

# NATS metrics
curl http://localhost:9092/metrics | head -5

# Health status shows sub-components
go run main.go client health status --json | jq .components
```

Verify `/metrics` is NOT served on port 8080:

```bash
curl http://localhost:8080/metrics  # should 404
```

- [ ] **Step 5: Integration tests**

```bash
just go::unit-int
```

- [ ] **Step 6: Final commit if fixups needed**

---

## Files Summary

| File                                                              | Change                                                        |
| ----------------------------------------------------------------- | ------------------------------------------------------------- |
| `internal/config/types.go`                                        | Add `OpsServer`, add `Metrics` to Controller/Agent/NATSServer |
| `internal/ops/server.go`                                          | New: ops server with isolated registry                        |
| `internal/ops/server_test.go`                                     | New: tests                                                    |
| `cmd/controller_start.go`                                         | Create/start/stop ops server                                  |
| `cmd/controller_setup.go`                                         | Remove metricsHandler params, add sub-components              |
| `cmd/agent_start.go`                                              | Create/start/stop ops server                                  |
| `cmd/nats_server_start.go`                                        | Create/start/stop ops server                                  |
| `cmd/start.go`                                                    | Wire all three ops servers                                    |
| `configs/osapi.yaml`                                              | Add metrics sections                                          |
| `test/integration/osapi.yaml`                                     | Add metrics (disabled)                                        |
| `internal/controller/api/handler_metrics.go`                      | Delete                                                        |
| `internal/controller/api/metrics/`                                | Delete entire package                                         |
| `internal/controller/api/handler_public_test.go`                  | Remove metrics test                                           |
| `internal/controller/api/health/health.go`                        | Add SubComponents field                                       |
| `internal/controller/api/health/health_status_get.go`             | Emit sub-components                                           |
| `internal/controller/api/health/health_status_get_public_test.go` | Add test                                                      |
| `CLAUDE.md`                                                       | Add `internal/ops/`                                           |
| `docs/docs/sidebar/usage/configuration.md`                        | Add metrics config                                            |
| `docs/docs/sidebar/features/metrics.md`                           | Rewrite for per-component                                     |
