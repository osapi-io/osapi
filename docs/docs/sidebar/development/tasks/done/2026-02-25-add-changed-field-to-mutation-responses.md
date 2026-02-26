---
title: Add `changed` field to mutation responses
status: done
created: 2026-02-25
updated: 2026-02-25
---

## Outcome

Implemented across all layers: DNS provider (compare-before-write idempotency),
command provider (always true), job response type, worker processor/handler, job
client, OpenAPI specs, API handlers, SDK, and CLI display. Full test coverage
including `extractChanged`, API response assertions, and CLI table rendering.

## Objective

Add a `changed` boolean field to all mutation (TypeModify) operation responses
so callers know whether the operation actually modified system state vs found it
already in the desired state. This is the foundational building block for
idempotent automation tooling (like Ansible's `changed: true/false` pattern).

The field must propagate through the full stack: provider → worker → job
response → API → SDK → CLI.

## Context

Currently, mutation responses (202 Accepted) report success/failure but not
whether state was actually modified. For example, a DNS update that writes the
same values already configured returns the same response as one that changes
them. An automation layer needs this distinction to report "3 of 5 operations
changed, 2 already converged."

### Current mutating operations

| Operation     | Provider      | Endpoint              |
| ------------- | ------------- | --------------------- |
| DNS update    | `network/dns` | `PUT /network/dns`    |
| Command exec  | `command`     | `POST /command/exec`  |
| Command shell | `command`     | `POST /command/shell` |

Note: command exec/shell always "change" (they execute arbitrary commands), so
`changed` is always `true` for them. DNS update is the first truly idempotent
provider where `changed` has meaningful semantics.

## Implementation Plan

### Layer 1: Provider interface

**Files:**

- `internal/provider/network/dns/types.go`
- `internal/provider/network/dns/dns.go` (or platform impl)

Add a `Result` type to the DNS provider that includes `Changed`:

```go
type Result struct {
    Changed bool `json:"changed"`
}
```

Update `UpdateResolvConfByInterface` to return `(*Result, error)`. The
implementation compares current resolv.conf contents against desired before
writing — if identical, return `Changed: false` without writing.

**Command provider** (`internal/provider/command/types.go`): Add `Changed bool`
to the existing `Result` struct. Always set to `true` (commands always mutate by
definition).

### Layer 2: Worker processor

**File:** `internal/job/worker/processor.go`

Update `processNetworkDNS` and command processors to read `Changed` from the
provider result and include it in the result map written to KV.

### Layer 3: Job response type

**File:** `internal/job/types.go`

Add `Changed *bool` to `Response`:

```go
type Response struct {
    JobID     string          `json:"job_id"`
    Status    Status          `json:"status"`
    Changed   *bool           `json:"changed,omitempty"`
    Data      json.RawMessage `json:"data,omitempty"`
    Error     string          `json:"error,omitempty"`
    Hostname  string          `json:"hostname"`
    Timestamp time.Time       `json:"timestamp"`
}
```

Use `*bool` so read-only queries omit the field entirely (nil) rather than
showing `false`.

### Layer 4: OpenAPI specs

**Files:**

- `internal/api/network/gen/api.yaml`
- `internal/api/command/gen/api.yaml`

Add `changed` boolean to `DNSUpdateResultItem` and `CommandResultItem` schemas:

```yaml
changed:
  type: boolean
  description: Whether the operation modified system state.
```

Run `just generate` to regenerate `*.gen.go`.

### Layer 5: API handlers

**Files:**

- `internal/api/network/network_dns_put_by_interface.go`
- `internal/api/command/command_exec_post.go`
- `internal/api/command/command_shell_post.go`

Populate the `Changed` field from the job response when building the 202
response.

### Layer 6: SDK

**Repo:** `osapi-sdk`

- Update the merged OpenAPI spec and regenerate the client
- Expose `Changed` in the service response types

### Layer 7: CLI

**Files:**

- `cmd/client_network_dns_update.go` (or equivalent)
- `cmd/client_command_exec.go`
- `cmd/client_command_shell.go`

Display changed status in output:

```
Hostname:  web-01
Status:    ok
Changed:   true
```

For `--json` output, the field comes through naturally from the API response.

### Layer 8: Tests

- Provider unit tests: verify `Changed: false` when state matches,
  `Changed: true` when state differs
- Worker processor tests: verify `changed` propagates into result
- API integration tests: verify `changed` field present in 202 responses
- CLI: verify `Changed` appears in output

## Verification

```bash
just generate
go build ./...
just go::unit
just go::vet
```

## Notes

- Future providers that implement idempotent mutations (e.g., hostname set,
  network interface config) should follow this same pattern from day one.
- The `changed` field is the foundation for the planned automation layer
  (`osapi-apply`) which needs to aggregate change status across multiple
  operations.
