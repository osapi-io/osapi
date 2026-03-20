# Rename API Server to Controller — Design Spec

## Goal

Rename the "API server" to "controller" to reflect its role as the control
plane process that owns multiple sub-components (REST API, notification watcher,
heartbeat, and future metrics server).

## Motivation

The API server today does more than serve HTTP:

- Runs the REST API (Echo)
- Runs the notification watcher (condition monitoring)
- Runs the component heartbeat
- Will run a metrics/ops server in the future

"Controller" captures what it actually is — the control plane process. The API
is just one thing it exposes.

## Config Changes

### Before

```yaml
api:
  client:
    url: 'http://localhost:8080'
    security:
      bearer_token: '<jwt>'
  server:
    port: 8080
    nats:
      host: localhost
      port: 4222
      client_name: osapi-api
      namespace: osapi
      auth:
        type: none
    security:
      signing_key: '<secret>'
      cors:
        allow_origins: [...]
      roles: {}
```

### After

```yaml
controller:
  client:
    url: 'http://localhost:8080'
    security:
      bearer_token: '<jwt>'
  api:
    port: 8080
    security:
      signing_key: '<secret>'
      cors:
        allow_origins: [...]
      roles: {}
  nats:
    host: localhost
    port: 4222
    client_name: osapi-api
    namespace: osapi
    auth:
      type: none
```

Key changes:

- `api` → `controller` (top-level)
- `api.server` → `controller.api` (the HTTP server config)
- `api.server.nats` → `controller.nats` (moved up, not nested under api)
- `api.client` → `controller.client` (unchanged structure)

### Go config types

```go
// Controller replaces the old API struct.
type Controller struct {
    Client Client         `mapstructure:"client"`
    API    APIServer      `mapstructure:"api"          mask:"struct"`
    NATS   NATSConnection `mapstructure:"nats"`
}

// APIServer holds the HTTP server config (port + security).
// Replaces the old Server struct minus the NATS connection
// (which moves to Controller.NATS).
type APIServer struct {
    Port     int            `mapstructure:"port"`
    Security ServerSecurity `mapstructure:"security" mask:"struct"`
}
```

The `validate:"required"` tags on `signing_key` and `bearer_token` remain
on their existing structs (`ServerSecurity`, `ClientSecurity`). Validation
works the same way — Viper unmarshals into the new structure and the
validator walks the nested structs.

### Environment variable mapping

| Config Key | Environment Variable |
|---|---|
| `controller.client.url` | `OSAPI_CONTROLLER_CLIENT_URL` |
| `controller.client.security.bearer_token` | `OSAPI_CONTROLLER_CLIENT_SECURITY_BEARER_TOKEN` |
| `controller.api.port` | `OSAPI_CONTROLLER_API_PORT` |
| `controller.api.security.signing_key` | `OSAPI_CONTROLLER_API_SECURITY_SIGNING_KEY` |
| `controller.api.security.cors.allow_origins` | `OSAPI_CONTROLLER_API_SECURITY_CORS_ALLOW_ORIGINS` |
| `controller.nats.host` | `OSAPI_CONTROLLER_NATS_HOST` |
| `controller.nats.port` | `OSAPI_CONTROLLER_NATS_PORT` |
| `controller.nats.client_name` | `OSAPI_CONTROLLER_NATS_CLIENT_NAME` |
| `controller.nats.namespace` | `OSAPI_CONTROLLER_NATS_NAMESPACE` |
| `controller.nats.auth.type` | `OSAPI_CONTROLLER_NATS_AUTH_TYPE` |

## CLI Changes

| Before | After |
|---|---|
| `osapi api server start` | `osapi controller start` |

Unchanged:

- `osapi client *` — all client commands stay the same
- `osapi agent start`
- `osapi nats server start`
- `osapi start` (all-in-one) — calls controller instead of API server

## Code Changes

### Directory moves

| From | To |
|---|---|
| `internal/api/` | `internal/controller/api/` |
| `internal/notify/` | `internal/controller/notify/` |

### New files

| File | Purpose |
|---|---|
| `internal/controller/controller.go` | Controller struct with Start/Stop. Owns the API server, heartbeat, and condition watcher. Implements `cli.Lifecycle`. |
| `cmd/controller.go` | `controllerCmd` parent command |
| `cmd/controller_start.go` | `controller start` subcommand |
| `cmd/controller_setup.go` | Setup logic (moved from `api_server_setup.go`), config paths updated |

### Removed files

| File | Reason |
|---|---|
| `cmd/api_server.go` | Replaced by `cmd/controller.go` |
| `cmd/api_server_start.go` | Replaced by `cmd/controller_start.go` |
| `cmd/api_server_setup.go` | Replaced by `cmd/controller_setup.go` |

### Modified files

| File | Change |
|---|---|
| `internal/config/types.go` | Replace `API` struct with `Controller` containing `APIServer`, `Client`, `NATSConnection`. Update `Config` struct field. |
| `cmd/start.go` | Call controller instead of API server |
| `cmd/client.go` | `appConfig.API` → `appConfig.Controller` |
| `cmd/client_*.go` | Update all config references |
| `test/integration/integration_test.go` | Update `api server start` → `controller start` in test harness |

### Heartbeat

`internal/api/heartbeat.go` moves to `internal/controller/heartbeat.go`.
The heartbeat registers the controller as a component in the registry KV.
It is a controller lifecycle concern — the API server doesn't need to know
about it. The agent has its own heartbeat in `internal/agent/` which is
unrelated.

### Notify

`internal/notify/` moves to `internal/controller/notify/`. The condition
watcher monitors the registry KV and dispatches notifications. It runs as
a goroutine owned by the controller — it has no reason to exist outside
the controller process.

## What doesn't change

- All REST API paths (`/node/`, `/job/`, `/health/`, etc.)
- `osapi client *` CLI commands
- SDK client (`pkg/sdk/client/`) — no config references, only HTTP
- Agent code (`internal/agent/`)
- NATS server code
- OpenAPI specs and generated code (stays under `internal/controller/api/`)

## Docs updates

- `docs/docs/sidebar/usage/configuration.md` — full config reference with
  new `controller.*` keys and env var table
- `docs/docs/sidebar/architecture/architecture.md` — rename "API Server"
  to "Controller" in process descriptions
- `docs/docs/sidebar/architecture/system-architecture.md` — update package
  layout, handler structure references
- `docs/docs/sidebar/usage/cli/` — update command docs for
  `controller start`
- `docs/docs/sidebar/development/development.md` — quick reference commands
- `docs/docs/sidebar/features/health-checks.md` — update references
- `docs/docs/sidebar/features/notifications.md` — update references
- `CLAUDE.md` — update architecture section (`internal/api/` →
  `internal/controller/api/`), cmd references, config references

## Breaking changes

- Config: `api.*` → `controller.*`
- Env vars: `OSAPI_API_*` → `OSAPI_CONTROLLER_*`
- CLI: `osapi api server start` → `osapi controller start`
- Integration tests: `api server start` → `controller start`
- `osapi.yaml`: must be updated before upgrading
