# Certificate Management Provider Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> superpowers:subagent-driven-development (recommended) or
> superpowers:executing-plans to implement this plan task-by-task. Steps use
> checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add CA certificate management to OSAPI — deploy custom CA certificates
to the system trust store via `file.Deployer`, list system + custom CAs, update
and remove custom CAs.

**Architecture:** Meta provider at `internal/provider/node/certificate/` using
`file.Deployer` for SHA-tracked deployment and `exec.Manager` for
`update-ca-certificates`. Registered as its own agent category (`certificate`)
following the cron/schedule pattern. Four API endpoints under
`/node/{hostname}/certificate/ca`.

**Tech Stack:** Go, file.Deployer, exec.Manager, update-ca-certificates,
oapi-codegen strict-server

---

## File Structure

### Provider Layer

- Create: `internal/provider/node/certificate/types.go` — Provider interface +
  domain types
- Create: `internal/provider/node/certificate/debian.go` — Debian implementation
  with file.Deployer
- Create: `internal/provider/node/certificate/debian_list.go` — List
  implementation (walks system + custom CA dirs)
- Create: `internal/provider/node/certificate/darwin.go` — macOS stub
- Create: `internal/provider/node/certificate/linux.go` — generic Linux stub
- Create: `internal/provider/node/certificate/mocks/generate.go` — mockgen
  directive
- Test: `internal/provider/node/certificate/debian_public_test.go`
- Test: `internal/provider/node/certificate/debian_list_public_test.go`
- Test: `internal/provider/node/certificate/darwin_public_test.go`
- Test: `internal/provider/node/certificate/linux_public_test.go`

### Agent Layer

- Create: `internal/agent/processor_certificate.go` — certificate operation
  dispatcher
- Modify: `cmd/agent_setup.go` — create certificate provider factory, register
  as own category
- Test: `internal/agent/processor_certificate_public_test.go`

### API Layer

- Create: `internal/controller/api/node/certificate/gen/api.yaml` — OpenAPI spec
- Create: `internal/controller/api/node/certificate/gen/cfg.yaml` — oapi-codegen
  config
- Create: `internal/controller/api/node/certificate/gen/generate.go` —
  go:generate
- Create: `internal/controller/api/node/certificate/types.go` — handler struct
- Create: `internal/controller/api/node/certificate/certificate.go` — New(),
  compile-time check
- Create: `internal/controller/api/node/certificate/validate.go` —
  validateHostname
- Create: `internal/controller/api/node/certificate/ca_list_get.go` — list
  handler
- Create: `internal/controller/api/node/certificate/ca_create_post.go` — create
  handler
- Create: `internal/controller/api/node/certificate/ca_update_put.go` — update
  handler
- Create: `internal/controller/api/node/certificate/ca_delete.go` — delete
  handler
- Create: `internal/controller/api/node/certificate/handler.go` — Handler()
  registration
- Modify: `cmd/controller_setup.go` — register certificate handler
- Test: `internal/controller/api/node/certificate/ca_list_get_public_test.go`
- Test: `internal/controller/api/node/certificate/ca_create_post_public_test.go`
- Test: `internal/controller/api/node/certificate/ca_update_put_public_test.go`
- Test: `internal/controller/api/node/certificate/ca_delete_public_test.go`
- Test: `internal/controller/api/node/certificate/handler_public_test.go`

### Operations & Permissions

- Modify: `pkg/sdk/client/operations.go` — add certificate operation constants
- Modify: `internal/job/types.go` — add certificate operation aliases
- Modify: `pkg/sdk/client/permissions.go` — add `PermCertificateRead`,
  `PermCertificateWrite`
- Modify: `internal/authtoken/permissions.go` — add to all roles

### SDK Layer

- Create: `pkg/sdk/client/certificate.go` — CertificateService methods
- Create: `pkg/sdk/client/certificate_types.go` — SDK result types
  - conversions
- Modify: `pkg/sdk/client/osapi.go` — add Certificate field
- Test: `pkg/sdk/client/certificate_public_test.go`
- Test: `pkg/sdk/client/certificate_types_public_test.go`

### CLI Layer

- Create: `cmd/client_node_certificate.go` — parent command
- Create: `cmd/client_node_certificate_list.go` — list subcommand
- Create: `cmd/client_node_certificate_create.go` — create subcommand
- Create: `cmd/client_node_certificate_update.go` — update subcommand
- Create: `cmd/client_node_certificate_delete.go` — delete subcommand

### Documentation

- Create: `docs/docs/sidebar/features/certificate-management.md` — feature page
- Create: `docs/docs/sidebar/usage/cli/client/node/certificate/certificate.md` —
  CLI landing
