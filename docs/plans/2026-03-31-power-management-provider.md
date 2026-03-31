# Power Management Provider Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add power management (reboot/shutdown) as action operations with a minimum 5-second implicit delay for agent lifecycle completion.

**Architecture:** Direct provider at `provider/node/power/` with two action methods (Reboot, Shutdown). Integrates into the node processor. API has two POST endpoints under `/node/{hostname}/power/`. SDK exposes `client.Power.Reboot()` and `client.Power.Shutdown()`. Permission is `power:execute` (admin only).

**Tech Stack:** Go 1.25, Echo, oapi-codegen (strict-server), gomock, testify/suite

**Coverage baseline:** 99.9% — must remain at or above this.

---

## Task 1: SDK Constants (Operations + Permissions)

**Files:**
- Modify: `pkg/sdk/client/operations.go`
- Modify: `pkg/sdk/client/permissions.go`
- Modify: `internal/job/types.go`
- Modify: `internal/authtoken/permissions.go`

- [ ] **Step 1: Add power operation constants**

In `pkg/sdk/client/operations.go`:

```go
// Power operations.
const (
	OpPowerReboot   JobOperation = "node.power.reboot"
	OpPowerShutdown JobOperation = "node.power.shutdown"
)
```

- [ ] **Step 2: Add permission constant**

In `pkg/sdk/client/permissions.go`:

```go
	PermPowerExecute Permission = "power:execute"
```

- [ ] **Step 3: Re-export in internal/job/types.go**

```go
// Power operations.
const (
	OperationPowerReboot   = client.OpPowerReboot
	OperationPowerShutdown = client.OpPowerShutdown
)
```

- [ ] **Step 4: Re-export permission in internal/authtoken/permissions.go**

Add constant, add to `AllPermissions`, add to `DefaultRolePermissions` for `RoleAdmin` ONLY (not write or read — power is destructive).

- [ ] **Step 5: Verify and commit**

```bash
go build ./...
git commit -m "feat(power): add operation and permission constants"
```

---

## Task 2: Provider Interface + Platform Stubs

**Files:**
- Create: `internal/provider/node/power/types.go`
- Create: `internal/provider/node/power/darwin.go`
- Create: `internal/provider/node/power/linux.go`
- Create: `internal/provider/node/power/mocks/generate.go`

- [ ] **Step 1: Create types.go**

```go
package power

import "context"

// Provider implements power management operations.
type Provider interface {
	Reboot(ctx context.Context, opts Opts) (*Result, error)
	Shutdown(ctx context.Context, opts Opts) (*Result, error)
}

// Opts contains optional parameters for power operations.
type Opts struct {
	Delay   int    `json:"delay,omitempty"`
	Message string `json:"message,omitempty"`
}

// Result represents the outcome of a power operation.
type Result struct {
	Action  string `json:"action"`
	Delay   int    `json:"delay"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}
```

- [ ] **Step 2: Create darwin.go and linux.go stubs**

Both return `fmt.Errorf("power: %w", provider.ErrUnsupported)`.

- [ ] **Step 3: Create mocks and generate**

- [ ] **Step 4: Verify and commit**

```bash
go generate ./internal/provider/node/power/mocks/...
go build ./...
git commit -m "feat(power): add provider interface and platform stubs"
```

---

## Task 3: Debian Provider Implementation

**Files:**
- Create: `internal/provider/node/power/debian.go`
- Create: `internal/provider/node/power/debian_public_test.go`
- Create: `internal/provider/node/power/darwin_public_test.go`
- Create: `internal/provider/node/power/linux_public_test.go`

The Debian provider:
- Enforces minimum 5-second delay: `actualDelay := max(userDelay, 5)`
- Runs shutdown command in background so the provider returns before
  the system goes down
- Reboot: `shutdown -r +N` or `sleep N && shutdown -r now &`
- Shutdown: `shutdown -h +N` or `sleep N && shutdown -h now &`
- Logs the message before executing if provided

- [ ] **Step 1: Write stub tests (Darwin, Linux)**

Verify all methods return `ErrUnsupported`.

- [ ] **Step 2: Write Debian tests**

Test cases for Reboot and Shutdown:
- success with default delay (5 seconds)
- success with user delay > 5
- success with user delay < 5 (clamped to 5)
- success with message
- exec error
- verify Changed is always true on success

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

- [ ] **Step 4: Verify 100% coverage and commit**

```bash
go test -coverprofile=/tmp/c.out ./internal/provider/node/power/...
git commit -m "feat(power): implement Debian power provider with tests"
```

---

## Task 4: Agent Processor + Wiring

**Files:**
- Create: `internal/agent/processor_power.go`
- Create: `internal/agent/processor_power_public_test.go`
- Modify: `internal/agent/processor.go`
- Modify: `cmd/agent_setup.go`

- [ ] **Step 1: Create processor with tests**

Two sub-operations: `power.reboot` and `power.shutdown`. Both
unmarshal `power.Opts` from `jobRequest.Data` (data may be nil
for default opts).

- [ ] **Step 2: Add to node processor**

Add `powerProvider power.Provider` parameter to
`NewNodeProcessor`. Add `case "power":` dispatch.

- [ ] **Step 3: Wire in agent_setup.go**

Create `createPowerProvider` function. Add to `NewNodeProcessor`
call and `registry.Register` providers list.

- [ ] **Step 4: Fix existing tests and verify**

```bash
go test ./internal/agent/... ./cmd/...
git commit -m "feat(power): add agent processor and wiring"
```

---

## Task 5: OpenAPI Spec + API Handlers

**Files:**
- Create: `internal/controller/api/node/power/gen/api.yaml`
- Create: `internal/controller/api/node/power/gen/cfg.yaml`
- Create: `internal/controller/api/node/power/gen/generate.go`
- Create: `internal/controller/api/node/power/types.go`
- Create: `internal/controller/api/node/power/power.go`
- Create: `internal/controller/api/node/power/validate.go`
- Create: `internal/controller/api/node/power/reboot_post.go`
- Create: `internal/controller/api/node/power/shutdown_post.go`
- Create: `internal/controller/api/node/power/handler.go`
- Create: test files for each handler
- Modify: `cmd/controller_setup.go`

- [ ] **Step 1: Create OpenAPI spec**

Two POST paths:
- `POST /node/{hostname}/power/reboot` — `PostNodePowerReboot`,
  security: `power:execute`
- `POST /node/{hostname}/power/shutdown` — `PostNodePowerShutdown`,
  security: `power:execute`

Request body (shared, optional):
- `PowerRequest` — delay (integer, min 0), message (string)

Response schemas:
- `PowerResult` — hostname (req), status (req, enum
  ok/failed/skipped), action, delay, changed, error
- `PowerRebootResponse` / `PowerShutdownResponse` — job_id +
  results array

- [ ] **Step 2: Generate code and create handler struct**

- [ ] **Step 3: Implement both handlers with broadcast support**

Category `"node"`, operations `job.OperationPowerReboot` and
`job.OperationPowerShutdown`. Use `JobClient.Modify` (these are
state-changing actions).

- [ ] **Step 4: Create handler.go (self-registration)**

- [ ] **Step 5: Write tests with RBAC for both handlers**

- [ ] **Step 6: Wire in controller_setup.go and verify**

```bash
go build ./...
go test ./internal/controller/api/node/power/... ./cmd/...
git commit -m "feat(power): add OpenAPI spec, API handlers, and server wiring"
```

---

## Task 6: SDK Service

**Files:**
- Create: `pkg/sdk/client/power.go`
- Create: `pkg/sdk/client/power_types.go`
- Create: `pkg/sdk/client/power_public_test.go`
- Create: `pkg/sdk/client/power_types_public_test.go`
- Modify: `pkg/sdk/client/osapi.go`

- [ ] **Step 1: Create types**

```go
type PowerResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Action   string `json:"action,omitempty"`
	Delay    int    `json:"delay,omitempty"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

type PowerOpts struct {
	Delay   int
	Message string
}
```

