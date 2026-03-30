# Sysctl Provider Design

## Overview

Add kernel parameter management (sysctl) to OSAPI. Users can query,
set, and remove sysctl parameters on managed nodes. Every set
operation is persistent (writes to `/etc/sysctl.d/` and applies
immediately) — there is no runtime-only mode.

## Architecture

Sysctl is a **meta-provider** under `internal/provider/node/sysctl/`.
It delegates file writes to `file.Deployer` for SHA tracking,
idempotency, and drift detection. After deploying the conf file, it
calls `sysctl -p <file>` to apply the change at runtime.

- **Category**: `node` (alongside host, disk, mem, load)
- **Path prefix**: `/node/{hostname}/sysctl`
- **Permissions**: `sysctl:read`, `sysctl:write`
- **Provider type**: meta-provider (delegates to file provider)

## Provider Interface

```go
// Package sysctl provides kernel parameter management via /etc/sysctl.d/.
package sysctl

type Provider interface {
    List(ctx context.Context) ([]Entry, error)
    Get(ctx context.Context, key string) (*Entry, error)
    Set(ctx context.Context, entry Entry) (*SetResult, error)
    Delete(ctx context.Context, key string) (*DeleteResult, error)
}
```

No separate Create/Update — `Set` is idempotent. If the key's conf
file exists with the same value, `Changed: false`. If different or
new, deploy and apply.

## Data Types

```go
type Entry struct {
    Key   string `json:"key"`   // e.g., "net.ipv4.ip_forward"
    Value string `json:"value"` // e.g., "1"
}

type SetResult struct {
    Key     string `json:"key"`
    Changed bool   `json:"changed"`
    Error   string `json:"error,omitempty"`
}

type DeleteResult struct {
    Key     string `json:"key"`
    Changed bool   `json:"changed"`
    Error   string `json:"error,omitempty"`
}
```

## File Layout

Each managed key gets its own conf file:

- **Path**: `/etc/sysctl.d/osapi-{sanitized-key}.conf`
- **Content**: `{key} = {value}\n`
- **Sanitization**: dots remain in the filename (e.g.,
  `osapi-net.ipv4.ip_forward.conf`)

The file provider tracks each file in the file-state KV bucket.
Domain-specific metadata (key, value) is stored in the
`FileState.Metadata` map.

## Operations Flow

### Set

1. Validate key format (dotted kernel parameter name)
2. Generate conf file content: `{key} = {value}\n`
3. Deploy via `file.Deployer.Deploy()` — handles SHA comparison,
   idempotency, and state persistence
4. If changed, apply via `sysctl -p /etc/sysctl.d/osapi-{key}.conf`
5. Return `SetResult` with `Changed` reflecting whether the file
   was actually modified

### Delete

1. Look up key in file-state KV to verify it is managed
2. Undeploy via `file.Deployer.Undeploy()` — removes file and state
3. Apply defaults via `sysctl --system` to reload all conf files
4. Return `DeleteResult`

### List

1. Scan file-state KV for sysctl-managed files (filter by metadata)
2. For each managed key, read current runtime value via
   `sysctl -n {key}`
3. Return list of `Entry` with current runtime values

### Get

1. Look up key in file-state KV
2. Read current runtime value via `sysctl -n {key}`
3. Return `Entry`

## API Endpoints

| Method   | Path                             | Permission     | Description                         |
| -------- | -------------------------------- | -------------- | ----------------------------------- |
| `GET`    | `/node/{hostname}/sysctl`        | `sysctl:read`  | List managed sysctl entries         |
| `GET`    | `/node/{hostname}/sysctl/{key}`  | `sysctl:read`  | Get a single entry by key           |
| `POST`   | `/node/{hostname}/sysctl`        | `sysctl:write` | Set a sysctl key (idempotent)       |
| `DELETE` | `/node/{hostname}/sysctl/{key}`  | `sysctl:write` | Remove managed entry, restore default |

All endpoints support broadcast targeting (`_all`, `_any`, hostname,
label selectors).

### Response Shape

All node-targeted operations return the standard collection response:

```json
{
  "job_id": "...",
  "results": [
    {"hostname": "web-01", "key": "net.ipv4.ip_forward", "value": "1", "error": ""},
    {"hostname": "web-02", "key": "net.ipv4.ip_forward", "value": "1", "error": ""}
  ]
}
```

Set/Delete results include `changed` field. Single-target returns
1 result; broadcast returns N results.

### POST Request Body

