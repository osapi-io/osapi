# NTP + Timezone Provider Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add NTP server management (chrony) and timezone configuration (timedatectl) as node providers with full API/CLI/SDK support.

**Architecture:** Two independent direct providers under `provider/node/`. NTP manages `/etc/chrony/sources.d/osapi.sources` and reads status via `chronyc`. Timezone reads/sets via `timedatectl`. Both integrate into the node processor, each with its own API package under `api/node/`, SDK service, and CLI commands.

**Tech Stack:** Go 1.25, Echo, oapi-codegen (strict-server), NATS JetStream, gomock, testify/suite, avfs

**Coverage baseline:** 99.9% — must remain at or above this.

---

## File Map

### NTP — New Files

```
internal/provider/node/ntp/
  types.go                          — Provider interface + Status, Config, results
  debian.go                         — Chrony implementation
  darwin.go                         — macOS stub (ErrUnsupported)
  linux.go                          — Generic Linux stub (ErrUnsupported)
  export_test.go                    — Expose unexported vars for testing
  mocks/generate.go                 — //go:generate mockgen directive

internal/agent/processor_ntp.go     — NTP operation dispatch

internal/controller/api/node/ntp/
  gen/ (api.yaml, cfg.yaml, generate.go)
  types.go                          — Handler struct
  ntp.go                            — New() factory + interface check
  ntp_get.go                        — GET handler + broadcast
  ntp_create.go                     — POST handler + broadcast
  ntp_update.go                     — PUT handler + broadcast
  ntp_delete.go                     — DELETE handler + broadcast
  handler.go                        — Handler() self-registration
  validate.go                       — validateHostname
  *_public_test.go                  — Tests for all handlers + RBAC

pkg/sdk/client/
  ntp.go                            — NTPService methods
  ntp_types.go                      — SDK result types + conversions
  ntp_public_test.go
  ntp_types_public_test.go

cmd/
  client_node_ntp.go                — Parent command
  client_node_ntp_get.go            — get subcommand
  client_node_ntp_create.go         — create subcommand
  client_node_ntp_update.go         — update subcommand
  client_node_ntp_delete.go         — delete subcommand

examples/sdk/client/ntp.go
test/integration/ntp_test.go
docs/docs/sidebar/features/ntp.md
docs/docs/sidebar/usage/cli/client/node/ntp/ntp.md
docs/docs/sidebar/usage/cli/client/node/ntp/get.md
docs/docs/sidebar/usage/cli/client/node/ntp/create.md
docs/docs/sidebar/usage/cli/client/node/ntp/update.md
docs/docs/sidebar/usage/cli/client/node/ntp/delete.md
docs/docs/sidebar/sdk/client/ntp.md
```

### Timezone — New Files

```
internal/provider/node/timezone/
  types.go                          — Provider interface + Info, UpdateResult
  debian.go                         — timedatectl implementation
  darwin.go                         — macOS stub (ErrUnsupported)
  linux.go                          — Generic Linux stub (ErrUnsupported)
  mocks/generate.go

internal/agent/processor_timezone.go

internal/controller/api/node/timezone/
  gen/ (api.yaml, cfg.yaml, generate.go)
  types.go
  timezone.go
  timezone_get.go                   — GET handler + broadcast
  timezone_update.go                — PUT handler + broadcast
  handler.go
  validate.go
  *_public_test.go

pkg/sdk/client/
  timezone.go
  timezone_types.go
  timezone_public_test.go
  timezone_types_public_test.go

cmd/
  client_node_timezone.go           — Parent command
  client_node_timezone_get.go
  client_node_timezone_update.go

examples/sdk/client/timezone.go
test/integration/timezone_test.go
docs/docs/sidebar/features/timezone.md
docs/docs/sidebar/usage/cli/client/node/timezone/timezone.md
docs/docs/sidebar/usage/cli/client/node/timezone/get.md
docs/docs/sidebar/usage/cli/client/node/timezone/update.md
docs/docs/sidebar/sdk/client/timezone.md
```

### Modified Files (shared)

