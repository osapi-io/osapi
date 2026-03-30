# API Directory Restructure Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> superpowers:subagent-driven-development (recommended) or
> superpowers:executing-plans to implement this plan task-by-task. Steps use
> checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move all node-targeted API handler packages under `api/node/` to
mirror the URL path structure, splitting the monolithic `api/node/` package into
domain-specific sub-packages.

**Architecture:** Each node-targeted domain gets its own sub-package under
`api/node/` with its own `gen/` directory and OpenAPI spec. Domains that already
have their own packages (sysctl, schedule, docker) are moved. Domains currently
embedded in the monolithic `api/node/` package (hostname, network, command,
file) are split out. Read-only single-GET endpoints (disk, memory, load, status,
uptime, os) remain flat in `api/node/`. Handler shims stay in `api/` with names
updated to match the nested path.

**Tech Stack:** Go 1.25, Echo, oapi-codegen (strict-server), redocly join

**Coverage baseline:** 99.9% — must remain at or above this.

---

## File Map

### Packages That Move (directory rename + import update)

```
api/sysctl/     → api/node/sysctl/
api/schedule/   → api/node/schedule/
api/docker/     → api/node/docker/
```

### Packages That Split Out of api/node/

```
api/node/hostname/    ← new (from node_hostname_get.go, node_hostname_put.go)
api/node/network/     ← new (from network_dns_*.go, network_ping_*.go)
api/node/command/     ← new (from command_exec_post.go, command_shell_post.go)
api/node/file/        ← new (from file_deploy_post.go, file_undeploy_post.go, file_status_post.go)
```

### Files Remaining in api/node/ (read-only, flat)

```
api/node/
  gen/                ← slimmed OpenAPI spec (disk, mem, load, status, uptime, os only)
  node.go             ← factory (slimmed — only read-only handlers)
  types.go            ← shared types (slimmed)
  validate.go         ← validateHostname (kept — each sub-pkg also has its own copy)
  node_disk_get.go
  node_memory_get.go
  node_load_get.go
  node_status_get.go
  node_uptime_get.go
  node_os_get.go
  export_test.go
  *_public_test.go    ← tests for the above
```

### Handler Shim Renames in api/

```
handler_node.go      → handler_node.go (slimmed — read-only only)
                     + handler_node_hostname.go (new)
                     + handler_node_network.go (new)
                     + handler_node_command.go (new)
                     + handler_node_file.go (new)
handler_sysctl.go    → handler_node_sysctl.go (import path change)
handler_schedule.go  → handler_node_schedule.go (import path change)
handler_docker.go    → handler_node_docker.go (import path change)
```

### Unchanged Packages

```
api/agent/    api/audit/    api/common/
api/facts/    api/file/     api/health/
api/job/      api/gen/
```

---

## Task Order

The safest approach is to do each domain independently so the project compiles
and tests pass after every task. Order:

1. Move sysctl (already its own package — simplest)
2. Move schedule (already its own package)
3. Move docker (already its own package)
4. Split hostname out of node/
5. Split network out of node/
6. Split command out of node/
7. Split file out of node/
8. Slim down remaining node/ package
9. Regenerate combined spec
10. Update CLAUDE.md
11. Final verification

---

## Task 1: Move sysctl under node/

**Files:**

- Move: `internal/controller/api/sysctl/` →
  `internal/controller/api/node/sysctl/`
- Rename: `internal/controller/api/handler_sysctl.go` →
  `internal/controller/api/handler_node_sysctl.go`
- Modify: all files that import `api/sysctl` or `api/sysctl/gen`
- Modify: `internal/controller/api/handler_public_test.go`

- [ ] **Step 1: Move the directory**

```bash
git mv internal/controller/api/sysctl internal/controller/api/node/sysctl
```

- [ ] **Step 2: Rename the handler shim**

```bash
git mv internal/controller/api/handler_sysctl.go internal/controller/api/handler_node_sysctl.go
```

- [ ] **Step 3: Update import paths**

Search the entire codebase for the old import path and update:

```bash
# Find all files importing the old path
grep -r "controller/api/sysctl" --include="*.go" .
```

Update every occurrence:

- `github.com/retr0h/osapi/internal/controller/api/sysctl` →
  `github.com/retr0h/osapi/internal/controller/api/node/sysctl`
- `github.com/retr0h/osapi/internal/controller/api/sysctl/gen` →
  `github.com/retr0h/osapi/internal/controller/api/node/sysctl/gen`

