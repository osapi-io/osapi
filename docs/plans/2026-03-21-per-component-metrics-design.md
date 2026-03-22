# Per-Component Metrics and Sub-Component Health — Design Spec

## Goal

Add per-component `/metrics` endpoints on dedicated ports for the controller,
agent, and NATS server. Each component gets its own Prometheus registry and OTEL
MeterProvider. Add sub-component health reporting to the controller's
`/health/status` endpoint.

## Motivation

Today `/metrics` is served on the controller's API port (8080) using a global
OTEL meter provider. This mixes API traffic with metrics scraping, doesn't
support per-component isolation, and provides no metrics for the agent or NATS
server. Operators need independent metrics endpoints for each component, and
visibility into whether internal services (notifier, heartbeat, consumers) are
running.

## Config

### Shared type

```go
// OpsServer configures the per-component metrics HTTP server.
type OpsServer struct {
    Enabled bool `mapstructure:"enabled"`
    Port    int  `mapstructure:"port"`
}
```

### YAML

```yaml
controller:
  metrics:
    enabled: true # default: true
    port: 9090 # default: 9090

agent:
  metrics:
    enabled: true # default: true
    port: 9091 # default: 9091

nats:
  server:
    metrics:
      enabled: true # default: true
      port: 9092 # default: 9092
```

### Environment variables

| Config Key                    | Environment Variable                |
| ----------------------------- | ----------------------------------- |
| `controller.metrics.enabled`  | `OSAPI_CONTROLLER_METRICS_ENABLED`  |
| `controller.metrics.port`     | `OSAPI_CONTROLLER_METRICS_PORT`     |
| `agent.metrics.enabled`       | `OSAPI_AGENT_METRICS_ENABLED`       |
| `agent.metrics.port`          | `OSAPI_AGENT_METRICS_PORT`          |
| `nats.server.metrics.enabled` | `OSAPI_NATS_SERVER_METRICS_ENABLED` |
| `nats.server.metrics.port`    | `OSAPI_NATS_SERVER_METRICS_PORT`    |

## Architecture

### New package: `internal/ops/`

A lightweight HTTP server that serves `/metrics` on a dedicated port. Each
component creates its own instance with isolated Prometheus registry and OTEL
MeterProvider.

```go
package ops

type Server struct { ... }

func New(port int, logger *slog.Logger) *Server
func (s *Server) MeterProvider() *sdkmetric.MeterProvider
func (s *Server) Start()
func (s *Server) Stop(ctx context.Context)
```

- Implements `cli.Lifecycle`
- Creates its own `prometheus.Registry` (not the global default)
- Registers Go runtime and process collectors on that registry
- Creates an OTEL `MeterProvider` backed by a `prometheus.Exporter` tied to that
  registry
- Does NOT call `otel.SetMeterProvider()` — no global state
- Serves `promhttp.HandlerFor(registry)` on `/metrics`

### Per-component wiring

Each component:

1. Checks `metrics.enabled` in config
2. If enabled, creates `ops.New(port, logger)`
3. Starts/stops alongside the main process
4. Uses `server.MeterProvider()` to create component-specific OTEL instruments

In single-process mode (`osapi start`), three ops servers run on three ports,
each with isolated metrics.

### Removal from API port

The controller's `/metrics` endpoint is removed from port 8080. The
`handler_metrics.go` file and metrics domain package
(`internal/controller/api/metrics/`) are deleted. Metrics are served exclusively
on the ops port.

`/health`, `/health/ready`, and `/health/status` remain on port 8080.

## Sub-component health

The controller's `/health/status` endpoint adds internal service status to the
existing `components` map:

```json
{
  "status": "ok",
  "components": {
    "nats": "ok",
    "kv": "ok",
    "notifier": "ok",
    "heartbeat": "ok"
  }
}
```

### Component status values

| Component   | When `ok`                        | When `disabled`                | When `error`      |
| ----------- | -------------------------------- | ------------------------------ | ----------------- |
| `nats`      | Connected                        | —                              | Connection failed |
| `kv`        | Accessible                       | —                              | Access failed     |
| `notifier`  | Watcher running                  | `notifications.enabled: false` | —                 |
| `heartbeat` | Always (started unconditionally) | —                              | —                 |

The `disabled` status is a new value. Today only `ok` and error strings exist. A
disabled component is not unhealthy — it was intentionally turned off.

### Agent and NATS sub-components

No `/health` endpoint on agent or NATS metrics ports. Their status is visible
through the controller's `/health/status` via the registry (heartbeat data).

Future work could add `/health` to the ops server if needed for k8s probes.

## Code changes

### New files

| File                          | Purpose                                         |
| ----------------------------- | ----------------------------------------------- |
| `internal/ops/server.go`      | Ops server: Start/Stop, registry, MeterProvider |
| `internal/ops/server_test.go` | Unit tests                                      |
| `internal/ops/types.go`       | Interface definitions                           |

### Modified files

| File                          | Change                                                                |
| ----------------------------- | --------------------------------------------------------------------- |
| `internal/config/types.go`    | Add `OpsServer` to `Controller`, `AgentConfig`, `NATSServer`          |
| `cmd/controller_start.go`     | Create and start ops server                                           |
| `cmd/controller_setup.go`     | Remove metrics handler from API, add notifier/heartbeat to components |
| `cmd/agent_start.go`          | Create and start ops server                                           |
| `cmd/nats_server_start.go`    | Create and start ops server                                           |
| `cmd/start.go`                | Wire all three ops servers, stop them on shutdown                     |
| `configs/osapi.yaml`          | Add metrics sections                                                  |
| `test/integration/osapi.yaml` | Add metrics sections (disabled or test ports)                         |

### Removed files

| File                                         | Reason                        |
| -------------------------------------------- | ----------------------------- |
| `internal/controller/api/handler_metrics.go` | Metrics moved to ops server   |
| `internal/controller/api/metrics/`           | Entire metrics domain package |

### Docs

| File                                             | Change                                |
| ------------------------------------------------ | ------------------------------------- |
| `docs/docs/sidebar/usage/configuration.md`       | Add metrics config for all components |
| `docs/docs/sidebar/features/metrics.md`          | Update for per-component metrics      |
| `docs/docs/sidebar/architecture/architecture.md` | Mention ops servers                   |
| `CLAUDE.md`                                      | Add `internal/ops/` to architecture   |

## What doesn't change

- `/health`, `/health/ready`, `/health/status` stay on port 8080
- Agent and NATS heartbeat mechanism unchanged
- SDK client unchanged
- All REST API endpoints unchanged
- Existing OTEL tracing unchanged

## Breaking changes

- `/metrics` removed from port 8080 — scrapers must update to port 9090
- `telemetry.metrics.path` config key becomes unused (path is always `/metrics`
  on the ops port)