```
pkg/sdk/client/operations.go          — Add OpNtp*, OpTimezone*
pkg/sdk/client/permissions.go         — Add PermNtp*, PermTimezone*
pkg/sdk/client/osapi.go               — Wire NTP + Timezone services
pkg/sdk/client/export_test.go         — Add conversion bridges
internal/job/types.go                  — Re-export operations
internal/authtoken/permissions.go      — Re-export + add to roles
internal/agent/processor.go           — Add ntp/timezone cases to node processor
internal/agent/export_test.go         — Export new providers for tests
internal/agent/fixture_public_test.go — Add provider params
cmd/agent_setup.go                     — Create + register providers
cmd/controller_setup.go                — Register handlers
docs/docusaurus.config.ts             — Add to Features navbar
docs/docs/sidebar/usage/configuration.md — Add permissions
docs/docs/sidebar/features/authentication.md — Add to permissions tables
docs/docs/sidebar/architecture/api-guidelines.md — Add endpoint table
docs/docs/sidebar/architecture/architecture.md — Add feature links
CLAUDE.md                             — Update provider list
```

---

## Task 1: SDK Constants (Operations + Permissions)

**Files:**
- Modify: `pkg/sdk/client/operations.go`
- Modify: `pkg/sdk/client/permissions.go`
- Modify: `internal/job/types.go`
- Modify: `internal/authtoken/permissions.go`

- [ ] **Step 1: Add NTP and timezone operation constants**

In `pkg/sdk/client/operations.go`, add after the Sysctl block:

```go
// NTP operations.
const (
	OpNtpGet    JobOperation = "node.ntp.get"
	OpNtpCreate JobOperation = "node.ntp.create"
	OpNtpUpdate JobOperation = "node.ntp.update"
	OpNtpDelete JobOperation = "node.ntp.delete"
)

// Timezone operations.
const (
	OpTimezoneGet    JobOperation = "node.timezone.get"
	OpTimezoneUpdate JobOperation = "node.timezone.update"
)
```

- [ ] **Step 2: Add permission constants**

In `pkg/sdk/client/permissions.go`, add after `PermSysctlWrite`:

```go
	PermNtpRead       Permission = "ntp:read"
	PermNtpWrite      Permission = "ntp:write"
	PermTimezoneRead  Permission = "timezone:read"
	PermTimezoneWrite Permission = "timezone:write"
```

- [ ] **Step 3: Re-export in internal/job/types.go**

Add after the Sysctl operations block:

```go
// NTP operations.
const (
	OperationNtpGet    = client.OpNtpGet
	OperationNtpCreate = client.OpNtpCreate
	OperationNtpUpdate = client.OpNtpUpdate
	OperationNtpDelete = client.OpNtpDelete
)

// Timezone operations.
const (
	OperationTimezoneGet    = client.OpTimezoneGet
	OperationTimezoneUpdate = client.OpTimezoneUpdate
)
```

- [ ] **Step 4: Re-export permissions in internal/authtoken/permissions.go**

Add constants, add to `AllPermissions`, add to `DefaultRolePermissions`:
- `RoleAdmin`: add all four
- `RoleWrite`: add all four
- `RoleRead`: add `PermNtpRead` and `PermTimezoneRead` only

- [ ] **Step 5: Verify build**

Run: `go build ./...`

- [ ] **Step 6: Commit**

```bash
git add pkg/sdk/client/operations.go pkg/sdk/client/permissions.go \
  internal/job/types.go internal/authtoken/permissions.go
git commit -m "feat(ntp,timezone): add operation and permission constants"
```

---

## Task 2: NTP Provider Interface + Platform Stubs

**Files:**
- Create: `internal/provider/node/ntp/types.go`
- Create: `internal/provider/node/ntp/darwin.go`
- Create: `internal/provider/node/ntp/linux.go`
- Create: `internal/provider/node/ntp/mocks/generate.go`

- [ ] **Step 1: Create types.go**

```go
// Package ntp provides NTP server management via chrony.
package ntp

import "context"

// Provider implements the methods to manage NTP configuration.
type Provider interface {
	// Get returns current NTP sync status and configured servers.
	Get(ctx context.Context) (*Status, error)
	// Create deploys a managed NTP server configuration. Fails if already managed.
	Create(ctx context.Context, config Config) (*CreateResult, error)
	// Update replaces the managed NTP server configuration. Fails if not managed.
	Update(ctx context.Context, config Config) (*UpdateResult, error)
	// Delete removes the managed NTP server configuration.
	Delete(ctx context.Context) (*DeleteResult, error)
}

// Config represents an NTP server configuration to deploy.
type Config struct {
	Servers []string `json:"servers"`
}

// Status represents the current NTP sync state and configured servers.
type Status struct {
	Synchronized  bool     `json:"synchronized"`
	Stratum       int      `json:"stratum,omitempty"`
	Offset        string   `json:"offset,omitempty"`
	CurrentSource string   `json:"current_source,omitempty"`
	Servers       []string `json:"servers,omitempty"`
}

// CreateResult represents the outcome of an NTP config create operation.
type CreateResult struct {
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// UpdateResult represents the outcome of an NTP config update operation.
type UpdateResult struct {
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// DeleteResult represents the outcome of an NTP config delete operation.
type DeleteResult struct {
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}
```

