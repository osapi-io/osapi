# Process Management Provider Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add process listing, details, and signal sending as a node provider with full API/CLI/SDK support.

**Architecture:** Direct provider at `provider/node/process/` using `gopsutil/process` for reading process info and `syscall.Kill` for signaling. Three endpoints: list all, get by PID, signal by PID. Integrates into the node processor. Permission: `process:read` (all roles) and `process:execute` (admin only).

**Tech Stack:** Go 1.25, gopsutil/v4, Echo, oapi-codegen (strict-server), gomock, testify/suite

**Coverage baseline:** 99.9% — must remain at or above this.

---

## Task 1: SDK Constants (Operations + Permissions)

**Files:**
- Modify: `pkg/sdk/client/operations.go`
- Modify: `pkg/sdk/client/permissions.go`
- Modify: `internal/job/types.go`
- Modify: `internal/authtoken/permissions.go`

- [ ] **Step 1: Add process operation constants**

In `pkg/sdk/client/operations.go`:

```go
// Process operations.
const (
	OpProcessList   JobOperation = "node.process.list"
	OpProcessGet    JobOperation = "node.process.get"
	OpProcessSignal JobOperation = "node.process.signal"
)
```

- [ ] **Step 2: Add permission constants**

In `pkg/sdk/client/permissions.go`:

```go
	PermProcessRead    Permission = "process:read"
	PermProcessExecute Permission = "process:execute"
```

- [ ] **Step 3: Re-export in internal/job/types.go**

```go
// Process operations.
const (
	OperationProcessList   = client.OpProcessList
	OperationProcessGet    = client.OpProcessGet
	OperationProcessSignal = client.OpProcessSignal
)
```

- [ ] **Step 4: Re-export permissions in internal/authtoken/permissions.go**

Add constants. Add to `AllPermissions`. Add to `DefaultRolePermissions`:
- `RoleAdmin`: add `PermProcessRead` and `PermProcessExecute`
- `RoleWrite`: add `PermProcessRead` only
- `RoleRead`: add `PermProcessRead` only

- [ ] **Step 5: Verify and commit**

```bash
go build ./...
git commit -m "feat(process): add operation and permission constants"
```

---

## Task 2: Provider Interface + Platform Stubs

**Files:**
- Create: `internal/provider/node/process/types.go`
- Create: `internal/provider/node/process/darwin.go`
- Create: `internal/provider/node/process/linux.go`
- Create: `internal/provider/node/process/mocks/generate.go`

- [ ] **Step 1: Create types.go**

```go
package process

import "context"

// Provider implements process management operations.
type Provider interface {
	List(ctx context.Context) ([]Info, error)
	Get(ctx context.Context, pid int) (*Info, error)
	Signal(ctx context.Context, pid int, signal string) (*SignalResult, error)
}

type Info struct {
	PID        int     `json:"pid"`
	Name       string  `json:"name"`
	User       string  `json:"user"`
	State      string  `json:"state"`
	CPUPercent float64 `json:"cpu_percent"`
	MemPercent float32 `json:"mem_percent"`
	MemRSS     int64   `json:"mem_rss"`
	Command    string  `json:"command"`
	StartTime  string  `json:"start_time"`
}

type SignalResult struct {
	PID     int    `json:"pid"`
	Signal  string `json:"signal"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}
```

- [ ] **Step 2: Create darwin.go and linux.go stubs**

All methods return `fmt.Errorf("process: %w", provider.ErrUnsupported)`.

- [ ] **Step 3: Create mocks and generate**

- [ ] **Step 4: Verify and commit**

```bash
go generate ./internal/provider/node/process/mocks/...
go build ./...
git commit -m "feat(process): add provider interface and platform stubs"
```

---

## Task 3: Debian Provider Implementation

**Files:**
- Create: `internal/provider/node/process/debian.go`
- Create: `internal/provider/node/process/debian_public_test.go`
- Create: `internal/provider/node/process/darwin_public_test.go`
- Create: `internal/provider/node/process/linux_public_test.go`

The Debian provider uses `gopsutil/v4/process` for reading process
info and `syscall.Kill` for sending signals.

- [ ] **Step 1: Write stub tests**

Verify Darwin and Linux return `ErrUnsupported` for all methods.

- [ ] **Step 2: Write Debian tests**

The provider wraps gopsutil — for testability, the provider should
accept an interface that wraps gopsutil calls so it can be mocked.
Alternatively, use the export_test.go pattern to swap the gopsutil
functions.

**TestList:**
- success (returns process list)
- gopsutil error

**TestGet:**
- success (PID exists)
- PID not found
- gopsutil error

**TestSignal:**
- success with TERM
- success with KILL
- invalid signal name
- PID not found (kill returns ESRCH)
- permission denied (kill returns EPERM)

- [ ] **Step 3: Implement debian.go**

```go
var (
	_ Provider             = (*Debian)(nil)
	_ provider.FactsSetter = (*Debian)(nil)
)

type Debian struct {
	provider.FactsAware
	logger *slog.Logger
}

