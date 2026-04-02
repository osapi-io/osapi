# Audit Stream Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> superpowers:subagent-driven-development (recommended) or
> superpowers:executing-plans to implement this plan task-by-task. Steps
> use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the audit KV store with a JetStream stream for
chronological ordering and efficient pagination, add `trace_id` to
audit entries.

**Architecture:** The audit `StreamStore` receives a `jetstream.Stream`
handle (for reads: `GetLastMsgForSubject`, `OrderedConsumer`, `Info`)
and uses `nc.Publish()` (via a `Publisher` interface) for writes. The
`Store` interface is unchanged — all consumers (handlers, middleware,
export) work without modification. Config changes from KV bucket
fields to stream fields.

**Tech Stack:** Go, NATS JetStream streams, OpenTelemetry trace
context

---

### Task 1: Update config types and YAML

**Files:**
- Modify: `internal/config/types.go:124-132`
- Modify: `configs/osapi.yaml` (default config)
- Modify: `configs/osapi.nerd.yaml` (dev config)

- [ ] **Step 1: Update the NATSAudit config struct**

Replace the KV bucket config with stream config:

```go
// NATSAudit configuration for the audit log stream.
type NATSAudit struct {
	// Stream is the JetStream stream name for audit log entries.
	Stream   string `mapstructure:"stream"`
	// Subject is the base subject prefix for audit messages.
	Subject  string `mapstructure:"subject"`
	MaxAge   string `mapstructure:"max_age"` // e.g. "720h" (30 days)
	MaxBytes int64  `mapstructure:"max_bytes"`
	Storage  string `mapstructure:"storage"` // "file" or "memory"
	Replicas int    `mapstructure:"replicas"`
}
```

- [ ] **Step 2: Update osapi.yaml configs**

In both `configs/osapi.yaml` and `configs/osapi.nerd.yaml`, change the
`nats.audit` section:

```yaml
nats:
  audit:
    stream: 'AUDIT'
    subject: 'audit'
    max_age: '720h'
    max_bytes: 52428800
    storage: 'file'
    replicas: 1
```

- [ ] **Step 3: Verify it compiles**

Run: `go build ./...`

Expect: compile errors in files that reference `NATSAudit.Bucket` and
`NATSAudit.TTL` — that's expected, we fix them in subsequent tasks.

- [ ] **Step 4: Commit**

```
chore(config): rename audit config from KV bucket to stream
```

---

### Task 2: Update CLI config builder and NATS setup

**Files:**
- Modify: `internal/cli/nats.go:128-142`
- Modify: `internal/cli/nats_public_test.go` (update test for renamed
  function)
- Modify: `cmd/nats_setup.go:150-155`

- [ ] **Step 1: Replace BuildAuditKVConfig with BuildAuditStreamConfig**

In `internal/cli/nats.go`, replace `BuildAuditKVConfig`:

```go
// BuildAuditStreamConfig builds a jetstream.StreamConfig from audit
// config values.
func BuildAuditStreamConfig(
	namespace string,
	auditCfg config.NATSAudit,
) jetstream.StreamConfig {
	streamName := job.ApplyNamespaceToInfraName(
		namespace,
		auditCfg.Stream,
	)
	subject := job.ApplyNamespaceToSubjects(
		namespace,
		auditCfg.Subject,
	)
	maxAge, _ := time.ParseDuration(auditCfg.MaxAge)

	return jetstream.StreamConfig{
		Name:     streamName,
		Subjects: []string{subject + ".>"},
		MaxAge:   maxAge,
		MaxBytes: auditCfg.MaxBytes,
		Storage:  ParseJetstreamStorageType(auditCfg.Storage),
		Replicas: auditCfg.Replicas,
		Discard:  jetstream.DiscardOld,
	}
}
```

- [ ] **Step 2: Update the test for the renamed function**

In `internal/cli/nats_public_test.go`, update the test that exercises
`BuildAuditKVConfig` to test `BuildAuditStreamConfig` instead. The
test should verify stream name, subjects with `.>` suffix, max age,
storage type, and replicas.

- [ ] **Step 3: Update nats_setup.go to create stream instead of KV**

In `cmd/nats_setup.go`, replace the audit KV bucket creation block:

```go
if appConfig.NATS.Audit.Stream != "" {
	auditStreamConfig := cli.BuildAuditStreamConfig(
		namespace,
		appConfig.NATS.Audit,
	)
	if err := nc.CreateOrUpdateStreamWithConfig(
		ctx,
		auditStreamConfig,
	); err != nil {
		return fmt.Errorf(
			"create audit stream %s: %w",
			auditStreamConfig.Name,
			err,
		)
	}
}
```