```json
{
  "key": "net.ipv4.ip_forward",
  "value": "1"
}
```

### Validation

- `key`: required, must match sysctl key format (dotted name, e.g.,
  `net.ipv4.ip_forward`)
- `value`: required, non-empty string
- Path parameter `{key}` on GET/DELETE uses the dotted key name
  directly

## Platform Implementations

| Platform | Implementation                                          |
| -------- | ------------------------------------------------------- |
| Debian   | Full — delegates to file provider, applies via sysctl   |
| Darwin   | Returns `ErrUnsupported` for all methods                |
| Linux    | Returns `ErrUnsupported` for all methods                |

### Container Behavior

No `DebianDocker` variant is needed. Unlike hostname or DNS, sysctl
works the same inside containers — reads always succeed (host kernel
values), and writes succeed or fail based on container capabilities.
The standard Debian provider handles both cases: if the agent lacks
permissions, `sysctl -w` returns an error and the provider reports
it in the result.

### Debian Dependencies

- `file.Deployer` — for conf file deployment and state tracking
- `jetstream.KeyValue` — file-state KV for listing managed entries
- `exec.Manager` — for running `sysctl -p` and `sysctl -n` commands
- `avfs.VFS` — filesystem access

## Orchestrator Integration

The OSAPI API is single-key CRUD. The orchestrator DSL handles
batching — a single sysctl block in the DSL can declare multiple
keys, and the orchestrator iterates, calling the API once per key.
This is the same pattern used for cron entries.

```yaml
# Example orchestrator DSL (future work in osapi-orchestrator)
sysctl:
  - key: net.ipv4.ip_forward
    value: "1"
  - key: net.core.somaxconn
    value: "4096"
```

## Files to Create/Modify

### New Files

```
internal/provider/node/sysctl/
  types.go        — Provider interface + Entry, SetResult, DeleteResult
  debian.go       — Debian implementation (meta-provider)
  darwin.go       — macOS stub
  linux.go        — Generic Linux stub
  mocks/
    generate.go   — //go:generate mockgen directive

internal/controller/api/sysctl/
  types.go        — Handler struct + dependency interfaces
  sysctl.go       — New() factory + interface check
  sysctl_list_get.go    — GET /sysctl handler
  sysctl_get.go         — GET /sysctl/{key} handler
  sysctl_set.go         — POST /sysctl handler
  sysctl_delete.go      — DELETE /sysctl/{key} handler
  validate.go           — Input validation helpers
  gen/
    api.yaml      — OpenAPI spec
    cfg.yaml      — oapi-codegen config
    generate.go   — //go:generate directive

internal/agent/processor_sysctl.go — Processor wiring

pkg/sdk/client/
  sysctl.go       — SysctlService methods
  sysctl_types.go — SDK result types + gen conversions

cmd/
  client_sysctl.go           — Parent command
  client_sysctl_list.go      — list subcommand
  client_sysctl_get.go       — get subcommand
  client_sysctl_set.go       — set subcommand
  client_sysctl_delete.go    — delete subcommand

examples/sdk/client/sysctl.go — SDK example

docs/docs/sidebar/features/sysctl.md — Feature docs
docs/docs/sidebar/usage/cli/client/sysctl/sysctl.md    — CLI parent
docs/docs/sidebar/usage/cli/client/sysctl/list.md       — CLI list
docs/docs/sidebar/usage/cli/client/sysctl/get.md        — CLI get
docs/docs/sidebar/usage/cli/client/sysctl/set.md        — CLI set
docs/docs/sidebar/usage/cli/client/sysctl/delete.md     — CLI delete

test/integration/sysctl_test.go — Integration tests
```

### Modified Files

```
internal/job/types.go           — Add OperationSysctl* constants
internal/agent/processor_sysctl.go — New processor
cmd/agent_setup.go              — Create and register sysctl provider
internal/controller/api/types.go     — Add sysctlHandler field
internal/controller/api/handler.go   — Wire handler in CreateHandlers
internal/controller/api/handler_sysctl.go — GetSysctlHandler method
cmd/controller_start.go         — Initialize sysctl handler
pkg/sdk/client/osapi.go         — Wire SysctlService
pkg/sdk/client/permissions.go   — Add PermSysctlRead/Write
internal/authtoken/permissions.go — Re-export + add to roles
docs/docusaurus.config.ts       — Add to Features navbar
docs/docs/sidebar/usage/configuration.md — Add permissions
docs/docs/sidebar/architecture/system-architecture.md — Add endpoints
```