- [ ] **Step 2: Create darwin.go and linux.go**

Both return `fmt.Errorf("ntp: %w", provider.ErrUnsupported)` for all methods. Follow the sysctl stub pattern exactly. Add license headers.

- [ ] **Step 3: Create mocks/generate.go**

```go
package mocks

//go:generate go tool github.com/golang/mock/mockgen -source=../types.go -destination=provider.gen.go -package=mocks
```

- [ ] **Step 4: Generate mocks and verify build**

```bash
go generate ./internal/provider/node/ntp/mocks/...
go build ./...
```

- [ ] **Step 5: Commit**

```bash
git add internal/provider/node/ntp/
git commit -m "feat(ntp): add provider interface and platform stubs"
```

---

## Task 3: NTP Debian Provider Implementation

**Files:**
- Create: `internal/provider/node/ntp/debian.go`
- Create: `internal/provider/node/ntp/export_test.go`
- Create: `internal/provider/node/ntp/debian_public_test.go`
- Create: `internal/provider/node/ntp/darwin_public_test.go`
- Create: `internal/provider/node/ntp/linux_public_test.go`

The Debian NTP provider:
- Writes `/etc/chrony/sources.d/osapi.sources` with server entries
- Reads status via `chronyc tracking` (parse output for sync state)
- Reads sources via `chronyc sources` (parse output for server list)
- Applies changes via `chronyc reload sources`
- Uses SHA-based idempotency (same approach as sysctl)

- [ ] **Step 1: Write tests for Darwin and Linux stubs**

Follow the sysctl pattern: one suite per platform, verify all methods return `ErrUnsupported`.

- [ ] **Step 2: Write failing tests for Debian provider**

Suite: `DebianPublicTestSuite` with mocks for `exec.Manager` and `avfs.VFS`.

**TestGet:**
- success (parse chronyc tracking + sources output)
- chronyc tracking error
- chronyc sources error

