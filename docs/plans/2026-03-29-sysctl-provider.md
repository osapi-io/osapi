# Sysctl Provider Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> superpowers:subagent-driven-development (recommended) or
> superpowers:executing-plans to implement this plan task-by-task. Steps use
> checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add kernel parameter management (sysctl) as a meta-provider that
delegates file writes to `file.Deployer` for SHA tracking and idempotency, with
full API/CLI/SDK support.

**Architecture:** Sysctl is a meta-provider under
`internal/provider/node/sysctl/` that generates `/etc/sysctl.d/osapi-{key}.conf`
files and applies them via `sysctl -p`. It follows the cron provider pattern:
delegate to `file.Deployer`, store domain metadata in `FileState.Metadata`, scan
file-state KV to list managed entries. The API lives under
`/node/{hostname}/sysctl` with broadcast support.

**Tech Stack:** Go 1.25, Echo, oapi-codegen (strict-server), NATS JetStream KV,
gomock, testify/suite, avfs

**Coverage baseline:** 99.9% — must remain at or above this after
implementation.

---

## File Map

### New Files

```
internal/provider/node/sysctl/
  types.go                          — Provider interface + Entry, SetResult, DeleteResult
  debian.go                         — Debian meta-provider (delegates to file.Deployer)
  darwin.go                         — macOS stub (ErrUnsupported)
  linux.go                          — Generic Linux stub (ErrUnsupported)
  export_test.go                    — Expose unexported vars for test package
  mocks/
    generate.go                     — //go:generate mockgen directive

internal/agent/processor_sysctl.go  — NewSysctlProcessor factory + operation dispatch

internal/controller/api/sysctl/
  types.go                          — Handler struct + dependencies
  sysctl.go                         — New() factory + interface check
  sysctl_list_get.go                — GET /sysctl handler + broadcast
  sysctl_get.go                     — GET /sysctl/{key} handler + broadcast
  sysctl_set.go                     — POST /sysctl handler + broadcast
  sysctl_delete.go                  — DELETE /sysctl/{key} handler + broadcast
  validate.go                       — validateHostname + validateSysctlKey
  sysctl_list_get_public_test.go    — Tests for list handler
  sysctl_get_public_test.go         — Tests for get handler
  sysctl_set_public_test.go         — Tests for set handler
  sysctl_delete_public_test.go      — Tests for delete handler
  gen/
    api.yaml                        — OpenAPI spec
    cfg.yaml                        — oapi-codegen config
    generate.go                     — //go:generate directive

internal/controller/api/handler_sysctl.go — GetSysctlHandler() method

pkg/sdk/client/
  sysctl.go                         — SysctlService methods
  sysctl_types.go                   — SDK result types + gen→SDK conversions
  sysctl_public_test.go             — SDK service tests
  sysctl_types_public_test.go       — SDK type conversion tests

cmd/
  client_node_sysctl.go             — Parent command
  client_node_sysctl_list.go        — list subcommand
  client_node_sysctl_get.go         — get subcommand
  client_node_sysctl_set.go         — set subcommand
  client_node_sysctl_delete.go      — delete subcommand

examples/sdk/client/sysctl.go       — SDK example

test/integration/sysctl_test.go     — Integration smoke tests

docs/docs/sidebar/features/sysctl.md
docs/docs/sidebar/usage/cli/client/node/sysctl/sysctl.md
docs/docs/sidebar/usage/cli/client/node/sysctl/list.md
docs/docs/sidebar/usage/cli/client/node/sysctl/get.md
docs/docs/sidebar/usage/cli/client/node/sysctl/set.md
docs/docs/sidebar/usage/cli/client/node/sysctl/delete.md
```

### Modified Files

```
pkg/sdk/client/operations.go          — Add OpSysctl* constants
pkg/sdk/client/permissions.go         — Add PermSysctlRead/Write
pkg/sdk/client/osapi.go               — Wire SysctlService into Client
internal/job/types.go                  — Add OperationSysctl* re-exports
internal/authtoken/permissions.go      — Re-export + add to roles
internal/controller/api/types.go       — Add sysctlHandler field (if needed)
internal/controller/api/handler.go     — Wire in CreateHandlers (if needed)
cmd/controller_setup.go                — Append GetSysctlHandler + register provider
cmd/agent_setup.go                     — Create + register sysctl provider
internal/controller/api/handler_public_test.go — Add TestGetSysctlHandler
docs/docusaurus.config.ts             — Add to Features navbar
docs/docs/sidebar/usage/configuration.md — Add sysctl permissions
```

---

## Task 1: SDK Constants (Operations + Permissions)

**Files:**

- Modify: `pkg/sdk/client/operations.go`
- Modify: `pkg/sdk/client/permissions.go`

- [ ] **Step 1: Add sysctl operation constants**

In `pkg/sdk/client/operations.go`, add after the Schedule/Cron block:

```go
// Sysctl operations.
const (
	OpSysctlList   JobOperation = "node.sysctl.list"
	OpSysctlGet    JobOperation = "node.sysctl.get"
	OpSysctlSet    JobOperation = "node.sysctl.set"
	OpSysctlDelete JobOperation = "node.sysctl.delete"
)
```

- [ ] **Step 2: Add sysctl permission constants**

In `pkg/sdk/client/permissions.go`, add after `PermCronWrite`:

```go
	PermSysctlRead  Permission = "sysctl:read"
	PermSysctlWrite Permission = "sysctl:write"
```