- Create: `docs/docs/sidebar/usage/cli/client/node/certificate/list.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/certificate/create.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/certificate/update.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/certificate/delete.md`
- Create: `docs/docs/sidebar/sdk/client/management/certificate.md` — SDK doc
- Create: `examples/sdk/client/certificate.go` — SDK example
- Modify: `docs/docs/sidebar/features/features.md` — add cert row
- Modify: `docs/docs/sidebar/features/authentication.md` — add permissions to
  role tables
- Modify: `docs/docs/sidebar/usage/configuration.md` — add permissions
- Modify: `docs/docs/sidebar/architecture/architecture.md` — add feature link
- Modify: `docs/docs/sidebar/architecture/api-guidelines.md` — add endpoints
- Modify: `docs/docusaurus.config.ts` — add to dropdowns
- Modify: `docs/docs/sidebar/sdk/client/client.md` — add service to table

### Integration Test

- Create: `test/integration/certificate_test.go` — smoke test

---

### Task 1: Provider Interface and Types

**Files:**

- Create: `internal/provider/node/certificate/types.go`

- [ ] **Step 1: Create provider interface and types**

```go
// Package certificate provides CA certificate management operations.
package certificate

import (
	"context"
)

// Provider implements CA certificate management operations.
type Provider interface {
	// List returns all CA certificates (system and custom).
	List(ctx context.Context) ([]Entry, error)
	// Create deploys a new custom CA certificate to the trust store.
	Create(ctx context.Context, entry Entry) (*CreateResult, error)
	// Update redeploys an existing custom CA certificate.
	Update(ctx context.Context, entry Entry) (*UpdateResult, error)
	// Delete removes a custom CA certificate from the trust store.
	Delete(ctx context.Context, name string) (*DeleteResult, error)
}

// Entry represents a CA certificate.
type Entry struct {
	Name   string `json:"name"`
	Source string `json:"source,omitempty"` // "system" or "custom"
	Object string `json:"object,omitempty"`
}

// CreateResult represents the outcome of a CA certificate creation.
type CreateResult struct {
	Name    string `json:"name"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// UpdateResult represents the outcome of a CA certificate update.
type UpdateResult struct {
	Name    string `json:"name"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// DeleteResult represents the outcome of a CA certificate deletion.
type DeleteResult struct {
	Name    string `json:"name"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/provider/node/certificate/...`

- [ ] **Step 3: Commit**

```bash
git add internal/provider/node/certificate/types.go
git commit -m "feat(certificate): add provider interface and types"
```

---

### Task 2: Platform Stubs (Darwin + Linux)

**Files:**

- Create: `internal/provider/node/certificate/darwin.go`
- Create: `internal/provider/node/certificate/linux.go`
- Test: `internal/provider/node/certificate/darwin_public_test.go`
- Test: `internal/provider/node/certificate/linux_public_test.go`

- [ ] **Step 1: Write stub tests**

Create `darwin_public_test.go` and `linux_public_test.go` with testify/suite.
Test all four methods (List, Create, Update, Delete) return
`provider.ErrUnsupported`. Follow the pattern in
`internal/provider/node/log/darwin_public_test.go` — one suite method per
provider method, all in a single table.

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -v ./internal/provider/node/certificate/...`

- [ ] **Step 3: Implement stubs**

Create `darwin.go`:

```go
type Darwin struct{}

func NewDarwinProvider() *Darwin { return &Darwin{} }

// All methods return:
// fmt.Errorf("certificate: %w", provider.ErrUnsupported)
```

Create `linux.go` — same pattern with `Linux` struct.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -v ./internal/provider/node/certificate/...`

- [ ] **Step 5: Commit**

```bash
git add internal/provider/node/certificate/
git commit -m "feat(certificate): add darwin and linux provider stubs"
```

---

### Task 3: Debian Provider Implementation

**Files:**

- Create: `internal/provider/node/certificate/debian.go`
- Create: `internal/provider/node/certificate/debian_list.go`
- Create: `internal/provider/node/certificate/mocks/generate.go`
- Test: `internal/provider/node/certificate/debian_public_test.go`
- Test: `internal/provider/node/certificate/debian_list_public_test.go`

This is a meta provider following the cron pattern. Read
`internal/provider/scheduled/cron/debian.go` as the primary reference.

- [ ] **Step 1: Create mock generator**

Create `internal/provider/node/certificate/mocks/generate.go`:

```go
package mocks

//go:generate go tool github.com/golang/mock/mockgen -source=../types.go -destination=provider.gen.go -package=mocks
```

Run: `go generate ./internal/provider/node/certificate/mocks/...`

- [ ] **Step 2: Write Debian provider tests**

Create `debian_public_test.go` with testify/suite and gomock. The Debian struct
needs:

- `provider.FactsAware` embedded
- `logger *slog.Logger`
- `fs avfs.VFS` (for listing system CAs)
- `fileDeployer file.Deployer` (mocked)
- `stateKV jetstream.KeyValue` (mocked)
- `execManager exec.Manager` (mocked, for `update-ca-certificates`)
- `hostname string`

**TestCreate** — table-driven with cases:

- success: mock fileDeployer.Deploy with path
  `/usr/local/share/ca-certificates/osapi-mycert.crt`, mode `0644`, returns
  `Changed: true`. Mock execManager.RunCmd for `update-ca-certificates`. Verify
  result.
- already exists: mock fs.Stat on the path returns nil (file exists). Verify
  error "already exists".
- deploy error: mock fileDeployer.Deploy returns error.
- update-ca-certificates error: deploy succeeds, RunCmd fails.
- invalid name (empty): verify error.
- invalid name (special chars): verify error.

**TestUpdate** — table-driven:

- success: mock fs.Stat finds file, mock fileDeployer.Deploy returns
  `Changed: true`, mock RunCmd succeeds.
- not found: mock fs.Stat returns error. Verify "does not exist".
- deploy error.
- update with same content: Deploy returns `Changed: false`, skip
  `update-ca-certificates`.

**TestDelete** — table-driven:

- success: mock fs.Stat finds file, mock fileDeployer.Undeploy returns
  `Changed: true`, mock RunCmd succeeds.
- not found: returns `Changed: false`, no error.
- undeploy error.

- [ ] **Step 3: Write List tests**

Create `debian_list_public_test.go` with testify/suite. Use `memfs.New()` for
filesystem. Set up:

- `/usr/share/ca-certificates/mozilla/DigiCert.crt` (system)
- `/usr/local/share/ca-certificates/osapi-mycert.crt` (custom, with matching
  file state KV entry)
- `/usr/local/share/ca-certificates/manual.crt` (not managed — no file state
  entry)

**TestList** — table-driven:

- success with system + custom certs
- empty directories
- fs.ReadDir error on system dir
- custom cert without file state (skipped)

- [ ] **Step 4: Implement debian.go**

```go
package certificate

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/avfs/avfs"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/retr0h/osapi/internal/exec"
	"github.com/retr0h/osapi/internal/provider"
	"github.com/retr0h/osapi/internal/provider/file"
)

