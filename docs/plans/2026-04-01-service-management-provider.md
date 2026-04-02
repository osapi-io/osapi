# Service Management Provider Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> superpowers:subagent-driven-development (recommended) or
> superpowers:executing-plans to implement this plan task-by-task. Steps use
> checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add systemd service management to OSAPI — list, inspect,
start/stop/restart, enable/disable services, and manage custom unit files via
Object Store deployment.

**Architecture:** Hybrid provider at `internal/provider/node/service/` using
`exec.Manager` for systemctl control operations and `file.Deployer` for unit
file CRUD. Registered under the `node` agent category. Container check enabled —
systemctl requires systemd.

**Tech Stack:** Go, exec.Manager, file.Deployer, avfs.VFS, systemctl,
oapi-codegen strict-server

---

## File Structure

### Provider Layer

- Create: `internal/provider/node/service/types.go` — Provider interface +
  domain types
- Create: `internal/provider/node/service/debian.go` — Debian struct,
  constructor, compile-time checks
- Create: `internal/provider/node/service/debian_list.go` — List implementation
  (systemctl list-units)
- Create: `internal/provider/node/service/debian_get.go` — Get implementation
  (systemctl show)
- Create: `internal/provider/node/service/debian_action.go` —
  Start/Stop/Restart/Enable/Disable
- Create: `internal/provider/node/service/debian_unit.go` — Unit file
  Create/Update/Delete via file.Deployer
- Create: `internal/provider/node/service/darwin.go` — macOS stub
- Create: `internal/provider/node/service/linux.go` — generic Linux stub
- Create: `internal/provider/node/service/mocks/generate.go`
- Test: `internal/provider/node/service/debian_list_public_test.go`
- Test: `internal/provider/node/service/debian_get_public_test.go`
- Test: `internal/provider/node/service/debian_action_public_test.go`
- Test: `internal/provider/node/service/debian_unit_public_test.go`
- Test: `internal/provider/node/service/darwin_public_test.go`
- Test: `internal/provider/node/service/linux_public_test.go`

### Agent Layer

- Create: `internal/agent/processor_service.go` — service operation dispatcher
- Modify: `internal/agent/processor.go` — add `service` case to NewNodeProcessor
- Modify: `cmd/agent_setup.go` — create service provider factory, wire into
  registry
- Test: `internal/agent/processor_service_public_test.go`

### Operations & Permissions

- Modify: `pkg/sdk/client/operations.go` — add service operation constants
- Modify: `internal/job/types.go` — add service operation aliases
- Modify: `pkg/sdk/client/permissions.go` — add `PermServiceRead`,
  `PermServiceWrite`
- Modify: `internal/authtoken/permissions.go` — add to all roles

### API Layer

- Create: `internal/controller/api/node/service/gen/api.yaml` — OpenAPI spec
- Create: `internal/controller/api/node/service/gen/cfg.yaml`
- Create: `internal/controller/api/node/service/gen/generate.go`
- Create: `internal/controller/api/node/service/types.go` — handler struct
- Create: `internal/controller/api/node/service/service.go` — New(),
  compile-time check
- Create: `internal/controller/api/node/service/validate.go` — validateHostname
- Create: `internal/controller/api/node/service/service_list_get.go` — list
  handler
- Create: `internal/controller/api/node/service/service_get.go` — get handler
- Create: `internal/controller/api/node/service/service_create_post.go` — create
  handler
- Create: `internal/controller/api/node/service/service_update_put.go` — update
  handler
- Create: `internal/controller/api/node/service/service_delete.go` — delete
  handler
- Create: `internal/controller/api/node/service/service_start_post.go` — start
  handler
- Create: `internal/controller/api/node/service/service_stop_post.go` — stop
  handler
- Create: `internal/controller/api/node/service/service_restart_post.go` —
  restart handler
- Create: `internal/controller/api/node/service/service_enable_post.go` — enable
  handler
- Create: `internal/controller/api/node/service/service_disable_post.go` —
  disable handler
- Create: `internal/controller/api/node/service/handler.go` — Handler()
  registration