Files that will need import updates:

- `internal/controller/api/handler_node_sysctl.go`
- `internal/controller/api/handler_public_test.go`
- `cmd/controller_setup.go`

- [ ] **Step 4: Update the import-mapping in sysctl's cfg.yaml**

The `cfg.yaml` has a relative import-mapping for the common spec. After moving
one level deeper, the relative path changes:

Read `internal/controller/api/node/sysctl/gen/cfg.yaml` and update:

```yaml
import-mapping:
  ../../../common/gen/api.yaml: github.com/retr0h/osapi/internal/controller/api/common/gen
```

(Was `../../common/gen/api.yaml` — now one level deeper.)

Also update the `$ref` paths in `api.yaml` that reference
`../../common/gen/api.yaml` — they become `../../../common/gen/api.yaml`.

- [ ] **Step 5: Rename the handler method**

In `handler_node_sysctl.go`, rename `GetSysctlHandler` to
`GetNodeSysctlHandler`.

Update the call site in `cmd/controller_setup.go`:

```go
// Before:
handlers = append(handlers, sm.GetSysctlHandler(jc)...)
// After:
handlers = append(handlers, sm.GetNodeSysctlHandler(jc)...)
```

Update the interface in `cmd/controller_setup.go` if `GetSysctlHandler` is
declared there.

Update the test in `handler_public_test.go`:

- Rename `TestGetSysctlHandler` → `TestGetNodeSysctlHandler`
- Update the method call inside the test

- [ ] **Step 6: Regenerate the sysctl spec**

```bash
go generate ./internal/controller/api/node/sysctl/gen/...
```

- [ ] **Step 7: Verify build and tests**

```bash
go build ./...
go test ./internal/controller/api/... ./cmd/...
```

- [ ] **Step 8: Commit**

```bash
git add -A
git commit -m "refactor(api): move sysctl handlers under node/"
```

---

## Task 2: Move schedule under node/

**Files:**

- Move: `internal/controller/api/schedule/` →
  `internal/controller/api/node/schedule/`
- Rename: `internal/controller/api/handler_schedule.go` →
  `internal/controller/api/handler_node_schedule.go`

Follow the exact same steps as Task 1:

- [ ] **Step 1: Move the directory**

```bash
git mv internal/controller/api/schedule internal/controller/api/node/schedule
```

- [ ] **Step 2: Rename the handler shim**

```bash
git mv internal/controller/api/handler_schedule.go internal/controller/api/handler_node_schedule.go
```

- [ ] **Step 3: Update import paths**

Search for `controller/api/schedule` and update to
`controller/api/node/schedule` in all `.go` files. Files that will need updates:

- `internal/controller/api/handler_node_schedule.go`
- `internal/controller/api/handler_public_test.go`
- `cmd/controller_setup.go`

- [ ] **Step 4: Update cfg.yaml and api.yaml relative paths**

In `internal/controller/api/node/schedule/gen/cfg.yaml`, update the
import-mapping relative path (one level deeper).

In `internal/controller/api/node/schedule/gen/api.yaml`, update all `$ref` paths
to common spec (one level deeper).

- [ ] **Step 5: Rename the handler method**

In `handler_node_schedule.go`:

- Rename `GetScheduleHandler` → `GetNodeScheduleHandler`

Update in `cmd/controller_setup.go`:

```go
handlers = append(handlers, sm.GetNodeScheduleHandler(jc)...)
```

Update the interface declaration if it exists.

Update in `handler_public_test.go`:

- Rename `TestGetScheduleHandler` → `TestGetNodeScheduleHandler`

- [ ] **Step 6: Regenerate**

```bash
go generate ./internal/controller/api/node/schedule/gen/...
```

- [ ] **Step 7: Verify build and tests**

```bash
go build ./...
go test ./internal/controller/api/... ./cmd/...
```

- [ ] **Step 8: Commit**

```bash
git add -A
git commit -m "refactor(api): move schedule handlers under node/"
```

---

## Task 3: Move docker under node/

**Files:**

- Move: `internal/controller/api/docker/` →
  `internal/controller/api/node/docker/`
- Rename: `internal/controller/api/handler_docker.go` →
  `internal/controller/api/handler_node_docker.go`

Follow the exact same pattern as Tasks 1-2:

- [ ] **Step 1: Move the directory**