const (
	systemCADir = "/usr/share/ca-certificates"
	customCADir = "/usr/local/share/ca-certificates"
	managedPrefix = "osapi-"
)

var validName = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// Compile-time checks.
var (
	_ Provider             = (*Debian)(nil)
	_ provider.FactsSetter = (*Debian)(nil)
)

type Debian struct {
	provider.FactsAware
	logger       *slog.Logger
	fs           avfs.VFS
	fileDeployer file.Deployer
	stateKV      jetstream.KeyValue
	execManager  exec.Manager
	hostname     string
}

func NewDebianProvider(
	logger *slog.Logger,
	fs avfs.VFS,
	fileDeployer file.Deployer,
	stateKV jetstream.KeyValue,
	execManager exec.Manager,
	hostname string,
) *Debian {
	return &Debian{
		logger:       logger.With(slog.String("subsystem", "provider.certificate")),
		fs:           fs,
		fileDeployer: fileDeployer,
		stateKV:      stateKV,
		execManager:  execManager,
		hostname:     hostname,
	}
}

func (d *Debian) Create(
	ctx context.Context,
	entry Entry,
) (*CreateResult, error) {
	if err := validateName(entry.Name); err != nil {
		return nil, err
	}

	filePath := customCADir + "/" + managedPrefix + entry.Name + ".crt"

	// Check if already exists on disk.
	if _, err := d.fs.Stat(filePath); err == nil {
		return nil, fmt.Errorf(
			"certificate %q already exists",
			entry.Name,
		)
	}

	d.logger.Debug("executing certificate.Create",
		slog.String("name", entry.Name),
	)

	result, err := d.fileDeployer.Deploy(ctx, file.DeployRequest{
		ObjectName:  entry.Object,
		Path:        filePath,
		Mode:        "0644",
		ContentType: "raw",
		Metadata: map[string]string{
			"source": "custom",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create certificate: %w", err)
	}

	if result.Changed {
		if _, err := d.execManager.RunCmd(
			"update-ca-certificates",
			nil,
		); err != nil {
			return nil, fmt.Errorf(
				"update-ca-certificates: %w",
				err,
			)
		}
	}

	return &CreateResult{
		Name:    entry.Name,
		Changed: result.Changed,
	}, nil
}

func (d *Debian) Update(
	ctx context.Context,
	entry Entry,
) (*UpdateResult, error) {
	if err := validateName(entry.Name); err != nil {
		return nil, err
	}

	filePath := customCADir + "/" + managedPrefix + entry.Name + ".crt"

	if _, err := d.fs.Stat(filePath); err != nil {
		return nil, fmt.Errorf(
			"certificate %q does not exist",
			entry.Name,
		)
	}

	d.logger.Debug("executing certificate.Update",
		slog.String("name", entry.Name),
	)

	result, err := d.fileDeployer.Deploy(ctx, file.DeployRequest{
		ObjectName:  entry.Object,
		Path:        filePath,
		Mode:        "0644",
		ContentType: "raw",
		Metadata: map[string]string{
			"source": "custom",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("update certificate: %w", err)
	}

	if result.Changed {
		if _, err := d.execManager.RunCmd(
			"update-ca-certificates",
			nil,
		); err != nil {
			return nil, fmt.Errorf(
				"update-ca-certificates: %w",
				err,
			)
		}
	}

	return &UpdateResult{
		Name:    entry.Name,
		Changed: result.Changed,
	}, nil
}

func (d *Debian) Delete(
	ctx context.Context,
	name string,
) (*DeleteResult, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}

	filePath := customCADir + "/" + managedPrefix + name + ".crt"

	if _, err := d.fs.Stat(filePath); err != nil {
		return &DeleteResult{
			Name:    name,
			Changed: false,
		}, nil
	}

	d.logger.Debug("executing certificate.Delete",
		slog.String("name", name),
	)

	result, err := d.fileDeployer.Undeploy(
		ctx,
		file.UndeployRequest{Path: filePath},
	)
	if err != nil {
		return nil, fmt.Errorf("delete certificate: %w", err)
	}

	if result.Changed {
		if _, err := d.execManager.RunCmd(
			"update-ca-certificates",
			nil,
		); err != nil {
			return nil, fmt.Errorf(
				"update-ca-certificates: %w",
				err,
			)
		}
	}

	return &DeleteResult{
		Name:    name,
		Changed: result.Changed,
	}, nil
}

func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("invalid certificate name: empty")
	}
	if !validName.MatchString(name) {
		return fmt.Errorf(
			"invalid certificate name %q: must match %s",
			name,
			validName.String(),
		)
	}
	return nil
}
```

- [ ] **Step 5: Implement debian_list.go**

```go
package certificate

import (
	"context"
	"encoding/json"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/provider/file"
)

// List returns all CA certificates — both system CAs from
// /usr/share/ca-certificates/ and custom CAs managed by OSAPI
// from /usr/local/share/ca-certificates/.
func (d *Debian) List(
	ctx context.Context,
) ([]Entry, error) {
	d.logger.Debug("executing certificate.List")

	var result []Entry

	// Walk system CA directory for system certs.
	systemEntries, err := d.listSystemCAs()
	if err != nil {
		return nil, fmt.Errorf("list system CAs: %w", err)
	}
	result = append(result, systemEntries...)

	// List custom CAs from file state KV.
	customEntries := d.listCustomCAs(ctx)
	result = append(result, customEntries...)

	return result, nil
}

// listSystemCAs walks /usr/share/ca-certificates/ and returns
// entries with source "system".
func (d *Debian) listSystemCAs() ([]Entry, error) {
	var entries []Entry

	err := d.fs.WalkDir(
		systemCADir,
		func(path string, dirEntry fs.DirEntry, err error) error {
			if err != nil {
				return nil // skip unreadable entries
			}
			if dirEntry.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".crt") {
				return nil
			}

			// Strip base dir and .crt extension for name.
			rel, _ := filepath.Rel(systemCADir, path)
			name := strings.TrimSuffix(rel, ".crt")

			entries = append(entries, Entry{
				Name:   name,
				Source: "system",
			})

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return entries, nil
}

// listCustomCAs reads custom CAs from the filesystem, checking
// the file-state KV to confirm they are OSAPI-managed.
func (d *Debian) listCustomCAs(
	ctx context.Context,
) []Entry {
	var entries []Entry

	dirEntries, err := d.fs.ReadDir(customCADir)
	if err != nil {
		return entries
	}

	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			continue
		}

		name := dirEntry.Name()
		if !strings.HasPrefix(name, managedPrefix) {
			continue
		}
		if !strings.HasSuffix(name, ".crt") {
			continue
		}

		path := customCADir + "/" + name

		// Verify this is managed via file state KV.
		stateKey := file.BuildStateKey(d.hostname, path)
		kvEntry, err := d.stateKV.Get(ctx, stateKey)
		if err != nil {
			continue
		}

		var state job.FileState
		if err := json.Unmarshal(
			kvEntry.Value(),
			&state,
		); err != nil {
			continue
		}
		if state.UndeployedAt != "" {
			continue
		}

		// Strip osapi- prefix and .crt suffix for clean name.
		cleanName := strings.TrimPrefix(name, managedPrefix)
		cleanName = strings.TrimSuffix(cleanName, ".crt")

		entries = append(entries, Entry{
			Name:   cleanName,
			Source: "custom",
			Object: state.ObjectName,
		})
	}

	return entries
}
```

Note: `debian_list.go` needs `"fmt"` in its imports for the `List` method error
wrapping.

- [ ] **Step 6: Run tests to verify they pass**

Run: `go test -v ./internal/provider/node/certificate/...`

- [ ] **Step 7: Verify 100% coverage**

Run:

```bash
go test -coverprofile=/tmp/cert_prov.cov \
  ./internal/provider/node/certificate/... && \
  go tool cover -func=/tmp/cert_prov.cov | \
  grep -v "100.0%" | grep -v "mocks"
