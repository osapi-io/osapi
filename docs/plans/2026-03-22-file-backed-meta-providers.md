# File-Backed Meta Providers Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development
> (if subagents available) or superpowers:executing-plans to implement this plan.
> Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactor the cron provider to delegate file writes to the file
provider for SHA tracking and idempotency, add file undeploy, and add protected
object support.

**Architecture:** Meta providers (cron, future systemd/sysctl) depend on a
narrow `FileDeployer` interface instead of writing files directly. The file
provider gains an `Undeploy` method. Objects with a `system/` prefix are
protected from deletion.

**Tech Stack:** Go 1.25, NATS JetStream (Object Store + KV), Echo, Cobra,
oapi-codegen (strict-server), testify/suite

---

## Chunk 1: File Provider — Undeploy + FileDeployer Interface

### Task 1: Add FileDeployer interface and Undeploy types

**Files:**
- Modify: `internal/provider/file/types.go`

- [ ] **Step 1: Add UndeployRequest, UndeployResult, and FileDeployer interface**

Add after the existing `Provider` interface:

```go
// UndeployRequest contains parameters for removing a deployed file from disk.
type UndeployRequest struct {
	// Path is the filesystem path to undeploy.
	Path string `json:"path"`
}

// UndeployResult contains the result of a file undeploy operation.
type UndeployResult struct {
	// Changed indicates whether the file was removed.
	Changed bool `json:"changed"`
	// Path is the filesystem path that was undeployed.
	Path string `json:"path"`
}

// FileDeployer is the narrow interface for providers that deploy files
// to well-known paths. Meta providers (cron, systemd, sysctl) depend
// on this instead of the full Provider interface.
type FileDeployer interface {
	// Deploy writes file content from the object store to the target
	// path with SHA tracking and idempotency.
	Deploy(
		ctx context.Context,
		req DeployRequest,
	) (*DeployResult, error)
	// Undeploy removes a deployed file from disk. The object store
	// entry and file-state KV record are preserved.
	Undeploy(
		ctx context.Context,
		req UndeployRequest,
	) (*UndeployResult, error)
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/provider/file/...`

- [ ] **Step 3: Commit**

```bash
git add internal/provider/file/types.go
git commit -m "feat: add FileDeployer interface and Undeploy types to file provider"
```

### Task 2: Implement Undeploy method

**Files:**
- Create: `internal/provider/file/undeploy.go`
- Create: `internal/provider/file/undeploy_public_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/provider/file/undeploy_public_test.go` with a testify/suite.
Follow the pattern from `deploy_public_test.go`. Test cases:

1. "when file exists on disk" — file exists, state exists → removes file,
   updates state with `Undeployed: true`, returns `Changed: true`
2. "when file does not exist" — no file on disk → returns `Changed: false`
3. "when file exists but no state entry" — file exists, no KV entry → removes
   file, returns `Changed: true` (no state to update)
4. "when fs remove fails" — `fs.Remove` returns error → returns error

Each test should:
- Set up afero.MemMapFs with files as needed
- Set up mock stateKV (or real in-memory KV)
- Call `provider.Undeploy(ctx, req)`
- Assert Changed, Path, and file absence on disk

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/provider/file/... -run TestUndeploy -v`
Expected: FAIL (method not implemented)

- [ ] **Step 3: Implement Undeploy**

Create `internal/provider/file/undeploy.go`:

```go
package file

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/retr0h/osapi/internal/job"
)