func NewDebianProvider(
	logger *slog.Logger,
) *Debian
```

Key implementation:
- **List**: call `process.Processes()`, for each PID gather info
  with `p.Name()`, `p.Username()`, `p.Status()`,
  `p.CPUPercent()`, `p.MemoryPercent()`, `p.MemoryInfo()` (for
  RSS), `p.Cmdline()`, `p.CreateTime()`. Skip processes that
  error (permission denied on some PIDs is normal).
- **Get**: `process.NewProcess(int32(pid))`, gather same info.
  Return error if PID doesn't exist.
- **Signal**: validate signal name against allowed map. Use
  `syscall.Kill(pid, sig)`. Handle `ESRCH` (no such process) and
  `EPERM` (permission denied) errors.

Allowed signals map:
```go
var allowedSignals = map[string]syscall.Signal{
	"TERM": syscall.SIGTERM,
	"KILL": syscall.SIGKILL,
	"HUP":  syscall.SIGHUP,
	"INT":  syscall.SIGINT,
	"USR1": syscall.SIGUSR1,
	"USR2": syscall.SIGUSR2,
}
```

- [ ] **Step 4: Verify 100% coverage and commit**

```bash
go test -coverprofile=/tmp/c.out ./internal/provider/node/process/...
git commit -m "feat(process): implement Debian process provider with tests"
```

---

## Task 4: Agent Processor + Wiring

**Files:**
- Create: `internal/agent/processor_process.go`
- Create: `internal/agent/processor_process_public_test.go`
- Modify: `internal/agent/processor.go`
- Modify: `cmd/agent_setup.go`

- [ ] **Step 1: Create processor with tests**

Three sub-operations:
- `process.list` — no data, call `provider.List(ctx)`
- `process.get` — unmarshal `{"pid": 1234}`, call `provider.Get(ctx, pid)`
- `process.signal` — unmarshal `{"pid": 1234, "signal": "TERM"}`, call `provider.Signal(ctx, pid, signal)`

- [ ] **Step 2: Add to node processor**

Add `processProvider process.Provider` parameter to
`NewNodeProcessor`. Add `case "process":` dispatch.

- [ ] **Step 3: Wire in agent_setup.go**

Create `createProcessProvider` function. Add container check.
Process provider only needs `logger` (no exec manager, no fs).

```go
func createProcessProvider(
	log *slog.Logger,
) processProv.Provider {
	plat := platform.Detect()
	switch plat {
	case "debian":
		if platform.IsContainer() {
			log.Info("running in container, process operations disabled")
			return processProv.NewLinuxProvider()
		}
		return processProv.NewDebianProvider(log)
	case "darwin":
		return processProv.NewDarwinProvider()
	default:
		return processProv.NewLinuxProvider()
	}
}
```

- [ ] **Step 4: Fix existing tests and verify**

```bash
go test ./internal/agent/... ./cmd/...
git commit -m "feat(process): add agent processor and wiring"
```

---

## Task 5: OpenAPI Spec + API Handlers

**Files:**
- Create: `internal/controller/api/node/process/gen/api.yaml`
- Create: `internal/controller/api/node/process/gen/cfg.yaml`
- Create: `internal/controller/api/node/process/gen/generate.go`
- Create: `internal/controller/api/node/process/types.go`
- Create: `internal/controller/api/node/process/process.go`
- Create: `internal/controller/api/node/process/validate.go`
- Create: `internal/controller/api/node/process/process_list_get.go`
- Create: `internal/controller/api/node/process/process_get.go`
- Create: `internal/controller/api/node/process/process_signal_post.go`
- Create: `internal/controller/api/node/process/handler.go`
- Create: test files for each handler
- Modify: `cmd/controller_setup.go`

- [ ] **Step 1: Create OpenAPI spec**

Three paths:
- `GET /node/{hostname}/process` — `GetNodeProcess`, security:
  `process:read`. Response: `ProcessCollectionResponse`
- `GET /node/{hostname}/process/{pid}` — `GetNodeProcessByPid`,
  security: `process:read`. PID is integer path param. Response:
  `ProcessGetResponse`
- `POST /node/{hostname}/process/{pid}/signal` —
  `PostNodeProcessSignal`, security: `process:execute`. Request
  body: `ProcessSignalRequest` with signal (required, enum of
  TERM/KILL/HUP/INT/USR1/USR2, validate `required,oneof=...`).
  Response: `ProcessSignalResponse`

Schemas:
- `ProcessEntry` — hostname (req), status (req, enum
  ok/failed/skipped), processes (array of ProcessInfo), error
- `ProcessInfo` — pid, name, user, state, cpu_percent,
  mem_percent, mem_rss, command, start_time
- `ProcessSignalResult` — hostname (req), status (req), pid,
  signal, changed, error

- [ ] **Step 2: Generate code and create handlers**

- [ ] **Step 3: Implement all handlers with broadcast support**

Category `"node"`. Use `JobClient.Query` for list/get (reads),
`JobClient.Modify` for signal (state change).

- [ ] **Step 4: Create handler.go (self-registration)**

- [ ] **Step 5: Write tests with RBAC**

- [ ] **Step 6: Wire in controller_setup.go and verify**

```bash
go build ./...
go test ./internal/controller/api/node/process/... ./cmd/...
git commit -m "feat(process): add OpenAPI spec, API handlers, and server wiring"
```

---

## Task 6: SDK Service

**Files:**
- Create: `pkg/sdk/client/process.go`
- Create: `pkg/sdk/client/process_types.go`
- Create: `pkg/sdk/client/process_public_test.go`
- Create: `pkg/sdk/client/process_types_public_test.go`
- Modify: `pkg/sdk/client/osapi.go`

- [ ] **Step 1: Create types**

```go
type ProcessInfoResult struct {
	Hostname  string        `json:"hostname"`
	Status    string        `json:"status"`
	Processes []ProcessInfo `json:"processes,omitempty"`
	Error     string        `json:"error,omitempty"`
}