- Modify: `cmd/controller_setup.go` — register service handler
- Test: one `*_public_test.go` per handler file + handler test

### SDK Layer

- Create: `pkg/sdk/client/service.go` — ServiceService methods
- Create: `pkg/sdk/client/service_types.go` — SDK result types + conversions
- Modify: `pkg/sdk/client/osapi.go` — add Service field
- Test: `pkg/sdk/client/service_public_test.go`
- Test: `pkg/sdk/client/service_types_public_test.go`

### CLI Layer

- Create: `cmd/client_node_service.go` — parent command
- Create: `cmd/client_node_service_list.go`
- Create: `cmd/client_node_service_get.go`
- Create: `cmd/client_node_service_create.go`
- Create: `cmd/client_node_service_update.go`
- Create: `cmd/client_node_service_delete.go`
- Create: `cmd/client_node_service_start.go`
- Create: `cmd/client_node_service_stop.go`
- Create: `cmd/client_node_service_restart.go`
- Create: `cmd/client_node_service_enable.go`
- Create: `cmd/client_node_service_disable.go`

### Documentation

- Create: `docs/docs/sidebar/features/service-management.md` — feature page
- Create: `docs/docs/sidebar/usage/cli/client/node/service/service.md` — CLI
  landing
- Create: CLI doc pages for all 10 subcommands
- Create: `docs/docs/sidebar/sdk/client/operations/service.md` — SDK doc
- Create: `examples/sdk/client/service.go` — SDK example
- Modify: `docs/docs/sidebar/features/features.md`
- Modify: `docs/docs/sidebar/features/authentication.md`
- Modify: `docs/docs/sidebar/usage/configuration.md`
- Modify: `docs/docs/sidebar/architecture/architecture.md`
- Modify: `docs/docs/sidebar/architecture/api-guidelines.md`
- Modify: `docs/docusaurus.config.ts`
- Modify: `docs/docs/sidebar/sdk/client/client.md`

### Integration Test

- Create: `test/integration/service_test.go`

---

### Task 1: Provider Interface and Types

**Files:**

- Create: `internal/provider/node/service/types.go`

- [ ] **Step 1: Create provider interface and types**

```go
// Package service provides systemd service management operations.
package service

import "context"

// Provider implements systemd service management operations.
type Provider interface {
	// Read
	List(ctx context.Context) ([]Info, error)
	Get(ctx context.Context, name string) (*Info, error)
	// Unit file CRUD (meta provider pattern)
	Create(ctx context.Context, entry Entry) (*CreateResult, error)
	Update(ctx context.Context, entry Entry) (*UpdateResult, error)
	Delete(ctx context.Context, name string) (*DeleteResult, error)
	// Control actions (direct provider pattern)
	Start(ctx context.Context, name string) (*ActionResult, error)
	Stop(ctx context.Context, name string) (*ActionResult, error)
	Restart(ctx context.Context, name string) (*ActionResult, error)
	Enable(ctx context.Context, name string) (*ActionResult, error)
	Disable(ctx context.Context, name string) (*ActionResult, error)
}

// Info represents a systemd service.
type Info struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Enabled     bool   `json:"enabled"`
	Description string `json:"description,omitempty"`
	PID         int    `json:"pid,omitempty"`
}

// Entry represents a unit file deployment request.
type Entry struct {
	Name   string `json:"name"`
	Object string `json:"object,omitempty"`
}

// CreateResult represents the outcome of a unit file creation.
type CreateResult struct {
	Name    string `json:"name"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// UpdateResult represents the outcome of a unit file update.