- [ ] **Step 3: Re-export in internal/job/types.go**

Add after the Schedule/Cron operations block:

```go
// Sysctl operations.
const (
	OperationSysctlList   = client.OpSysctlList
	OperationSysctlGet    = client.OpSysctlGet
	OperationSysctlSet    = client.OpSysctlSet
	OperationSysctlDelete = client.OpSysctlDelete
)
```

- [ ] **Step 4: Re-export permissions in internal/authtoken/permissions.go**

Add constants:

```go
	PermSysctlRead  = client.PermSysctlRead
	PermSysctlWrite = client.PermSysctlWrite
```

Add to `AllPermissions` slice:

```go
	PermSysctlRead,
	PermSysctlWrite,
```

Add to `DefaultRolePermissions`:

- `RoleAdmin`: add `PermSysctlRead, PermSysctlWrite`
- `RoleWrite`: add `PermSysctlRead, PermSysctlWrite`
- `RoleRead`: add `PermSysctlRead`

- [ ] **Step 5: Verify it compiles**

Run: `go build ./...` Expected: clean build

- [ ] **Step 6: Commit**

```bash
git add pkg/sdk/client/operations.go pkg/sdk/client/permissions.go \
  internal/job/types.go internal/authtoken/permissions.go
git commit -m "feat(sysctl): add operation and permission constants"
```

---

## Task 2: Provider Interface + Platform Stubs

**Files:**

- Create: `internal/provider/node/sysctl/types.go`
- Create: `internal/provider/node/sysctl/darwin.go`
- Create: `internal/provider/node/sysctl/linux.go`
- Create: `internal/provider/node/sysctl/mocks/generate.go`

- [ ] **Step 1: Create types.go**

```go
// Package sysctl provides kernel parameter management via /etc/sysctl.d/.
// It is a meta-provider that delegates file writes to the file provider
// for SHA tracking, idempotency, and drift detection.
package sysctl

import "context"

// Provider implements the methods to manage sysctl entries.
type Provider interface {
	// List returns all osapi-managed sysctl entries with current runtime values.
	List(ctx context.Context) ([]Entry, error)
	// Get returns a single sysctl entry by key with current runtime value.
	Get(ctx context.Context, key string) (*Entry, error)
	// Set deploys a sysctl conf file and applies it. Idempotent.
	Set(ctx context.Context, entry Entry) (*SetResult, error)
	// Delete removes a managed sysctl conf file and reloads defaults.
	Delete(ctx context.Context, key string) (*DeleteResult, error)
}

// Entry represents a sysctl kernel parameter.
type Entry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// SetResult represents the outcome of a sysctl set operation.
type SetResult struct {
	Key     string `json:"key"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// DeleteResult represents the outcome of a sysctl delete operation.