- [ ] **Step 4: Run tests and verify build**

Run: `go test ./internal/cli/... -count=1`
Run: `go build ./...`

Expect: cli tests pass, build still has errors in controller_setup.go
(expected — fixed in Task 4).

- [ ] **Step 5: Commit**

```
feat(audit): replace KV bucket setup with stream creation
```

---

### Task 3: Add TraceID to audit entry and OpenAPI spec

**Files:**
- Modify: `internal/audit/types.go:27-48`
- Modify: `internal/controller/api/audit/gen/api.yaml:190-247`
- Modify: `internal/controller/api/audit/audit_list.go:75-94`
  (mapEntryToGen)
- Modify: `internal/controller/api/middleware_audit.go`
- Modify: `internal/controller/api/export_test.go` (if needed)
- Modify: `pkg/sdk/client/audit_types.go`

- [ ] **Step 1: Add TraceID to the audit Entry struct**

In `internal/audit/types.go`, add the field:

```go
type Entry struct {
	ID           string    `json:"id"`
	Timestamp    time.Time `json:"timestamp"`
	User         string    `json:"user"`
	Roles        []string  `json:"roles"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	OperationID  string    `json:"operation_id,omitempty"`
	SourceIP     string    `json:"source_ip"`
	ResponseCode int       `json:"response_code"`
	DurationMs   int64     `json:"duration_ms"`
	TraceID      string    `json:"trace_id,omitempty"`
}
```

- [ ] **Step 2: Add trace_id to the OpenAPI spec**

In `internal/controller/api/audit/gen/api.yaml`, add `trace_id` to the
`AuditEntry` schema properties (after `duration_ms`):

```yaml
        trace_id:
          type: string
          description: OpenTelemetry trace ID for correlation.
          example: "4bf92f3577b34da6a3ce929d0e0e4736"
```

Do NOT add it to `required` — it's optional (empty when tracing is
disabled).

- [ ] **Step 3: Regenerate OpenAPI code**

Run: `just generate`

- [ ] **Step 4: Update mapEntryToGen in audit_list.go**

Add the `TraceID` mapping in `mapEntryToGen`:

```go
if e.TraceID != "" {
	entry.TraceId = &e.TraceID
}
```

- [ ] **Step 5: Add trace ID extraction to the audit middleware**

In `internal/controller/api/middleware_audit.go`, add the import for
`go.opentelemetry.io/otel/trace` and extract the trace ID:

```go
spanCtx := trace.SpanContextFromContext(
	c.Request().Context(),
)
if spanCtx.HasTraceID() {
	entry.TraceID = spanCtx.TraceID().String()
}
```

Add this after building the `entry` struct and before the goroutine
that writes it.

- [ ] **Step 6: Add TraceID to SDK audit types**

In `pkg/sdk/client/audit_types.go`, add to `AuditEntry`:

```go
TraceID string `json:"trace_id,omitempty"`
```

Update `auditEntryFromGen` to map the field:

```go
if g.TraceId != nil {
	a.TraceID = *g.TraceId
}
```

- [ ] **Step 7: Run tests**

Run: `go test ./internal/controller/api/audit/... -count=1`
Run: `go test ./internal/controller/api/ -run Audit -count=1`
Run: `go test ./pkg/sdk/client/ -run Audit -count=1`

Expect: all pass. The middleware test uses a hand-written spy that
already accepts the new field (it stores the full `Entry`). The
trace ID will be empty in tests since there's no OTel span — that's
correct.

- [ ] **Step 8: Commit**

```
feat(audit): add trace_id field for OpenTelemetry correlation
```

---

### Task 4: Implement the stream store

**Files:**
- Create: `internal/audit/stream_store.go`
- Create: `internal/audit/stream_store_public_test.go`
- Modify: `internal/audit/export_test.go` (keep marshalJSON export)
- Delete: `internal/audit/kv_store.go`
- Delete: `internal/audit/kv_store_test.go`
- Delete: `internal/audit/kv_store_public_test.go`

- [ ] **Step 1: Create the StreamStore**

Create `internal/audit/stream_store.go`:

```go
package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go/jetstream"
)

// ensure StreamStore implements Store at compile time.
var _ Store = (*StreamStore)(nil)

// marshalJSON is a package-level variable for testing the marshal
// error path.
var marshalJSON = json.Marshal

// Publisher publishes messages to a NATS subject.
type Publisher interface {
	Publish(
		ctx context.Context,
		subject string,
		data []byte,
	) error
}