```bash
git mv internal/controller/api/docker internal/controller/api/node/docker
```

- [ ] **Step 2: Rename the handler shim**

```bash
git mv internal/controller/api/handler_docker.go internal/controller/api/handler_node_docker.go
```

- [ ] **Step 3: Update import paths**

Search for `controller/api/docker` and update to `controller/api/node/docker`.

- [ ] **Step 4: Update cfg.yaml and api.yaml relative paths**

One level deeper for the common spec reference.

- [ ] **Step 5: Rename the handler method**

`GetDockerHandler` → `GetNodeDockerHandler` in shim, controller_setup.go,
interface, and test.

- [ ] **Step 6: Regenerate**

```bash
go generate ./internal/controller/api/node/docker/gen/...
```

- [ ] **Step 7: Verify build and tests**

```bash
go build ./...
go test ./internal/controller/api/... ./cmd/...
```

- [ ] **Step 8: Commit**

```bash
git add -A
git commit -m "refactor(api): move docker handlers under node/"
```

---

## Task 4: Split hostname out of node/

This is the first split task. The current `api/node/` package has
`node_hostname_get.go` and `node_hostname_put.go` that need to become their own
package at `api/node/hostname/`.

**Files:**

- Create: `internal/controller/api/node/hostname/`
- Create: `internal/controller/api/node/hostname/gen/api.yaml`
- Create: `internal/controller/api/node/hostname/gen/cfg.yaml`
- Create: `internal/controller/api/node/hostname/gen/generate.go`
- Move + modify: `node_hostname_get.go` → `hostname/hostname_get.go`
- Move + modify: `node_hostname_put.go` → `hostname/hostname_put.go`
- Move + modify: corresponding `*_public_test.go` files
- Create: `internal/controller/api/node/hostname/types.go`
- Create: `internal/controller/api/node/hostname/hostname.go`
- Create: `internal/controller/api/node/hostname/validate.go`
- Create: `internal/controller/api/handler_node_hostname.go`
- Modify: `internal/controller/api/node/gen/api.yaml` (remove hostname
  endpoints)
- Modify: `internal/controller/api/node/node.go` (remove hostname from
  interface)

- [ ] **Step 1: Extract hostname paths from node's api.yaml**

Read `internal/controller/api/node/gen/api.yaml`. Find the hostname endpoints
(GET and PUT on `/node/{hostname}/hostname`). Extract them into a new
`internal/controller/api/node/hostname/gen/api.yaml`.

Include the hostname-specific request/response schemas. Reference common schemas
via `$ref` to `../../../common/gen/api.yaml`.

Remove the hostname endpoints from the node spec.

- [ ] **Step 2: Create cfg.yaml and generate.go**

```yaml
# cfg.yaml
---
package: gen
output: hostname.gen.go
generate:
  models: true
  echo-server: true
  strict-server: true
import-mapping:
  ../../../common/gen/api.yaml: github.com/retr0h/osapi/internal/controller/api/common/gen
output-options:
  skip-prune: true
```

```go
// generate.go
package gen

//go:generate go tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config cfg.yaml api.yaml
```

- [ ] **Step 3: Generate code**

```bash
go generate ./internal/controller/api/node/hostname/gen/...
```

- [ ] **Step 4: Create hostname.go (factory + interface check)**

```go
package hostname

import (
    "log/slog"

    client "github.com/retr0h/osapi/internal/job/client"
    gen "github.com/retr0h/osapi/internal/controller/api/node/hostname/gen"
)

var _ gen.StrictServerInterface = (*Hostname)(nil)

func New(
    logger *slog.Logger,
    jobClient client.JobClient,
    appConfig config.Config,
) *Hostname {
    return &Hostname{
        JobClient: jobClient,
        logger:    logger.With(slog.String("subsystem", "api.hostname")),
        appConfig: appConfig,
    }
}
```

Check what the existing hostname handlers need (the PUT handler may need
`appConfig` for labels). Read the existing handler code to determine the
constructor signature.

- [ ] **Step 5: Create types.go and validate.go**

```go
// types.go
package hostname

type Hostname struct {
    JobClient client.JobClient
    logger    *slog.Logger
    appConfig config.Config
}
```

```go
// validate.go
package hostname

func validateHostname(hostname string) (string, bool) {
    return validation.Var(hostname, "required,min=1,valid_target")
}
```

- [ ] **Step 6: Move and update handler files**