type DeleteResult struct {
	Key     string `json:"key"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}
```

- [ ] **Step 2: Create darwin.go**

```go
package sysctl

import (
	"context"
	"fmt"

	"github.com/retr0h/osapi/internal/provider"
)

// Darwin implements the sysctl Provider interface for macOS.
// All methods return ErrUnsupported.
type Darwin struct{}

// NewDarwinProvider factory to create a new Darwin instance.
func NewDarwinProvider() *Darwin {
	return &Darwin{}
}

// List returns ErrUnsupported on macOS.
func (d *Darwin) List(
	_ context.Context,
) ([]Entry, error) {
	return nil, fmt.Errorf("sysctl: %w", provider.ErrUnsupported)
}

// Get returns ErrUnsupported on macOS.
func (d *Darwin) Get(
	_ context.Context,
	_ string,
) (*Entry, error) {
	return nil, fmt.Errorf("sysctl: %w", provider.ErrUnsupported)
}

// Set returns ErrUnsupported on macOS.
func (d *Darwin) Set(
	_ context.Context,
	_ Entry,
) (*SetResult, error) {
	return nil, fmt.Errorf("sysctl: %w", provider.ErrUnsupported)
}

// Delete returns ErrUnsupported on macOS.
func (d *Darwin) Delete(
	_ context.Context,
	_ string,
) (*DeleteResult, error) {
	return nil, fmt.Errorf("sysctl: %w", provider.ErrUnsupported)
}
```

- [ ] **Step 3: Create linux.go** (same pattern as darwin.go with Linux
      struct/factory)

```go
package sysctl

import (
	"context"
	"fmt"

	"github.com/retr0h/osapi/internal/provider"
)

// Linux implements the sysctl Provider interface for generic Linux.
// All methods return ErrUnsupported.
type Linux struct{}

// NewLinuxProvider factory to create a new Linux instance.
func NewLinuxProvider() *Linux {
	return &Linux{}
}

// List returns ErrUnsupported on generic Linux.
func (l *Linux) List(
	_ context.Context,
) ([]Entry, error) {
	return nil, fmt.Errorf("sysctl: %w", provider.ErrUnsupported)
}

// Get returns ErrUnsupported on generic Linux.
func (l *Linux) Get(
	_ context.Context,
	_ string,
) (*Entry, error) {
	return nil, fmt.Errorf("sysctl: %w", provider.ErrUnsupported)
}

// Set returns ErrUnsupported on generic Linux.
func (l *Linux) Set(
	_ context.Context,
	_ Entry,
) (*SetResult, error) {
	return nil, fmt.Errorf("sysctl: %w", provider.ErrUnsupported)
}

// Delete returns ErrUnsupported on generic Linux.
func (l *Linux) Delete(
	_ context.Context,
	_ string,
) (*DeleteResult, error) {
	return nil, fmt.Errorf("sysctl: %w", provider.ErrUnsupported)
}
```

- [ ] **Step 4: Create mocks/generate.go**

```go
// Package mocks contains generated mocks for the sysctl provider.
package mocks

//go:generate go tool github.com/golang/mock/mockgen -source=../types.go -destination=provider.gen.go -package=mocks
```

- [ ] **Step 5: Generate mocks**

Run: `go generate ./internal/provider/node/sysctl/mocks/...` Expected:
`mocks/provider.gen.go` created

- [ ] **Step 6: Verify it compiles**

Run: `go build ./...` Expected: clean build

- [ ] **Step 7: Commit**

```bash
git add internal/provider/node/sysctl/
git commit -m "feat(sysctl): add provider interface and platform stubs"
```

---

## Task 3: Debian Provider Implementation

**Files:**

- Create: `internal/provider/node/sysctl/debian.go`
- Create: `internal/provider/node/sysctl/export_test.go`
- Test: `internal/provider/node/sysctl/debian_public_test.go`
- Test: `internal/provider/node/sysctl/darwin_public_test.go`
- Test: `internal/provider/node/sysctl/linux_public_test.go`

The Debian provider is a meta-provider. It:

- Generates conf file content from key/value pairs
- Delegates file writes to `file.Deployer`
- Stores `key` and `value` in `FileState.Metadata`
- Uses `exec.Manager` to run `sysctl -p` and `sysctl -n`
- Scans file-state KV to list managed entries

- [ ] **Step 1: Write failing tests for Darwin and Linux stubs**

Create `internal/provider/node/sysctl/darwin_public_test.go`:

```go
package sysctl_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/provider"
	"github.com/retr0h/osapi/internal/provider/node/sysctl"
)

type DarwinPublicTestSuite struct {
	suite.Suite
	provider *sysctl.Darwin
}

func (s *DarwinPublicTestSuite) SetupTest() {
	s.provider = sysctl.NewDarwinProvider()
}

func (s *DarwinPublicTestSuite) TestAllMethodsReturnErrUnsupported() {
	tests := []struct {
		name string
		fn   func() error
	}{
		{
			name: "List",
			fn: func() error {
				_, err := s.provider.List(context.Background())
				return err
			},
		},
		{
			name: "Get",
			fn: func() error {
				_, err := s.provider.Get(context.Background(), "net.ipv4.ip_forward")
				return err
			},
		},
		{
			name: "Set",
			fn: func() error {
				_, err := s.provider.Set(context.Background(), sysctl.Entry{
					Key:   "net.ipv4.ip_forward",
					Value: "1",
				})
				return err
			},
		},
		{
			name: "Delete",
			fn: func() error {
				_, err := s.provider.Delete(context.Background(), "net.ipv4.ip_forward")
				return err
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := tt.fn()
			s.Require().Error(err)
			s.ErrorIs(err, provider.ErrUnsupported)
		})
	}
}

func TestDarwinPublicTestSuite(t *testing.T) {
	suite.Run(t, new(DarwinPublicTestSuite))
}
```

Create `internal/provider/node/sysctl/linux_public_test.go` with the same
pattern using `Linux`/`NewLinuxProvider`.

- [ ] **Step 2: Run stub tests to verify they pass**

Run: `go test -v ./internal/provider/node/sysctl/...` Expected: all pass

- [ ] **Step 3: Write failing tests for Debian provider**

Create `internal/provider/node/sysctl/debian_public_test.go` with test suite:

```go
package sysctl_test

type DebianPublicTestSuite struct {
	suite.Suite
	mockCtrl     *gomock.Controller
	mockDeployer *filemocks.MockDeployer
	mockExec     *execmocks.MockManager
	mockStateKV  *natsmocks.MockKeyValue
	provider     *sysctl.Debian
	ctx          context.Context
}
```

Tests to cover for each method:

**TestList:**

- success with managed entries (mock stateKV.ListKeys, stateKV.Get for each,
  exec for runtime values)
- empty list (no managed keys)
- stateKV.ListKeys error
- exec error reading runtime value (still returns entry with empty value)

**TestGet:**

- success (entry exists in stateKV, runtime value read via exec)
- not found (stateKV returns no entry for the key)
- exec error reading value

**TestSet:**

- success creates new entry (deploy returns Changed=true, exec sysctl -p
  succeeds)
- idempotent no change (deploy returns Changed=false, skip sysctl -p)
- deploy error
- sysctl -p error (file deployed but apply failed)

**TestDelete:**

- success (undeploy succeeds, sysctl --system succeeds)
- not found (entry not in stateKV)
- undeploy error
- sysctl --system error

Each method is ONE suite method with all scenarios as table rows.

- [ ] **Step 4: Run tests to verify they fail**

Run: `go test -v ./internal/provider/node/sysctl/...` Expected: FAIL — `Debian`
type does not exist yet

- [ ] **Step 5: Implement debian.go**

Create `internal/provider/node/sysctl/debian.go`:

Key implementation details:

- Struct embeds `provider.FactsAware`, holds `logger`, `fs`,
  `fileDeployer file.Deployer`, `stateKV jetstream.KeyValue`,
  `execManager exec.Manager`, `hostname string`
- Compile-time checks: `var _ Provider = (*Debian)(nil)` and
  `var _ provider.FactsSetter = (*Debian)(nil)`
- `NewDebianProvider(logger, fs, fileDeployer, stateKV, execManager, hostname) *Debian`
- `confPath(key)` returns `/etc/sysctl.d/osapi-{key}.conf`
- `confContent(key, value)` returns `{key} = {value}\n`
- `buildMetadata(entry)` returns
  `map[string]string{"key": entry.Key, "value": entry.Value}`
- `isManagedFile(stateKey)` checks stateKV for existence and metadata containing
  "key"
- `Set` deploys via `file.Deployer.Deploy()` with `ContentType: "raw"`, then
  runs `sysctl -p <path>` if changed
- `Delete` undeploys via `file.Deployer.Undeploy()`, then runs `sysctl --system`
- `List` scans stateKV keys matching hostname prefix, filters for sysctl
  metadata, reads runtime values
- `Get` looks up single key in stateKV, reads runtime value

For deploying, the content is generated inline (not from object store), so use a
`DeployRequest` with the content embedded. Check how the cron provider generates
content — it uses `ObjectName` pointing to an object store entry. For sysctl,
the content is trivial (`key = value\n`), so we need to check if `file.Deployer`
supports inline content or if we need to write to the object store first.

Read `internal/provider/file/types.go` and `internal/provider/file/deploy.go` to
understand the `DeployRequest` fields. The `ObjectName` field references an
object in NATS Object Store. For sysctl, we may need to:

1. Write the content to a temp object in the object store, then deploy from it,
   OR
2. Add support for inline content in `DeployRequest`

Since the cron provider always deploys from object store objects, and the sysctl
content is a single line, the simplest approach is to write the content directly
to the filesystem (bypassing `file.Deployer`) and manage state in the KV
manually. However, that loses SHA tracking and idempotency.

**Alternative approach**: The Debian provider can write the conf file directly
using `avfs.VFS`, track state in the file-state KV itself, and use `sysctl -p`
to apply. This is simpler than using `file.Deployer` for trivial content. The
provider manages its own state entries in the same KV bucket using the same key
format (`hostname.sha256(path)`).

The implementer should read the cron provider's `debian.go` carefully to decide
which approach fits best. If `file.Deployer.Deploy()` requires an object store
reference, use the direct-write approach with manual KV state tracking. If it
can accept inline content, use the deployer.

- [ ] **Step 6: Create export_test.go for testing internals**

```go
package sysctl

// SetConfPath overrides the confPath function for testing.
var SetConfPath = func(fn func(string) string) {
	confPathFn = fn
}

// ResetConfPath restores the default confPath function.
var ResetConfPath = func() {
	confPathFn = defaultConfPath
}
```

Adjust as needed based on which internal functions need test injection.

- [ ] **Step 7: Run all tests**

Run: `go test -v ./internal/provider/node/sysctl/...` Expected: all pass

- [ ] **Step 8: Check coverage**

Run: `go test -coverprofile=cover.out ./internal/provider/node/sysctl/...`
Expected: 100% on non-generated code

- [ ] **Step 9: Commit**

```bash
git add internal/provider/node/sysctl/
git commit -m "feat(sysctl): implement Debian meta-provider with tests"
```

---

## Task 4: Agent Processor

**Files:**

- Create: `internal/agent/processor_sysctl.go`
- Test: `internal/agent/processor_sysctl_public_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/agent/processor_sysctl_public_test.go` following the pattern in
`processor_schedule_public_test.go`:

Test all operations:

- `sysctl.list` — success, provider error
- `sysctl.get` — success, unmarshal error, provider error
- `sysctl.set` — success, unmarshal error, provider error
- `sysctl.delete` — success, unmarshal error, provider error
- unknown sub-operation — error

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -run TestProcessorSysctl -v ./internal/agent/...` Expected: FAIL

- [ ] **Step 3: Implement processor_sysctl.go**

```go
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/provider/node/sysctl"
)

