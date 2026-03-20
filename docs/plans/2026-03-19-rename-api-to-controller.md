# Rename API Server to Controller Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development
> (if subagents available) or superpowers:executing-plans to implement this plan.
> Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rename the "API server" to "controller" across config, CLI, code
structure, and docs to reflect its role as the control plane process.

**Architecture:** Move `internal/api/` to `internal/controller/api/` and
`internal/notify/` to `internal/controller/notify/`. Create a new
`internal/controller/controller.go` that owns the API server, heartbeat, and
condition watcher. Rename `osapi api server start` to `osapi controller start`.
Update all config from `api.*` to `controller.*`.

**Tech Stack:** Go 1.25, Cobra CLI, Viper config, Echo HTTP

---

## Chunk 1: Config types and YAML

### Task 1: Update config types

**Files:**
- Modify: `internal/config/types.go`

- [ ] **Step 1: Replace API struct with Controller**

Replace the `API` struct and `Server` struct with `Controller` and `APIServer`:

```go
// Controller holds the control plane configuration.
type Controller struct {
	Client Client         `mapstructure:"client"`
	API    APIServer      `mapstructure:"api"          mask:"struct"`
	NATS   NATSConnection `mapstructure:"nats"`
}

// APIServer holds the HTTP server config (port + security).
type APIServer struct {
	Port     int            `mapstructure:"port"`
	Security ServerSecurity `mapstructure:"security" mask:"struct"`
}
```

Update the `Config` struct field:

```go
type Config struct {
	Controller    Controller          `mapstructure:"controller"              mask:"struct"`
	Agent         AgentConfig         `mapstructure:"agent,omitempty"`
	// ... rest unchanged
}
```

Remove the old `API` and `Server` structs. Keep `Client`, `ClientSecurity`,
`ServerSecurity`, `CORS`, `CustomRole`, `NATSConnection` unchanged.