// Undeploy removes a deployed file from disk. The object store entry
// is preserved. The file-state KV is updated to record the undeploy.
func (p *Service) Undeploy(
	ctx context.Context,
	req UndeployRequest,
) (*UndeployResult, error) {
	// Check if file exists on disk.
	_, err := p.fs.Stat(req.Path)
	if err != nil {
		p.logger.Debug(
			"file not on disk, nothing to undeploy",
			slog.String("path", req.Path),
		)

		return &UndeployResult{
			Changed: false,
			Path:    req.Path,
		}, nil
	}

	if err := p.fs.Remove(req.Path); err != nil {
		return nil, fmt.Errorf("failed to remove file %q: %w", req.Path, err)
	}

	// Update file-state KV if entry exists.
	stateKey := buildStateKey(p.hostname, req.Path)
	entry, err := p.stateKV.Get(ctx, stateKey)
	if err == nil {
		var state job.FileState
		if unmarshalErr := json.Unmarshal(entry.Value(), &state); unmarshalErr == nil {
			state.UndeployedAt = time.Now().UTC().Format(time.RFC3339)

			stateBytes, marshalErr := marshalJSON(state)
			if marshalErr == nil {
				_, _ = p.stateKV.Put(ctx, stateKey, stateBytes)
			}
		}
	}

	p.logger.Info(
		"file undeployed",
		slog.String("path", req.Path),
		slog.Bool("changed", true),
	)

	return &UndeployResult{
		Changed: true,
		Path:    req.Path,
	}, nil
}
```

- [ ] **Step 4: Add UndeployedAt to FileState**

In `internal/job/types.go`, add `UndeployedAt` field to `FileState`:

```go
type FileState struct {
	ObjectName   string `json:"object_name"`
	Path         string `json:"path"`
	SHA256       string `json:"sha256"`
	Mode         string `json:"mode,omitempty"`
	Owner        string `json:"owner,omitempty"`
	Group        string `json:"group,omitempty"`
	DeployedAt   string `json:"deployed_at"`
	ContentType  string `json:"content_type"`
	UndeployedAt string `json:"undeployed_at,omitempty"`
}
```

- [ ] **Step 5: Run tests**

Run: `go test ./internal/provider/file/... -v`
Expected: All pass

- [ ] **Step 6: Commit**

```bash
git add internal/provider/file/undeploy.go \
        internal/provider/file/undeploy_public_test.go \
        internal/job/types.go
git commit -m "feat: implement Undeploy method on file provider"
```

### Task 3: Add file undeploy to agent processor

**Files:**
- Modify: `internal/agent/processor_file.go`
- Modify: `pkg/sdk/client/operations.go`
- Modify: `internal/job/types.go` (operation constant alias)

- [ ] **Step 1: Add operation constant**

In `pkg/sdk/client/operations.go`, add:

```go
OpFileUndeploy JobOperation = "file.undeploy.execute"
```

In `internal/job/types.go`, add alias:

```go
OperationFileUndeployExecute = client.OpFileUndeploy
```

- [ ] **Step 2: Add undeploy case to processor**

In `internal/agent/processor_file.go`, add `case "undeploy"` to the switch:

```go
case "undeploy":
	return a.processFileUndeploy(jobRequest)
```

Add the handler method:

```go
func (a *Agent) processFileUndeploy(
	jobRequest job.Request,
) (json.RawMessage, error) {
	var req fileProv.UndeployRequest
	if err := json.Unmarshal(jobRequest.Data, &req); err != nil {
		return nil, fmt.Errorf("failed to parse file undeploy data: %w", err)
	}

	result, err := a.fileProvider.Undeploy(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("file undeploy failed: %w", err)
	}

	return json.Marshal(result)
}
```

- [ ] **Step 3: Update file Provider interface**

In `internal/provider/file/types.go`, add `Undeploy` to the existing `Provider`
interface:

```go
type Provider interface {
	Deploy(ctx context.Context, req DeployRequest) (*DeployResult, error)
	Undeploy(ctx context.Context, req UndeployRequest) (*UndeployResult, error)
	Status(ctx context.Context, req StatusRequest) (*StatusResult, error)
}
```

- [ ] **Step 4: Regenerate mocks if needed, verify compile**

Run: `go build ./internal/agent/...`

- [ ] **Step 5: Commit**

```bash
git add internal/agent/processor_file.go \
        pkg/sdk/client/operations.go \
        internal/job/types.go \
        internal/provider/file/types.go
git commit -m "feat: add file undeploy operation to agent processor"
```

### Task 4: Add file undeploy API endpoint

**Files:**
- Modify: `internal/controller/api/node/gen/api.yaml`
- Create: `internal/controller/api/node/file_undeploy_post.go`
- Create: `internal/controller/api/node/file_undeploy_post_public_test.go`
- Modify: `internal/job/client/file.go` (add `ModifyFileUndeploy`)

- [ ] **Step 1: Add endpoint to OpenAPI spec**

In `internal/controller/api/node/gen/api.yaml`, add a new path:

```yaml
/node/{hostname}/file/undeploy:
  post:
    summary: Undeploy a file
    description: >
      Remove a deployed file from disk on the target node. The object
      stays in the store for redeployment or audit.
    tags:
      - file_operations
    operationId: PostNodeFileUndeploy
    security:
      - BearerAuth:
          - file:write
    parameters:
      - $ref: '#/components/parameters/Hostname'
    requestBody:
      required: true
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/FileUndeployRequest'
    responses:
      '200':
        description: File undeployed.
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/FileUndeployResponse'
      '400':
        description: Invalid request.
        content:
          application/json:
            schema:
              $ref: '../../common/gen/api.yaml#/components/schemas/ErrorResponse'
      '401': ...
      '403': ...
      '500': ...