// NewSysctlProcessor creates a processor for sysctl operations.
func NewSysctlProcessor(
	provider sysctl.Provider,
	logger *slog.Logger,
) ProcessorFunc {
	return func(req job.Request) (json.RawMessage, error) {
		if provider == nil {
			return nil, fmt.Errorf("sysctl provider not available")
		}

		parts := strings.Split(req.Operation, ".")
		if len(parts) < 3 {
			return nil, fmt.Errorf("invalid sysctl operation: %s", req.Operation)
		}
		subOp := parts[2]

		ctx := context.Background()

		switch subOp {
		case "list":
			return processSysctlList(ctx, provider, logger)
		case "get":
			return processSysctlGet(ctx, provider, logger, req)
		case "set":
			return processSysctlSet(ctx, provider, logger, req)
		case "delete":
			return processSysctlDelete(ctx, provider, logger, req)
		default:
			return nil, fmt.Errorf("unsupported sysctl operation: %s", req.Operation)
		}
	}
}

func processSysctlList(
	ctx context.Context,
	provider sysctl.Provider,
	logger *slog.Logger,
) (json.RawMessage, error) {
	logger.Debug("executing sysctl.list")

	entries, err := provider.List(ctx)
	if err != nil {
		return nil, err
	}

	return json.Marshal(entries)
}