- [ ] **Step 2: Create service**

```go
type PowerService struct {
	client *gen.ClientWithResponses
}
```

Methods: `Reboot(ctx, hostname, opts)` and
`Shutdown(ctx, hostname, opts)`. Both return
`*Response[Collection[PowerResult]]`.

- [ ] **Step 3: Wire into osapi.go, write tests**

Run `just generate` first to get the combined spec updated.

- [ ] **Step 4: Verify and commit**

```bash
go test ./pkg/sdk/client/...
git commit -m "feat(power): add SDK service with tests"
```

---

## Task 7: CLI Commands

**Files:**
- Create: `cmd/client_node_power.go`
- Create: `cmd/client_node_power_reboot.go`
- Create: `cmd/client_node_power_shutdown.go`

- [ ] **Step 1: Create parent command**

```go
var clientNodePowerCmd = &cobra.Command{
	Use:   "power",
	Short: "Manage power state",
}

func init() {
	clientNodeCmd.AddCommand(clientNodePowerCmd)
}
```

- [ ] **Step 2: Create reboot command**

Flags: `--delay` (int, optional), `--message` (string, optional).
Call `sdkClient.Power.Reboot(ctx, host, opts)`.
Mutation table output with ACTION and DELAY columns.

- [ ] **Step 3: Create shutdown command**

Same flags. Call `sdkClient.Power.Shutdown(ctx, host, opts)`.

- [ ] **Step 4: Verify and commit**

```bash
go build ./...
git commit -m "feat(power): add CLI commands"
```

---

## Task 8: Docs + Example + Integration Test

**Files:**
- Create: `examples/sdk/client/power.go`
- Create: `test/integration/power_test.go`
- Create: `docs/docs/sidebar/features/power-management.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/power/power.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/power/reboot.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/power/shutdown.md`
- Create: `docs/docs/sidebar/sdk/client/power.md`
- Modify: `docs/docs/sidebar/features/features.md`
- Modify: `docs/docs/sidebar/features/authentication.md`
- Modify: `docs/docs/sidebar/usage/configuration.md`
- Modify: `docs/docs/sidebar/architecture/api-guidelines.md`
- Modify: `docs/docs/sidebar/architecture/architecture.md`
- Modify: `docs/docusaurus.config.ts` (Features + SDK dropdowns)

- [ ] **Step 1: Create SDK example**

Demonstrate `client.Power.Reboot()` with delay and message.

- [ ] **Step 2: Create integration test**

`PowerSmokeSuite` — read-only test only (don't actually reboot
in CI). Test that the endpoint responds (even if skipped on
macOS).

- [ ] **Step 3: Create feature page**

`power-management.md` — what it does, how the delay works,
CLI examples, permissions (admin only), platforms.

- [ ] **Step 4: Create CLI doc pages**

Directory with landing page + reboot.md + shutdown.md.

- [ ] **Step 5: Create SDK doc page**

Title: `# Power`. Methods: `Reboot`, `Shutdown`.

- [ ] **Step 6: Update all shared docs**

Features table, authentication permissions, configuration
roles table, API guidelines endpoints, architecture feature
link, docusaurus dropdowns (Features + SDK).

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
git commit -m "feat(power): add docs, SDK example, and integration tests"
```