```

Fix any gaps before proceeding.

- [ ] **Step 8: Commit**

```bash
git add internal/provider/node/certificate/
git commit -m "feat(certificate): add meta provider with file.Deployer"
```

---

### Task 4: Operations, Permissions, and Agent Wiring

**Files:**

- Modify: `pkg/sdk/client/operations.go`
- Modify: `internal/job/types.go`
- Modify: `pkg/sdk/client/permissions.go`
- Modify: `internal/authtoken/permissions.go`
- Create: `internal/agent/processor_certificate.go`
- Modify: `cmd/agent_setup.go`
- Test: `internal/agent/processor_certificate_public_test.go`

- [ ] **Step 1: Add operation constants**

In `pkg/sdk/client/operations.go`, add after Log operations:

```go
// Certificate operations.
const (
	OpCertificateCAList   JobOperation = "certificate.ca.list"
	OpCertificateCACreate JobOperation = "certificate.ca.create"
	OpCertificateCAUpdate JobOperation = "certificate.ca.update"
	OpCertificateCADelete JobOperation = "certificate.ca.delete"
)
```

In `internal/job/types.go`, add corresponding aliases:

```go
// Certificate operations.
const (
	OperationCertificateCAList   = client.OpCertificateCAList
	OperationCertificateCACreate = client.OpCertificateCACreate
	OperationCertificateCAUpdate = client.OpCertificateCAUpdate
	OperationCertificateCADelete = client.OpCertificateCADelete
)
```

- [ ] **Step 2: Add permission constants**

In `pkg/sdk/client/permissions.go`, add:

```go
	PermCertificateRead  Permission = "certificate:read"
	PermCertificateWrite Permission = "certificate:write"