func processSysctlGet(
	ctx context.Context,
	provider sysctl.Provider,
	logger *slog.Logger,
	req job.Request,
) (json.RawMessage, error) {
	var data struct {
		Key string `json:"key"`
	}
	if err := json.Unmarshal(req.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal sysctl get data: %w", err)
	}

	logger.Debug("executing sysctl.get", slog.String("key", data.Key))

	entry, err := provider.Get(ctx, data.Key)
	if err != nil {
		return nil, err
	}

	return json.Marshal(entry)
}

func processSysctlSet(
	ctx context.Context,
	provider sysctl.Provider,
	logger *slog.Logger,
	req job.Request,
) (json.RawMessage, error) {
	var entry sysctl.Entry
	if err := json.Unmarshal(req.Data, &entry); err != nil {
		return nil, fmt.Errorf("unmarshal sysctl set data: %w", err)
	}

	logger.Debug("executing sysctl.set", slog.String("key", entry.Key))

	result, err := provider.Set(ctx, entry)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

func processSysctlDelete(
	ctx context.Context,
	provider sysctl.Provider,
	logger *slog.Logger,
	req job.Request,
) (json.RawMessage, error) {
	var data struct {
		Key string `json:"key"`
	}
	if err := json.Unmarshal(req.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal sysctl delete data: %w", err)
	}

	logger.Debug("executing sysctl.delete", slog.String("key", data.Key))

	result, err := provider.Delete(ctx, data.Key)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}
```

- [ ] **Step 4: Run tests**

Run: `go test -run TestProcessorSysctl -v ./internal/agent/...` Expected: all
pass

- [ ] **Step 5: Commit**

```bash
git add internal/agent/processor_sysctl.go internal/agent/processor_sysctl_public_test.go
git commit -m "feat(sysctl): add agent processor with tests"
```

---

## Task 5: Agent Wiring

**Files:**

- Modify: `cmd/agent_setup.go`

- [ ] **Step 1: Add sysctl provider creation in setupAgent**

In `cmd/agent_setup.go`, after the cron provider creation
(`createCronProvider`), add:

```go
	// --- Sysctl provider ---
	sysctlProvider := createSysctlProvider(log, appFs, fileProvider, fileStateKV, execManager, hostname)
```

Add the `createSysctlProvider` helper function:

```go
func createSysctlProvider(
	log *slog.Logger,
	fs avfs.VFS,
	fileProvider fileProv.Provider,
	fileStateKV jetstream.KeyValue,
	execManager exec.Manager,
	hostname string,
) sysctlProv.Provider {
	plat := platform.Detect()

	switch plat {
	case "debian":
		if fileProvider == nil {
			log.Warn("file provider not available, sysctl operations disabled")
			return sysctlProv.NewLinuxProvider()
		}
		return sysctlProv.NewDebianProvider(log, fs, fileProvider, fileStateKV, execManager, hostname)
	case "darwin":
		return sysctlProv.NewDarwinProvider()
	default:
		return sysctlProv.NewLinuxProvider()
	}
}
```

Add the import:

```go
sysctlProv "github.com/retr0h/osapi/internal/provider/node/sysctl"
```

Note: `execManager` is already created earlier in `setupAgent` as
`execManager := exec.New(log)`. If it doesn't exist, check if it's named
`execManager` or just created inline. The sysctl provider needs it for
`sysctl -p` and `sysctl -n` commands.

- [ ] **Step 2: Register sysctl processor in the registry**

After `registry.Register("schedule", ...)`, add:

```go
	registry.Register("sysctl",
		agent.NewSysctlProcessor(sysctlProvider, log),
		sysctlProvider,
	)
```

Note: The sysctl operations use `node.sysctl.*` format. The registry dispatches
on the first segment. Since the operation format is `node.sysctl.list`, the
registry key should match how the processor splits the operation. Check if the
node processor already handles `node.*` operations. If so, the sysctl processor
should be integrated INTO the node processor OR use a different category key.
Read `internal/agent/processor_node.go` to confirm.

If the node processor dispatches `node.hostname.*`, `node.disk.*`, etc., then
sysctl should be added as another case in the node processor rather than a
separate registry entry. In that case, modify `processor_node.go` to handle
`sysctl` sub-operations and delegate to the sysctl provider. The registry key
would remain `"node"`.

The implementer MUST read `internal/agent/processor_node.go` to determine the
correct integration approach.

- [ ] **Step 3: Verify it compiles**

Run: `go build ./...` Expected: clean build

- [ ] **Step 4: Commit**

```bash
git add cmd/agent_setup.go internal/agent/processor_node.go  # or processor_sysctl.go
git commit -m "feat(sysctl): wire provider into agent"
```

---

## Task 6: OpenAPI Spec + Code Generation

**Files:**

- Create: `internal/controller/api/sysctl/gen/api.yaml`
- Create: `internal/controller/api/sysctl/gen/cfg.yaml`
- Create: `internal/controller/api/sysctl/gen/generate.go`

- [ ] **Step 1: Create api.yaml**

Follow the cron spec structure
(`internal/controller/api/schedule/gen/api.yaml`).

Paths:

- `GET /node/{hostname}/sysctl` — list, security `sysctl:read`, responses
  200/401/403/500
- `POST /node/{hostname}/sysctl` — set, security `sysctl:write`, responses
  200/400/401/403/500
- `GET /node/{hostname}/sysctl/{key}` — get, security `sysctl:read`, responses
  200/401/403/404/500
- `DELETE /node/{hostname}/sysctl/{key}` — delete, security `sysctl:write`,
  responses 200/401/403/404/500

Parameters:

- `Hostname` — same as cron spec (reuse `$ref` to common)
- `SysctlKey` — path param, `type: string`, `minLength: 1`,
  `pattern: '^[a-z0-9._]+$'`

Request schemas:

- `SysctlSetRequest` — required: `key`, `value`. Both strings with validate
  tags.

Response schemas:

- `SysctlEntry` — hostname (required), status (required, enum
  ok/failed/skipped), key, value, error
- `SysctlMutationResult` — hostname (required), status (required, enum
  ok/failed/skipped), key, changed, error
- `SysctlCollectionResponse` — job_id (uuid), results (array of SysctlEntry)
- `SysctlGetResponse` — same structure as collection
- `SysctlSetResponse` — job_id (uuid), results (array of SysctlMutationResult)
- `SysctlDeleteResponse` — job_id (uuid), results (array of
  SysctlMutationResult)

- [ ] **Step 2: Create cfg.yaml**

```yaml
---
package: gen
output: sysctl.gen.go
generate:
  models: true
  echo-server: true
  strict-server: true
import-mapping:
  ../../common/gen/api.yaml: github.com/retr0h/osapi/internal/controller/api/common/gen
output-options:
  skip-prune: true
```

- [ ] **Step 3: Create generate.go**

```go
// Package gen contains generated code for the sysctl API.
package gen

//go:generate go tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config cfg.yaml api.yaml
```

- [ ] **Step 4: Generate code**

Run: `go generate ./internal/controller/api/sysctl/gen/...` Expected:
`sysctl.gen.go` created

- [ ] **Step 5: Verify it compiles**

Run: `go build ./internal/controller/api/sysctl/...` Expected: clean build

- [ ] **Step 6: Commit**

```bash
git add internal/controller/api/sysctl/gen/
git commit -m "feat(sysctl): add OpenAPI spec and generate code"
```

---

## Task 7: API Handler Implementation

**Files:**

- Create: `internal/controller/api/sysctl/types.go`
- Create: `internal/controller/api/sysctl/sysctl.go`
- Create: `internal/controller/api/sysctl/validate.go`
- Create: `internal/controller/api/sysctl/sysctl_list_get.go`
- Create: `internal/controller/api/sysctl/sysctl_get.go`
- Create: `internal/controller/api/sysctl/sysctl_set.go`
- Create: `internal/controller/api/sysctl/sysctl_delete.go`
- Test: `internal/controller/api/sysctl/sysctl_list_get_public_test.go`
- Test: `internal/controller/api/sysctl/sysctl_get_public_test.go`
- Test: `internal/controller/api/sysctl/sysctl_set_public_test.go`
- Test: `internal/controller/api/sysctl/sysctl_delete_public_test.go`

Follow the cron handler pattern exactly. Each handler file:

1. Validates hostname via `validateHostname()`
2. Validates request body via `validation.Struct()` (for POST)
3. Checks `job.IsBroadcastTarget()` and routes to broadcast function
4. Calls `JobClient.Query` (reads) or `JobClient.Modify` (writes) with category
   `"node"` and the sysctl operation constant
5. Handles skipped/failed status
6. Converts provider results to gen response types

The category string passed to `JobClient.Query/Modify` MUST match the registry
key used in `agent_setup.go`. If sysctl is registered under the `"node"`
category (integrated into the node processor), use `"node"`. If it's a separate
registry entry, use `"sysctl"`.

- [ ] **Step 1: Create types.go, sysctl.go, validate.go**

Follow the schedule handler pattern. `validate.go` should include both
`validateHostname` and `validateSysctlKey` (validates the dotted key format).

- [ ] **Step 2: Write failing tests for list handler**

Follow `cron_list_get_public_test.go` pattern with table-driven tests covering:

- success single target
- success broadcast
- skipped status
- query error
- invalid hostname

Include `TestSysctlListGetHTTP` and `TestSysctlListGetRBACHTTP` methods.

- [ ] **Step 3: Implement list handler**

- [ ] **Step 4: Run list tests**

Run: `go test -run TestSysctlListGet -v ./internal/controller/api/sysctl/...`

- [ ] **Step 5: Repeat steps 2-4 for get, set, delete handlers**

Each handler test file needs table-driven tests + HTTP wiring + RBAC tests.

- [ ] **Step 6: Run all handler tests**

Run: `go test -v ./internal/controller/api/sysctl/...` Expected: all pass

- [ ] **Step 7: Commit**

```bash
git add internal/controller/api/sysctl/
git commit -m "feat(sysctl): implement API handlers with tests"
```

---

## Task 8: Server Wiring

**Files:**

- Create: `internal/controller/api/handler_sysctl.go`
- Modify: `cmd/controller_setup.go`
- Modify: `internal/controller/api/handler_public_test.go`

- [ ] **Step 1: Create handler_sysctl.go**

Follow `handler_schedule.go` pattern:

```go
func (s *Server) GetSysctlHandler(
	jobClient client.JobClient,
) []func(e *echo.Echo) {
	var tokenManager TokenValidator = authtoken.New(s.logger)

	sysctlHandler := sysctlAPI.New(s.logger, jobClient)

	strictHandler := sysctlGen.NewStrictHandler(
		sysctlHandler,
		[]sysctlGen.StrictMiddlewareFunc{
			func(handler strictecho.StrictEchoHandlerFunc, _ string) strictecho.StrictEchoHandlerFunc {
				return scopeMiddleware(
					handler,
					tokenManager,
					s.appConfig.Controller.API.Security.SigningKey,
					sysctlGen.BearerAuthScopes,
					s.customRoles,
				)
			},
		},
	)

	return []func(e *echo.Echo){
		func(e *echo.Echo) {
			sysctlGen.RegisterHandlers(e, strictHandler)
		},
	}
}
```

- [ ] **Step 2: Wire in controller_setup.go**

Add after `GetScheduleHandler`:

```go
handlers = append(handlers, sm.GetSysctlHandler(jc)...)
```

Add the interface method to the `HandlerFactory` interface (or wherever
`GetScheduleHandler` is declared):

```go
GetSysctlHandler(jobClient jobclient.JobClient) []func(e *echo.Echo)
```

- [ ] **Step 3: Add test in handler_public_test.go**

Add `TestGetSysctlHandler` following the `TestGetScheduleHandler` pattern.

- [ ] **Step 4: Run tests**

Run: `go test -v ./internal/controller/api/...` Expected: all pass

- [ ] **Step 5: Verify full build**

Run: `go build ./...` Expected: clean build

- [ ] **Step 6: Commit**

```bash
git add internal/controller/api/handler_sysctl.go \
  internal/controller/api/handler_public_test.go \
  cmd/controller_setup.go
git commit -m "feat(sysctl): wire handler into server"
```

---

## Task 9: Regenerate Combined Spec

**Files:**

- Modified by generation: `internal/controller/api/gen/api.yaml`
- Modified by generation: `pkg/sdk/client/gen/`

- [ ] **Step 1: Regenerate combined spec**

Run: `just generate` Expected: combined `api.yaml` includes sysctl paths, SDK
client regenerated

- [ ] **Step 2: Verify build**

Run: `go build ./...` Expected: clean build

- [ ] **Step 3: Commit**

```bash
git add internal/controller/api/gen/ pkg/sdk/client/gen/
git commit -m "chore: regenerate combined spec with sysctl endpoints"
```

---

## Task 10: SDK Service

**Files:**

- Create: `pkg/sdk/client/sysctl.go`
- Create: `pkg/sdk/client/sysctl_types.go`
- Modify: `pkg/sdk/client/osapi.go`
- Test: `pkg/sdk/client/sysctl_public_test.go`
- Test: `pkg/sdk/client/sysctl_types_public_test.go`

- [ ] **Step 1: Create sysctl_types.go**

Define SDK types (never expose gen types):

```go
package client

// SysctlEntryResult represents a sysctl entry from a query operation.
type SysctlEntryResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Key      string `json:"key,omitempty"`
	Value    string `json:"value,omitempty"`
	Error    string `json:"error,omitempty"`
}

