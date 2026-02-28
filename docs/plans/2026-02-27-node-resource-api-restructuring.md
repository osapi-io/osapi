# Node-Centric API Restructuring Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to
> implement this plan task-by-task.

**Goal:** Restructure the REST API so that `/node/{hostname}` is the
top-level resource for all machine operations, moving target from query
params to path params and nesting network/command under node.

**Architecture:** Replace flat domain paths (`/node/status`, `/network/ping`,
`/command/exec`) with resource-oriented paths (`/node/{hostname}`,
`/node/{hostname}/network/ping`, `/node/{hostname}/command/exec`). The
`{hostname}` path segment replaces the `target_hostname` query parameter,
accepting the same values: `_any`, `_all`, literal hostnames, or
`key:value` label selectors. Individual sub-resource endpoints (disk,
memory, load, os, uptime) are added alongside the composite status.

**Tech Stack:** Go 1.25, oapi-codegen (strict-server), Echo v4, NATS
JetStream, testify/suite

---

## New API Layout

```
# Node (resource-oriented, target in path)
GET  /node/{hostname}                              composite status
GET  /node/{hostname}/disk                         disk usage
GET  /node/{hostname}/memory                       memory stats
GET  /node/{hostname}/load                         load averages
GET  /node/{hostname}/os                           OS info
GET  /node/{hostname}/uptime                       uptime
GET  /node/{hostname}/hostname                     hostname + labels

# Network (nested under node)
GET  /node/{hostname}/network/dns/{interfaceName}  DNS config
PUT  /node/{hostname}/network/dns                  update DNS
POST /node/{hostname}/network/ping                 ping

# Command (nested under node)
POST /node/{hostname}/command/exec                 execute command
POST /node/{hostname}/command/shell                shell command

# Unchanged domains
GET  /agent                                        list agents
GET  /agent/{hostname}                             agent details
POST /job                                          create job
GET  /job                                          list jobs
GET  /job/{id}                                     get job
DELETE /job/{id}                                   delete job
POST /job/{id}/retry                               retry job
GET  /job/status                                   queue stats
GET  /health                                       liveness
GET  /health/ready                                 readiness
GET  /health/status                                detailed status
GET  /audit                                        list entries
GET  /audit/{id}                                   get entry
GET  /audit/export                                 export all
```

### Query Param Policy

After this change, query params are ONLY used for **filtering and
pagination** on collection endpoints (`/job`, `/audit`). Resource
identification always uses path parameters. Complex input data uses
request bodies.

### Reserved Hostname Values

The `{hostname}` path segment accepts:

| Value              | Meaning                          |
| ------------------ | -------------------------------- |
| `_any`             | Load-balanced to any agent       |
| `_all`             | Broadcast to all agents          |
| `web-01`           | Direct routing to specific host  |
| `group:web.dev`    | Label-based routing              |

---

## Phasing

### Phase 1 — API Guidelines & Task File

Update `api-guidelines.md` to document the node-centric resource model,
path parameter conventions, and query param policy. Create in-progress
task file.

### Phase 2 — OpenAPI Specs & Code Generation

Rewrite the three OpenAPI specs (node, network, command) with the new
path structure. Merge network and command specs into the node spec since
they now share the `/node/{hostname}` prefix. Regenerate `*.gen.go`.

### Phase 3 — Handlers

Update handlers to extract `{hostname}` from path instead of query
params. Add new handlers for the individual sub-resources (disk, memory,
load, os, uptime). Update handler wiring.

### Phase 4 — JobClient Methods

Add new `QueryNodeDisk`, `QueryNodeMemory`, `QueryNodeLoad`,
`QueryNodeOS`, `QueryNodeUptime` methods (plus broadcast variants) to
the job client. The agent already handles these operations.

### Phase 5 — CLI & SDK

Update CLI commands to build paths with target instead of query params.
Restructure CLI command tree (`client node disk --target web-01` →
builds `GET /node/web-01/disk`). Update SDK.

### Phase 6 — Tests & Documentation

Update all integration tests, feature docs, CLI docs, architecture
docs. Remove stale `target_hostname` references.

---

## Task 1: Update API Guidelines

**Files:**
- Modify: `docs/docs/sidebar/architecture/api-guidelines.md`

**Step 1: Update the API guidelines document**

Add two new sections after the existing four guidelines:

```markdown
5. **Node as Top-Level Resource**

All operations that target a managed machine are nested under
`/node/{hostname}`. The `{hostname}` path segment identifies the
target and accepts literal hostnames, reserved routing values
(`_any`, `_all`), or label selectors (`key:value`).

Sub-resources represent distinct capabilities of the node:

| Path Pattern                                    | Domain  |
| ----------------------------------------------- | ------- |
| `/node/{hostname}`                              | Status  |
| `/node/{hostname}/disk`                         | Node    |
| `/node/{hostname}/memory`                       | Node    |
| `/node/{hostname}/network/dns/{interfaceName}`  | Network |
| `/node/{hostname}/command/exec`                 | Command |

6. **Path Parameters Over Query Parameters**

Use path parameters for **resource identification and targeting**.
Use query parameters only for **filtering and pagination** on
collection endpoints (e.g., `/job?status=completed&limit=20`).

Never use query parameters to identify which resource to act on.
Complex input data belongs in request bodies.
```

**Step 2: Commit**

```
docs: update API guidelines with node-centric resource model
```

---

## Task 2: Create In-Progress Task File

**Files:**
- Create: `docs/docs/sidebar/development/tasks/in-progress/2026-02-27-node-resource-api-restructuring.md`

**Step 1: Create the task file**

```markdown
---
title: Node-centric API restructuring
status: in-progress
created: 2026-02-27
updated: 2026-02-27
---

## Objective

Restructure the REST API so `/node/{hostname}` is the top-level
resource for all machine operations. Move target from query params to
path params. Nest network and command under node. Add individual
sub-resource endpoints (disk, memory, load, os, uptime).

## Notes

- Plan: `docs/plans/2026-02-27-node-resource-api-restructuring.md`
- Agent already handles individual operations (disk, memory, load, etc.)
- Job routing logic unchanged — only HTTP layer changes
- Reserved hostnames: `_any`, `_all`, label selectors (`key:value`)

## Outcome

_To be filled in when complete._
```

**Step 2: Commit**

```
docs: add node-centric API restructuring task
```

---

## Task 3: Rewrite Node OpenAPI Spec

This is the core spec change. The node spec absorbs network and command
paths since they all share the `/node/{hostname}` prefix.

**Files:**
- Modify: `internal/api/node/gen/api.yaml`
- Modify: `internal/api/node/gen/cfg.yaml` (add import mappings for
  network/command shared schemas if needed)
- Delete: `internal/api/network/gen/api.yaml` (paths move to node spec)
- Delete: `internal/api/command/gen/api.yaml` (paths move to node spec)

**Step 1: Rewrite `internal/api/node/gen/api.yaml`**

The new spec must include:

**Paths:**
- `GET /node/{hostname}` — composite status (was `/node/status`)
- `GET /node/{hostname}/hostname` — hostname + labels
- `GET /node/{hostname}/disk` — disk usage (NEW)
- `GET /node/{hostname}/memory` — memory stats (NEW)
- `GET /node/{hostname}/load` — load averages (NEW)
- `GET /node/{hostname}/os` — OS info (NEW)
- `GET /node/{hostname}/uptime` — uptime (NEW)
- `GET /node/{hostname}/network/dns/{interfaceName}` — DNS config
- `PUT /node/{hostname}/network/dns` — update DNS
- `POST /node/{hostname}/network/ping` — ping
- `POST /node/{hostname}/command/exec` — exec command
- `POST /node/{hostname}/command/shell` — shell command

**Path parameter `hostname`:**
```yaml
parameters:
  - name: hostname
    in: path
    required: true
    description: >-
      Target agent hostname, reserved routing value (_any, _all),
      or label selector (key:value).
    schema:
      type: string
      minLength: 1
    x-oapi-codegen-extra-tags:
      validate: required,min=1,valid_target
```

**Key design decisions:**
- `target_hostname` query param removed from ALL endpoints
- `hostname` path param added to ALL endpoints
- Network and command schemas (request/response types) copied into
  node spec or imported via `$ref` from `common/gen/api.yaml`
- Existing response schemas preserved (no API response breaking changes)
- Security scopes remain: `node:read`, `network:read`, `network:write`,
  `command:execute`

**New response schemas for individual sub-resources:**

```yaml
DiskUsageCollectionResponse:
  type: object
  properties:
    job_id:
      type: string
      format: uuid
    results:
      type: array
      items:
        $ref: '#/components/schemas/DiskUsageResponse'

DiskUsageResponse:
  type: object
  properties:
    hostname:
      type: string
    disks:
      type: array
      items:
        $ref: './common/DiskUsageItem'
    error:
      type: string
```