// StreamStore implements Store backed by a NATS JetStream stream.
type StreamStore struct {
	stream    jetstream.Stream
	publisher Publisher
	subject   string
	logger    *slog.Logger
}

// NewStreamStore creates a new StreamStore with the given
// dependencies. The subject is the base prefix (e.g., "audit");
// messages are published to "audit.{id}".
func NewStreamStore(
	logger *slog.Logger,
	stream jetstream.Stream,
	publisher Publisher,
	subject string,
) *StreamStore {
	return &StreamStore{
		stream:    stream,
		publisher: publisher,
		subject:   subject,
		logger:    logger.With(slog.String("subsystem", "audit")),
	}
}

// Write persists an audit entry to the stream.
func (s *StreamStore) Write(
	ctx context.Context,
	entry Entry,
) error {
	data, err := marshalJSON(entry)
	if err != nil {
		return fmt.Errorf("marshal audit entry: %w", err)
	}

	subject := s.subject + "." + entry.ID
	if err := s.publisher.Publish(ctx, subject, data); err != nil {
		return fmt.Errorf("publish audit entry: %w", err)
	}

	return nil
}

// Get retrieves a single audit entry by ID using subject lookup.
func (s *StreamStore) Get(
	ctx context.Context,
	id string,
) (*Entry, error) {
	subject := s.subject + "." + id

	msg, err := s.stream.GetLastMsgForSubject(ctx, subject)
	if err != nil {
		return nil, fmt.Errorf(
			"get audit entry: not found: %w",
			err,
		)
	}

	var entry Entry
	if err := json.Unmarshal(msg.Data, &entry); err != nil {
		return nil, fmt.Errorf("unmarshal audit entry: %w", err)
	}

	return &entry, nil
}

// List retrieves audit entries with pagination, newest first.
// Uses the stream's message count for total and an ordered consumer
// for efficient sequential reads.
func (s *StreamStore) List(
	ctx context.Context,
	limit int,
	offset int,
) ([]Entry, int, error) {
	info, err := s.stream.Info(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("get stream info: %w", err)
	}

	total := int(info.State.Msgs)
	if total == 0 || offset >= total {
		return []Entry{}, total, nil
	}

	// For newest-first: read from the end.
	// We want entries at positions [total-offset-limit .. total-offset)
	// mapped to stream sequences [first .. last].
	startIdx := total - offset - limit
	if startIdx < 0 {
		startIdx = 0
	}
	count := total - offset - startIdx

	startSeq := info.State.FirstSeq + uint64(startIdx)

	consumer, err := s.stream.OrderedConsumer(
		ctx,
		jetstream.OrderedConsumerConfig{
			DeliverPolicy: jetstream.DeliverByStartSequencePolicy,
			OptStartSeq:   startSeq,
		},
	)
	if err != nil {
		return nil, 0, fmt.Errorf("create ordered consumer: %w", err)
	}

	entries := make([]Entry, 0, count)

	fetchCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	batch, err := consumer.Fetch(count, jetstream.FetchMaxWait(fetchTimeout))
	if err != nil {
		return nil, 0, fmt.Errorf("fetch audit entries: %w", err)
	}

	for msg := range batch.Messages() {
		var entry Entry
		if err := json.Unmarshal(msg.Data(), &entry); err != nil {
			s.logger.Warn(
				"failed to unmarshal audit entry",
				slog.String("error", err.Error()),
			)

			continue
		}

		entries = append(entries, entry)
	}

	if batchErr := batch.Error(); batchErr != nil {
		s.logger.Warn(
			"batch fetch error",
			slog.String("error", batchErr.Error()),
		)
	}

	_ = fetchCtx

	// Reverse for newest-first order.
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	return entries, total, nil
}

