# Package Management Provider Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> superpowers:subagent-driven-development (recommended) or
> superpowers:executing-plans to implement this plan task-by-task. Steps use
> checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add package management (list, get, install, remove, update sources,
list updates) via apt as a node provider with full API/CLI/SDK support.

**Architecture:** Direct provider at `provider/node/apt/` using `apt-get` and
`dpkg-query` via `exec.Manager`. API path is `/node/{hostname}/package`. One SDK
service (`PackageService`). Permissions: `package:read` (all roles),
`package:write` (admin + write).

**Tech Stack:** Go 1.25, Echo, oapi-codegen (strict-server), gomock,
testify/suite

**Coverage baseline:** 99.9% — must remain at or above this.

---

## Task 1: SDK Constants (Operations + Permissions)

**Files:**

- Modify: `pkg/sdk/client/operations.go`
- Modify: `pkg/sdk/client/permissions.go`
- Modify: `internal/job/types.go`
- Modify: `internal/authtoken/permissions.go`

- [ ] **Step 1: Add package operation constants**

In `pkg/sdk/client/operations.go`:

```go
// Package operations.
const (
	OpPackageList        JobOperation = "node.package.list"
	OpPackageGet         JobOperation = "node.package.get"
	OpPackageInstall     JobOperation = "node.package.install"
	OpPackageRemove      JobOperation = "node.package.remove"
	OpPackageUpdate      JobOperation = "node.package.update"
	OpPackageListUpdates JobOperation = "node.package.listUpdates"
)
```

- [ ] **Step 2: Add permission constants**

```go
	PermPackageRead  Permission = "package:read"
	PermPackageWrite Permission = "package:write"
```

- [ ] **Step 3: Re-export in internal/job/types.go**

- [ ] **Step 4: Re-export permissions in internal/authtoken/permissions.go**

Add to `DefaultRolePermissions`:

- `RoleAdmin`: both
- `RoleWrite`: both
- `RoleRead`: `PermPackageRead` only

- [ ] **Step 5: Verify and commit**

```bash
go build ./...
git commit -m "feat(package): add operation and permission constants"
```

---

## Task 2: Provider Interface + Platform Stubs

**Files:**

- Create: `internal/provider/node/apt/types.go`
- Create: `internal/provider/node/apt/darwin.go`
- Create: `internal/provider/node/apt/linux.go`
- Create: `internal/provider/node/apt/mocks/generate.go`

- [ ] **Step 1: Create types.go**

Package name is `apt`. Provider interface with 6 methods: List, Get, Install,
Remove, Update, ListUpdates.

Data types: Package, Update, Result (see design spec).

- [ ] **Step 2: Create darwin.go and linux.go stubs**

All methods return `fmt.Errorf("package: %w", provider.ErrUnsupported)`.

- [ ] **Step 3: Create mocks and generate**

- [ ] **Step 4: Verify and commit**

```bash
go generate ./internal/provider/node/apt/mocks/...
go build ./...
git commit -m "feat(package): add provider interface and platform stubs"
```

---

## Task 3: Debian Provider Implementation

**Files:**

- Create: `internal/provider/node/apt/debian.go`
- Create: `internal/provider/node/apt/debian_public_test.go`
- Create: `internal/provider/node/apt/darwin_public_test.go`
- Create: `internal/provider/node/apt/linux_public_test.go`

The Debian provider uses `exec.Manager` for all operations.

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

### Operations

- **List**: run
  `dpkg-query -W -f '${Package}\t${Version}\t${binary:Summary}\t${db:Status-Abbrev}\t${Installed-Size}\n'`.
  Parse tab-separated output. Filter lines where status starts with `ii`
  (installed).
- **Get**: same command with package name argument. Error if not installed.
- **Install**: run `apt-get install -y <name>`. Return
  `Result{Name: name, Changed: true}`.
- **Remove**: run `apt-get remove -y <name>`. Return
  `Result{Name: name, Changed: true}`.
- **Update**: run `apt-get update`. Return `Result{Changed: true}`.
- **ListUpdates**: run `apt list --upgradable 2>/dev/null`. Parse output lines
  like `package/source version [upgradable from: oldversion]`.

### Tests

Mock `exec.Manager` with gomock.

**TestList:** success (parse dpkg output), exec error, empty output **TestGet:**
success, not found, exec error **TestInstall:** success, exec error (package not
found) **TestRemove:** success, exec error **TestUpdate:** success, exec error
**TestListUpdates:** success (parse apt list output), no updates, exec error

Stub tests: Darwin and Linux return ErrUnsupported.

Target: 100% coverage.

