# API Directory Restructure Design

## Overview

Reorganize `internal/controller/api/` so that all node-targeted API handler
packages live under `api/node/`, matching the URL path structure
(`/node/{hostname}/...`). Read-only single-endpoint handlers stay flat in
`api/node/`. Domains with mutations (CRUD) get their own subdirectory with
`gen/`, `types.go`, and handler files.

## Motivation

Currently, node-targeted domains are scattered at the top level of `api/`:
`api/sysctl/`, `api/schedule/`, `api/docker/`, while `api/node/` holds hostname,
disk, mem, load, status, uptime, os, DNS, ping, command, and file handlers in
one package. Every one of these routes to `/node/{hostname}/...`, so they belong
under `api/node/`.

As more domains are added (power, process, user, ntp, ssh), the flat top-level
layout will grow confusing. Nesting under `api/node/` keeps things organized and
makes it clear what's node-targeted vs controller-only.

## Design Rule

- **Read-only, single-endpoint handlers** stay flat in `api/node/` (disk,
  memory, load, status, uptime, os)
- **Domains with mutations** (create, update, delete) get their own subdirectory
  under `api/node/` with `gen/`, `types.go`, handler files, and tests

## Directory Structure

### Before

```
internal/controller/api/
  handler.go
  handler_agent.go
  handler_audit.go
  handler_docker.go
  handler_facts.go
  handler_file.go
  handler_health.go
  handler_node.go
  handler_schedule.go
  handler_sysctl.go
  handler_public_test.go
  types.go
  agent/
  audit/
  common/
  docker/         ← node-targeted, at top level
  facts/
  file/
  health/
  job/
  node/           ← catch-all for many domains
  schedule/       ← node-targeted, at top level
  sysctl/         ← node-targeted, at top level
  gen/
```

### After

```
internal/controller/api/
  handler.go
  handler_node.go              ← flat read-only node handlers
  handler_node_hostname.go     ← renamed from part of handler_node.go
  handler_node_sysctl.go       ← renamed from handler_sysctl.go
  handler_node_schedule.go     ← renamed from handler_schedule.go
  handler_node_docker.go       ← renamed from handler_docker.go
  handler_node_command.go      ← new (split from handler_node.go)
  handler_node_file.go         ← new (split from handler_node.go)
  handler_node_network.go      ← new (split from handler_node.go)
  handler_agent.go
  handler_audit.go
  handler_facts.go
  handler_file.go              ← controller-only file CRUD
  handler_health.go
  handler_job.go               ← renamed from handler_node.go job part
  handler_public_test.go
  types.go
  node/
    gen/                       ← OpenAPI spec for read-only node endpoints
    node.go                    ← factory for flat read-only handlers
    types.go                   ← shared types
    validate.go                ← validateHostname (shared by all node subpkgs)
    disk_get.go
    memory_get.go
    load_get.go
    status_get.go
    uptime_get.go
    os_get.go
    *_public_test.go
    hostname/                  ← has PUT, gets own dir
      gen/
      types.go
      hostname.go
      hostname_get.go
      hostname_put.go
      validate.go
      *_public_test.go
    sysctl/                    ← moved from api/sysctl/
      gen/
      types.go
      sysctl.go
      sysctl_list_get.go
      sysctl_get.go
      sysctl_create.go
      sysctl_update.go
      sysctl_delete.go
      validate.go
      *_public_test.go
    schedule/                  ← moved from api/schedule/
      gen/
      types.go
      schedule.go
      cron_list_get.go
      cron_get.go
      cron_create.go
      cron_update.go
      cron_delete.go
      validate.go
      *_public_test.go
    docker/                    ← moved from api/docker/
      gen/
      types.go
      docker.go
      container_list.go
      container_create.go
      container_inspect.go
      container_start.go
      container_stop.go
      container_remove.go
      container_exec.go
      container_pull.go
      container_image_remove.go
      validate.go
      *_public_test.go
    command/                   ← split from api/node/
      gen/
      types.go
      command.go
      exec_post.go
      shell_post.go
      validate.go
      *_public_test.go
    file/                      ← split from api/node/ (deploy/undeploy/status)
      gen/
      types.go
      file.go
      deploy_post.go
      undeploy_post.go
      status_post.go
      validate.go
      *_public_test.go
    network/                   ← split from api/node/ (dns/ping)
      gen/
      types.go
      network.go
      dns_get.go
      dns_put.go
      ping_post.go
      validate.go
      *_public_test.go
  agent/                       ← unchanged
  audit/                       ← unchanged
  common/                      ← unchanged
  facts/                       ← unchanged
  file/                        ← unchanged (controller-only file CRUD)
  health/                      ← unchanged
  job/                         ← unchanged
  gen/                         ← combined spec
```

