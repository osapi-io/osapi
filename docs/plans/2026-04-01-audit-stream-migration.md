# Audit Stream Migration

Migrate the audit store from NATS KV to a JetStream stream for
chronological ordering and efficient pagination.

## Problem

Audit entries are stored in a NATS KV bucket keyed by random UUID v4.
`List()` fetches all keys into memory, sorts them (incorrectly — UUIDs
don't sort chronologically), then paginates. With 1400+ entries and a
30-day TTL, this gets progressively slower and returns entries in
random order.

## Solution

Replace the KV bucket with a JetStream stream. Use ULIDs as message
subjects for chronological ordering and direct lookup. Add `trace_id`
to audit entries for OpenTelemetry correlation.

## Design

### Config Change

```yaml
# Before
nats:
  audit:
    bucket: 'audit-log'
    ttl: '720h'
    max_bytes: 52428800
    storage: 'file'
    replicas: 1

# After
nats:
  audit:
    stream: 'AUDIT'
    subject: 'audit'
    max_age: '720h'
    max_bytes: 52428800
    storage: 'file'
    replicas: 1
```

Fields renamed: `bucket` -> `stream`, `ttl` -> `max_age`. New field:
`subject` (base subject for audit messages). Drop `bucket` entirely.

### Audit Entry

Add one field to `Entry`:

```go
type Entry struct {
    ID           string    `json:"id"`
    Timestamp    time.Time `json:"timestamp"`
    User         string    `json:"user"`
    Roles        []string  `json:"roles,omitempty"`
    Method       string    `json:"method"`
    Path         string    `json:"path"`
    SourceIP     string    `json:"source_ip"`
    ResponseCode int       `json:"response_code"`
    DurationMs   int64     `json:"duration_ms"`
    OperationID  string    `json:"operation_id,omitempty"`
    TraceID      string    `json:"trace_id,omitempty"` // NEW
}
```

The `ID` field changes from UUID to ULID. This is the only breaking
change — no backward compatibility needed.

### Store Interface

The `Store` interface stays the same:

```go
type Store interface {
    Write(ctx context.Context, entry Entry) error
    Get(ctx context.Context, id string) (*Entry, error)
    List(ctx context.Context, limit int, offset int) ([]Entry, int, error)
    ListAll(ctx context.Context) ([]Entry, error)
}
```

### Stream Store Operations

| Operation | Implementation |
|-----------|---------------|
| Write | `js.Publish("audit.{ulid}", data)` |
| Get | `stream.GetMsg(ctx, &GetMsgRequest{NextFor: "audit.{id}"})` |
| List | `stream.Info()` for total count; ordered consumer with `DeliverByStartSequence` for pagination; read newest-first by computing start sequence from total - offset |
| ListAll | Ordered consumer from sequence 1, read all messages forward |
| Count | `stream.Info().State.Msgs` |

### Middleware Change

Extract trace ID from OpenTelemetry span context in the audit
middleware:

```go
spanCtx := trace.SpanContextFromContext(c.Request().Context())
if spanCtx.HasTraceID() {
    entry.TraceID = spanCtx.TraceID().String()
}
```

### Files Changed

Production code:
- `internal/audit/types.go` — add `TraceID` field to `Entry`
- `internal/audit/stream_store.go` — new stream-based `Store` impl
- `internal/audit/kv_store.go` — delete
- `internal/audit/mocks/` — regenerate
- `internal/config/types.go` — update audit config struct
- `internal/controller/api/middleware_audit.go` — add trace ID, use ULID
- `cmd/nats_setup.go` — create stream instead of KV bucket
- `cmd/controller_setup.go` — wire stream store
- `internal/controller/api/audit/gen/api.yaml` — add `trace_id` field
- `internal/controller/api/audit/audit_list.go` — update `mapEntryToGen`
- `internal/controller/api/audit/audit_get.go` — update if needed
- `pkg/sdk/client/audit_types.go` — add `TraceID` to SDK types
- `docs/docs/sidebar/usage/configuration.md` — update config reference

Test code:
- `internal/audit/stream_store_public_test.go` — new, 100% coverage
- `internal/audit/kv_store_test.go` — delete
- `internal/audit/kv_store_public_test.go` — delete
- `internal/controller/api/middleware_audit_public_test.go` — update
- Update all existing test files that reference changed types

### Coverage Baseline

All files below are currently at 100% coverage. The new
implementation must maintain 100%:

| File | Current |
|------|---------|
| `internal/audit/kv_store.go` | 100% -> new `stream_store.go` |
| `internal/audit/export/` | 100% |
| `internal/controller/api/audit/` | 100% |
| `internal/controller/api/middleware_audit.go` | 100% |
| `pkg/sdk/client/audit.go` | 100% |
| `pkg/sdk/client/audit_types.go` | 100% |

### Not Changing

- `internal/audit/export/` — export uses `ListAll()` via the `Store`
  interface, no changes needed
- OpenAPI spec for audit list/get/export — response shapes stay the
  same, just add `trace_id` field
- CLI commands — they consume SDK types, pick up `trace_id` via
  `--json` automatically
- Job KV, registry KV, state KV, facts KV — no changes