```

Add schemas:

```yaml
FileUndeployRequest:
  type: object
  required:
    - path
  properties:
    path:
      type: string
      description: Filesystem path to undeploy.
      x-oapi-codegen-extra-tags:
        validate: "required,min=1"

FileUndeployResponse:
  type: object
  properties:
    job_id:
      type: string
    hostname:
      type: string
    changed:
      type: boolean
```

- [ ] **Step 2: Regenerate code**

Run: `just generate`

- [ ] **Step 3: Add ModifyFileUndeploy to job client**

In `internal/job/client/file.go`, add:

```go
func (c *Client) ModifyFileUndeploy(
	ctx context.Context,
	hostname string,
	path string,
) (string, string, bool, error) {
	req := fileProv.UndeployRequest{Path: path}
	// ... same pattern as ModifyFileDeploy but with OperationFileUndeployExecute
}
```

- [ ] **Step 4: Implement handler**

Create `internal/controller/api/node/file_undeploy_post.go` following the
pattern from `file_deploy_post.go`.

- [ ] **Step 5: Write tests**

Create `internal/controller/api/node/file_undeploy_post_public_test.go`.

- [ ] **Step 6: Verify**

Run: `go test ./internal/controller/api/node/... -v`

- [ ] **Step 7: Commit**

```bash
git add internal/controller/api/node/gen/api.yaml \
        internal/controller/api/node/file_undeploy_post.go \
        internal/controller/api/node/file_undeploy_post_public_test.go \
        internal/job/client/file.go
git commit -m "feat: add file undeploy API endpoint"
```

### Task 5: Add file undeploy to SDK and CLI

**Files:**
- Modify: `pkg/sdk/client/node.go` (or `file.go` if separate)
- Modify: `pkg/sdk/client/file_types.go`
- Create: `cmd/client_node_file_undeploy.go`

- [ ] **Step 1: Add SDK method**

Add `FileUndeploy` method to the appropriate service following existing
`FileDeploy` pattern. Add `FileUndeployOpts` and `FileUndeployResult` types.

- [ ] **Step 2: Add CLI command**

Create `cmd/client_node_file_undeploy.go` with `--path` flag following the
pattern from `client_node_file_deploy.go`.

- [ ] **Step 3: Regenerate SDK client**

Run: `go generate ./pkg/sdk/client/gen/...`

- [ ] **Step 4: Write tests and verify**

Run: `go test ./pkg/sdk/client/... -v && go build ./cmd/...`

- [ ] **Step 5: Commit**

```bash
git add pkg/sdk/client/ cmd/client_node_file_undeploy.go
git commit -m "feat: add file undeploy to SDK and CLI"
```

---

## Chunk 2: Protected Objects

### Task 6: Add protected object check to file delete handler

**Files:**
- Modify: `internal/controller/api/file/file_delete.go`
- Modify: `internal/controller/api/file/file_delete_public_test.go`
- Modify: `internal/controller/api/file/file_list.go` (add source column)

- [ ] **Step 1: Add protected check to DeleteFileByName**

In `file_delete.go`, add check before deletion:

```go
if strings.HasPrefix(request.Name, "system/") {
	errMsg := fmt.Sprintf("cannot delete system file: %s", request.Name)
	return gen.DeleteFileByName403JSONResponse{Error: &errMsg}, nil
}
```

- [ ] **Step 2: Add source field to file list response**

In `file_list.go`, when building the response, set source based on prefix:

```go
source := "user"
if strings.HasPrefix(info.Name, "system/") {
	source = "system"
}
```

Update the OpenAPI spec to include `source` field in the file list response.

- [ ] **Step 3: Write tests**

Add test cases:
- "when deleting system file returns 403"
- "when deleting user file succeeds"
- "when listing shows source column"

- [ ] **Step 4: Verify**

Run: `go test ./internal/controller/api/file/... -v`

- [ ] **Step 5: Update SDK file list types**

Add `Source` field to `FileItem` in `pkg/sdk/client/file_types.go`.

- [ ] **Step 6: Update CLI file list**

Add SOURCE column to `cmd/client_file_list.go` table output.

- [ ] **Step 7: Commit**

```bash
git add internal/controller/api/file/ pkg/sdk/client/file_types.go \
        cmd/client_file_list.go