## Handler Shim Pattern

Shims stay in `api/` as methods on `Server`. They are renamed to match the
nested path:

| Before                | After                         |
| --------------------- | ----------------------------- |
| `handler_node.go`     | split into multiple files     |
| `handler_sysctl.go`   | `handler_node_sysctl.go`      |
| `handler_schedule.go` | `handler_node_schedule.go`    |
| `handler_docker.go`   | `handler_node_docker.go`      |
| (in handler_node.go)  | `handler_node_hostname.go`    |
| (in handler_node.go)  | `handler_node_command.go`     |
| (in handler_node.go)  | `handler_node_file.go`        |
| (in handler_node.go)  | `handler_node_network.go`     |
| (new)                 | `handler_node.go` (read-only) |

Shims stay in `api/` because:

- They are methods on `Server` which owns middleware config
- Moving them into domain packages would couple domains to auth/middleware
  internals
- Domain packages stay clean: only know about `JobClient` and their own `gen`
  types

## OpenAPI Spec Split

The current monolithic `api/node/gen/api.yaml` must be split. Each subdirectory
gets its own `gen/api.yaml` containing only its endpoints. The read-only node
endpoints (disk, mem, load, status, uptime, os) stay in `api/node/gen/api.yaml`.

The combined spec (`api/gen/api.yaml`) is still generated by `redocly join` from
all individual specs — no change to how it works, just more input specs.

## Validation Sharing

`validateHostname()` is needed by every node sub-package. Options:

1. Each sub-package defines its own copy (current cron/sysctl pattern — simple,
   no cross-package imports)
2. Put it in `api/node/validate.go` and import from sub-packages

Option 1 is simpler and avoids circular imports. Each sub-package already has
its own `validate.go` with `validateHostname()`. This is a one-liner that calls
`validation.Var()` — duplication is acceptable.

## Controller-Only Packages

These packages are NOT under `/node/{hostname}/` and stay at the top level:

| Package   | Routes                               |
| --------- | ------------------------------------ |
| `agent/`  | `/node` (list/get/drain/undrain)     |
| `job/`    | `/job/...`                           |
| `health/` | `/health/...`                        |
| `audit/`  | `/audit/...`                         |
| `file/`   | `/file/...` (upload/list/get/delete) |
| `facts/`  | `/facts/...`                         |

## What Does NOT Change

- Provider directory structure — stays as-is
- Agent processor structure — stays as-is
- SDK client structure — stays as-is
- CLI command structure — stays as-is
- Job operation constants — stay as-is
- URL paths — stay exactly the same
- Handler logic — no changes to handler implementations, only package paths and
  imports

## Migration Approach

This is a purely mechanical refactor:

1. Create new directory structure
2. Move files, update package declarations
3. Update imports across the codebase
4. Split the monolithic `api/node/gen/api.yaml` into per-domain specs
5. Regenerate all specs
6. Rename shim files
7. Update `controller_setup.go` method names
8. Run tests, lint, verify

No logic changes anywhere. Every handler, test, and validation function stays
identical — only the package path changes.