```

In `internal/authtoken/permissions.go`:

- Add constants `PermCertificateRead`, `PermCertificateWrite`
- Add both to `AllPermissions`
- Add `PermCertificateRead`, `PermCertificateWrite` to admin role
- Add `PermCertificateRead`, `PermCertificateWrite` to write role
- Add `PermCertificateRead` to read role

- [ ] **Step 3: Write processor tests**

Create `internal/agent/processor_certificate_public_test.go`. Follow the
`processor_schedule_public_test.go` pattern — the certificate provider gets its
own `NewCertificateProcessor`.

**TestProcessCertificateOperation** — dispatch-level:

- nil provider returns error
- invalid operation format
- unsupported sub-operation

**TestProcessCertificateCAList** — table-driven:

- success
- provider error

**TestProcessCertificateCACreate** — table-driven:

- success
- unmarshal error
- provider error

**TestProcessCertificateCAUpdate** — table-driven:

- success
- unmarshal error
- provider error

**TestProcessCertificateCADelete** — table-driven:

- success
- unmarshal error
- provider error

- [ ] **Step 4: Implement processor**

Create `internal/agent/processor_certificate.go`. Follow `processor_schedule.go`
pattern:

```go
func NewCertificateProcessor(
	certProvider certProv.Provider,
	logger *slog.Logger,
) ProcessorFunc {
	return func(req job.Request) (json.RawMessage, error) {
		if certProvider == nil {
			return nil, fmt.Errorf(
				"certificate provider not available",
			)
		}
		baseOperation := strings.Split(
			req.Operation, ".")[0]
		switch baseOperation {
		case "ca":
			return processCertificateCAOperation(
				certProvider, logger, req)
		default:
			return nil, fmt.Errorf(
				"unsupported certificate operation: %s",
				req.Operation)
		}
	}
}
```

`processCertificateCAOperation` splits on `.` to get sub-op (`list`, `create`,
`update`, `delete`) and dispatches.

- [ ] **Step 5: Wire in agent_setup.go**

Add import:

```go
certProv "github.com/retr0h/osapi/internal/provider/node/certificate"
```

Add factory function `createCertificateProvider` — on Debian, needs
`fileProvider`, `fileStateKV`, `execManager`, `hostname`. If
`fileProvider == nil`, log warning and return Linux stub. No container check.
Darwin/Linux return stubs.

Register as its own category:

```go
registry.Register("certificate",
	agent.NewCertificateProcessor(certProvider, log),
	certProvider,
)
```

- [ ] **Step 6: Run tests and verify**

```bash
go test -v ./internal/agent/...
go build ./...
```

- [ ] **Step 7: Verify 100% coverage on processor**

```bash
go test -coverprofile=/tmp/cert_proc.cov \
  ./internal/agent/... && \
  go tool cover -func=/tmp/cert_proc.cov | \
  grep "processor_certificate"