// SysctlMutationResult represents the result of a sysctl set or delete.
type SysctlMutationResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Key      string `json:"key,omitempty"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// SysctlSetOpts contains options for setting a sysctl parameter.
type SysctlSetOpts struct {
	// Key is the sysctl parameter name (e.g., "net.ipv4.ip_forward"). Required.
	Key string
	// Value is the parameter value. Required.
	Value string
}
```

Add gen→SDK conversion functions following the cron pattern.

- [ ] **Step 2: Create sysctl.go**

```go
package client

// SysctlService provides sysctl management operations.
type SysctlService struct {
	client *gen.ClientWithResponses
}
```

Methods: `SysctlList`, `SysctlGet`, `SysctlSet`, `SysctlDelete` — each following
the schedule service pattern with proper error checking, nil response guards,
and response wrapping.

- [ ] **Step 3: Wire into osapi.go**

Add `Sysctl *SysctlService` to `Client` struct and initialize in `New()`:

```go
c.Sysctl = &SysctlService{client: httpClient}
```

- [ ] **Step 4: Write tests**

Create `pkg/sdk/client/sysctl_public_test.go` using `httptest.Server` mocks.
Cover all status code paths (200, 400, 401, 403, 404, 500), nil response body,
transport errors.

Create `pkg/sdk/client/sysctl_types_public_test.go` testing conversion
functions.