- [ ] **Step 1: Write stub tests**
- [ ] **Step 2: Write Debian tests**
- [ ] **Step 3: Implement debian.go**
- [ ] **Step 4: Verify 100% coverage and commit**

```bash
go test -coverprofile=/tmp/c.out ./internal/provider/node/apt/...
git commit -m "feat(package): implement Debian apt provider with tests"
```

---

## Task 4: Agent Processor + Wiring

**Files:**

- Create: `internal/agent/processor_package.go`
- Create: `internal/agent/processor_package_public_test.go`
- Modify: `internal/agent/processor.go`
- Modify: `cmd/agent_setup.go`

- [ ] **Step 1: Create processor with tests**

Six sub-operations: package.list, package.get, package.install, package.remove,
package.update, package.listUpdates.

Get/install/remove unmarshal `{"name": "..."}` from Data.
List/update/listUpdates need no data.

- [ ] **Step 2: Add to node processor**

Add `packageProvider apt.Provider` param to `NewNodeProcessor`. Add
`case "package":` dispatch.

- [ ] **Step 3: Wire in agent_setup.go**

```go
func createPackageProvider(
	log *slog.Logger,
	execManager exec.Manager,
) aptProv.Provider
```

Container check: return ErrUnsupported in containers.

- [ ] **Step 4: Fix existing tests and verify**

```bash
go test ./internal/agent/... ./cmd/...
git commit -m "feat(package): add agent processor and wiring"
```

---

## Task 5: OpenAPI Spec + API Handlers

**Files:**

- Create: `internal/controller/api/node/package/gen/api.yaml`
- Create: `internal/controller/api/node/package/gen/cfg.yaml`
- Create: `internal/controller/api/node/package/gen/generate.go`
- Create: `internal/controller/api/node/package/types.go`
- Create: `internal/controller/api/node/package/package.go`
- Create: `internal/controller/api/node/package/validate.go`
- Create: `internal/controller/api/node/package/package_list_get.go`
- Create: `internal/controller/api/node/package/package_get.go`
- Create: `internal/controller/api/node/package/package_install.go`
- Create: `internal/controller/api/node/package/package_remove.go`
- Create: `internal/controller/api/node/package/package_update_post.go`
- Create: `internal/controller/api/node/package/package_update_get.go`
- Create: `internal/controller/api/node/package/handler.go`
- Create: test files for each handler
- Modify: `cmd/controller_setup.go`

Note: the API handler package is `package` which is a Go reserved word. Use
`pkg` as the Go package name: `package pkg` at the top of each file. Import
alias in controller_setup.go: `packageAPI "...api/node/package"`.

Actually — Go does NOT allow `package` as a directory name for a package
declaration. The directory can be named `package` but the Go package must be
something else. Use `package pkgmgmt` or just put it under a different directory
name.

Simplest: name the directory `pkg` to match the Go package name. So:
`internal/controller/api/node/pkg/`. But `pkg` is confusing.

Better: `internal/controller/api/node/package/` with `package packagemgmt` at
the top. Import as `packageAPI "...api/node/package"`.

Wait — Go actually CAN have a directory named `package` with a different package
declaration. The directory name doesn't need to match the Go package name. Use
directory `package/` with `package packageapi` or similar.

The implementer should read how this is handled and choose the cleanest
approach. Follow the pattern: directory name matches the URL path segment
(`package`), Go package name avoids the reserved word.

- [ ] **Step 1: Create OpenAPI spec**

Six paths under `/node/{hostname}/package`:

- `GET /` — list installed, security: `package:read`
- `POST /` — install, security: `package:write`
- `GET /{name}` — get details, security: `package:read`
- `DELETE /{name}` — remove, security: `package:write`
- `POST /update` — refresh sources, security: `package:write`
- `GET /update` — list updates, security: `package:read`

Request schemas:

- `PackageInstallRequest` — name (required, validate required)

Response schemas:

- `PackageEntry` — hostname, status, packages (array of PackageInfo)
- `PackageInfo` — name, version, description, status, size
- `PackageMutationResult` — hostname, status, name, changed, error
- `UpdateEntry` — hostname, status, updates (array of UpdateInfo)
- `UpdateInfo` — name, current_version, new_version

- [ ] **Step 2: Generate code and implement handlers**

List/Get use `JobClient.Query`. Install/Remove/Update use `JobClient.Modify`.
ListUpdates uses `JobClient.Query`.

Category: `"node"`.

- [ ] **Step 3: Create handler.go and write tests with RBAC**

- [ ] **Step 4: Wire in controller_setup.go and verify**

```bash
go build ./...
go test ./internal/controller/api/node/package/... ./cmd/...
git commit -m "feat(package): add OpenAPI spec, API handlers, and server wiring"
```