```

- [ ] **Step 8: Commit**

```bash
git add pkg/sdk/client/operations.go internal/job/types.go \
  pkg/sdk/client/permissions.go \
  internal/authtoken/permissions.go \
  internal/agent/processor_certificate.go \
  internal/agent/processor_certificate_public_test.go \
  cmd/agent_setup.go
git commit -m "feat(certificate): add operations, permissions, and agent wiring"
```

---

### Task 5: OpenAPI Spec and Code Generation

**Files:**

- Create: `internal/controller/api/node/certificate/gen/api.yaml`
- Create: `internal/controller/api/node/certificate/gen/cfg.yaml`
- Create: `internal/controller/api/node/certificate/gen/generate.go`

- [ ] **Step 1: Create OpenAPI spec**

Read `internal/controller/api/node/schedule/gen/api.yaml` as the reference.
Create `api.yaml` with:

- Tag: `certificate_operations`, displayName `Node/Certificate`
- Paths:
  - `GET /node/{hostname}/certificate/ca` (operationId: `GetNodeCertificateCa`,
    security: `certificate:read`)
  - `POST /node/{hostname}/certificate/ca` (operationId:
    `PostNodeCertificateCa`, security: `certificate:write`)
  - `PUT /node/{hostname}/certificate/ca/{name}` (operationId:
    `PutNodeCertificateCa`, security: `certificate:write`)
  - `DELETE /node/{hostname}/certificate/ca/{name}` (operationId:
    `DeleteNodeCertificateCa`, security: `certificate:write`)
- Parameters: Hostname (path), CertName (path, `name`)
- Request body for POST: `CertificateCACreateRequest` with `name` (required,
  validate: required,min=1) and `object` (required, validate: required,min=1)
- Request body for PUT: `CertificateCAUpdateRequest` with `object` (required,
  validate: required,min=1). Name comes from path.
- Schemas:
  - `CertificateCAInfo` — name (string), source (string enum system/custom),
    object (string)
  - `CertificateCAEntry` — hostname, status (ok/failed/skipped), certificates
    (array of CertificateCAInfo), error
  - `CertificateCACollectionResponse` — job_id, results (array of
    CertificateCAEntry)
  - `CertificateCAMutationEntry` — hostname, status, name, changed (boolean),
    error
  - `CertificateCAMutationResponse` — job_id, results (array of
    CertificateCAMutationEntry)
- Responses: 200, 400 (for POST/PUT), 401, 403, 404 (for PUT/DELETE), 500

- [ ] **Step 2: Create cfg.yaml and generate.go**

`cfg.yaml`:

```yaml
package: gen
output: certificate.gen.go
generate:
  models: true
  echo-server: true
  strict-server: true
import-mapping:
  ../../../common/gen/api.yaml: github.com/retr0h/osapi/internal/controller/api/common/gen
output-options:
  skip-prune: true
```

`generate.go`:

```go
package gen
//go:generate go tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config cfg.yaml api.yaml
```

- [ ] **Step 3: Generate code and rebuild combined spec**

```bash
go generate ./internal/controller/api/node/certificate/gen/...
just generate
go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add internal/controller/api/node/certificate/gen/
git commit -m "feat(certificate): add OpenAPI spec and generated code"
```

---

### Task 6: API Handler Implementation

**Files:**

- Create all handler files under `internal/controller/api/node/certificate/`
- Modify: `cmd/controller_setup.go`
- Test: all `*_public_test.go` files

- [ ] **Step 1: Create handler scaffolding**

Create `types.go`, `certificate.go`, `validate.go`, `handler.go` following the
same pattern as `internal/controller/api/node/schedule/`. The handler struct is
`Certificate` with `JobClient` and `logger`.

Compile-time check:

```go
var _ gen.StrictServerInterface = (*Certificate)(nil)
```

Subsystem: `"api.certificate"`.

- [ ] **Step 2: Implement list handler**

Create `ca_list_get.go` — `GetNodeCertificateCa` method.

- Validate hostname
- If broadcast: `QueryBroadcast` with category `"certificate"` and
  `job.OperationCertificateCAList`
- Single target: `Query` with same
- Parse response: unmarshal `[]certProv.Entry` from resp.Data, convert to
  `[]gen.CertificateCAInfo`
- Return `gen.GetNodeCertificateCa200JSONResponse`

- [ ] **Step 3: Implement create handler**

Create `ca_create_post.go` — `PostNodeCertificateCa` method.

- Validate hostname
- Validate request body
- Build entry from request body (name + object)
- If broadcast: `ModifyBroadcast` with category `"certificate"` and
  `job.OperationCertificateCACreate`
- Single target: `Modify` with same
- Parse mutation response (name, changed)
- Handle 400 (validation), 401, 403, 500

- [ ] **Step 4: Implement update handler**

Create `ca_update_put.go` — `PutNodeCertificateCa` method.

- Validate hostname
- Name from path param `request.Name`
- Object from request body
- Build entry with name + object
- `Modify` with `job.OperationCertificateCAUpdate`
- Handle 404 (not found)

- [ ] **Step 5: Implement delete handler**

Create `ca_delete.go` — `DeleteNodeCertificateCa` method.

- Validate hostname
- Name from path param
- `Modify` with `job.OperationCertificateCADelete` and
  `map[string]string{"name": name}`

- [ ] **Step 6: Write handler tests**

Create test files for all four handlers. Follow the pattern in
`internal/controller/api/node/schedule/cron_create_public_test.go` and similar.
Each test file needs:

- Table-driven tests with success, error, skipped, broadcast cases
- HTTP wiring tests (`TestXxxHTTP`)
- RBAC tests (`TestXxxRBACHTTP`) — 401, 403, 200

- [ ] **Step 7: Register in controller_setup.go**

Add import:

```go
certAPI "github.com/retr0h/osapi/internal/controller/api/node/certificate"
```

Add:

```go
handlers = append(handlers,
    certAPI.Handler(log, jc, signingKey, customRoles)...)