- [ ] **Step 5: Run tests**

Run: `go test -v ./pkg/sdk/client/...` Expected: all pass, 100% coverage on
sysctl files

- [ ] **Step 6: Commit**

```bash
git add pkg/sdk/client/sysctl.go pkg/sdk/client/sysctl_types.go \
  pkg/sdk/client/osapi.go \
  pkg/sdk/client/sysctl_public_test.go pkg/sdk/client/sysctl_types_public_test.go
git commit -m "feat(sysctl): add SDK service with tests"
```

---

## Task 11: CLI Commands

**Files:**

- Create: `cmd/client_node_sysctl.go`
- Create: `cmd/client_node_sysctl_list.go`
- Create: `cmd/client_node_sysctl_get.go`
- Create: `cmd/client_node_sysctl_set.go`
- Create: `cmd/client_node_sysctl_delete.go`

- [ ] **Step 1: Create parent command**

```go
var clientNodeSysctlCmd = &cobra.Command{
	Use:   "sysctl",
	Short: "Manage kernel parameters",
}

func init() {
	clientNodeCmd.AddCommand(clientNodeSysctlCmd)
}
```

- [ ] **Step 2: Create list command**

Follow `client_node_schedule_cron_list.go` pattern:

- Call `sdkClient.Sysctl.SysctlList(ctx, host)`
- Handle `--json` output
- Build table with fields: KEY, VALUE
- Use `cli.BuildBroadcastTable` + `cli.PrintCompactTable`