// ListAll retrieves all audit entries, newest first.
func (s *StreamStore) ListAll(
	ctx context.Context,
) ([]Entry, error) {
	info, err := s.stream.Info(ctx)
	if err != nil {
		return nil, fmt.Errorf("get stream info: %w", err)
	}

	total := int(info.State.Msgs)
	if total == 0 {
		return []Entry{}, nil
	}

	consumer, err := s.stream.OrderedConsumer(
		ctx,
		jetstream.OrderedConsumerConfig{
			DeliverPolicy: jetstream.DeliverAllPolicy,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create ordered consumer: %w", err)
	}

	entries := make([]Entry, 0, total)

	batch, err := consumer.Fetch(
		total,
		jetstream.FetchMaxWait(fetchTimeout),
	)
	if err != nil {
		return nil, fmt.Errorf("fetch audit entries: %w", err)
	}

	for msg := range batch.Messages() {
		var entry Entry
		if err := json.Unmarshal(msg.Data(), &entry); err != nil {
			s.logger.Warn(
				"failed to unmarshal audit entry",
				slog.String("error", err.Error()),
			)

			continue
		}

		entries = append(entries, entry)
	}

	if batchErr := batch.Error(); batchErr != nil {
		s.logger.Warn(
			"batch fetch error",
			slog.String("error", batchErr.Error()),
		)
	}

	// Reverse for newest-first order.
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	return entries, nil
}
```

Also add a `fetchTimeout` constant at the top of the file:

```go
const fetchTimeout = 5 * time.Second
```

And add `"time"` to the imports.

- [ ] **Step 2: Update export_test.go**

The `export_test.go` file currently exports `SetMarshalJSON` /
`ResetMarshalJSON` for the `marshalJSON` var in `kv_store.go`. Since
`stream_store.go` declares the same `marshalJSON` var, the export
test file works unchanged. Verify it still compiles.

- [ ] **Step 3: Write the stream store tests**

Create `internal/audit/stream_store_public_test.go`. Use gomock to
mock `jetstream.Stream` and the `Publisher` interface. Follow the
exact same test structure as `kv_store_public_test.go`:

Test `Write`:
- successfully publishes entry (mock publisher expects
  `Publish(ctx, "audit.{id}", data)`)
- returns error when publish fails
- returns error when marshal fails (via `SetMarshalJSON`)

Test `Get`:
- successfully gets entry (mock stream `GetLastMsgForSubject` returns
  `RawStreamMsg` with valid JSON data)
- returns error containing "not found" when subject not found
- returns error when unmarshal fails (bad JSON data)

Test `List`:
- returns all entries newest-first when within limit
- applies pagination correctly (offset + limit)
- returns empty when offset exceeds total
- returns empty for empty stream (Info returns `State.Msgs == 0`)
- returns error when stream info fails
- skips entries when unmarshal fails (bad JSON in batch)

Test `ListAll`:
- returns all entries newest-first
- returns empty for empty stream
- returns error when stream info fails
- skips entries when unmarshal fails

For mocking `jetstream.Stream`: create a mock interface in
`internal/audit/mocks/` using `go:generate mockgen`. The mock needs
`Info`, `GetLastMsgForSubject`, and `OrderedConsumer` methods.

For mocking the `Publisher` interface: add a `go:generate mockgen`
directive for the `Publisher` interface defined in `stream_store.go`.

For mocking `jetstream.Consumer` (returned by `OrderedConsumer`): mock
its `Fetch` method which returns `MessageBatch`. Mock `MessageBatch`
for its `Messages()` channel and `Error()` method.

Target: **100% coverage** on `stream_store.go`.

- [ ] **Step 4: Delete old KV store files**

Delete:
- `internal/audit/kv_store.go`
- `internal/audit/kv_store_test.go`
- `internal/audit/kv_store_public_test.go`

- [ ] **Step 5: Regenerate mocks**

Update `internal/audit/mocks/generate.go` to generate mocks for the
new interfaces:

```go
//go:generate go tool github.com/golang/mock/mockgen -source=../store.go -destination=store.gen.go -package=mocks
//go:generate go tool github.com/golang/mock/mockgen -source=../stream_store.go -destination=publisher.gen.go -package=mocks -mock_names=Publisher=MockPublisher
```

Run: `go generate ./internal/audit/mocks/...`

Also generate mocks for the `jetstream.Stream` and
`jetstream.Consumer` interfaces used in tests. These can live in
`internal/audit/mocks/` or use `gomock`'s reflect mode for the
`jetstream` package interfaces. Check how existing tests in the
codebase mock `jetstream` interfaces (e.g., the `job/mocks` package)
and follow the same pattern.

- [ ] **Step 6: Run tests and check coverage**

Run: `go test ./internal/audit/... -count=1 -coverprofile=/tmp/audit.out`
Run: `go tool cover -func=/tmp/audit.out | grep stream_store`

Expect: 100% coverage on `stream_store.go`.

- [ ] **Step 7: Commit**

```
feat(audit): implement stream-based audit store
```

---

### Task 5: Wire stream store in controller setup

**Files:**
- Modify: `cmd/controller_setup.go:640-658`

- [ ] **Step 1: Replace createAuditStore to use stream**

Replace the `createAuditStore` function:

```go
func createAuditStore(
	ctx context.Context,
	log *slog.Logger,
	nc NATSClient,
	namespace string,
) (audit.Store, []api.Option) {
	if appConfig.NATS.Audit.Stream == "" {
		return nil, nil
	}

	auditStreamConfig := cli.BuildAuditStreamConfig(
		namespace,
		appConfig.NATS.Audit,
	)
	if err := nc.CreateOrUpdateStreamWithConfig(
		ctx,
		auditStreamConfig,
	); err != nil {
		cli.LogFatal(log, "failed to create audit stream", err)
	}

	streamName := job.ApplyNamespaceToInfraName(
		namespace,
		appConfig.NATS.Audit.Stream,
	)
	stream, err := nc.Stream(ctx, streamName)
	if err != nil {
		cli.LogFatal(log, "failed to get audit stream", err)
	}

	subject := job.ApplyNamespaceToSubjects(
		namespace,
		appConfig.NATS.Audit.Subject,
	)

	store := audit.NewStreamStore(log, stream, nc, subject)

	return store, []api.Option{api.WithAuditStore(store)}
}
```

Note: `nc` (the `NATSClient`) satisfies `audit.Publisher` since it
has a `Publish(ctx, subject, data) error` method. Verify the method
signature matches. If the `NATSClient` interface's `Publish` method
signature matches `audit.Publisher`, pass it directly. If not, create
a thin adapter.

- [ ] **Step 2: Verify build**

Run: `go build ./...`

Expect: clean build.

- [ ] **Step 3: Run full test suite**

Run: `just go::unit`

Expect: all tests pass.

- [ ] **Step 4: Commit**

```
feat(audit): wire stream store in controller setup
```

---

### Task 6: Update documentation

**Files:**
- Modify: `docs/docs/sidebar/usage/configuration.md`
- Modify: `docs/docs/sidebar/features/audit-logging.md`

- [ ] **Step 1: Update configuration.md**

Replace the `nats.audit` section in the full reference YAML:

```yaml
  audit:
    # JetStream stream name for audit log entries.
    stream: 'AUDIT'
    # Base subject prefix for audit messages.
    subject: 'audit'
    # Maximum age of audit entries (Go duration). Default 30 days.
    max_age: '720h'
    # Maximum total size of the audit stream in bytes.
    max_bytes: 52428800 # 50 MiB
    # Storage backend: "file" or "memory".
    storage: 'file'
    # Number of stream replicas.
    replicas: 1
```

Update the `nats.audit` section reference table:

| Key         | Type   | Description                           |
| ----------- | ------ | ------------------------------------- |
| `stream`    | string | JetStream stream name for audit logs  |
| `subject`   | string | Base subject prefix for audit msgs    |
| `max_age`   | string | Maximum entry age (Go duration)       |
| `max_bytes` | int    | Maximum stream size in bytes          |
| `storage`   | string | `"file"` or `"memory"`                |
| `replicas`  | int    | Number of stream replicas             |

Update the environment variable table — replace `OSAPI_NATS_AUDIT_BUCKET`
and `OSAPI_NATS_AUDIT_TTL` with `OSAPI_NATS_AUDIT_STREAM`,
`OSAPI_NATS_AUDIT_SUBJECT`, and `OSAPI_NATS_AUDIT_MAX_AGE`.

- [ ] **Step 2: Update audit-logging.md if it mentions KV**

Check `docs/docs/sidebar/features/audit-logging.md` for references to
"KV bucket" or "bucket" and update to "stream". Add a note about the
`trace_id` field.

- [ ] **Step 3: Commit**

```
docs(audit): update config reference for stream migration
```

---

### Task 7: Final verification

- [ ] **Step 1: Run full test suite with coverage**

```bash
go test ./internal/audit/... -coverprofile=/tmp/audit.out -count=1
go tool cover -func=/tmp/audit.out | grep -v mocks
```

Verify 100% on `stream_store.go`.

```bash
go test ./internal/controller/api/audit/... -count=1
go test ./internal/controller/api/ -run Audit -count=1
go test ./pkg/sdk/client/ -run Audit -count=1
```

All must pass.

- [ ] **Step 2: Build and lint**

```bash
go build ./...
just go::vet
```

- [ ] **Step 3: Verify no references to old KV audit remain**

```bash
grep -r "AuditKV\|BuildAuditKVConfig\|NewKVStore\|audit.*bucket\|kv_store" \
  --include="*.go" internal/ cmd/ pkg/ | grep -v _test.go | grep -v mocks
```

Expect: no matches.

- [ ] **Step 4: Run docs formatting**

```bash
just docs::fmt-check
```

Fix any formatting issues.