```

- [ ] **Step 8: Run tests and verify coverage**

```bash
go test -v ./internal/controller/api/node/certificate/...
go build ./...
go test -coverprofile=/tmp/cert_handler.cov \
  ./internal/controller/api/node/certificate/... && \
  go tool cover -func=/tmp/cert_handler.cov | \
  grep -v "100.0%" | grep -v "gen/"
```

- [ ] **Step 9: Commit**

```bash
git add internal/controller/api/node/certificate/ \
  cmd/controller_setup.go
git commit -m "feat(certificate): add API handlers with broadcast support"
```

---

### Task 7: SDK Service

**Files:**

- Create: `pkg/sdk/client/certificate.go`
- Create: `pkg/sdk/client/certificate_types.go`
- Modify: `pkg/sdk/client/osapi.go`
- Test: `pkg/sdk/client/certificate_public_test.go`
- Test: `pkg/sdk/client/certificate_types_public_test.go`

- [ ] **Step 1: Implement types**

Create `certificate_types.go` with:

- `CertificateCAResult` — Hostname, Status, Certificates ([]CertificateCA),
  Error
- `CertificateCA` — Name, Source, Object
- `CertificateCAMutationResult` — Hostname, Status, Name, Changed, Error
- `CertificateCreateOpts` — Name, Object
- `CertificateUpdateOpts` — Object
- Conversion functions from gen types

- [ ] **Step 2: Implement service**

Create `certificate.go` with `CertificateService`:

- `List(ctx, hostname)` → `*Response[Collection[CertificateCAResult]]`
- `Create(ctx, hostname, opts CertificateCreateOpts)` →
  `*Response[Collection[CertificateCAMutationResult]]`
- `Update(ctx, hostname, name, opts CertificateUpdateOpts)` →
  `*Response[Collection[CertificateCAMutationResult]]`
- `Delete(ctx, hostname, name)` →
  `*Response[Collection[CertificateCAMutationResult]]`

Each method: build gen params/body, call generated client, checkError, nil
guard, convert, return.

- [ ] **Step 3: Wire in osapi.go**

Add `Certificate *CertificateService` to Client struct and initialize in
`New()`.

- [ ] **Step 4: Regenerate SDK client**

```bash
go generate ./pkg/sdk/client/gen/...
```

- [ ] **Step 5: Write tests**

`certificate_public_test.go` — httptest.Server tests for all 4 methods, covering
200, 400, 401, 403, 404, 500, nil body, transport error.

`certificate_types_public_test.go` — conversion function tests.

- [ ] **Step 6: Verify 100% coverage**

```bash
go test -coverprofile=/tmp/cert_sdk.cov \
  ./pkg/sdk/client/... && \
  go tool cover -func=/tmp/cert_sdk.cov | \
  grep "certificate" | grep -v "100.0%"
```

- [ ] **Step 7: Commit**

```bash
git add pkg/sdk/client/certificate.go \
  pkg/sdk/client/certificate_types.go \
  pkg/sdk/client/certificate_public_test.go \
  pkg/sdk/client/certificate_types_public_test.go \
  pkg/sdk/client/osapi.go pkg/sdk/client/gen/