Follow the same pattern for Memory, Load, OS, Uptime responses.

**Step 2: Regenerate code**

```bash
go generate ./internal/api/node/gen/...
```

**Step 3: Verify generated code compiles**

```bash
go build ./internal/api/node/...
```

**Step 4: Commit**

```
feat: rewrite node OpenAPI spec with resource-oriented paths
```

---

## Task 4: Consolidate Handler Packages

Since network and command paths are now under `/node/{hostname}/...`,
the handlers can either stay in their packages (with the node spec
importing their logic) or be consolidated. The cleanest approach:
keep separate handler structs but register them all through the node
spec's generated interface.

**Files:**
- Modify: `internal/api/node/types.go` — add network/command
  dependencies
- Modify: `internal/api/node/node.go` — expand factory to accept
  all dependencies
- Create: `internal/api/node/node_disk_get.go` — new handler
- Create: `internal/api/node/node_memory_get.go` — new handler
- Create: `internal/api/node/node_load_get.go` — new handler
- Create: `internal/api/node/node_os_get.go` — new handler
- Create: `internal/api/node/node_uptime_get.go` — new handler
- Move: network handler logic into `internal/api/node/network_*.go`
- Move: command handler logic into `internal/api/node/command_*.go`
- Modify: `internal/api/handler_node.go` — register all handlers
- Remove: `internal/api/handler_network.go` (absorbed by node)
- Remove: `internal/api/handler_command.go` (absorbed by node)
- Modify: `internal/api/types.go` — remove network/command handler
  fields
- Modify: `internal/api/handler.go` — remove network/command from
  `CreateHandlers`
- Modify: `cmd/api_helpers.go` — remove `GetNetworkHandler`,
  `GetCommandHandler` from `ServerManager` interface

**Step 1: Update `internal/api/node/types.go`**

The Node struct needs the JobClient (already has it). The JobClient
interface already has all Query/Modify methods for network and command.
No new dependencies needed — just new handler methods.

**Step 2: Write failing tests for new sub-resource handlers**

For each new handler (disk, memory, load, os, uptime), write a public
test following the existing `node_status_get_public_test.go` pattern:

```go
// node_disk_get_public_test.go
func (suite *NodeDiskGetPublicTestSuite) TestGetNodeDisk() {
    tests := []struct {
        name     string
        hostname string
        // ...
    }{
        // success, validation error, job client error cases
    }
}
```

**Step 3: Implement handlers**

Each new handler follows the same pattern as `GetNodeStatus`:
1. Extract `hostname` from `request.Hostname` (path param)
2. Validate with `validation.Struct`
3. Check `job.IsBroadcastTarget(hostname)` for broadcast fork
4. Call appropriate `JobClient.QueryNode*` method
5. Build and return response

**Step 4: Update existing handlers**

Modify `node_status_get.go` and `node_hostname_get.go` to extract
hostname from `request.Hostname` (path param) instead of
`request.Params.TargetHostname` (query param).

Move network handlers (`network_ping_post.go`, `network_dns_*.go`) and
command handlers (`command_exec_post.go`, `command_shell_post.go`) into
the node package. Update them to extract hostname from path param.

**Step 5: Update handler wiring**

In `internal/api/handler_node.go`, the `GetNodeHandler` method now
registers ALL handlers (node + network + command) since they all come
from the same generated spec.

Remove `GetNetworkHandler` and `GetCommandHandler` from the
`ServerManager` interface and handler.go.

**Step 6: Run tests**

```bash
go test ./internal/api/node/...
```

**Step 7: Commit**

```
feat: consolidate node/network/command handlers under node package
```

---

## Task 5: Add JobClient Methods for New Sub-Resources

The agent already handles `node.disk.get`, `node.memory.get`, etc.
We need JobClient methods to create and wait for these jobs.

**Files:**
- Modify: `internal/job/client/query.go` — add new Query methods
- Modify: `internal/job/client/types.go` (or interface file) — add to
  interface
- Create: `internal/job/client/query_node_test.go` — tests for new
  methods

**Step 1: Write failing tests**

Follow the existing `TestQueryNodeStatus` pattern in
`query_public_test.go`.

**Step 2: Add methods**

New methods needed:
- `QueryNodeDisk(ctx, hostname)` + `QueryNodeDiskBroadcast`
- `QueryNodeMemory(ctx, hostname)` + `QueryNodeMemoryBroadcast`
- `QueryNodeLoad(ctx, hostname)` + `QueryNodeLoadBroadcast`
- `QueryNodeOS(ctx, hostname)` + `QueryNodeOSBroadcast`
- `QueryNodeUptime(ctx, hostname)` + `QueryNodeUptimeBroadcast`