- [ ] **Step 3: Create get command**

Flags: `--key` (required)

- Call `sdkClient.Sysctl.SysctlGet(ctx, host, key)`
- Same output pattern

- [ ] **Step 4: Create set command**

Flags: `--key` (required), `--value` (required)

- Build `client.SysctlSetOpts{Key: key, Value: value}`
- Call `sdkClient.Sysctl.SysctlSet(ctx, host, opts)`
- Use `cli.BuildMutationTable` with CHANGED field

- [ ] **Step 5: Create delete command**

Flags: `--key` (required)

- Call `sdkClient.Sysctl.SysctlDelete(ctx, host, key)`
- Mutation table output

- [ ] **Step 6: Verify CLI builds**

Run: `go build ./cmd/...` Expected: clean build

- [ ] **Step 7: Commit**

```bash
git add cmd/client_node_sysctl*.go
git commit -m "feat(sysctl): add CLI commands"
```

---

## Task 12: SDK Example

**Files:**

- Create: `examples/sdk/client/sysctl.go`

- [ ] **Step 1: Create example**

Follow the conventions in CLAUDE.md (one domain per file, self-contained, print
results, handle errors inline, under ~100 lines):

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/retr0h/osapi/pkg/sdk/client"
)

func sysctlExample() {
	c, err := client.New("http://localhost:8080",
		client.WithBearerToken("..."),
	)
	if err != nil {
		log.Fatalf("create client: %v", err)
	}

	ctx := context.Background()

	// List managed sysctl entries
	listResp, err := c.Sysctl.SysctlList(ctx, "_any")
	if err != nil {
		log.Fatalf("sysctl list: %v", err)
	}
	fmt.Printf("Managed entries: %d\n", len(listResp.Data.Results))
	for _, e := range listResp.Data.Results {
		fmt.Printf("  %s = %s\n", e.Key, e.Value)
	}

	// Set a sysctl parameter
	setResp, err := c.Sysctl.SysctlSet(ctx, "_any", client.SysctlSetOpts{
		Key:   "net.ipv4.ip_forward",
		Value: "1",
	})
	if err != nil {
		log.Fatalf("sysctl set: %v", err)
	}
	if first := setResp.Data.First(); first != nil {
		fmt.Printf("Set %s: changed=%v\n", first.Key, first.Changed)
	}
}
```

- [ ] **Step 2: Commit**

```bash
git add examples/sdk/client/sysctl.go
git commit -m "feat(sysctl): add SDK example"
```

---

## Task 13: Documentation

**Files:**

- Create: `docs/docs/sidebar/features/sysctl.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/sysctl/sysctl.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/sysctl/list.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/sysctl/get.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/sysctl/set.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/sysctl/delete.md`
- Modify: `docs/docusaurus.config.ts`
- Modify: `docs/docs/sidebar/usage/configuration.md`

- [ ] **Step 1: Create feature page**

Follow the template from `docs/docs/sidebar/features/cron-management.md`:
overview, how it works, operations table, CLI examples, permissions, platforms.

- [ ] **Step 2: Create CLI doc pages**

Parent page with `<DocCardList />`, one page per subcommand with usage examples.

- [ ] **Step 3: Update docusaurus.config.ts**

Add sysctl to the Features navbar dropdown.

- [ ] **Step 4: Update configuration.md**

Add `sysctl:read` and `sysctl:write` to the permissions/roles tables.

- [ ] **Step 5: Commit**

```bash
git add docs/
git commit -m "docs(sysctl): add feature and CLI documentation"
```

---

## Task 14: Integration Tests

**Files:**

- Create: `test/integration/sysctl_test.go`

- [ ] **Step 1: Create integration smoke tests**

Follow the `node_test.go` pattern. Guard writes with `skipWrite(s.T())`.

```go
//go:build integration

package integration_test

type SysctlSmokeSuite struct {
	suite.Suite
}

func (s *SysctlSmokeSuite) TestSysctlList() {
	// Test --json output, verify results array
}

func TestSysctlSmokeSuite(t *testing.T) {
	suite.Run(t, new(SysctlSmokeSuite))
}
```

- [ ] **Step 2: Commit**

```bash
git add test/integration/sysctl_test.go
git commit -m "test(sysctl): add integration smoke tests"
```

---

## Task 15: Final Verification

- [ ] **Step 1: Regenerate all**

Run: `just generate` Expected: clean

- [ ] **Step 2: Build**

Run: `go build ./...` Expected: clean

- [ ] **Step 3: Run all unit tests**

Run: `just go::unit` Expected: all pass

- [ ] **Step 4: Check coverage**

Run: `just go::unit-cov` Expected: >= 99.9% total coverage

- [ ] **Step 5: Lint**

Run: `just go::vet` Expected: clean

- [ ] **Step 6: Format**

Run: `just go::fmt` Expected: no changes

- [ ] **Step 7: Ready check**

Run: `just ready` Expected: all green

- [ ] **Step 8: Final commit (if formatting changed anything)**

```bash
git add -A
git commit -m "chore(sysctl): formatting and lint fixes"
```