type ProcessInfo struct {
	PID        int     `json:"pid"`
	Name       string  `json:"name,omitempty"`
	User       string  `json:"user,omitempty"`
	State      string  `json:"state,omitempty"`
	CPUPercent float64 `json:"cpu_percent,omitempty"`
	MemPercent float32 `json:"mem_percent,omitempty"`
	MemRSS     int64   `json:"mem_rss,omitempty"`
	Command    string  `json:"command,omitempty"`
	StartTime  string  `json:"start_time,omitempty"`
}

type ProcessSignalResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	PID      int    `json:"pid,omitempty"`
	Signal   string `json:"signal,omitempty"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

type ProcessSignalOpts struct {
	Signal string
}
```

- [ ] **Step 2: Create service**

```go
type ProcessService struct {
	client *gen.ClientWithResponses
}
```

Methods: `List(ctx, hostname)`, `Get(ctx, hostname, pid)`,
`Signal(ctx, hostname, pid, opts)`.

- [ ] **Step 3: Wire into osapi.go, run `just generate`, write tests**

- [ ] **Step 4: Verify and commit**

```bash
go test ./pkg/sdk/client/...
git commit -m "feat(process): add SDK service with tests"
```

---

## Task 7: CLI Commands

**Files:**
- Create: `cmd/client_node_process.go`
- Create: `cmd/client_node_process_list.go`
- Create: `cmd/client_node_process_get.go`
- Create: `cmd/client_node_process_signal.go`

- [ ] **Step 1: Create parent command**

- [ ] **Step 2: Create list command**

No extra flags. Table fields: PID, NAME, USER, STATE, CPU%, MEM%,
COMMAND. Use `cli.BuildBroadcastTable` + `cli.PrintCompactTable`.
Format CPU% and MEM% with `fmt.Sprintf("%.1f%%", val)`.

- [ ] **Step 3: Create get command**

Flag: `--pid` (int, required). Same table output as list.

- [ ] **Step 4: Create signal command**

Flags: `--pid` (int, required), `--signal` (string, required).
Mutation table output with PID and SIGNAL fields.

- [ ] **Step 5: Verify and commit**

```bash
go build ./...
git commit -m "feat(process): add CLI commands"
```

---

## Task 8: Docs + Example + Integration Test

**Files:**
- Create: `examples/sdk/client/process.go`
- Create: `test/integration/process_test.go`
- Create: `docs/docs/sidebar/features/process-management.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/process/process.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/process/list.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/process/get.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/process/signal.md`
- Create: `docs/docs/sidebar/sdk/client/management/process.md`
- Modify: `docs/docs/sidebar/features/features.md`
- Modify: `docs/docs/sidebar/features/authentication.md`
- Modify: `docs/docs/sidebar/usage/configuration.md`
- Modify: `docs/docs/sidebar/architecture/api-guidelines.md`
- Modify: `docs/docs/sidebar/architecture/architecture.md`
- Modify: `docs/docs/sidebar/sdk/client/client.md`
- Modify: `docs/docusaurus.config.ts` (Features + SDK dropdowns)

SDK doc goes under `management/` category (same as Agent, Job,
Health, Audit — process management is an operational concern).

- [ ] **Step 1: Create SDK example**

Demonstrate `client.Process.List()` and `client.Process.Get()`.
Don't demonstrate Signal in the example (destructive).

- [ ] **Step 2: Create integration test**

`ProcessSmokeSuite` with `TestProcessList`. Guard signal test
with `skipWrite`.

- [ ] **Step 3: Create feature page**

`process-management.md` — list, get, signal operations. CLI
examples. Permissions. Platforms.

- [ ] **Step 4: Create CLI doc pages**

Directory with landing page + list.md, get.md, signal.md.

- [ ] **Step 5: Create SDK doc page**

Under `management/`. Title: `# Process`. Methods: List, Get,
Signal. Add to `client.md` Management table.

- [ ] **Step 6: Update all shared docs**

Features table, authentication permissions, configuration roles,
API guidelines endpoints, architecture feature link, docusaurus
dropdowns (Features + SDK under Management group).

- [ ] **Step 7: Regenerate and verify**

```bash
just generate
go build ./...
just go::unit
just go::unit-cov  # >= 99.9%
just go::vet
```

- [ ] **Step 8: Commit**

```bash
git commit -m "feat(process): add docs, SDK example, and integration tests"
```