---

## Task 6: SDK Service

**Files:**

- Create: `pkg/sdk/client/package.go`
- Create: `pkg/sdk/client/package_types.go`
- Create: `pkg/sdk/client/package_public_test.go`
- Create: `pkg/sdk/client/package_types_public_test.go`
- Modify: `pkg/sdk/client/osapi.go`

Note: `package.go` is fine as a filename — Go filenames don't conflict with
reserved words, only package declarations do.

- [ ] **Step 1: Create types**

```go
type PackageInfoResult struct {
	Hostname string        `json:"hostname"`
	Status   string        `json:"status"`
	Packages []PackageInfo `json:"packages,omitempty"`
	Error    string        `json:"error,omitempty"`
}

type PackageInfo struct {
	Name        string `json:"name,omitempty"`
	Version     string `json:"version,omitempty"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
	Size        int64  `json:"size,omitempty"`
}

type PackageMutationResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Name     string `json:"name,omitempty"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

type PackageUpdateResult struct {
	Hostname string       `json:"hostname"`
	Status   string       `json:"status"`
	Updates  []UpdateInfo `json:"updates,omitempty"`
	Error    string       `json:"error,omitempty"`
}

type UpdateInfo struct {
	Name           string `json:"name,omitempty"`
	CurrentVersion string `json:"current_version,omitempty"`
	NewVersion     string `json:"new_version,omitempty"`
}
```

- [ ] **Step 2: Create service**

```go
type PackageService struct {
	client *gen.ClientWithResponses
}
```

Methods (clean verbs):

- `List(ctx, hostname)`
- `Get(ctx, hostname, name)`
- `Install(ctx, hostname, name)`
- `Remove(ctx, hostname, name)`
- `Update(ctx, hostname)`
- `ListUpdates(ctx, hostname)`

Wire as `Package *PackageService` in Client.

Run `just generate` first.

- [ ] **Step 3: Write tests and verify**

```bash
go test ./pkg/sdk/client/...
git commit -m "feat(package): add SDK service with tests"
```

---

## Task 7: CLI Commands

**Files:**

- Create: `cmd/client_node_package.go` — parent
- Create: `cmd/client_node_package_list.go`
- Create: `cmd/client_node_package_get.go`
- Create: `cmd/client_node_package_install.go`
- Create: `cmd/client_node_package_remove.go`
- Create: `cmd/client_node_package_update.go`
- Create: `cmd/client_node_package_updates.go`

- [ ] **Step 1: Create parent + list/get commands**

List: table fields NAME, VERSION, STATUS, SIZE. Format SIZE with
`cli.FormatBytes`.

Get: flag `--name` (required). Same table.

- [ ] **Step 2: Create install/remove commands**

Install: flag `--name` (required). Mutation table. Remove: flag `--name`
(required). Mutation table.

- [ ] **Step 3: Create update + list-updates commands**

Update (refresh sources): no extra flags. Mutation table. List updates: table
fields NAME, CURRENT, NEW.

The CLI command for listing updates could be `updates` as a subcommand:
`osapi client node package updates`.

- [ ] **Step 4: Verify and commit**

```bash
go build ./...
git commit -m "feat(package): add CLI commands"
```

---

## Task 8: Docs + Example + Integration Test

**Files:**

- Create: `examples/sdk/client/package.go`
- Create: `test/integration/package_test.go`
- Create: `docs/docs/sidebar/features/package-management.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/package/package.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/package/list.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/package/get.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/package/install.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/package/remove.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/package/update.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/package/updates.md`
- Create: `docs/docs/sidebar/sdk/client/system-config/package.md`
- Modify: shared docs (features table, auth, config, api guidelines,
  architecture, client.md, docusaurus.config.ts)

SDK doc goes under `system-config/` category — package management is system
configuration alongside sysctl, ntp, timezone.

- [ ] **Step 1: Create SDK example**

Demonstrate List and Get. Don't demonstrate Install in example (destructive).

- [ ] **Step 2: Create integration test**

`PackageSmokeSuite` with `TestPackageList` (read-only). Guard install/remove
with `skipWrite`.

- [ ] **Step 3: Create feature page + CLI docs + SDK doc**

- [ ] **Step 4: Update all shared docs**

Features table, auth permissions, config roles, API guidelines (6 endpoints),
architecture feature link, client.md system-config table, docusaurus dropdowns
(Features + SDK).

- [ ] **Step 5: Regenerate and verify**

```bash
just generate
go build ./...
just go::unit
just go::unit-cov  # >= 99.9%
just go::vet
```

- [ ] **Step 6: Commit**

```bash
git commit -m "feat(package): add docs, SDK example, and integration tests"
```