git commit -m "feat: add protected object support and source column to file list"
```

---

## Chunk 3: Cron Provider Refactor

### Task 7: Update cron provider interface and Entry type

**Files:**
- Modify: `internal/provider/scheduled/cron/types.go`

- [ ] **Step 1: Update Entry struct**

Replace `Command string` with `Object string`, add `ContentType` and `Vars`:

```go
type Entry struct {
	Name        string         `json:"name"`
	Object      string         `json:"object,omitempty"`
	Schedule    string         `json:"schedule,omitempty"`
	Interval    string         `json:"interval,omitempty"`
	Source      string         `json:"source,omitempty"`
	User        string         `json:"user,omitempty"`
	ContentType string         `json:"content_type,omitempty"`
	Vars        map[string]any `json:"vars,omitempty"`
}
```

- [ ] **Step 2: Update Provider interface**

Add `context.Context` to all methods that will now call the file provider:

```go
type Provider interface {
	List(ctx context.Context) ([]Entry, error)
	Get(ctx context.Context, name string) (*Entry, error)
	Create(ctx context.Context, entry Entry) (*CreateResult, error)
	Update(ctx context.Context, entry Entry) (*UpdateResult, error)
	Delete(ctx context.Context, name string) (*DeleteResult, error)
}
```

- [ ] **Step 3: Commit**

```bash
git add internal/provider/scheduled/cron/types.go
git commit -m "refactor: update cron Entry type and Provider interface for file-backed pattern"
```

### Task 8: Refactor Debian cron provider to use FileDeployer

**Files:**
- Modify: `internal/provider/scheduled/cron/debian.go`
- Modify: `internal/provider/scheduled/cron/debian_public_test.go`

- [ ] **Step 1: Update Debian struct and constructor**

```go
type Debian struct {
	logger       *slog.Logger
	fs           afero.Fs
	fileDeployer file.FileDeployer
	stateKV      jetstream.KeyValue
	hostname     string
}

func NewDebianProvider(
	logger *slog.Logger,
	fs afero.Fs,
	fileDeployer file.FileDeployer,
	stateKV jetstream.KeyValue,
	hostname string,
) *Debian {
	return &Debian{
		logger:       logger,
		fs:           fs,
		fileDeployer: fileDeployer,
		stateKV:      stateKV,
		hostname:     hostname,
	}
}
```

Note: `stateKV` and `hostname` are needed for `List`/`Get` to check which files
are managed by osapi (replacing `# Managed by osapi` header).

- [ ] **Step 2: Refactor Create**

Replace direct `afero.WriteFile` with `fileDeployer.Deploy()`:

```go
func (d *Debian) Create(
	ctx context.Context,
	entry Entry,
) (*CreateResult, error) {
	if err := validateName(entry.Name); err != nil {
		return nil, err
	}

	if existingPath, _ := d.findEntryPath(entry.Name); existingPath != "" {
		return nil, fmt.Errorf("cron entry %q already exists", entry.Name)
	}

	filePath, perm := d.entryFilePath(entry)

	result, err := d.fileDeployer.Deploy(ctx, file.DeployRequest{
		ObjectName:  entry.Object,
		Path:        filePath,
		Mode:        fmt.Sprintf("%04o", perm),
		ContentType: entry.ContentType,
		Vars:        entry.Vars,
	})
	if err != nil {
		return nil, fmt.Errorf("create cron entry: %w", err)
	}

	return &CreateResult{
		Name:    entry.Name,
		Changed: result.Changed,
	}, nil
}
```

- [ ] **Step 3: Refactor Update**

Same pattern — call `fileDeployer.Deploy()` (idempotent, SHA check handles
unchanged content):

```go
func (d *Debian) Update(
	ctx context.Context,
	entry Entry,
) (*UpdateResult, error) {
	if err := validateName(entry.Name); err != nil {
		return nil, err
	}

	filePath, perm := d.findEntryPath(entry.Name)
	if filePath == "" {
		return nil, fmt.Errorf("cron entry %q does not exist", entry.Name)
	}

	result, err := d.fileDeployer.Deploy(ctx, file.DeployRequest{
		ObjectName:  entry.Object,
		Path:        filePath,
		Mode:        fmt.Sprintf("%04o", perm),
		ContentType: entry.ContentType,
		Vars:        entry.Vars,
	})
	if err != nil {
		return nil, fmt.Errorf("update cron entry: %w", err)
	}

	return &UpdateResult{
		Name:    entry.Name,
		Changed: result.Changed,
	}, nil
}
```

- [ ] **Step 4: Refactor Delete**

Use `fileDeployer.Undeploy()`:

```go
func (d *Debian) Delete(
	ctx context.Context,
	name string,
) (*DeleteResult, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}

	filePath, _ := d.findEntryPath(name)
	if filePath == "" {
		return &DeleteResult{
			Name:    name,
			Changed: false,
		}, nil
	}

	result, err := d.fileDeployer.Undeploy(ctx, file.UndeployRequest{
		Path: filePath,
	})
	if err != nil {
		return nil, fmt.Errorf("delete cron entry: %w", err)
	}

	return &DeleteResult{
		Name:    name,
		Changed: result.Changed,
	}, nil
}
```

- [ ] **Step 5: Refactor List**

Replace `# Managed by osapi` header check with file-state KV lookup:

```go
func (d *Debian) List(
	ctx context.Context,
) ([]Entry, error) {
	var result []Entry

	cronDirEntries, err := afero.ReadDir(d.fs, cronDir)
	if err != nil {
		return nil, fmt.Errorf("list cron entries: %w", err)
	}

	for _, entry := range cronDirEntries {
		if entry.IsDir() {
			continue
		}

		if !d.isManagedFile(ctx, cronDir+"/"+entry.Name()) {
			continue
		}

		cronEntry := d.buildEntryFromState(ctx, entry.Name(), cronDir, "cron.d")
		if cronEntry != nil {
			result = append(result, *cronEntry)
		}
	}

	for _, interval := range periodicIntervals {
		dir := periodicDirs[interval]
		dirEntries, err := afero.ReadDir(d.fs, dir)
		if err != nil {
			continue
		}

		for _, entry := range dirEntries {
			if entry.IsDir() {
				continue
			}

			if !d.isManagedFile(ctx, dir+"/"+entry.Name()) {
				continue
			}

			cronEntry := d.buildEntryFromState(ctx, entry.Name(), dir, interval)
			if cronEntry != nil {
				result = append(result, *cronEntry)
			}
		}
	}

	return result, nil
}
```

Add helper methods:

```go
// isManagedFile checks if the file at path has a file-state KV entry.
func (d *Debian) isManagedFile(
	ctx context.Context,
	path string,
) bool {
	stateKey := file.BuildStateKey(d.hostname, path)
	_, err := d.stateKV.Get(ctx, stateKey)
	return err == nil
}

// buildEntryFromState creates an Entry from file-state KV metadata.
func (d *Debian) buildEntryFromState(
	ctx context.Context,
	name string,
	dir string,
	source string,
) *Entry {
	// Build entry from state metadata and filesystem
	// ...
}
```

Note: `file.BuildStateKey` needs to be exported (currently `buildStateKey`
is unexported). Export it in `internal/provider/file/deploy.go`.

- [ ] **Step 6: Refactor Get**

Replace header-based parsing with state-based lookup:

```go
func (d *Debian) Get(
	ctx context.Context,
	name string,
) (*Entry, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}

	filePath, _ := d.findEntryPath(name)
	if filePath == "" {
		return nil, fmt.Errorf("cron entry %q: not found", name)
	}

	if !d.isManagedFile(ctx, filePath) {
		return nil, fmt.Errorf("cron entry %q is not managed by osapi", name)
	}

	return d.buildEntryFromPath(ctx, name, filePath)
}
```

- [ ] **Step 7: Remove dead code**

Delete `buildFileContent()`, `readCronFile()`, `readPeriodicFile()`, and the
`managedHeader` constant.

- [ ] **Step 8: Update tests**

Rewrite `debian_public_test.go` to use mock `FileDeployer` and `KeyValue`
interfaces instead of checking file content for `# Managed by osapi`.

- [ ] **Step 9: Verify**

Run: `go test ./internal/provider/scheduled/cron/... -v`

- [ ] **Step 10: Commit**

```bash
git add internal/provider/scheduled/cron/ internal/provider/file/deploy.go
git commit -m "refactor: cron provider uses FileDeployer instead of direct file writes"
```

### Task 9: Update Darwin and Linux stubs

**Files:**
- Modify: `internal/provider/scheduled/cron/darwin.go`
- Modify: `internal/provider/scheduled/cron/linux.go`

- [ ] **Step 1: Add context.Context to stub methods**

Both stubs return `ErrUnsupported` but need updated signatures:

```go
func (d *Darwin) List(ctx context.Context) ([]Entry, error) {
	return nil, provider.ErrUnsupported
}
// ... same for Get, Create, Update, Delete
```

- [ ] **Step 2: Verify**

Run: `go build ./internal/provider/scheduled/cron/...`

- [ ] **Step 3: Commit**

```bash
git add internal/provider/scheduled/cron/darwin.go \
        internal/provider/scheduled/cron/linux.go
git commit -m "refactor: update cron provider stubs for context parameter"
```

