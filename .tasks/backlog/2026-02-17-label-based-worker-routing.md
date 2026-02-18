---
title: Label-based worker routing with NATS subject wildcards
status: backlog
created: 2026-02-17
updated: 2026-02-17
---

## Objective

Extend worker targeting beyond `_any`, `_all`, and exact hostname. Admins need
to target groups of servers (e.g., all web servers, all prod machines, all
servers in rack-3). The solution should leverage NATS native subject wildcards
for zero-overhead routing.

## Problem

Today the `--target` flag accepts only:

- `_any` — load-balanced to one random worker
- `_all` — broadcast to every worker
- `server1` — direct to a specific hostname

There is no way to say "all web servers" or "every prod machine in us-east-1."

## Proposed Architecture: Labels as Subject Segments

### Core idea

Workers publish their identity through the subjects they subscribe to. Instead
of a flat `jobs.{type}.{hostname}`, add label segments that NATS wildcards can
match against.

### Subject format

```
jobs.{type}.host.{hostname}     — direct to specific host
jobs.{type}.label.{key}.{value} — all hosts with that label
jobs.{type}._any                — any worker (load-balanced)
jobs.{type}._all                — broadcast all workers
```

Examples:

```
jobs.query.host.web-prod-01
jobs.query.label.role.web
jobs.query.label.env.prod
jobs.query.label.rack.us-east-1a
jobs.modify.label.role.db
```

### Worker config

```yaml
job:
  worker:
    hostname: web-prod-01    # optional, auto-detected if empty
    labels:                  # NEW
      role: web
      env: prod
      rack: us-east-1a
```

### Worker subscription behavior

On startup, a worker with the above config subscribes to:

```
# Existing patterns
jobs.*.host.web-prod-01          — direct messages
jobs.*._any                      — load-balanced (queue group)
jobs.*._all                      — broadcasts

# NEW: one subscription per label
jobs.*.label.role.web            — all "role=web" jobs
jobs.*.label.env.prod            — all "env=prod" jobs
jobs.*.label.rack.us-east-1a    — all "rack=us-east-1a" jobs
```

Label subscriptions use **no queue group** so every matching worker gets the
message (broadcast semantics within the label group). If the admin wants
load-balanced label routing (send to one web server, not all), they could use
a queue group per label — but that's a future enhancement.

### Client targeting

The `--target` flag (and `target_hostname` query param) expands:

| Target value            | Resolves to subject               | Semantics          |
| ----------------------- | --------------------------------- | -------------------|
| `_any`                  | `jobs.{type}._any`                | Load-balanced      |
| `_all`                  | `jobs.{type}._all`                | Broadcast all      |
| `web-prod-01`           | `jobs.{type}.host.web-prod-01`    | Direct to host     |
| `role:web`              | `jobs.{type}.label.role.web`      | Broadcast to label |
| `env:prod`              | `jobs.{type}.label.env.prod`      | Broadcast to label |

The `key:value` syntax is unambiguous — hostnames cannot contain `:`.

### Why labels over hierarchical hostnames

Hierarchical hostnames (`prod.web.server1`) force a single taxonomy. A server
can only be in one hierarchy. Labels are multi-dimensional — a server can be
`role:web` AND `env:prod` AND `rack:us-east-1a` simultaneously. The admin
can target any dimension without restructuring naming conventions.

### Why not a registration/discovery service

NATS subject routing IS the discovery mechanism. Workers self-register by
subscribing to their label subjects. No external registry, no heartbeats, no
consistency problem. If a worker is running, its subscriptions are active. If
it dies, NATS removes the subscriptions. This is the simplest architecture
that could possibly work.

## Implementation Plan

### 1. Config changes

**File:** `internal/config/types.go`

```go
type JobWorker struct {
    // ... existing fields ...
    Labels map[string]string `mapstructure:"labels"` // NEW
}
```

### 2. Subject routing

**File:** `internal/job/subjects.go`

Add:

```go
func BuildLabelSubject(key, value string) string {
    return fmt.Sprintf("jobs.*.label.%s.%s", key, value)
}

func BuildHostSubject(hostname string) string {
    return fmt.Sprintf("jobs.*.host.%s", hostname)
}

func ParseTarget(target string) (subjectType, key, value string) {
    // "_any" → ("_any", "", "")
    // "_all" → ("_all", "", "")
    // "role:web" → ("label", "role", "web")
    // "server1" → ("host", "server1", "")
}
```

**Validation:** Label keys and values must be `[a-zA-Z0-9_-]+` (NATS subject
token safe). Reject dots, spaces, wildcards.

### 3. Worker subscriptions

**File:** `internal/job/worker/consumer.go`

Extend consumer creation to loop over `w.appConfig.Job.Worker.Labels` and
create a consumer + goroutine for each `label.{key}.{value}` pattern.

### 4. Stream subjects

**File:** Config / stream setup

Update JetStream stream subjects to include the new patterns:

```
jobs.query.>
jobs.modify.>
```

The `>` wildcard already covers any depth, so this should already work if the
stream is configured with `jobs.>` or similar. Verify the current
`StreamSubjects` config value.

### 5. Client-side targeting

**File:** `internal/job/client/query.go`, `modify.go`

Update `BuildQuerySubject` / `BuildModifySubject` calls to parse the target
and build the correct subject. The `publishAndCollect` method already handles
multi-response collection, so label targeting works like `_all`.

**File:** `cmd/` CLI files

Update `--target` flag help text and validation.

### 6. API changes

**File:** OpenAPI specs

Update `target_hostname` parameter description and validation to accept
`key:value` label syntax. Consider renaming to `target` in a future version.

### 7. Worker discovery

**File:** `internal/job/client/query.go` — `ListWorkers`

Extend worker discovery to optionally filter by label. When listing workers,
each worker's response should include its labels so the admin can see the
topology.

## Migration

- **Backwards compatible:** Existing configs with no `labels` key work
  unchanged. The bare hostname targeting becomes `host.{hostname}` internally
  but the `--target server1` syntax stays the same.
- **Stream subjects:** If currently `jobs.query.*` (single token wildcard),
  must widen to `jobs.query.>` (multi-token). This is a one-time migration on
  stream recreation.

## Future Enhancements (out of scope)

- **Multi-label targeting:** `role:web,env:prod` (AND semantics — requires
  client-side intersection of results)
- **Load-balanced label routing:** `role:web:any` to pick one web server
  (queue group per label)
- **Label wildcards:** `env:*` to target all environments
- **Admin CLI for label management:** `osapi admin workers list --label
  role:web`
- **Dynamic label registration:** Workers can update labels at runtime via
  NATS

## Notes

- NATS subject tokens cannot contain `.` so label values like `us-east-1` are
  fine but `us.east.1` would break subject parsing. Use hyphens or underscores.
- Label-based subscriptions scale linearly with the number of unique labels
  per worker. A worker with 5 labels creates 5 additional consumers. This is
  well within NATS limits.
- Consider documenting recommended label taxonomies: `role`, `env`, `region`,
  `rack`, `team`.