git commit -m "feat(certificate): add SDK service with tests"
```

---

### Task 8: CLI Commands

**Files:**

- Create: `cmd/client_node_certificate.go`
- Create: `cmd/client_node_certificate_list.go`
- Create: `cmd/client_node_certificate_create.go`
- Create: `cmd/client_node_certificate_update.go`
- Create: `cmd/client_node_certificate_delete.go`

- [ ] **Step 1: Create parent command**

```go
var clientNodeCertificateCmd = &cobra.Command{
	Use:   "certificate",
	Short: "Manage CA certificates",
}

func init() {
	clientNodeCmd.AddCommand(clientNodeCertificateCmd)
}
```

- [ ] **Step 2: Create list subcommand**

`client_node_certificate_list.go`:

- Calls `sdkClient.Certificate.List(ctx, host)`
- Table headers: `NAME`, `SOURCE`
- Uses `BuildBroadcastTable`

- [ ] **Step 3: Create create subcommand**

`client_node_certificate_create.go`:

- Flags: `--name` (required), `--object` (required)
- Calls `sdkClient.Certificate.Create(ctx, host, opts)`
- Uses `BuildMutationTable` with headers `NAME`, `CHANGED`

- [ ] **Step 4: Create update subcommand**

`client_node_certificate_update.go`:

- Flags: `--name` (required), `--object` (required)
- Calls `sdkClient.Certificate.Update(ctx, host, name, opts)`
- Uses `BuildMutationTable`

- [ ] **Step 5: Create delete subcommand**

`client_node_certificate_delete.go`:

- Flags: `--name` (required)
- Calls `sdkClient.Certificate.Delete(ctx, host, name)`
- Uses `BuildMutationTable`

- [ ] **Step 6: Verify build**

```bash
go build ./...
```

- [ ] **Step 7: Commit**

```bash
git add cmd/client_node_certificate*.go
git commit -m "feat(certificate): add CLI commands for CA cert management"
```

---

### Task 9: Documentation and SDK Example

**Files:**

- Create all doc files listed in File Structure
- Modify all cross-reference files

- [ ] **Step 1: Create feature page**

`docs/docs/sidebar/features/certificate-management.md`:

- How It Works (List, Create, Update, Delete)
- Operations table (4 operations)
- CLI Usage examples
- Broadcast Support
- Supported Platforms (Debian: Full, Darwin: Skipped, Linux: Skipped)
- No container restriction
- Permissions: `certificate:read` (list), `certificate:write` (create, update,
  delete)
- Related links

- [ ] **Step 2: Create CLI doc pages**

Landing page + list.md, create.md, update.md, delete.md with usage, flags, and
output examples.

- [ ] **Step 3: Create SDK doc page**

`docs/docs/sidebar/sdk/client/management/certificate.md`:

- Methods table (List, Create, Update, Delete)
- Request/result types
- Usage examples
- Permissions

- [ ] **Step 4: Create SDK example**

`examples/sdk/client/certificate.go` — demonstrate List, Create, Delete with
cleanup-first pattern. Under ~100 lines.

- [ ] **Step 5: Update cross-references**

- `features/features.md` — add row
- `features/authentication.md` — add `certificate:read`, `certificate:write` to
  role tables
- `usage/configuration.md` — add permissions
- `architecture/architecture.md` — add feature link
- `architecture/api-guidelines.md` — add 4 endpoint rows
- `docusaurus.config.ts` — add to Features + SDK dropdowns
- `sdk/client/client.md` — add Certificate to Management table

- [ ] **Step 6: Commit**

```bash
git add docs/ examples/sdk/client/certificate.go
git commit -m "docs: add certificate management feature docs, SDK example, and cross-references"
```

---

### Task 10: Integration Test

**Files:**

- Create: `test/integration/certificate_test.go`

- [ ] **Step 1: Write integration test**

`//go:build integration` tag. Test:

- `osapi client node certificate list --target _any --json` — verify JSON with
  results array containing system certs
- Optionally test create/delete if `OSAPI_INTEGRATION_WRITES=1` is set (guarded
  by `skipWrite`)

- [ ] **Step 2: Commit**

```bash
git add test/integration/certificate_test.go
git commit -m "test(certificate): add integration test"
```

---

### Task 11: Final Verification

- [ ] **Step 1: Run full suite**

```bash
just generate
go build ./...
just go::unit
just go::vet
```

- [ ] **Step 2: Verify coverage on all new code**

```bash
go test -coverprofile=/tmp/cert_all.cov \
  ./internal/provider/node/certificate/... \
  ./internal/agent/... \
  ./internal/controller/api/node/certificate/... \
  ./pkg/sdk/client/...
go tool cover -func=/tmp/cert_all.cov | \
  grep "certificate" | grep -v "100.0%" | \
  grep -v "mocks\|gen/"
```

All new certificate code must be at 100%.

- [ ] **Step 3: Commit any fixes**

```bash
git add -A
git commit -m "chore(certificate): fix formatting and lint"
```