type UpdateResult struct {
	Name    string `json:"name"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// DeleteResult represents the outcome of a unit file deletion.
type DeleteResult struct {
	Name    string `json:"name"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// ActionResult represents the outcome of a service control action.
type ActionResult struct {
	Name    string `json:"name"`
	Changed bool   `json:"changed"`
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/provider/node/service/...`

- [ ] **Step 3: Commit**

```bash
git add internal/provider/node/service/types.go
git commit -m "feat(service): add provider interface and types"
```

---

### Task 2: Platform Stubs (Darwin + Linux)

**Files:**

- Create: `internal/provider/node/service/darwin.go`
- Create: `internal/provider/node/service/linux.go`
- Test: `internal/provider/node/service/darwin_public_test.go`
- Test: `internal/provider/node/service/linux_public_test.go`

- [ ] **Step 1: Write stub tests**

All 10 methods must return `provider.ErrUnsupported`. One suite method per
provider method, all in a single table. Follow
`internal/provider/node/certificate/darwin_public_test.go`.

- [ ] **Step 2: Implement stubs**

All methods return `fmt.Errorf("service: %w", provider.ErrUnsupported)`.

- [ ] **Step 3: Create mocks**

Create `internal/provider/node/service/mocks/generate.go` and run
`go generate ./internal/provider/node/service/mocks/...`.

- [ ] **Step 4: Run tests**

Run: `go test -v ./internal/provider/node/service/...`

- [ ] **Step 5: Commit**

```bash
git add internal/provider/node/service/
git commit -m "feat(service): add darwin and linux stubs"
```

---

### Task 3: Debian Implementation — Read Operations

**Files:**

- Create: `internal/provider/node/service/debian.go`
- Create: `internal/provider/node/service/debian_list.go`
- Create: `internal/provider/node/service/debian_get.go`
- Test: `internal/provider/node/service/debian_list_public_test.go`
- Test: `internal/provider/node/service/debian_get_public_test.go`

- [ ] **Step 1: Implement debian.go**

Debian struct with all dependencies:

```go
type Debian struct {
	provider.FactsAware
	logger       *slog.Logger
	fs           avfs.VFS
	fileDeployer file.Deployer
	stateKV      jetstream.KeyValue
	execManager  exec.Manager
	hostname     string
}
```

Constructor
`NewDebianProvider(logger, fs, fileDeployer, stateKV, execManager, hostname)`
with subsystem `"provider.service"`.

Compile-time checks for Provider and FactsSetter.

- [ ] **Step 2: Implement debian_list.go**

`List` runs `systemctl list-units --type=service --all --output=json`. Parse
JSON output — each entry has `unit` (name), `active` (state), `sub` (sub-state),
`description`. Map `active` to status and determine enabled via a separate
`systemctl is-enabled {name}` call (or batch via
`systemctl list-unit-files --type=service --output=json`).

For efficiency, use two commands:

1. `systemctl list-units --type=service --all --output=json` for active/inactive
   status
2. `systemctl list-unit-files --type=service --output=json` for enabled/disabled
   status

Merge the two into `[]Info`.

- [ ] **Step 3: Implement debian_get.go**

`Get` runs
`systemctl show {name}.service --property=ActiveState,UnitFileState,Description,MainPID`.
Parse key=value output. Map `ActiveState` to status, `UnitFileState=enabled` to
`Enabled: true`, `MainPID` to PID (0 if inactive), `Description` to description.

If service not found (exit code non-zero), return error.

- [ ] **Step 4: Write tests**

**TestList** — table-driven with gomock for execManager:

- success (mock returns JSON for 2 services)
- exec error on list-units
- exec error on list-unit-files
- empty service list
- malformed JSON

**TestGet** — table-driven:

- success (active, enabled service)
- success (inactive, disabled)
- service not found (exec error)
- malformed output

Use `memfs.New()` for fs (not needed for read ops but required by the struct).

- [ ] **Step 5: Verify 100% coverage**

- [ ] **Step 6: Commit**

```bash
git add internal/provider/node/service/
git commit -m "feat(service): add list and get operations"
```

---

### Task 4: Debian Implementation — Control Actions

**Files:**

- Create: `internal/provider/node/service/debian_action.go`
- Test: `internal/provider/node/service/debian_action_public_test.go`

- [ ] **Step 1: Implement control actions**

Each action checks current state for idempotency, then runs the systemctl
command:

**Start**: `systemctl is-active {name}` → if "active", return `changed: false`.
Otherwise `systemctl start {name}`.

**Stop**: `systemctl is-active {name}` → if not "active", return
`changed: false`. Otherwise `systemctl stop {name}`.

**Restart**: Always `systemctl restart {name}`, return `changed: true`. No
idempotency — restart is intentional.

**Enable**: `systemctl is-enabled {name}` → if "enabled", return
`changed: false`. Otherwise `systemctl enable {name}`.

**Disable**: `systemctl is-enabled {name}` → if not "enabled", return
`changed: false`. Otherwise `systemctl disable {name}`.

All actions validate the service name first using the same `validateName` regex
from the certificate provider: `^[a-zA-Z0-9_@.-]+$` (service names allow dots
and `@`).

Error wrapping: `fmt.Errorf("service: start: %w", err)` etc.

- [ ] **Step 2: Write tests**

**TestStart** — table-driven:

- success (is-active returns "inactive", start succeeds)
- already active → changed: false
- start error
- is-active check error
- invalid name

**TestStop** — table-driven:

- success (is-active returns "active", stop succeeds)
- already stopped → changed: false
- stop error
- invalid name

**TestRestart** — table-driven:

- success (always changed: true)
- restart error
- invalid name

**TestEnable** — table-driven:

- success (is-enabled returns "disabled", enable succeeds)
- already enabled → changed: false
- enable error
- invalid name

**TestDisable** — table-driven:

- success (is-enabled returns "enabled", disable succeeds)
- already disabled → changed: false
- disable error
- invalid name

- [ ] **Step 3: Verify 100% coverage**

- [ ] **Step 4: Commit**

```bash
git add internal/provider/node/service/
git commit -m "feat(service): add start/stop/restart/enable/disable"
```

---

### Task 5: Debian Implementation — Unit File CRUD

**Files:**

- Create: `internal/provider/node/service/debian_unit.go`
- Test: `internal/provider/node/service/debian_unit_public_test.go`

Follow the certificate provider pattern exactly. Read
`internal/provider/node/certificate/debian.go`.

- [ ] **Step 1: Implement unit file CRUD**

**Create**: Validate name. Check file doesn't exist at
`/etc/systemd/system/osapi-{name}.service`. Deploy via `fileDeployer.Deploy`
with mode `0644`, contentType `raw`, metadata `{"source": "custom"}`. If
changed, run `systemctl daemon-reload`.

**Update**: Validate name. Check file EXISTS. Deploy same path with new object.
If object empty, preserve from state KV (like cron). If changed, run
`daemon-reload`.

**Delete**: Validate name. Check file exists (if not, return changed: false).
Stop and disable service first (best-effort — log warnings on error). Undeploy
via `fileDeployer.Undeploy`. If changed, run `daemon-reload`.

Helper: `daemonReload()` runs `systemctl daemon-reload` via execManager.

Helper: `isManagedFile(ctx, path)` checks file state KV, same pattern as
cron/certificate.

Helper: `buildEntryFromState(ctx, name, path)` reads state KV and reconstructs
Entry.

- [ ] **Step 2: Write tests**

Use gomock for fileDeployer, stateKV, execManager. Use `memfs.New()` for
filesystem.

**TestCreate** — table-driven:

- success (deploy + daemon-reload)
- already exists (fs.Stat succeeds → error)
- deploy error
- daemon-reload error
- invalid name
- deploy returns changed:false → skip daemon-reload

**TestUpdate** — table-driven:

- success (stat finds file, deploy + daemon-reload)
- not found → error
- deploy error
- unchanged (changed:false, skip daemon-reload)
- daemon-reload error
- invalid name
- preserve object when not specified

**TestDelete** — table-driven:

- success (stop + disable + undeploy + daemon-reload)
- not found → changed:false
- undeploy error
- daemon-reload error
- stop/disable failures are non-fatal
- invalid name

- [ ] **Step 3: Verify 100% coverage**

- [ ] **Step 4: Commit**

```bash
git add internal/provider/node/service/
git commit -m "feat(service): add unit file CRUD via file.Deployer"
```

---

### Task 6: Operations, Permissions, and Agent Wiring

**Files:**

- Modify: `pkg/sdk/client/operations.go`
- Modify: `internal/job/types.go`
- Modify: `pkg/sdk/client/permissions.go`
- Modify: `internal/authtoken/permissions.go`
- Create: `internal/agent/processor_service.go`
- Modify: `internal/agent/processor.go`
- Modify: `cmd/agent_setup.go`
- Test: `internal/agent/processor_service_public_test.go`

- [ ] **Step 1: Add operation constants**

```go
// Service operations.
const (
	OpServiceList    JobOperation = "node.service.list"
	OpServiceGet     JobOperation = "node.service.get"
	OpServiceCreate  JobOperation = "node.service.create"
	OpServiceUpdate  JobOperation = "node.service.update"
	OpServiceDelete  JobOperation = "node.service.delete"
	OpServiceStart   JobOperation = "node.service.start"
	OpServiceStop    JobOperation = "node.service.stop"
	OpServiceRestart JobOperation = "node.service.restart"
	OpServiceEnable  JobOperation = "node.service.enable"
	OpServiceDisable JobOperation = "node.service.disable"
)
```

Plus aliases in `internal/job/types.go`.

- [ ] **Step 2: Add permissions**

`PermServiceRead` and `PermServiceWrite`. Add to AllPermissions. Add both to
admin, both to write, read-only to read role.

- [ ] **Step 3: Implement processor**

`processServiceOperation` dispatches 10 sub-operations. Follow
`processor_certificate.go` for CRUD and `processor_power.go` for action sub-ops.

Sub-ops: `list`, `get`, `create`, `update`, `delete`, `start`, `stop`,
`restart`, `enable`, `disable`.

For list: no data needed. For get/start/stop/restart/enable/ disable: unmarshal
`{"name":"..."}`. For create/update: unmarshal `service.Entry`.

- [ ] **Step 4: Wire into node processor**

Add `serviceProvider service.Provider` parameter to `NewNodeProcessor`. Add
`case "service"` to the switch.

- [ ] **Step 5: Wire in agent_setup.go**

Create `createServiceProvider` — on Debian, needs `fileProvider`, `fileStateKV`,
`execManager`, `hostname`. Container check: return Linux stub if
`platform.IsContainer()`. If `fileProvider == nil`, log warning, return Linux
stub.

- [ ] **Step 6: Write processor tests**

**TestProcessServiceOperation** — dispatch-level table:

- nil provider, invalid operation, unsupported sub-op

One suite method per sub-operation (list, get, create, update, delete, start,
stop, restart, enable, disable). Each with success, unmarshal error (where
applicable), and provider error cases.

- [ ] **Step 7: Verify 100% coverage**

- [ ] **Step 8: Commit**

```bash
git commit -m "feat(service): add operations, permissions, and agent wiring"
```

---

### Task 7: OpenAPI Spec and Code Generation

**Files:**

- Create: `internal/controller/api/node/service/gen/api.yaml`
- Create: `internal/controller/api/node/service/gen/cfg.yaml`
- Create: `internal/controller/api/node/service/gen/generate.go`

- [ ] **Step 1: Create OpenAPI spec**

10 endpoints. Parameters: Hostname, ServiceName (path).

Request schemas:

- `ServiceCreateRequest` — name + object (both required, validate tags)
- `ServiceUpdateRequest` — object (required, validate tag)

Response schemas:

- `ServiceInfo` — name, status, enabled, description, pid
- `ServiceListEntry` — hostname, status (ok/failed/skipped), services (array of
  ServiceInfo), error
- `ServiceGetEntry` — hostname, status, service (ServiceInfo), error
- `ServiceMutationEntry` — hostname, status, name, changed, error
- `ServiceListResponse` — job_id, results (ServiceListEntry[])
- `ServiceGetResponse` — job_id, results (ServiceGetEntry[])
- `ServiceMutationResponse` — job_id, results (ServiceMutationEntry[])

Action endpoints (start/stop/restart/enable/disable) use POST with no request
body. They share the `ServiceMutationResponse`.

Security: `service:read` for GET, `service:write` for POST/PUT/DELETE.

- [ ] **Step 2: Generate code and rebuild**

```bash
go generate ./internal/controller/api/node/service/gen/...
just generate
go build ./...
```

- [ ] **Step 3: Commit**

```bash
git commit -m "feat(service): add OpenAPI spec and generated code"
```

---

### Task 8: API Handler Implementation

**Files:**

- Create all handler files under `internal/controller/api/node/service/`
- Modify: `cmd/controller_setup.go`

- [ ] **Step 1: Create handler scaffolding**

`types.go`, `service.go` (New + compile-time check, subsystem `"api.service"`),
`validate.go`, `handler.go`.

- [ ] **Step 2: Implement list handler**

`service_list_get.go` — Query with `"node"` category,
`job.OperationServiceList`. Parse `[]serviceProv.Info` from response. Broadcast
support.

- [ ] **Step 3: Implement get handler**

`service_get.go` — Query with `job.OperationServiceGet` and data
`{"name": name}`. Parse single `serviceProv.Info`.

- [ ] **Step 4: Implement CRUD handlers**

`service_create_post.go` — Modify with `job.OperationServiceCreate`. Validate
body. Parse mutation response.

`service_update_put.go` — Modify with `job.OperationServiceUpdate`. Name from
path param. Handle 404.

`service_delete.go` — Modify with `job.OperationServiceDelete`.

- [ ] **Step 5: Implement action handlers**

`service_start_post.go`, `service_stop_post.go`, `service_restart_post.go`,
`service_enable_post.go`, `service_disable_post.go` — all Modify with no request
body, name from path. Data: `{"name": name}`.

- [ ] **Step 6: Register in controller_setup.go**

Add import and handler registration.

- [ ] **Step 7: Write tests**

One test file per handler file. Each needs success, error, skipped, broadcast,
validation, HTTP wiring, and RBAC tests (401/403/200). One suite method per
handler function, all scenarios as table rows.

- [ ] **Step 8: Verify 100% coverage**

- [ ] **Step 9: Commit**

```bash
git commit -m "feat(service): add API handlers with broadcast support"
```

---

### Task 9: SDK Service

**Files:**

- Create: `pkg/sdk/client/service.go`
- Create: `pkg/sdk/client/service_types.go`
- Modify: `pkg/sdk/client/osapi.go`
- Test: `pkg/sdk/client/service_public_test.go`
- Test: `pkg/sdk/client/service_types_public_test.go`

- [ ] **Step 1: Implement types**

```go
type ServiceInfoResult struct {
	Hostname string        `json:"hostname"`
	Status   string        `json:"status"`
	Services []ServiceInfo `json:"services,omitempty"`
	Error    string        `json:"error,omitempty"`
}

type ServiceInfo struct {
	Name        string `json:"name,omitempty"`
	Status      string `json:"status,omitempty"`
	Enabled     bool   `json:"enabled"`
	Description string `json:"description,omitempty"`
	PID         int    `json:"pid,omitempty"`
}

type ServiceGetResult struct {
	Hostname string       `json:"hostname"`
	Status   string       `json:"status"`
	Service  *ServiceInfo `json:"service,omitempty"`
	Error    string       `json:"error,omitempty"`
}

type ServiceMutationResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Name     string `json:"name"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

type ServiceCreateOpts struct {
	Name   string
	Object string
}

type ServiceUpdateOpts struct {
	Object string
}
```

Plus conversion functions.

- [ ] **Step 2: Implement service methods**

10 methods on `ServiceService`:

- `List(ctx, hostname)` → `*Response[Collection[ServiceInfoResult]]`
- `Get(ctx, hostname, name)` → `*Response[Collection[ServiceGetResult]]`
- `Create(ctx, hostname, opts)` → `*Response[Collection[ServiceMutationResult]]`
- `Update(ctx, hostname, name, opts)` →
  `*Response[Collection[ServiceMutationResult]]`
- `Delete(ctx, hostname, name)` → `*Response[Collection[ServiceMutationResult]]`
- `Start/Stop/Restart/Enable/Disable(ctx, hostname, name)` →
  `*Response[Collection[ServiceMutationResult]]`

Error wrapping: `"service list: %w"`, `"service start: %w"`, etc.

- [ ] **Step 3: Wire in osapi.go**

Add `Service *ServiceService` to Client and init in New().

- [ ] **Step 4: Regenerate SDK client**

- [ ] **Step 5: Write tests**

httptest.Server tests for all 10 methods. Each covers 200, 401, 403, 500, nil
body, transport error. Create/Update also cover 400. Update/Delete also
cover 404.

Conversion function tests for all converters.

- [ ] **Step 6: Verify 100% coverage**

- [ ] **Step 7: Commit**

```bash
git commit -m "feat(service): add SDK service with tests"
```

---

### Task 10: CLI Commands

**Files:**

- Create: 11 CLI command files

- [ ] **Step 1: Create parent command**

`cmd/client_node_service.go` — `Use: "service"`,
`Short: "Manage systemd services"`. Register under `clientNodeCmd`.

- [ ] **Step 2: Create list command**

`cmd/client_node_service_list.go` — table headers: `NAME`, `STATUS`, `ENABLED`,
`DESCRIPTION`. Uses `BuildBroadcastTable`.

- [ ] **Step 3: Create get command**

`cmd/client_node_service_get.go` — `--name` flag (required). Shows single
service details. Uses `BuildBroadcastTable` with `NAME`, `STATUS`, `ENABLED`,
`DESCRIPTION`, `PID`.

- [ ] **Step 4: Create CRUD commands**

`client_node_service_create.go` — `--name`, `--object` (both required). Uses
`BuildMutationTable`.

`client_node_service_update.go` — `--name`, `--object` (both required).

`client_node_service_delete.go` — `--name` (required).

- [ ] **Step 5: Create action commands**

`client_node_service_start.go`, `client_node_service_stop.go`,
`client_node_service_restart.go`, `client_node_service_enable.go`,
`client_node_service_disable.go` — each has `--name` (required). Uses
`BuildMutationTable` with `NAME`, `CHANGED`.

- [ ] **Step 6: Verify build**

```bash
go build ./...
```

- [ ] **Step 7: Commit**

```bash
git commit -m "feat(service): add CLI commands for service management"
```

---

### Task 11: Documentation and SDK Example

**Files:**

- Create all doc files listed in File Structure
- Modify all cross-reference files

- [ ] **Step 1: Create feature page**

`docs/docs/sidebar/features/service-management.md` with:

- How It Works: List, Get, Start/Stop/Restart, Enable/Disable,
  Create/Update/Delete unit files
- Operations table (10 operations)
- CLI Usage examples
- Broadcast Support
- Supported Platforms (Debian: Full, Darwin/Linux: Skipped)
- Container Behavior (ErrUnsupported in containers)
- Permissions

- [ ] **Step 2: Create CLI doc pages**

Landing page + 10 subcommand docs.

- [ ] **Step 3: Create SDK doc + example**

SDK doc at `docs/docs/sidebar/sdk/client/operations/service.md`. Example at
`examples/sdk/client/service.go`.

- [ ] **Step 4: Update cross-references**

features.md, authentication.md, configuration.md, architecture.md,
api-guidelines.md, docusaurus.config.ts, client.md.

- [ ] **Step 5: Commit**

```bash
git commit -m "docs: add service management feature docs, SDK example, and cross-references"
```

---

### Task 12: Integration Test and Final Verification

- [ ] **Step 1: Create integration test**

`test/integration/service_test.go` — test
`osapi client node service list --target _any --json`.

- [ ] **Step 2: Run full suite**

```bash
just generate
go build ./...
just go::unit
just go::vet
```

- [ ] **Step 3: Verify 100% coverage on all new code**

```bash
go test -coverprofile=/tmp/svc.cov \
  ./internal/provider/node/service/... \
  ./internal/agent/... \
  ./internal/controller/api/node/service/... \
  ./pkg/sdk/client/...
go tool cover -func=/tmp/svc.cov | \
  grep "service" | grep -v "100.0%" | \
  grep -v "mocks\|gen/"
```

- [ ] **Step 4: Commit any fixes**

```bash
git commit -m "chore(service): fix formatting and lint"
```