Each follows the exact same pattern as `QueryNodeStatus` but uses
different operation constants (`OperationNodeDiskGet`, etc.).

**Step 3: Run tests**

```bash
go test ./internal/job/client/...
```

**Step 4: Commit**

```
feat: add job client methods for individual node sub-resources
```

---

## Task 6: Update CLI Commands

**Files:**
- Modify: `cmd/client_node.go` — add subcommands for disk, memory,
  load, os, uptime
- Create: `cmd/client_node_disk_get.go`
- Create: `cmd/client_node_memory_get.go`
- Create: `cmd/client_node_load_get.go`
- Create: `cmd/client_node_os_get.go`
- Create: `cmd/client_node_uptime_get.go`
- Modify: `cmd/client_node_status_get.go` — update for path-based
  target
- Modify: `cmd/client_node_hostname_get.go` — same
- Move: `cmd/client_network_*.go` → `cmd/client_node_network_*.go`
- Move: `cmd/client_command_*.go` → `cmd/client_node_command_*.go`
- Modify: `cmd/client.go` — update command tree

**Step 1: Update CLI command tree**

New structure:
```
osapi client node status     --target web-01
osapi client node hostname   --target web-01
osapi client node disk       --target web-01
osapi client node memory     --target web-01
osapi client node load       --target web-01
osapi client node os         --target web-01
osapi client node uptime     --target web-01
osapi client node network dns get --interface eth0 --target web-01
osapi client node network dns update ... --target web-01
osapi client node network ping --address 1.1.1.1 --target web-01
osapi client node command exec --command ls --target web-01
osapi client node command shell --command "ls -la" --target web-01
```

The `--target` flag is still a CLI flag — it gets placed into the URL
path by the SDK, not as a query param.

**Step 2: Commit**

```
feat: restructure CLI commands under node with sub-resources
```

---

## Task 7: Update SDK

**Files (in osapi-sdk repo):**
- Sync new `api.yaml` specs
- Regenerate client code
- Update service wrappers

**Step 1: Copy updated api.yaml to SDK**

The SDK pulls specs via gilt. During development, manually copy
`internal/api/node/gen/api.yaml` to the SDK.

**Step 2: Regenerate SDK**

```bash
cd ../osapi-sdk && just generate
```

**Step 3: Update SDK service wrappers**

The `Node` service now includes network and command methods. Path
construction changes from query params to path segments.

**Step 4: Commit (in SDK repo)**

```
feat: update SDK for node-centric API paths
```

---

## Task 8: Integration Tests

**Files:**
- Modify: all `*_integration_test.go` in `internal/api/node/`
- Create: new integration tests for disk, memory, load, os, uptime
- Remove: `internal/api/network/*_integration_test.go` (moved to node)
- Remove: `internal/api/command/*_integration_test.go` (moved to node)

Every integration test must verify:
- Valid input returns correct response
- Invalid `{hostname}` returns 400
- Missing token returns 401
- Wrong permissions return 403
- Valid token with correct scope returns 200/202

**Step 1: Write integration tests for new endpoints**

**Step 2: Update existing integration tests for path change**

**Step 3: Run full test suite**

```bash
just test
```

**Step 4: Commit**

```
test: update integration tests for node-centric API paths
```

---

## Task 9: Update Documentation

**Files:**
- Modify: `docs/docs/sidebar/features/node-management.md`
- Modify: `docs/docs/sidebar/features/network-management.md`
- Modify: `docs/docs/sidebar/features/command-execution.md`
- Modify: `docs/docs/sidebar/architecture/system-architecture.md`
- Modify: `docs/docs/sidebar/usage/configuration.md` (permissions table)
- Modify: `docs/docs/sidebar/usage/cli/client/node/` docs
- Create: CLI docs for new subcommands (disk, memory, load, os, uptime)
- Move: network/command CLI docs under node
- Modify: `CLAUDE.md` — update architecture quick reference

**Step 1: Update feature docs**

**Step 2: Update CLI docs**

**Step 3: Update architecture docs**

**Step 4: Run docs build**

```bash
just docs::build
```

**Step 5: Commit**

```
docs: update documentation for node-centric API restructuring
```

---

## Task 10: Cleanup