### Task 10: Update agent processor and wiring

**Files:**
- Modify: `internal/agent/processor_schedule.go`
- Modify: `internal/agent/factory.go`
- Modify: `internal/agent/types.go`
- Modify: `cmd/agent_setup.go`

- [ ] **Step 1: Pass context to cron provider calls**

In `processor_schedule.go`, pass `context.Background()` (or job context) to all
cron provider calls.

- [ ] **Step 2: Update factory to accept file provider**

In `factory.go`, `CreateProviders` needs to accept the file provider and pass it
to `NewDebianProvider`. Alternatively, create the cron provider in
`cmd/agent_setup.go` after both providers are initialized.

The cleanest approach: create the cron provider in `agent_setup.go` after the
file provider, since the cron provider depends on the file provider + stateKV +
hostname which are all available there:

```go
var cronProvider cronProv.Provider
switch platform.Detect() {
case "debian":
	cronProvider = cronProv.NewDebianProvider(
		logger, appFs, fileProvider, fileStateKV, hostname)
// ...
}
```

- [ ] **Step 3: Verify**

Run: `go build ./... && go test ./internal/agent/... -v`

- [ ] **Step 4: Commit**

```bash
git add internal/agent/ cmd/agent_setup.go
git commit -m "refactor: wire cron provider with file deployer dependency"
```

---

## Chunk 4: Cron API, Handler, Job Client Changes

### Task 11: Update cron OpenAPI spec

**Files:**
- Modify: `internal/controller/api/schedule/gen/api.yaml`

- [ ] **Step 1: Update CronCreateRequest**

Replace `command` with `object`, add `content_type` and `vars`:

```yaml
CronCreateRequest:
  type: object
  required:
    - name
    - object
  properties:
    name:
      type: string
      x-oapi-codegen-extra-tags:
        validate: "required,min=1,max=64"
    object:
      type: string
      description: >
        Name of the uploaded file in the object store.
      x-oapi-codegen-extra-tags:
        validate: "required,min=1"
    schedule:
      type: string
      x-oapi-codegen-extra-tags:
        validate: "required_without=Interval,excluded_with=Interval,omitempty,cron_schedule"
    interval:
      type: string
      enum: [hourly, daily, weekly, monthly]
      x-oapi-codegen-extra-tags:
        validate: "required_without=Schedule,excluded_with=Schedule,omitempty,oneof=hourly daily weekly monthly"
    user:
      type: string
    content_type:
      type: string
      enum: [raw, template]
      x-oapi-codegen-extra-tags:
        validate: "omitempty,oneof=raw template"
    vars:
      type: object
      additionalProperties: true
```

- [ ] **Step 2: Update CronUpdateRequest**

Replace `command` with `object`, add `content_type` and `vars`.

- [ ] **Step 3: Update CronEntry response schema**

Add `object` field, remove `command`. Add `content_type`.

- [ ] **Step 4: Regenerate**

Run: `just generate`

- [ ] **Step 5: Commit**

```bash
git add internal/controller/api/schedule/gen/
git commit -m "feat: update cron OpenAPI spec for file-backed pattern"
```

### Task 12: Update cron handler

**Files:**
- Modify: `internal/controller/api/schedule/cron_create.go`
- Modify: `internal/controller/api/schedule/cron_update.go`
- Modify: `internal/controller/api/schedule/cron_get.go`
- Modify: `internal/controller/api/schedule/cron_list_get.go`
- Modify: all corresponding `*_public_test.go` files

- [ ] **Step 1: Update PostNodeScheduleCron**

Map `Object`, `ContentType`, `Vars` from request body to `cronProv.Entry`
instead of `Command`.

- [ ] **Step 2: Update PutNodeScheduleCron**

Same mapping changes for update.

- [ ] **Step 3: Update response mapping in list/get**

Map `Object`, `ContentType` from response instead of `Command`.

- [ ] **Step 4: Update tests**

Fix all handler test fixtures to use `object` instead of `command`.

- [ ] **Step 5: Verify**

Run: `go test ./internal/controller/api/schedule/... -v`

- [ ] **Step 6: Commit**

```bash
git add internal/controller/api/schedule/
git commit -m "refactor: update cron handlers for file-backed pattern"
```

### Task 13: Update job client for cron

**Files:**
- Modify: `internal/job/client/schedule_cron.go`

- [ ] **Step 1: Update cron job client methods**

The `Entry` type has changed (no `Command`, has `Object`, `ContentType`,
`Vars`). The job client marshals this into the job request data. The methods
themselves don't need logic changes since they pass the full `Entry` — but the
`Entry` struct has changed, so this should just compile.

