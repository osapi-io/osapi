---
title: Add consistent host targeting to all CLI commands
status: done
created: 2026-02-17
updated: 2026-02-17
---

## Objective

The CLI has an inconsistent targeting story. Job commands (`job add`,
`job run`) support `--target-hostname` for `_any`/`_all`/specific host
routing. All other commands (`system status`, `network dns get`, etc.)
have no targeting — they hardcode `_any` through the REST API layer.

This means users cannot:

- Query system status from a specific host
- Query system status from all hosts
- Run a DNS lookup or ping from a specific host
- Run any network operation against all hosts

The job client already supports all routing modes internally. The gap is
purely in the CLI/REST API surface.

## Design

### Approach: `--host` flag on client subcommand

Add a `--host` persistent flag to the `client` parent command, inherited
by all subcommands:

```bash
# Default: _any (load-balanced, one response)
osapi client system status

# Target specific host
osapi client --host server1 system status

# Query all hosts
osapi client --host _all system status
```

This gives consistent behavior across all commands with minimal flag
clutter. The `--host` flag would:

- Default to `_any` when omitted
- Accept `_all` for broadcast queries
- Accept any specific hostname
- Be inherited by all subcommands automatically

### Implementation Layers

#### Layer 1: REST API endpoints accept hostname parameter

Add optional `target_hostname` query parameter to REST API endpoints:

```
GET /system/status?target_hostname=server1
GET /system/status?target_hostname=_all
GET /network/dns/{interface}?target_hostname=server1
POST /network/ping?target_hostname=_all
```

The API handlers would pass this to the appropriate job client method
instead of hardcoding `_any`. When `_all` is specified, use the
`QuerySystemStatusAll()` / broadcast methods.

#### Layer 2: REST client handler interfaces accept hostname

Update the `SystemHandler` and `NetworkHandler` interfaces to accept a
hostname parameter:

```go
type SystemHandler interface {
    GetSystemStatus(ctx context.Context, hostname string) (...)
    GetSystemHostname(ctx context.Context, hostname string) (...)
}
```

#### Layer 3: CLI commands pass --host flag value

Each CLI command reads the `--host` flag from the parent and passes it
through:

```go
host, _ := cmd.Flags().GetString("host")
resp, err := systemHandler.GetSystemStatus(ctx, host)
```

#### Layer 4: Response formatting for _all

When `--host _all` is used, the CLI needs to format multi-host
responses. For system status, this could be a table:

```
HOSTNAME    STATUS     UPTIME    LOAD
server1     ok         3d 2h     0.5
server2     ok         1d 8h     1.2
server3     degraded   5d 0h     4.8
```

### Migration

- `--target-hostname` on `job add`/`job run` stays as-is (different
  semantic — raw job targeting)
- `--host` on `client` is the user-friendly equivalent for high-level
  commands
- Both use the same routing infrastructure underneath

### Open Questions

- Should `--host` also be configurable via `osapi.yaml` as a default?
  (e.g., `default_target_host: _any`)
- Should `_all` responses stream as they arrive or wait for timeout?
- For modify operations with `_all` (e.g., DNS update on all hosts),
  should we require explicit confirmation?
- Should response formatting differ between `_any` (single result)
  and `_all` (table/list)?

## Notes

- The `publishAndCollect()` method and `QuerySystemStatusAll()` are
  already implemented as of 2026-02-17.
- The REST API `GET /job/{id}` already exposes per-worker `responses`
  and `worker_states` for broadcast jobs.
- Each REST API domain (system, network) would need its own `_all`
  query method or a generic pattern.