Move `node_hostname_get.go` → `hostname/hostname_get.go`:

- Change `package node` → `package hostname`
- Update receiver type from `(s *Node)` → `(s *Hostname)`
- Update gen type references from `gen.GetNodeHostname*` to the new generated
  types (check the generated interface)
- Update import paths

Same for `node_hostname_put.go` and all corresponding test files.

- [ ] **Step 7: Create handler shim**

Create `internal/controller/api/handler_node_hostname.go` following the existing
shim pattern. Method: `GetNodeHostnameHandler`.

- [ ] **Step 8: Remove hostname from node's StrictServerInterface**

The node `gen/api.yaml` no longer has hostname endpoints, so after regeneration
the node's `StrictServerInterface` won't include them. Remove the hostname
methods from `node.go` and update `handler_node.go`.

Regenerate the node spec:

```bash
go generate ./internal/controller/api/node/gen/...
```

- [ ] **Step 9: Wire in controller_setup.go**

Add the new handler call:

```go
handlers = append(handlers, sm.GetNodeHostnameHandler(jc)...)
```

- [ ] **Step 10: Update handler_public_test.go**

Add `TestGetNodeHostnameHandler` test.

- [ ] **Step 11: Verify build and tests**

```bash
go build ./...
go test ./internal/controller/api/... ./cmd/...
```

- [ ] **Step 12: Commit**

```bash
git add -A
git commit -m "refactor(api): split hostname handlers into node/hostname/"
```

---

## Task 5: Split network out of node/

Same pattern as Task 4. Extract DNS and ping endpoints.

**Files to move from api/node/:**

- `network_dns_get_by_interface.go`
- `network_dns_put_by_interface.go`
- `network_ping_post.go`
- Corresponding `*_public_test.go` files

**New package:** `internal/controller/api/node/network/`

- [ ] **Step 1: Extract network paths from node's api.yaml**

Find DNS (GET/PUT `/node/{hostname}/network/dns/{interfaceName}`) and ping (POST
`/node/{hostname}/network/ping`) endpoints. Create
`api/node/network/gen/api.yaml`.

Remove from node spec.

- [ ] **Step 2: Create cfg.yaml, generate.go, generate code**

- [ ] **Step 3: Create network.go, types.go, validate.go**

Factory: `New(logger, jobClient) *Network` Struct fields: `JobClient`, `logger`

- [ ] **Step 4: Move and update handler files**

Change package, receiver type, gen references, imports.

- [ ] **Step 5: Create handler shim handler_node_network.go**

Method: `GetNodeNetworkHandler`

- [ ] **Step 6: Regenerate node spec, wire in controller_setup.go**

- [ ] **Step 7: Update handler_public_test.go**

- [ ] **Step 8: Verify build and tests**

- [ ] **Step 9: Commit**

```bash
git add -A
git commit -m "refactor(api): split network handlers into node/network/"
```

---

## Task 6: Split command out of node/

Same pattern. Extract exec and shell endpoints.

**Files to move:**

- `command_exec_post.go`
- `command_shell_post.go`
- Corresponding `*_public_test.go` files

**New package:** `internal/controller/api/node/command/`

- [ ] **Step 1-8:** Follow the same steps as Task 5.

Extract POST `/node/{hostname}/command/exec` and POST
`/node/{hostname}/command/shell`.

Factory: `New(logger, jobClient) *Command` Handler shim: `GetNodeCommandHandler`

- [ ] **Step 9: Commit**

```bash
git add -A
git commit -m "refactor(api): split command handlers into node/command/"
```

---

## Task 7: Split file (deploy/undeploy/status) out of node/

Same pattern. Extract the node-targeted file operations (NOT the controller-only
`api/file/` package which stays at the top level).

**Files to move:**

- `file_deploy_post.go`
- `file_undeploy_post.go`
- `file_status_post.go`
- Corresponding `*_public_test.go` files

**New package:** `internal/controller/api/node/filedeploy/`

Note: Cannot use `api/node/file/` because Go doesn't allow a package `file`
under `node` when there's already an `api/file/` package — the import paths
would be unambiguous (`api/node/file` vs `api/file`) but the package name `file`
would collide in files that import both. Use `filedeploy` to avoid confusion.

Actually — check if any handler file imports both `api/file` and the node file
handlers. If not, `api/node/file/` is fine since Go resolves by full import
path. The package name would be `file` in both cases but they're different
packages. Import aliases handle any collision: `nodeFile "...api/node/file"`.