- [ ] **Step 2: Verify**

Run: `go build ./internal/job/...`

- [ ] **Step 3: Commit**

```bash
git add internal/job/client/schedule_cron.go
git commit -m "refactor: update cron job client for updated Entry type"
```

---

## Chunk 5: SDK, CLI, and Docs

### Task 14: Update SDK cron types and methods

**Files:**
- Modify: `pkg/sdk/client/schedule_types.go`
- Modify: `pkg/sdk/client/schedule.go`

- [ ] **Step 1: Update SDK types**

```go
type CronCreateOpts struct {
	Name        string
	Object      string
	Schedule    string
	Interval    string
	User        string
	ContentType string
	Vars        map[string]any
}

type CronUpdateOpts struct {
	Object      string
	Schedule    string
	User        string
	ContentType string
	Vars        map[string]any
}

type CronEntryResult struct {
	Name        string         `json:"name"`
	Object      string         `json:"object,omitempty"`
	Schedule    string         `json:"schedule,omitempty"`
	Interval    string         `json:"interval,omitempty"`
	Source      string         `json:"source,omitempty"`
	User        string         `json:"user,omitempty"`
	ContentType string         `json:"content_type,omitempty"`
	Error       string         `json:"error,omitempty"`
}
```

- [ ] **Step 2: Update SDK methods**

In `schedule.go`, update `CronCreate` and `CronUpdate` to map `Object`,
`ContentType`, `Vars` to the gen request types. Update conversion functions
to handle new fields.

- [ ] **Step 3: Regenerate SDK client**

Run: `go generate ./pkg/sdk/client/gen/...`

- [ ] **Step 4: Write tests and verify**

Run: `go test ./pkg/sdk/client/... -v`

- [ ] **Step 5: Commit**

```bash
git add pkg/sdk/client/
git commit -m "refactor: update SDK cron types for file-backed pattern"
```

### Task 15: Update CLI cron commands

**Files:**
- Modify: `cmd/client_node_schedule_cron_create.go`
- Modify: `cmd/client_node_schedule_cron_update.go`
- Modify: `cmd/client_node_schedule_cron_get.go`
- Modify: `cmd/client_node_schedule_cron_list.go`

- [ ] **Step 1: Update create command**

Remove `--command` flag, add `--object` (required), `--content-type` (optional,
default "raw"), `--var` (repeatable key=value).

- [ ] **Step 2: Update update command**

Remove `--command`, add `--object`, `--content-type`, `--var`.

- [ ] **Step 3: Update list output**

Replace COMMAND column with OBJECT column.

- [ ] **Step 4: Update get output**

Replace Command field with Object field.

- [ ] **Step 5: Verify**

Run: `go build ./cmd/...`

- [ ] **Step 6: Commit**

```bash
git add cmd/client_node_schedule_cron_*.go
git commit -m "refactor: update CLI cron commands for file-backed pattern"
```

### Task 16: Update SDK example and docs

**Files:**
- Modify: `examples/sdk/client/cron.go`
- Modify: `docs/docs/sidebar/features/cron-management.md`
- Modify: CLI docs for cron commands
- Modify: `docs/docs/sidebar/features/file-management.md` (undeploy section)

- [ ] **Step 1: Update SDK example**

Update `examples/sdk/client/cron.go` to use `Object` instead of `Command`:

```go
createResp, err := c.Schedule.CronCreate(ctx, target, client.CronCreateOpts{
	Name:     "backup-daily",
	Schedule: "0 2 * * *",
	Object:   "backup-script",
	User:     "root",
})
```

- [ ] **Step 2: Update cron feature docs**

Update `docs/docs/sidebar/features/cron-management.md` to document:
- Object-based workflow (upload → create)
- Template support with content_type and vars
- No `# Managed by osapi` header
- File-state KV tracking

- [ ] **Step 3: Update file management docs**

Add undeploy section to `docs/docs/sidebar/features/file-management.md`.
Document protected objects with `system/` prefix.

- [ ] **Step 4: Update CLI docs**

Update CLI reference docs for cron create, update, get, list with new flags.
Add CLI docs for `client node file undeploy`.

- [ ] **Step 5: Commit**

```bash
git add examples/sdk/client/cron.go docs/
git commit -m "docs: update cron and file docs for file-backed meta provider pattern"
```

---

## Chunk 6: System Template Seeding

### Task 17: Add system template seeding on agent startup