- [ ] **Step 2: Verify it compiles (it won't — many references to fix)**

Run: `go build ./internal/config/...`
Expected: PASS (the config package itself should compile)

- [ ] **Step 3: Commit**

```
refactor(config): rename API struct to Controller
```

---

### Task 2: Update config YAML files

**Files:**
- Modify: `osapi.yaml`
- Modify: `test/integration/osapi.yaml`

- [ ] **Step 1: Update osapi.yaml**

Replace the `api:` section:

```yaml
controller:
  client:
    url: 'http://0.0.0.0:8080'
    security:
      bearer_token: '<jwt>'
  api:
    port: 8080
    security:
      signing_key: '<secret>'
      cors:
        allow_origins:
          - 'http://localhost:3001'
          - 'https://osapi-io.github.io'
  nats:
    host: 'localhost'
    port: 4222
    client_name: 'osapi-api'
    namespace: 'osapi'
    auth:
      type: 'none'
```

- [ ] **Step 2: Update test/integration/osapi.yaml**

Same structure change:

```yaml
controller:
  client:
    url: http://127.0.0.1:8080
    security:
      bearer_token: placeholder
  api:
    port: 8080
    security:
      signing_key: 111fdb0cfd9788fa6af8815f856a0374bf7a0174ad62fa8b98ec07a55f68d8d8
      cors:
        allow_origins: []
  nats:
    host: localhost
    port: 4222
    client_name: osapi-api-integration
    namespace: ""
    auth:
      type: none
```

- [ ] **Step 3: Commit**

```
refactor(config): rename api to controller in YAML files
```

---

## Chunk 2: Directory moves

### Task 3: Move internal/api/ to internal/controller/api/

**Files:**
- Move: `internal/api/` → `internal/controller/api/`
- Move: `internal/api/heartbeat.go` → `internal/controller/heartbeat.go`
- Move: `internal/api/heartbeat_test.go` → `internal/controller/heartbeat_test.go`

- [ ] **Step 1: Create directory and move files**

```bash
mkdir -p internal/controller
git mv internal/api internal/controller/api
git mv internal/controller/api/heartbeat.go internal/controller/heartbeat.go
git mv internal/controller/api/heartbeat_test.go internal/controller/heartbeat_test.go
```

- [ ] **Step 2: Update package declaration in heartbeat files**

Change `package api` to `package controller` in:
- `internal/controller/heartbeat.go`
- `internal/controller/heartbeat_test.go`

Update imports in heartbeat.go to reference `internal/controller/api` where
needed.

- [ ] **Step 3: Update all import paths project-wide**

Find and replace all occurrences:
- `"github.com/retr0h/osapi/internal/api"` → `"github.com/retr0h/osapi/internal/controller/api"`
- `"github.com/retr0h/osapi/internal/api/` → `"github.com/retr0h/osapi/internal/controller/api/`

Files that import `internal/api`:
- `cmd/api_server_setup.go` (will become `cmd/controller_setup.go` in Task 5)
- `cmd/nats_heartbeat.go`
- `cmd/start.go`
- All `internal/api/handler_*.go` files (now under `internal/controller/api/`)
- All domain packages (`internal/controller/api/health/`, etc.) — these use
  relative imports within the api package so they may not need changes

- [ ] **Step 4: Verify compilation**

Run: `go build ./...`

- [ ] **Step 5: Commit**

```
refactor: move internal/api to internal/controller/api
```

---

### Task 4: Move internal/notify/ to internal/controller/notify/

**Files:**
- Move: `internal/notify/` → `internal/controller/notify/`

- [ ] **Step 1: Move directory**

```bash
git mv internal/notify internal/controller/notify
```

- [ ] **Step 2: Update all import paths**

Find and replace:
- `"github.com/retr0h/osapi/internal/notify"` → `"github.com/retr0h/osapi/internal/controller/notify"`

Files that import `internal/notify`:
- `cmd/api_server_setup.go` (will become `cmd/controller_setup.go`)

- [ ] **Step 3: Verify compilation**

Run: `go build ./...`

- [ ] **Step 4: Commit**

```
refactor: move internal/notify to internal/controller/notify
```

---

## Chunk 3: Controller struct and CMD files

### Task 5: Create controller.go

**Files:**
- Create: `internal/controller/controller.go`

- [ ] **Step 1: Create the Controller struct**

```go
package controller

// Controller is the control plane process. It owns the API server,
// component heartbeat, and condition watcher.
type Controller struct {
	apiServer *api.Server
	// Additional fields will be added as heartbeat and watcher
	// are refactored into this struct in future work.
}
```

For now this is a thin wrapper. The existing setup logic in
`cmd/api_server_setup.go` already manages the lifecycle. The controller
struct provides a home for future sub-component ownership.

- [ ] **Step 2: Commit**

```
feat: add internal/controller/controller.go
```

---

### Task 6: Rename CMD files

**Files:**
- Remove: `cmd/api_server.go`
- Remove: `cmd/api_server_start.go`
- Remove: `cmd/api_server_setup.go`
- Create: `cmd/controller.go`
- Create: `cmd/controller_start.go`
- Create: `cmd/controller_setup.go`

- [ ] **Step 1: Move and rename files**

```bash
git mv cmd/api_server.go cmd/controller.go
git mv cmd/api_server_start.go cmd/controller_start.go
git mv cmd/api_server_setup.go cmd/controller_setup.go
```

- [ ] **Step 2: Update cmd/controller.go**

Rename `apiServerCmd` to `controllerCmd`. Change:
- `Use: "server"` → `Use: "start"`
- Parent command: registered under `rootCmd` not `apiCmd`
- Remove the `apiCmd` parent entirely
- Update all `appConfig.API` references to `appConfig.Controller`
- Update log messages from "api server" to "controller"

The command becomes `osapi controller start` (two levels: controller → start).
Actually per the spec it's just `osapi controller start` where `controller`
is the parent and `start` is the subcommand. Keep the parent for future
subcommands (e.g., `controller status`).

- [ ] **Step 3: Update cmd/controller_start.go**

- Rename `apiServerStartCmd` to `controllerStartCmd`
- Update `Use` and `Short` descriptions
- Change `appConfig.API.NATS` to `appConfig.Controller.NATS`
- Update import from `internal/api` to `internal/controller/api`

- [ ] **Step 4: Update cmd/controller_setup.go**

- Rename `setupAPIServer` to `setupController`
- Rename `registerAPIHandlers` to `registerControllerHandlers`
- Rename `startAPIHeartbeat` to `startControllerHeartbeat`
- Update all `appConfig.API` to `appConfig.Controller`:
  - `appConfig.API.Port` → `appConfig.Controller.API.Port`
  - `appConfig.API.NATS` → `appConfig.Controller.NATS`
  - `appConfig.API.Server.Security.SigningKey` →
    `appConfig.Controller.API.Security.SigningKey`
  - `appConfig.API.Server.Security.CORS.AllowOrigins` →
    `appConfig.Controller.API.Security.CORS.AllowOrigins`
  - `appConfig.API.Server.Security.Roles` →
    `appConfig.Controller.API.Security.Roles`
- Update import paths:
  - `internal/api` → `internal/controller/api`
  - `internal/notify` → `internal/controller/notify`

- [ ] **Step 5: Verify compilation**

Run: `go build ./...`

- [ ] **Step 6: Commit**

```
refactor: rename api server cmd to controller
```

---

### Task 7: Update cmd/start.go

**Files:**
- Modify: `cmd/start.go`

- [ ] **Step 1: Update references**

- `setupAPIServer` → `setupController`
- `appConfig.API.NATS` → `appConfig.Controller.NATS`
- `apiBundle` → `controllerBundle`
- Update `Short` and `Long` descriptions: "API server" → "controller"
- Update log component label: `"component", "api"` → `"component", "controller"`

- [ ] **Step 2: Verify compilation**

Run: `go build ./...`

- [ ] **Step 3: Commit**

```
refactor: update start.go for controller rename
```

---

### Task 8: Update client and token commands

**Files:**
- Modify: `cmd/client.go`
- Modify: `cmd/token_generate.go`
- Modify: `cmd/token_validate.go`

- [ ] **Step 1: Update cmd/client.go**

- `appConfig.API.URL` → `appConfig.Controller.Client.URL`
- `appConfig.API.Client.Security.BearerToken` →
  `appConfig.Controller.Client.Security.BearerToken`
- Viper binding: `"api.client.url"` → `"controller.client.url"`
- Log message: `"api.client.url"` → `"controller.client.url"`

- [ ] **Step 2: Update cmd/token_generate.go**

- All `appConfig.API` references → `appConfig.Controller`

- [ ] **Step 3: Update cmd/token_validate.go**

- All `appConfig.API` references → `appConfig.Controller`

- [ ] **Step 4: Update cmd/nats_heartbeat.go**

- Update import from `internal/api` to `internal/controller/api`

- [ ] **Step 5: Verify compilation and run tests**

Run: `go build ./... && go test ./cmd/... -count=1`

- [ ] **Step 6: Commit**

```
refactor: update client and token commands for controller config
```

---

## Chunk 4: Integration tests and verification

### Task 9: Update integration tests

**Files:**
- Modify: `test/integration/integration_test.go`
- Modify: `test/integration/osapi.yaml` (already done in Task 2)

- [ ] **Step 1: Update serverEnv()**

```go
func serverEnv() []string {
	return append(os.Environ(),
		fmt.Sprintf("OSAPI_NATS_SERVER_PORT=%d", natsPort),
		fmt.Sprintf("OSAPI_NATS_SERVER_STORE_DIR=%s", storeDir),
		fmt.Sprintf("OSAPI_CONTROLLER_API_PORT=%d", apiPort),
		fmt.Sprintf("OSAPI_CONTROLLER_NATS_PORT=%d", natsPort),
		fmt.Sprintf("OSAPI_AGENT_NATS_PORT=%d", natsPort),
		fmt.Sprintf("OSAPI_CONTROLLER_CLIENT_SECURITY_BEARER_TOKEN=%s", token),
	)
}
```

- [ ] **Step 2: Update clientEnv()**

```go
func clientEnv() []string {
	return append(os.Environ(),
		fmt.Sprintf("OSAPI_CONTROLLER_CLIENT_URL=http://127.0.0.1:%d", apiPort),
		fmt.Sprintf("OSAPI_CONTROLLER_CLIENT_SECURITY_BEARER_TOKEN=%s", token),
	)
}
```

- [ ] **Step 3: Verify full build and unit tests**

```bash
go build ./...
go test ./... -count=1
```

- [ ] **Step 4: Run integration tests**

```bash
just go::unit-int
```

- [ ] **Step 5: Commit**

```
refactor: update integration tests for controller config
```

---

## Chunk 5: Documentation

### Task 10: Update CLAUDE.md

**Files:**
- Modify: `CLAUDE.md`

- [ ] **Step 1: Update architecture section**

- `cmd/` description: replace "api server" with "controller"
- `internal/api/` → `internal/controller/api/`
- Add `internal/controller/` description
- Add `internal/controller/notify/` description
- Update "Adding a New API Domain" section paths
- Update config references throughout

- [ ] **Step 2: Commit**

```
docs: update CLAUDE.md for controller rename
```

---

### Task 11: Update Docusaurus docs

**Files:**
- Modify: `docs/docs/sidebar/usage/configuration.md`
- Modify: `docs/docs/sidebar/architecture/architecture.md`
- Modify: `docs/docs/sidebar/architecture/system-architecture.md`
- Modify: `docs/docs/sidebar/development/development.md`
- Modify: `docs/docs/sidebar/features/health-checks.md`
- Modify: `docs/docs/sidebar/features/notifications.md`
- Modify: `docs/docs/sidebar/intro.md`

- [ ] **Step 1: Update configuration.md**

Replace all `api.*` config keys with `controller.*` in:
- YAML examples
- Environment variable table
- Section reference tables

- [ ] **Step 2: Update architecture.md**

- "API Server" → "Controller" in process descriptions
- `osapi api server start` → `osapi controller start`

- [ ] **Step 3: Update system-architecture.md**

- Package layout: `internal/api/` → `internal/controller/api/`
- Handler structure references

- [ ] **Step 4: Update development.md**

- Quick reference: `osapi api server start` → `osapi controller start`

- [ ] **Step 5: Update feature docs**

- health-checks.md: update any "API server" references
- notifications.md: update any "API server" references

- [ ] **Step 6: Update intro.md**

- Quickstart section: update startup commands if they reference
  `api server start`

- [ ] **Step 7: Verify docs build**

```bash
just docs::build
```

- [ ] **Step 8: Commit**

```
docs: update all docs for controller rename
```

---

## Chunk 6: Final verification

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

- [ ] **Step 4: Integration tests**

```bash
just go::unit-int
```

- [ ] **Step 5: Verify CLI**

```bash
go run main.go controller start --help
go run main.go start --help
go run main.go client --help
```

- [ ] **Step 6: Docs build**

```bash
just docs::build
```

- [ ] **Step 7: Final commit if any fixups needed**

---

## Files Modified Summary

| File | Change |
|---|---|
| `internal/config/types.go` | `API` → `Controller`, `Server` → `APIServer` |
| `osapi.yaml` | `api:` → `controller:` |
| `test/integration/osapi.yaml` | `api:` → `controller:` |
| `internal/api/` → `internal/controller/api/` | Directory move |
| `internal/notify/` → `internal/controller/notify/` | Directory move |
| `internal/controller/heartbeat.go` | Moved from `internal/api/`, package rename |
| `internal/controller/heartbeat_test.go` | Moved from `internal/api/`, package rename |
| `internal/controller/controller.go` | New file |
| `cmd/api_server.go` → `cmd/controller.go` | Rename + update |
| `cmd/api_server_start.go` → `cmd/controller_start.go` | Rename + update |
| `cmd/api_server_setup.go` → `cmd/controller_setup.go` | Rename + update |
| `cmd/start.go` | Config path updates |
| `cmd/client.go` | Config path + viper binding updates |
| `cmd/token_generate.go` | Config path updates |
| `cmd/token_validate.go` | Config path updates |
| `cmd/nats_heartbeat.go` | Import path update |
| `test/integration/integration_test.go` | Env var updates |
| `CLAUDE.md` | Architecture references |
| `docs/docs/sidebar/usage/configuration.md` | Full config reference |
| `docs/docs/sidebar/architecture/architecture.md` | Process descriptions |
| `docs/docs/sidebar/architecture/system-architecture.md` | Package layout |
| `docs/docs/sidebar/development/development.md` | Quick reference |
| `docs/docs/sidebar/features/health-checks.md` | References |
| `docs/docs/sidebar/features/notifications.md` | References |
| `docs/docs/sidebar/intro.md` | Startup commands |