The implementer should check whether `api/node/file/` or `api/node/filedeploy/`
is cleaner. Read existing code to see if there are cross-imports.

- [ ] **Step 1-8:** Follow the same steps as Task 5.

Extract POST endpoints for deploy, undeploy, status under
`/node/{hostname}/file/...`.

Factory: `New(logger, jobClient) *File` Handler shim: `GetNodeFileHandler`

Rename existing `handler_file.go` (controller-only file CRUD) to be explicit —
it's already named `handler_file.go` and handles `api/file/` which is NOT
moving. Just make sure the new shim is `handler_node_file.go` to avoid
confusion.

- [ ] **Step 9: Commit**

```bash
git add -A
git commit -m "refactor(api): split file deploy handlers into node/file/"
```

---

## Task 8: Slim down remaining node/ package

After tasks 4-7, the `api/node/` package should only contain the read-only GET
handlers: disk, memory, load, status, uptime, os.

**Files:**

- Modify: `internal/controller/api/node/node.go` — slim factory
- Modify: `internal/controller/api/node/types.go` — remove unused types
- Modify: `internal/controller/api/node/gen/api.yaml` — should only have 6 GET
  endpoints
- Modify: `internal/controller/api/handler_node.go` — slim to read-only
- Delete: any leftover moved files

- [ ] **Step 1: Verify node/ only has read-only handlers**

List files in `internal/controller/api/node/`. Should only have:

```
gen/
node.go
types.go
validate.go
export_test.go
node_disk_get.go
node_memory_get.go
node_load_get.go
node_status_get.go
node_uptime_get.go
node_os_get.go
*_public_test.go (for the above)
```

And subdirectories: `hostname/`, `sysctl/`, `schedule/`, `docker/`, `network/`,
`command/`, `file/`.

- [ ] **Step 2: Regenerate the slimmed node spec**

```bash
go generate ./internal/controller/api/node/gen/...
```

- [ ] **Step 3: Clean up node.go**

Remove any methods from the `Node` struct that were moved to sub-packages. The
`Node` struct should only implement the slimmed `gen.StrictServerInterface` with
read-only methods.

- [ ] **Step 4: Clean up types.go**

Remove any types that are no longer used (they moved to sub-package types.go
files).

- [ ] **Step 5: Update handler_node.go**

The `GetNodeHandler` method should only register the read-only node routes. The
other handlers are now separate shims (`GetNodeHostnameHandler`, etc.).

- [ ] **Step 6: Verify build and tests**

```bash
go build ./...
go test ./internal/controller/api/... ./cmd/...
```

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "refactor(api): slim node package to read-only handlers"
```

---

## Task 9: Regenerate combined spec

- [ ] **Step 1: Run just generate**

```bash
just generate
```

This runs `redocly join` across all individual specs to produce the combined
`api/gen/api.yaml` and regenerates the SDK client.

- [ ] **Step 2: Verify build**

```bash
go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add -A
git commit -m "chore: regenerate combined spec after API restructure"
```

---

## Task 10: Update CLAUDE.md

- [ ] **Step 1: Update the "Adding a New API Domain" guide**

Update Step 1 (OpenAPI Spec) to reference the nested path:

```
internal/controller/api/node/{domain}/gen/
```

Update Step 2 (Handler Implementation) to reference:

```
internal/controller/api/node/{domain}/
```

Update Step 3 (Server Wiring) to reference:

```
handler_node_{domain}.go
GetNode{Domain}Handler
```

Update the file structure examples throughout to show the nested layout.

- [ ] **Step 2: Update the architecture quick reference**

Update the `internal/controller/api/` description to reflect the nested
structure.

- [ ] **Step 3: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: update CLAUDE.md for nested API directory structure"
```

---

## Task 11: Final Verification

- [ ] **Step 1: Full regeneration**

```bash
just generate
```

- [ ] **Step 2: Build**

```bash
go build ./...
```

- [ ] **Step 3: All unit tests**

```bash
just go::unit
```

- [ ] **Step 4: Coverage**

```bash
just go::unit-cov
```

Expected: >= 99.9%

- [ ] **Step 5: Lint**

```bash
just go::vet
```

- [ ] **Step 6: Format**

```bash
just go::fmt
```

- [ ] **Step 7: Commit any fixes**

```bash
git add -A
git commit -m "chore: formatting and lint fixes after restructure"
```