**Files:**
- Create: `internal/agent/templates/` (embedded templates)
- Modify: `cmd/agent_setup.go` (seed on startup)

- [ ] **Step 1: Create embedded templates directory**

Create `internal/agent/templates/` with initial system templates. For now, just
a placeholder `system/cron-wrapper.tmpl` that can be used as a reference:

```
#!/bin/sh
# {{ .Vars.description }}
{{ .Vars.command }}
```

- [ ] **Step 2: Add seeding function**

Create `internal/agent/seed.go` with:

```go
//go:embed templates/*
var systemTemplates embed.FS

func SeedSystemTemplates(
	ctx context.Context,
	objStore jetstream.ObjectStore,
) error {
	// Walk embedded templates, upload each with "system/" prefix
	// Skip if already present (idempotent)
}
```

- [ ] **Step 3: Call from agent startup**

In `cmd/agent_setup.go`, call `agent.SeedSystemTemplates()` after Object Store
is initialized.

- [ ] **Step 4: Write tests**

- [ ] **Step 5: Commit**

```bash
git add internal/agent/seed.go internal/agent/templates/ cmd/agent_setup.go
git commit -m "feat: seed system templates on agent startup"
```

---

## Chunk 7: Verification

### Task 18: Full verification

- [ ] **Step 1: Build**

Run: `go build ./...`

- [ ] **Step 2: Unit tests**

Run: `just go::unit`

- [ ] **Step 3: Lint**

Run: `just go::vet`

- [ ] **Step 4: Format**

Run: `just go::fmt`

- [ ] **Step 5: Generate**

Run: `just generate` — verify no diff

- [ ] **Step 6: Commit any fixes**

---

## Files Modified Summary

| File | Change |
| --- | --- |
| `internal/provider/file/types.go` | Add FileDeployer, UndeployRequest/Result, update Provider |
| `internal/provider/file/undeploy.go` | New: Undeploy method |
| `internal/provider/file/undeploy_public_test.go` | New: Undeploy tests |
| `internal/provider/file/deploy.go` | Export BuildStateKey |
| `internal/job/types.go` | Add UndeployedAt to FileState, operation alias |
| `internal/agent/processor_file.go` | Add undeploy case |
| `internal/agent/processor_schedule.go` | Pass context to cron calls |
| `internal/agent/factory.go` | Update cron provider creation |
| `internal/agent/types.go` | No change if factory handles wiring |
| `internal/agent/seed.go` | New: system template seeding |
| `internal/agent/templates/` | New: embedded templates |
| `internal/controller/api/node/gen/api.yaml` | Add undeploy endpoint |
| `internal/controller/api/node/file_undeploy_post.go` | New: undeploy handler |
| `internal/controller/api/node/file_undeploy_post_public_test.go` | New: tests |
| `internal/controller/api/file/file_delete.go` | Protected object check |
| `internal/controller/api/file/file_list.go` | Source column |
| `internal/controller/api/schedule/gen/api.yaml` | command→object, add content_type/vars |
| `internal/controller/api/schedule/cron_create.go` | Map new fields |
| `internal/controller/api/schedule/cron_update.go` | Map new fields |
| `internal/controller/api/schedule/cron_get.go` | Map new fields |
| `internal/controller/api/schedule/cron_list_get.go` | Map new fields |
| `internal/provider/scheduled/cron/types.go` | Update Entry, add ctx to interface |
| `internal/provider/scheduled/cron/debian.go` | Full refactor to FileDeployer |
| `internal/provider/scheduled/cron/darwin.go` | Add ctx to stubs |
| `internal/provider/scheduled/cron/linux.go` | Add ctx to stubs |
| `internal/job/client/file.go` | Add ModifyFileUndeploy |
| `pkg/sdk/client/operations.go` | Add OpFileUndeploy |
| `pkg/sdk/client/schedule_types.go` | command→object, add content_type/vars |
| `pkg/sdk/client/schedule.go` | Map new fields |
| `pkg/sdk/client/file_types.go` | Add FileUndeployResult, Source to FileItem |
| `cmd/agent_setup.go` | Cron provider wiring, template seeding |
| `cmd/client_node_file_undeploy.go` | New: undeploy CLI command |
| `cmd/client_node_schedule_cron_create.go` | --command→--object, add flags |
| `cmd/client_node_schedule_cron_update.go` | Same flag changes |
| `cmd/client_node_schedule_cron_list.go` | OBJECT column |
| `cmd/client_node_schedule_cron_get.go` | Object field |
| `examples/sdk/client/cron.go` | Use Object instead of Command |
| `docs/` | Feature pages, CLI docs |