**Files:**
- Remove: `internal/api/network/` package (if fully absorbed)
- Remove: `internal/api/command/` package (if fully absorbed)
- Remove: stale CLI command files
- Verify: no remaining `target_hostname` query param references

**Step 1: Search for stale references**

```bash
grep -r "target_hostname" --include="*.go" --include="*.yaml"
grep -r "TargetHostname" --include="*.go"
```

**Step 2: Remove dead code**

**Step 3: Final verification**

```bash
just generate
go build ./...
just test
just go::vet
```

**Step 4: Commit**

```
refactor: remove stale network/command packages and target_hostname refs
```

---

## Resolved Decisions

- **Colon in path segments** — keep `:` as the label selector delimiter.
  It is a valid URL path character and does not conflict with Echo routing
  (oapi-codegen uses `{hostname}` style, not `:hostname`).
- **Path param validation** — confirmed: oapi-codegen strict-server mode
  does NOT generate validate tags on path param request object structs.
  Each handler must manually validate `request.Hostname` using a shared
  helper: `validateHostname(request.Hostname)` → calls
  `validation.Var()` with `required,min=1,valid_target`. Add YAML
  comments in OpenAPI specs wherever path param `x-oapi-codegen-extra-tags`
  appears, noting the tags are not generated in strict-server mode and
  validation is handled manually in handlers.
- **RBAC permissions** — keep existing granular permissions unchanged
  (`node:read`, `network:read`, `network:write`, `command:execute`).
  Permissions map to capabilities, not URL structure. Existing JWT tokens
  continue to work.
- **Spec merging** — merge network and command specs into the node spec
  since all paths share `/node/{hostname}` prefix.

## Task 11: File oapi-codegen Feature Request

**Step 1: Open an issue on oapi-codegen**

File a feature request on `github.com/oapi-codegen/oapi-codegen` asking for
`x-oapi-codegen-extra-tags` on path parameters to be propagated to
`RequestObject` struct fields in strict-server mode. Currently the tags
only work for request body properties and query parameter `*Params`
structs. Path params in strict mode become plain struct fields with no
extra tags.

Include a minimal reproducer (OpenAPI spec + cfg.yaml with
`strict-server: true`) showing the expected vs actual generated code.

---

## Outcome

All tasks complete. The restructuring was implemented across three
Claude Code sessions (context exhausted twice).

### What was done

- **OpenAPI spec**: Merged network and command specs into a single node
  spec with 12 endpoints under `/node/{hostname}/...`. Regenerated
  `*.gen.go`.
- **Handlers**: Consolidated `internal/api/network/` and
  `internal/api/command/` into `internal/api/node/`. Added 5 new
  sub-resource handlers (disk, memory, load, os, uptime). Removed
  stale handler wiring (`GetNetworkHandler`, `GetCommandHandler`).
- **JobClient**: Added `QueryNodeDisk`, `QueryNodeMemory`,
  `QueryNodeLoad`, `QueryNodeOS`, `QueryNodeUptime` methods (plus
  broadcast variants).
- **CLI**: Restructured command tree — `client network *` →
  `client node network *`, `client command *` →
  `client node command *`. All commands use `sdkClient.Node.*`.
- **SDK**: Consolidated `NetworkService` and `CommandService` into
  `NodeService`. Merged and pushed as
  [osapi-sdk#9](https://github.com/osapi-io/osapi-sdk/pull/9).
  Fixed README example link in
  [osapi-sdk#10](https://github.com/osapi-io/osapi-sdk/pull/10).
- **Tests**: All unit tests (testify/suite, table-driven), integration
  tests (RBAC + validation), and bats CLI tests updated. 26 packages
  pass.
- **Documentation**: CLI docs moved under `node/` directory. Feature
  docs, architecture docs, API guidelines, CLAUDE.md, and audit
  example paths updated. No stale references remain.
- **Cleanup**: Deleted `internal/api/network/`, `internal/api/command/`,
  old CLI files, old SDK services. Removed `target_hostname` query
  param references.

### oapi-codegen feature request

Filed [oapi-codegen#2261](https://github.com/oapi-codegen/oapi-codegen/issues/2261):
`x-oapi-codegen-extra-tags` on path parameters are not propagated to
`RequestObject` struct fields in strict-server mode. Workaround:
manual `validation.Var()` calls in each handler.

### Verification

```
go build ./...    # passes
just go::unit     # 26 packages pass
just go::vet      # lint clean
```

## Risk Notes

- **Breaking API change** — all clients must update. Document migration.
- **SDK sync** — SDK must be updated before CLI works with new paths.