**TestCreate:**
- success (file doesn't exist, writes it, runs reload)
- already exists error
- write error (failfs)
- reload error (non-fatal, still returns success)
- idempotent (same content, Changed: false)

**TestUpdate:**
- success (file exists, overwrites, runs reload)
- not managed error (file doesn't exist)
- write error (failfs)
- idempotent (same content, Changed: false)

**TestDelete:**
- success (file exists, removes it, runs reload)
- not found error
- remove error (failfs)

- [ ] **Step 3: Implement debian.go**

```go
var (
	_ Provider             = (*Debian)(nil)
	_ provider.FactsSetter = (*Debian)(nil)
)

type Debian struct {
	provider.FactsAware
	logger      *slog.Logger
	fs          avfs.VFS
	execManager exec.Manager
}

func NewDebianProvider(
	logger *slog.Logger,
	fs avfs.VFS,
	execManager exec.Manager,
) *Debian {
	return &Debian{
		logger:      logger.With(slog.String("subsystem", "provider.ntp")),
		fs:          fs,
		execManager: execManager,
	}
}
```

Key implementation details:
- Config file path: `/etc/chrony/sources.d/osapi.sources`
- Config content format: `server <addr> iburst\n` per server
- **Get**: run `chronyc tracking` and parse output for `Leap status`, `Stratum`, `System time` (offset), `Reference ID`. Run `chronyc sources` and parse for server addresses. Return Status struct.
- **Create**: check if osapi.sources exists → error if yes. Write file. Run `chronyc reload sources`.
- **Update**: check if osapi.sources exists → error if no. Compare SHA of new content → skip if same. Write file. Run `chronyc reload sources`.
- **Delete**: check if exists → error if no. Remove file. Run `chronyc reload sources`.

Read `internal/provider/node/sysctl/debian.go` as the reference for the file write + idempotency pattern.

- [ ] **Step 4: Create export_test.go**

Expose any unexported variables needed for testing (e.g., `marshalJSON` or chronyc command path overrides).

- [ ] **Step 5: Verify all tests pass with 100% coverage**

```bash
go test -coverprofile=/tmp/c.out -v ./internal/provider/node/ntp/...
go tool cover -func=/tmp/c.out | grep ntp
```

- [ ] **Step 6: Commit**

```bash
git add internal/provider/node/ntp/
git commit -m "feat(ntp): implement Debian chrony provider with tests"
```

---

## Task 4: NTP Agent Processor + Wiring

**Files:**
- Create: `internal/agent/processor_ntp.go`
- Create: `internal/agent/processor_ntp_public_test.go`
- Modify: `internal/agent/processor.go` — add ntp provider param + case
- Modify: `cmd/agent_setup.go` — create + register provider

- [ ] **Step 1: Write processor tests**

Test all operations: ntp.get, ntp.create, ntp.update, ntp.delete, unsupported, nil provider. Follow `processor_sysctl_public_test.go` pattern.

- [ ] **Step 2: Implement processor_ntp.go**

```go
func processNtpOperation(
	ntpProvider ntp.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	// nil check, parse sub-operation, dispatch
}
```

Sub-operations: get (no data), create (unmarshal Config), update (unmarshal Config), delete (no data).

- [ ] **Step 3: Add to node processor**

In `processor.go`, add `ntpProvider ntp.Provider` parameter to `NewNodeProcessor` and add `case "ntp":`.

- [ ] **Step 4: Wire in agent_setup.go**

Create `createNtpProvider` function. Add provider to `NewNodeProcessor` call and `registry.Register` providers list.

NTP provider needs: logger, fs (avfs.VFS), execManager. No KV needed — it manages files directly like sysctl.

- [ ] **Step 5: Fix existing tests**

Add nil NTP provider parameter to any test that calls `NewNodeProcessor`.

- [ ] **Step 6: Verify all tests pass**

```bash
go test ./internal/agent/... ./cmd/...
```

- [ ] **Step 7: Commit**

```bash
git add internal/agent/ cmd/agent_setup.go
git commit -m "feat(ntp): add agent processor and wiring"
```

---

## Task 5: NTP OpenAPI Spec + API Handlers

**Files:**
- Create: `internal/controller/api/node/ntp/gen/api.yaml`
- Create: `internal/controller/api/node/ntp/gen/cfg.yaml`
- Create: `internal/controller/api/node/ntp/gen/generate.go`
- Create: `internal/controller/api/node/ntp/types.go`
- Create: `internal/controller/api/node/ntp/ntp.go`
- Create: `internal/controller/api/node/ntp/validate.go`
- Create: `internal/controller/api/node/ntp/ntp_get.go`
- Create: `internal/controller/api/node/ntp/ntp_create.go`
- Create: `internal/controller/api/node/ntp/ntp_update.go`
- Create: `internal/controller/api/node/ntp/ntp_delete.go`
- Create: `internal/controller/api/node/ntp/handler.go`
- Create: all `*_public_test.go` files

- [ ] **Step 1: Create OpenAPI spec**

Paths:
- `GET /node/{hostname}/ntp` — operationId: `GetNodeNtp`, security: `ntp:read`
- `POST /node/{hostname}/ntp` — operationId: `PostNodeNtp`, security: `ntp:write`
- `PUT /node/{hostname}/ntp` — operationId: `PutNodeNtp`, security: `ntp:write`
- `DELETE /node/{hostname}/ntp` — operationId: `DeleteNodeNtp`, security: `ntp:write`

Schemas:
- `NtpCreateRequest` — required: servers (array of strings)
- `NtpUpdateRequest` — required: servers (array of strings)
- `NtpStatusEntry` — hostname (req), status (req, enum ok/failed/skipped), synchronized, stratum, offset, current_source, servers, error
- `NtpMutationResult` — hostname (req), status (req), changed, error
- Collection responses wrapping results + job_id

Reference common ErrorResponse via `../../../common/gen/api.yaml`.

- [ ] **Step 2: Create cfg.yaml, generate.go, generate code**

- [ ] **Step 3: Create handler struct, factory, validate**

Follow sysctl pattern. Handler struct has JobClient + logger.

- [ ] **Step 4: Implement all handlers with broadcast support**

Follow sysctl handler pattern exactly. Category is `"node"`. Operations are `job.OperationNtpGet`, etc.

- [ ] **Step 5: Create handler.go (self-registration)**

Follow sysctl handler.go pattern.

- [ ] **Step 6: Write tests for all handlers**

Each handler needs: success, broadcast, skipped, error, not-found (for update/delete), validation (for create/update). Include RBAC HTTP tests.

- [ ] **Step 7: Wire in controller_setup.go**

```go
handlers = append(handlers, ntpAPI.Handler(log, jc, signingKey, customRoles)...)
```

- [ ] **Step 8: Verify**

```bash
go build ./...
go test ./internal/controller/api/node/ntp/... ./cmd/...
```

- [ ] **Step 9: Commit**

```bash
git add internal/controller/api/node/ntp/ cmd/controller_setup.go
git commit -m "feat(ntp): add OpenAPI spec, API handlers, and server wiring"
```

---

## Task 6: NTP SDK Service

**Files:**
- Create: `pkg/sdk/client/ntp.go`
- Create: `pkg/sdk/client/ntp_types.go`
- Create: `pkg/sdk/client/ntp_public_test.go`
- Create: `pkg/sdk/client/ntp_types_public_test.go`
- Modify: `pkg/sdk/client/osapi.go`
- Modify: `pkg/sdk/client/export_test.go`

- [ ] **Step 1: Create SDK types**

```go
type NtpStatusResult struct {
	Hostname      string   `json:"hostname"`
	Status        string   `json:"status"`
	Synchronized  bool     `json:"synchronized,omitempty"`
	Stratum       int      `json:"stratum,omitempty"`
	Offset        string   `json:"offset,omitempty"`
	CurrentSource string   `json:"current_source,omitempty"`
	Servers       []string `json:"servers,omitempty"`
	Error         string   `json:"error,omitempty"`
}

type NtpMutationResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

type NtpCreateOpts struct {
	Servers []string
}

type NtpUpdateOpts struct {
	Servers []string
}
```

Add gen→SDK conversion functions.

- [ ] **Step 2: Create SDK service**

Methods: `NtpGet`, `NtpCreate`, `NtpUpdate`, `NtpDelete`. Follow sysctl service pattern.

- [ ] **Step 3: Wire into osapi.go**

- [ ] **Step 4: Write tests**

Cover all status code paths. Use httptest.Server mocks.

- [ ] **Step 5: Verify**

```bash
go test -v ./pkg/sdk/client/...
```

- [ ] **Step 6: Commit**

```bash
git add pkg/sdk/client/
git commit -m "feat(ntp): add SDK service with tests"
```

---

## Task 7: NTP CLI Commands

**Files:**
- Create: `cmd/client_node_ntp.go`
- Create: `cmd/client_node_ntp_get.go`
- Create: `cmd/client_node_ntp_create.go`
- Create: `cmd/client_node_ntp_update.go`
- Create: `cmd/client_node_ntp_delete.go`

- [ ] **Step 1: Create parent command**

```go
var clientNodeNtpCmd = &cobra.Command{
	Use:   "ntp",
	Short: "Manage NTP configuration",
}

func init() {
	clientNodeCmd.AddCommand(clientNodeNtpCmd)
}
```

- [ ] **Step 2: Create get command**

No extra flags. Call `sdkClient.Ntp.NtpGet(ctx, host)`.
Table fields: SYNCHRONIZED, STRATUM, OFFSET, SOURCE, SERVERS.
Format `Servers` as comma-separated string.

- [ ] **Step 3: Create create command**

Flags: `--servers` (required, string slice).
Call `sdkClient.Ntp.NtpCreate(ctx, host, opts)`.
Mutation table output.

- [ ] **Step 4: Create update command**

Flags: `--servers` (required, string slice).
Call `sdkClient.Ntp.NtpUpdate(ctx, host, opts)`.

- [ ] **Step 5: Create delete command**

No extra flags. Call `sdkClient.Ntp.NtpDelete(ctx, host)`.

- [ ] **Step 6: Verify build**

- [ ] **Step 7: Commit**

```bash
git add cmd/client_node_ntp*.go
git commit -m "feat(ntp): add CLI commands"
```

---

## Task 8: NTP Docs + Example + Integration Test

**Files:**
- Create: `examples/sdk/client/ntp.go`
- Create: `test/integration/ntp_test.go`
- Create: `docs/docs/sidebar/features/ntp.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/ntp/ntp.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/ntp/get.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/ntp/create.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/ntp/update.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/ntp/delete.md`
- Create: `docs/docs/sidebar/sdk/client/ntp.md`

- [ ] **Step 1: Create SDK example**

Follow sysctl example pattern: list status, create config.

- [ ] **Step 2: Create integration test**

Follow sysctl integration test pattern: `NtpSmokeSuite` with `TestNtpGet`.

- [ ] **Step 3: Create feature doc**

Follow cron-management.md template.

- [ ] **Step 4: Create CLI doc pages**

Parent with `<DocCardList />`, one page per subcommand.

- [ ] **Step 5: Create SDK doc page**

Follow sysctl SDK doc pattern.

- [ ] **Step 6: Commit**

```bash
git add examples/ test/ docs/
git commit -m "feat(ntp): add docs, SDK example, and integration tests"
```

---

## Task 9: Timezone Provider Interface + Platform Stubs

**Files:**
- Create: `internal/provider/node/timezone/types.go`
- Create: `internal/provider/node/timezone/darwin.go`
- Create: `internal/provider/node/timezone/linux.go`
- Create: `internal/provider/node/timezone/mocks/generate.go`

- [ ] **Step 1: Create types.go**

```go
// Package timezone provides system timezone management via timedatectl.
package timezone

import "context"

// Provider implements the methods to manage the system timezone.
type Provider interface {
	// Get returns the current system timezone.
	Get(ctx context.Context) (*Info, error)
	// Update sets the system timezone. Idempotent.
	Update(ctx context.Context, timezone string) (*UpdateResult, error)
}

// Info represents the current timezone configuration.
type Info struct {
	Timezone  string `json:"timezone"`
	UTCOffset string `json:"utc_offset,omitempty"`
}

// UpdateResult represents the outcome of a timezone update.
type UpdateResult struct {
	Timezone string `json:"timezone"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}
```

- [ ] **Step 2: Create darwin.go and linux.go stubs**

All methods return `fmt.Errorf("timezone: %w", provider.ErrUnsupported)`.

- [ ] **Step 3: Create mocks and generate**

- [ ] **Step 4: Verify and commit**

```bash
git add internal/provider/node/timezone/
git commit -m "feat(timezone): add provider interface and platform stubs"
```

---

## Task 10: Timezone Debian Provider Implementation

**Files:**
- Create: `internal/provider/node/timezone/debian.go`
- Create: `internal/provider/node/timezone/debian_public_test.go`
- Create: `internal/provider/node/timezone/darwin_public_test.go`
- Create: `internal/provider/node/timezone/linux_public_test.go`

- [ ] **Step 1: Write stub tests**

- [ ] **Step 2: Write Debian tests**

**TestGet:**
- success (parse timedatectl output)
- timedatectl error

**TestUpdate:**
- success (different timezone, runs timedatectl set-timezone)
- idempotent (same timezone, Changed: false)
- timedatectl error
- invalid timezone (validate against known list or let timedatectl fail)

- [ ] **Step 3: Implement debian.go**

```go
type Debian struct {
	provider.FactsAware
	logger      *slog.Logger
	execManager exec.Manager
}

func NewDebianProvider(
	logger *slog.Logger,
	execManager exec.Manager,
) *Debian
```

- **Get**: run `timedatectl show -p Timezone --value` for name, `date +%:z` for UTC offset.
- **Update**: read current timezone first (for idempotency), run `timedatectl set-timezone <tz>` if different.

No file management needed — timedatectl handles everything.

- [ ] **Step 4: Verify 100% coverage and commit**

```bash
git add internal/provider/node/timezone/
git commit -m "feat(timezone): implement Debian timedatectl provider with tests"
```

---

## Task 11: Timezone Agent Processor + Wiring

**Files:**
- Create: `internal/agent/processor_timezone.go`
- Create: `internal/agent/processor_timezone_public_test.go`
- Modify: `internal/agent/processor.go`
- Modify: `cmd/agent_setup.go`

- [ ] **Step 1: Write processor tests**

Operations: timezone.get, timezone.update, unsupported, nil provider.

- [ ] **Step 2: Implement processor_timezone.go**

Two sub-operations: get (no data), update (unmarshal `{"timezone": "..."}` from Data).

- [ ] **Step 3: Add to node processor and agent_setup**

Add `timezoneProvider timezone.Provider` to `NewNodeProcessor`. Add `case "timezone":`.

Create `createTimezoneProvider` in agent_setup.go. Needs: logger, execManager.

- [ ] **Step 4: Fix existing tests, verify, commit**

```bash
git add internal/agent/ cmd/agent_setup.go
git commit -m "feat(timezone): add agent processor and wiring"
```

---

## Task 12: Timezone OpenAPI Spec + API Handlers

**Files:**
- Create: `internal/controller/api/node/timezone/gen/...`
- Create: `internal/controller/api/node/timezone/*.go`
- Modify: `cmd/controller_setup.go`

- [ ] **Step 1: Create OpenAPI spec**

Paths:
- `GET /node/{hostname}/timezone` — `GetNodeTimezone`, security: `timezone:read`
- `PUT /node/{hostname}/timezone` — `PutNodeTimezone`, security: `timezone:write`

Schemas:
- `TimezoneUpdateRequest` — required: timezone (string, validate: required)
- `TimezoneEntry` — hostname (req), status (req), timezone, utc_offset, error
- `TimezoneMutationResult` — hostname (req), status (req), timezone, changed, error
- Collection responses

- [ ] **Step 2: Create handlers + handler.go + tests**

Follow NTP handler pattern. Two handlers: get and update.
Include RBAC HTTP tests.

- [ ] **Step 3: Wire in controller_setup.go**

- [ ] **Step 4: Verify and commit**

```bash
git add internal/controller/api/node/timezone/ cmd/controller_setup.go
git commit -m "feat(timezone): add OpenAPI spec, API handlers, and server wiring"
```

---

## Task 13: Timezone SDK Service

**Files:**
- Create: `pkg/sdk/client/timezone.go`
- Create: `pkg/sdk/client/timezone_types.go`
- Create: `pkg/sdk/client/timezone_public_test.go`
- Create: `pkg/sdk/client/timezone_types_public_test.go`
- Modify: `pkg/sdk/client/osapi.go`

- [ ] **Step 1: Create types, service, wire, test**

Two methods: `TimezoneGet`, `TimezoneUpdate`.
`TimezoneUpdateOpts` has one field: `Timezone string`.

- [ ] **Step 2: Verify and commit**

```bash
git add pkg/sdk/client/
git commit -m "feat(timezone): add SDK service with tests"
```

---

## Task 14: Timezone CLI + Docs

**Files:**
- Create: `cmd/client_node_timezone.go`
- Create: `cmd/client_node_timezone_get.go`
- Create: `cmd/client_node_timezone_update.go`
- Create: `examples/sdk/client/timezone.go`
- Create: `test/integration/timezone_test.go`
- Create: `docs/docs/sidebar/features/timezone.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/timezone/*.md`
- Create: `docs/docs/sidebar/sdk/client/timezone.md`

- [ ] **Step 1: Create CLI commands**

Get: table fields TIMEZONE, UTC_OFFSET.
Update: flags `--timezone` (required).

- [ ] **Step 2: Create SDK example, integration test, docs**

- [ ] **Step 3: Verify and commit**

```bash
git add cmd/ examples/ test/ docs/
git commit -m "feat(timezone): add CLI commands, docs, and integration tests"
```

---

## Task 15: Shared Docs + Regeneration + Final Verification

**Files:**
- Modify: `docs/docusaurus.config.ts`
- Modify: `docs/docs/sidebar/usage/configuration.md`
- Modify: `docs/docs/sidebar/features/authentication.md`
- Modify: `docs/docs/sidebar/architecture/api-guidelines.md`
- Modify: `docs/docs/sidebar/architecture/architecture.md`
- Modify: `CLAUDE.md`

- [ ] **Step 1: Update shared docs**

Add NTP and timezone to:
- Features navbar dropdown
- Permissions/roles tables in configuration.md and authentication.md
- Endpoint table in api-guidelines.md
- Feature links in architecture.md
- Provider list in CLAUDE.md

- [ ] **Step 2: Regenerate combined spec**

```bash
just generate
```

- [ ] **Step 3: Full verification**

```bash
go build ./...
just go::unit
just go::unit-cov   # must be >= 99.9%
just go::vet
just go::fmt
```

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "docs(ntp,timezone): update shared docs and regenerate specs"
```
