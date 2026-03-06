# File Deploy & Template Rendering Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to
> implement this plan task-by-task.

**Goal:** Add file management (upload/list/get/delete via Object Store),
file deployment with SHA-based idempotency, and Go template rendering
with per-host facts.

**Architecture:** NATS Object Store as shared blob storage, dedicated
`file-state` KV for SHA tracking, single `file.deploy` job operation
with `content_type` flag (raw/template), agent-side `text/template`
rendering. Object Store is a shared primitive — future providers
(firmware, packages, certs) reuse the same infrastructure.

**Tech Stack:** Go 1.25, NATS JetStream Object Store, `text/template`,
oapi-codegen, testify/suite, gomock.

**Design doc:** `docs/plans/2026-03-06-file-deploy-template-design.md`

---

## Prerequisites

### nats-client Object Store Support

The `github.com/osapi-io/nats-client` package needs Object Store methods
before this plan can start. Add to the nats-client repo:

```go
// In pkg/client/types.go or new objectstore.go
func (c *Client) CreateOrUpdateObjectStore(
    ctx context.Context,
    cfg jetstream.ObjectStoreConfig,
) (jetstream.ObjectStore, error)

func (c *Client) ObjectStore(
    ctx context.Context,
    name string,
) (jetstream.ObjectStore, error)
```

Then update `internal/messaging/types.go` in osapi to add:

```go
CreateOrUpdateObjectStore(
    ctx context.Context,
    cfg jetstream.ObjectStoreConfig,
) (jetstream.ObjectStore, error)

ObjectStore(
    ctx context.Context,
    name string,
) (jetstream.ObjectStore, error)
```

This is a separate PR on the nats-client repo. Once merged, `go get`
the new version before starting Task 1.

---

## Task 1: NATS Configuration for Object Store + File-State KV

Add config structs, builder functions, and startup creation for the two
new NATS resources.

**Files:**
- Modify: `internal/config/types.go`
- Modify: `internal/cli/nats.go`
- Modify: `internal/cli/nats_public_test.go`
- Modify: `cmd/nats_helpers.go`
- Modify: `internal/messaging/types.go`
- Modify: `docs/docs/sidebar/usage/configuration.md`

### Step 1: Add config structs

In `internal/config/types.go`, add two new types and fields to `NATS`:

```go
// NATSObjects configuration for the NATS Object Store bucket.
type NATSObjects struct {
	// Bucket is the Object Store bucket name for file content.
	Bucket   string `mapstructure:"bucket"`
	MaxBytes int64  `mapstructure:"max_bytes"`
	Storage  string `mapstructure:"storage"` // "file" or "memory"
	Replicas int    `mapstructure:"replicas"`
}

// NATSFileState configuration for the file deployment state KV bucket.
// No TTL — deployed file state persists until explicitly removed.
type NATSFileState struct {
	// Bucket is the KV bucket name for file deployment SHA tracking.
	Bucket   string `mapstructure:"bucket"`
	Storage  string `mapstructure:"storage"` // "file" or "memory"
	Replicas int    `mapstructure:"replicas"`
}
```

Add to `NATS` struct:

```go
type NATS struct {
	// ... existing fields ...
	Objects   NATSObjects   `mapstructure:"objects,omitempty"`
	FileState NATSFileState `mapstructure:"file_state,omitempty"`
}
```

### Step 2: Add NATSClient Object Store methods

In `internal/messaging/types.go`, add to the `NATSClient` interface:

```go
// Object Store operations
CreateOrUpdateObjectStore(
    ctx context.Context,
    cfg jetstream.ObjectStoreConfig,
) (jetstream.ObjectStore, error)
ObjectStore(
    ctx context.Context,
    name string,
) (jetstream.ObjectStore, error)
```

### Step 3: Add builder functions

In `internal/cli/nats.go`, add:

```go
// BuildObjectStoreConfig builds a jetstream.ObjectStoreConfig from
// objects config values.
func BuildObjectStoreConfig(
	namespace string,
	objectsCfg config.NATSObjects,
) jetstream.ObjectStoreConfig {
	bucket := job.ApplyNamespaceToInfraName(namespace, objectsCfg.Bucket)

	return jetstream.ObjectStoreConfig{
		Bucket:   bucket,
		MaxBytes: objectsCfg.MaxBytes,
		Storage:  ParseJetstreamStorageType(objectsCfg.Storage),
		Replicas: objectsCfg.Replicas,
	}
}

// BuildFileStateKVConfig builds a jetstream.KeyValueConfig from
// file state config values. No TTL — deployed state persists.
func BuildFileStateKVConfig(
	namespace string,
	fileStateCfg config.NATSFileState,
) jetstream.KeyValueConfig {
	bucket := job.ApplyNamespaceToInfraName(namespace, fileStateCfg.Bucket)

	return jetstream.KeyValueConfig{
		Bucket:   bucket,
		Storage:  ParseJetstreamStorageType(fileStateCfg.Storage),
		Replicas: fileStateCfg.Replicas,
	}
}
```

### Step 4: Add startup creation

In `cmd/nats_helpers.go` `setupJetStream()`, add after the state KV
block:

```go
// Create Object Store bucket for file content
if appConfig.NATS.Objects.Bucket != "" {
    objStoreConfig := cli.BuildObjectStoreConfig(namespace, appConfig.NATS.Objects)
    if _, err := nc.CreateOrUpdateObjectStore(ctx, objStoreConfig); err != nil {
        return fmt.Errorf("create Object Store bucket %s: %w", objStoreConfig.Bucket, err)
    }
}

// Create file-state KV bucket for deployment SHA tracking
if appConfig.NATS.FileState.Bucket != "" {
    fileStateKVConfig := cli.BuildFileStateKVConfig(namespace, appConfig.NATS.FileState)
    if _, err := nc.CreateOrUpdateKVBucketWithConfig(ctx, fileStateKVConfig); err != nil {
        return fmt.Errorf("create file-state KV bucket %s: %w", fileStateKVConfig.Bucket, err)
    }
}
```

### Step 5: Add default config values

Add to `osapi.yaml` and the configuration docs the new sections:

```yaml
nats:
  objects:
    bucket: "file-objects"
    max_bytes: 524288000  # 500 MiB
    storage: "file"
    replicas: 1

  file_state:
    bucket: "file-state"
    storage: "file"
    replicas: 1
```

### Step 6: Run tests and verify

```bash
go build ./...
just go::unit
```

### Step 7: Commit

```bash
git add internal/config/types.go internal/cli/nats.go \
    internal/messaging/types.go cmd/nats_helpers.go
git commit -m "feat(config): add Object Store and file-state KV config"
```

---

## Task 2: Permissions — Add file:read and file:write

Add file permissions to the auth system before creating API endpoints.

**Files:**
- Modify: `internal/authtoken/permissions.go`
- Modify: `internal/authtoken/permissions_public_test.go`

### Step 1: Add permission constants

In `internal/authtoken/permissions.go`, add:

```go
const (
	// ... existing ...
	PermFileRead  Permission = "file:read"
	PermFileWrite Permission = "file:write"
)
```

Add to `AllPermissions`:

```go
var AllPermissions = []Permission{
	// ... existing ...
	PermFileRead,
	PermFileWrite,
}
```

Add to `DefaultRolePermissions`:

```go
"admin": {
    // ... existing ...
    PermFileRead,
    PermFileWrite,
},
"write": {
    // ... existing ...
    PermFileRead,
    PermFileWrite,
},
"read": {
    // ... existing ...
    PermFileRead,
},
```

### Step 2: Update permission tests

Add test cases to the existing permissions test suite to verify the new
permissions resolve correctly for admin, write, and read roles.

### Step 3: Run tests

```bash
go test ./internal/authtoken/... -count=1 -v
```

### Step 4: Commit

```bash
git add internal/authtoken/permissions.go \
    internal/authtoken/permissions_public_test.go
git commit -m "feat(auth): add file:read and file:write permissions"
```

---

## Task 3: File API Domain — OpenAPI Spec + Code Generation

Create the `/file` REST API domain for Object Store management.

**Files:**
- Create: `internal/api/file/gen/api.yaml`
- Create: `internal/api/file/gen/cfg.yaml`
- Create: `internal/api/file/gen/generate.go`
- Generated: `internal/api/file/gen/file.gen.go`

### Step 1: Write OpenAPI spec

Create `internal/api/file/gen/api.yaml`:

```yaml
openapi: "3.0.0"
info:
  title: File Management API
  version: 1.0.0

tags:
  - name: file
    x-displayName: File
    description: Manage files in the Object Store.

paths:
  /file:
    post:
      operationId: PostFile
      summary: Upload a file to Object Store
      description: >
        Stores file content in NATS Object Store. Returns the object
        reference with SHA256 and size.
      tags: [file]
      security:
        - BearerAuth:
            - "file:write"
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/FileUploadRequest"
      responses:
        "201":
          description: File uploaded successfully.
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/FileUploadResponse"
        "400":
          description: Invalid input.
          content:
            application/json:
              schema:
                $ref: "../../common/gen/api.yaml#/components/schemas/ErrorResponse"
        "401":
          description: Unauthorized.
          content:
            application/json:
              schema:
                $ref: "../../common/gen/api.yaml#/components/schemas/ErrorResponse"
        "403":
          description: Forbidden.
          content:
            application/json:
              schema:
                $ref: "../../common/gen/api.yaml#/components/schemas/ErrorResponse"
        "500":
          description: Internal server error.
          content:
            application/json:
              schema:
                $ref: "../../common/gen/api.yaml#/components/schemas/ErrorResponse"

    get:
      operationId: GetFiles
      summary: List stored files
      description: Returns metadata for all files in the Object Store.
      tags: [file]
      security:
        - BearerAuth:
            - "file:read"
      responses:
        "200":
          description: List of stored files.
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/FileListResponse"
        "401":
          description: Unauthorized.
          content:
            application/json:
              schema:
                $ref: "../../common/gen/api.yaml#/components/schemas/ErrorResponse"
        "403":
          description: Forbidden.
          content:
            application/json:
              schema:
                $ref: "../../common/gen/api.yaml#/components/schemas/ErrorResponse"
        "500":
          description: Internal server error.
          content:
            application/json:
              schema:
                $ref: "../../common/gen/api.yaml#/components/schemas/ErrorResponse"

  /file/{name}:
    get:
      operationId: GetFileByName
      summary: Get file metadata
      description: Returns metadata for a specific file in the Object Store.
      tags: [file]
      security:
        - BearerAuth:
            - "file:read"
      parameters:
        - $ref: "#/components/parameters/FileName"
      responses:
        "200":
          description: File metadata.
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/FileInfoResponse"
        "401":
          description: Unauthorized.
          content:
            application/json:
              schema:
                $ref: "../../common/gen/api.yaml#/components/schemas/ErrorResponse"
        "403":
          description: Forbidden.
          content:
            application/json:
              schema:
                $ref: "../../common/gen/api.yaml#/components/schemas/ErrorResponse"
        "404":
          description: File not found.
          content:
            application/json:
              schema:
                $ref: "../../common/gen/api.yaml#/components/schemas/ErrorResponse"
        "500":
          description: Internal server error.
          content:
            application/json:
              schema:
                $ref: "../../common/gen/api.yaml#/components/schemas/ErrorResponse"

    delete:
      operationId: DeleteFile
      summary: Delete a file from Object Store
      description: Removes a file from the Object Store.
      tags: [file]
      security:
        - BearerAuth:
            - "file:write"
      parameters:
        - $ref: "#/components/parameters/FileName"
      responses:
        "200":
          description: File deleted.
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/FileDeleteResponse"
        "401":
          description: Unauthorized.
          content:
            application/json:
              schema:
                $ref: "../../common/gen/api.yaml#/components/schemas/ErrorResponse"
        "403":
          description: Forbidden.
          content:
            application/json:
              schema:
                $ref: "../../common/gen/api.yaml#/components/schemas/ErrorResponse"
        "404":
          description: File not found.
          content:
            application/json:
              schema:
                $ref: "../../common/gen/api.yaml#/components/schemas/ErrorResponse"
        "500":
          description: Internal server error.
          content:
            application/json:
              schema:
                $ref: "../../common/gen/api.yaml#/components/schemas/ErrorResponse"

components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT

  parameters:
    FileName:
      name: name
      in: path
      required: true
      schema:
        type: string
      description: The name of the file in the Object Store.
      # NOTE: path param x-oapi-codegen-extra-tags does not generate
      # tags on RequestObject structs in strict-server mode.
      # Validated manually in handler.
      x-oapi-codegen-extra-tags:
        validate: required,min=1,max=255

  schemas:
    FileUploadRequest:
      type: object
      properties:
        name:
          type: string
          description: >
            Name to store the file under in the Object Store.
          x-oapi-codegen-extra-tags:
            validate: required,min=1,max=255
        content:
          type: string
          format: byte
          description: >
            Base64-encoded file content.
          x-oapi-codegen-extra-tags:
            validate: required
      required: [name, content]

    FileUploadResponse:
      type: object
      properties:
        name:
          type: string
        sha256:
          type: string
        size:
          type: integer
          format: int64
      required: [name, sha256, size]

    FileListResponse:
      type: object
      properties:
        files:
          type: array
          items:
            $ref: "#/components/schemas/FileInfo"
      required: [files]

    FileInfo:
      type: object
      properties:
        name:
          type: string
        sha256:
          type: string
        size:
          type: integer
          format: int64
      required: [name, size]

    FileInfoResponse:
      type: object
      properties:
        name:
          type: string
        sha256:
          type: string
        size:
          type: integer
          format: int64
      required: [name, sha256, size]

    FileDeleteResponse:
      type: object
      properties:
        name:
          type: string
        deleted:
          type: boolean
      required: [name, deleted]
```

### Step 2: Write codegen config

Create `internal/api/file/gen/cfg.yaml`:

```yaml
package: gen
generate:
  strict-server: true
  echo-server: true
  models: true
import-mapping:
  ../../common/gen/api.yaml: github.com/retr0h/osapi/internal/api/common/gen
output: file.gen.go
```

### Step 3: Write generate directive

Create `internal/api/file/gen/generate.go`:

```go
package gen

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config cfg.yaml api.yaml
```

### Step 4: Generate code

```bash
go generate ./internal/api/file/gen/...
```

### Step 5: Commit

```bash
git add internal/api/file/gen/
git commit -m "feat(api): add file domain OpenAPI spec and codegen"
```

---

## Task 4: File API Handler — Upload, List, Get, Delete

Implement the file API handler with all four endpoints.

**Files:**
- Create: `internal/api/file/types.go`
- Create: `internal/api/file/file.go`
- Create: `internal/api/file/file_upload.go`
- Create: `internal/api/file/file_list.go`
- Create: `internal/api/file/file_get.go`
- Create: `internal/api/file/file_delete.go`
- Create: `internal/api/file/file_upload_public_test.go`
- Create: `internal/api/file/file_list_public_test.go`
- Create: `internal/api/file/file_get_public_test.go`
- Create: `internal/api/file/file_delete_public_test.go`

### Step 1: Write types.go

```go
package file

import (
	"context"
	"log/slog"

	"github.com/nats-io/nats.go/jetstream"
)

// ObjectStoreManager abstracts NATS Object Store operations for testing.
type ObjectStoreManager interface {
	PutBytes(
		ctx context.Context,
		name string,
		data []byte,
	) (*jetstream.ObjectInfo, error)
	GetBytes(
		ctx context.Context,
		name string,
	) ([]byte, error)
	GetInfo(
		ctx context.Context,
		name string,
	) (*jetstream.ObjectInfo, error)
	Delete(
		ctx context.Context,
		name string,
	) error
	List(
		ctx context.Context,
	) ([]*jetstream.ObjectInfo, error)
}

// File handles file management REST API endpoints.
type File struct {
	objStore ObjectStoreManager
	logger   *slog.Logger
}
```

**Note:** The `ObjectStoreManager` interface wraps `jetstream.ObjectStore`
so handlers can be tested with mocks. The actual `jetstream.ObjectStore`
satisfies this interface. Verify that the `jetstream.ObjectStore`
interface matches — the `List` method may return a lister instead of a
slice; adapt accordingly.

### Step 2: Write file.go factory

```go
package file

import (
	"log/slog"

	gen "github.com/retr0h/osapi/internal/api/file/gen"
)

var _ gen.StrictServerInterface = (*File)(nil)

// New creates a new File handler.
func New(
	logger *slog.Logger,
	objStore ObjectStoreManager,
) *File {
	return &File{
		objStore: objStore,
		logger:   logger,
	}
}
```

### Step 3: Write upload handler (file_upload.go)

Decode base64 content from request body, store in Object Store, return
reference. Use `validation.Struct(request.Body)` for input validation.

### Step 4: Write failing tests for upload

Create `file_upload_public_test.go` with table-driven suite:
- when valid upload succeeds (201)
- when name is empty (400, validation error)
- when content is empty (400, validation error)
- when Object Store put fails (500)

Include `TestPostFileHTTP` and `TestPostFileRBACHTTP` methods.

### Step 5: Implement remaining handlers

Follow the same test-first pattern for list, get, delete:
- `file_list.go` — iterate Object Store, return file info array
- `file_get.go` — get info by name, return 404 if not found
- `file_delete.go` — delete by name, return 404 if not found

### Step 6: Run tests

```bash
go test ./internal/api/file/... -count=1 -v
```

### Step 7: Commit

```bash
git add internal/api/file/
git commit -m "feat(api): implement file upload, list, get, delete handlers"
```

---

## Task 5: File API Server Wiring

Wire the file handler into the API server.

**Files:**
- Create: `internal/api/handler_file.go`
- Create: `internal/api/handler_file_public_test.go`
- Modify: `internal/api/types.go`
- Modify: `internal/api/handler.go`
- Modify: `cmd/api_helpers.go`

### Step 1: Create handler_file.go

Follow the pattern from `handler_node.go`. All file endpoints require
authentication. The handler factory takes an `ObjectStoreManager`:

```go
func (s *Server) GetFileHandler(
	objStore file.ObjectStoreManager,
) []func(e *echo.Echo) {
	var tokenManager TokenValidator = authtoken.New(s.logger)

	fileHandler := file.New(s.logger, objStore)

	strictHandler := fileGen.NewStrictHandler(
		fileHandler,
		[]fileGen.StrictMiddlewareFunc{
			func(handler strictecho.StrictEchoHandlerFunc, _ string) strictecho.StrictEchoHandlerFunc {
				return scopeMiddleware(
					handler,
					tokenManager,
					s.appConfig.API.Server.Security.SigningKey,
					fileGen.BearerAuthScopes,
					s.customRoles,
				)
			},
		},
	)

	return []func(e *echo.Echo){
		func(e *echo.Echo) {
			fileGen.RegisterHandlers(e, strictHandler)
		},
	}
}
```

### Step 2: Update types.go

No new fields needed on `Server` — the Object Store is passed directly
to `GetFileHandler()`.

### Step 3: Update handler.go

In `registerAPIHandlers()` (or equivalent), add:

```go
handlers = append(handlers, sm.GetFileHandler(objStore)...)
```

### Step 4: Update startup wiring

In `cmd/api_helpers.go`, create the Object Store handle at startup and
pass it to the file handler:

```go
// Create Object Store handle for file management API
var objStore jetstream.ObjectStore
if appConfig.NATS.Objects.Bucket != "" {
    objStoreName := job.ApplyNamespaceToInfraName(namespace, appConfig.NATS.Objects.Bucket)
    objStore, err = nc.ObjectStore(ctx, objStoreName)
    // handle error
}
```

### Step 5: Add handler test

Create `handler_file_public_test.go` following the pattern of
`handler_node_public_test.go`.

### Step 6: Update combined OpenAPI spec

Add the file spec to `internal/api/gen/api.yaml` merged spec.

### Step 7: Run tests and verify

```bash
go build ./...
go test ./internal/api/... -count=1 -v
```

### Step 8: Commit

```bash
git add internal/api/handler_file.go internal/api/handler_file_public_test.go \
    internal/api/types.go internal/api/handler.go \
    cmd/api_helpers.go internal/api/gen/api.yaml
git commit -m "feat(api): wire file handler into API server"
```

---

## Task 6: Job Types + File Provider Interface

Define operation constants, request/response types, and the file
provider interface.

**Files:**
- Modify: `internal/job/types.go`
- Create: `internal/provider/file/types.go`
- Create: `internal/provider/file/mocks/types.gen.go`
- Create: `internal/provider/file/mocks/mocks.go`

### Step 1: Add operation constants

In `internal/job/types.go`:

```go
// File operations
const (
	OperationFileDeployExecute = "file.deploy.execute"
	OperationFileStatusGet     = "file.status.get"
)
```

### Step 2: Define file state type

In `internal/job/types.go`, add the file state KV entry structure:

```go
// FileState represents a deployed file's state in the file-state KV.
// Keyed by <hostname>.<sha256-of-path>.
type FileState struct {
	ObjectName  string `json:"object_name"`
	Path        string `json:"path"`
	SHA256      string `json:"sha256"`
	Mode        string `json:"mode,omitempty"`
	Owner       string `json:"owner,omitempty"`
	Group       string `json:"group,omitempty"`
	DeployedAt  string `json:"deployed_at"`
	ContentType string `json:"content_type"`
}
```

### Step 3: Define provider interface

Create `internal/provider/file/types.go`:

```go
package file

import "context"

// DeployRequest contains parameters for deploying a file to disk.
type DeployRequest struct {
	ObjectName  string         `json:"object_name"`
	Path        string         `json:"path"`
	Mode        string         `json:"mode,omitempty"`
	Owner       string         `json:"owner,omitempty"`
	Group       string         `json:"group,omitempty"`
	ContentType string         `json:"content_type"` // "raw" or "template"
	Vars        map[string]any `json:"vars,omitempty"`
}

// DeployResult contains the result of a file deploy operation.
type DeployResult struct {
	Changed bool   `json:"changed"`
	SHA256  string `json:"sha256"`
	Path    string `json:"path"`
}

// StatusRequest contains parameters for checking file status.
type StatusRequest struct {
	Path string `json:"path"`
}

// StatusResult contains the result of a file status check.
type StatusResult struct {
	Path   string `json:"path"`
	Status string `json:"status"` // "in-sync", "drifted", "missing"
	SHA256 string `json:"sha256,omitempty"`
}

// Provider defines the interface for file operations.
type Provider interface {
	Deploy(
		ctx context.Context,
		req DeployRequest,
	) (*DeployResult, error)
	Status(
		ctx context.Context,
		req StatusRequest,
	) (*StatusResult, error)
}
```

### Step 4: Generate mocks

Create `internal/provider/file/mocks/mocks.go`:

```go
package mocks

//go:generate mockgen -source=../types.go -destination=types.gen.go -package=mocks
```

Run:

```bash
go generate ./internal/provider/file/mocks/...
```

### Step 5: Commit

```bash
git add internal/job/types.go internal/provider/file/
git commit -m "feat(file): add job operation constants and provider interface"
```

---

## Task 7: File Provider Implementation — Deploy with SHA Idempotency

Implement the core deploy logic: pull from Object Store, SHA compare,
write file, set permissions, update state KV.

**Files:**
- Create: `internal/provider/file/provider.go`
- Create: `internal/provider/file/deploy.go`
- Create: `internal/provider/file/deploy_public_test.go`
- Create: `internal/provider/file/status.go`
- Create: `internal/provider/file/status_public_test.go`

### Step 1: Write provider constructor

Create `internal/provider/file/provider.go`:

```go
package file

import (
	"context"
	"log/slog"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/spf13/afero"

	"github.com/retr0h/osapi/internal/job"
)

// FileProvider implements file deploy and status operations.
type FileProvider struct {
	logger     *slog.Logger
	fs         afero.Fs
	objStore   jetstream.ObjectStore
	stateKV    jetstream.KeyValue
	hostname   string
	cachedFacts *job.FactsRegistration
}

// New creates a new FileProvider.
func New(
	logger *slog.Logger,
	fs afero.Fs,
	objStore jetstream.ObjectStore,
	stateKV jetstream.KeyValue,
	hostname string,
	cachedFacts *job.FactsRegistration,
) *FileProvider {
	return &FileProvider{
		logger:      logger,
		fs:          fs,
		objStore:    objStore,
		stateKV:     stateKV,
		hostname:    hostname,
		cachedFacts: cachedFacts,
	}
}
```

**Note:** The provider uses `afero.Fs` for filesystem abstraction
(testable without writing real files). The `objStore` and `stateKV` are
NATS JetStream interfaces — mock them in tests.

### Step 2: Write failing deploy tests

Create `deploy_public_test.go` with table-driven cases:

| Case | Setup | Expected |
|------|-------|----------|
| when deploy succeeds (new file) | Mock: objStore returns content, stateKV has no entry | changed: true, file written |
| when deploy succeeds (changed) | Mock: objStore returns content, stateKV has different SHA | changed: true, file written |
| when deploy skips (unchanged) | Mock: objStore returns content, stateKV has same SHA | changed: false, no write |
| when Object Store get fails | Mock: objStore returns error | error |
| when file write fails | Mock: fs write fails | error |
| when state KV put fails | Mock: stateKV put fails | error |
| when mode is set | Mock: success | file written with correct mode |

### Step 3: Implement deploy

Create `deploy.go`. Core logic:

1. Pull content from Object Store: `objStore.GetBytes(ctx, req.ObjectName)`
2. If `content_type == "template"`, render (delegate to Task 8)
3. Compute SHA256 of final content
4. Build state key: `hostname + "." + sha256(req.Path)`
5. Check `stateKV.Get(ctx, stateKey)` — if SHA matches, return
   `{changed: false}`
6. Write file using `afero.WriteFile(fs, req.Path, content, mode)`
7. If owner/group set, `fs.Chown` (skip if not root or on macOS)
8. Update stateKV with new `FileState`
9. Return `{changed: true, sha256: sha}`

```go
func (p *FileProvider) Deploy(
	ctx context.Context,
	req DeployRequest,
) (*DeployResult, error) {
	// 1. Pull content from Object Store
	content, err := p.objStore.GetBytes(ctx, req.ObjectName)
	if err != nil {
		return nil, fmt.Errorf("failed to get object %q: %w", req.ObjectName, err)
	}

	// 2. Template rendering (if applicable)
	if req.ContentType == "template" {
		content, err = p.renderTemplate(content, req.Vars)
		if err != nil {
			return nil, fmt.Errorf("failed to render template: %w", err)
		}
	}

	// 3. Compute SHA of final content
	sha := computeSHA256(content)

	// 4. Check state for idempotency
	stateKey := buildStateKey(p.hostname, req.Path)
	existing, _ := p.stateKV.Get(ctx, stateKey)
	if existing != nil {
		var state job.FileState
		if json.Unmarshal(existing.Value(), &state) == nil && state.SHA256 == sha {
			return &DeployResult{Changed: false, SHA256: sha, Path: req.Path}, nil
		}
	}

	// 5. Write file
	mode := parseFileMode(req.Mode)
	if err := afero.WriteFile(p.fs, req.Path, content, mode); err != nil {
		return nil, fmt.Errorf("failed to write file %q: %w", req.Path, err)
	}

	// 6. Update state KV
	state := job.FileState{
		ObjectName:  req.ObjectName,
		Path:        req.Path,
		SHA256:      sha,
		Mode:        req.Mode,
		Owner:       req.Owner,
		Group:       req.Group,
		DeployedAt:  time.Now().UTC().Format(time.RFC3339),
		ContentType: req.ContentType,
	}
	stateBytes, _ := json.Marshal(state)
	if _, err := p.stateKV.Put(ctx, stateKey, stateBytes); err != nil {
		return nil, fmt.Errorf("failed to update file state: %w", err)
	}

	return &DeployResult{Changed: true, SHA256: sha, Path: req.Path}, nil
}
```

### Step 4: Implement helper functions

```go
func computeSHA256(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func buildStateKey(hostname, path string) string {
	pathHash := computeSHA256([]byte(path))
	return hostname + "." + pathHash
}

func parseFileMode(mode string) os.FileMode {
	if mode == "" {
		return 0o644
	}
	m, err := strconv.ParseUint(mode, 8, 32)
	if err != nil {
		return 0o644
	}
	return os.FileMode(m)
}
```

### Step 5: Write failing status tests

Create `status_public_test.go`:

| Case | Setup | Expected |
|------|-------|----------|
| when file in sync | Local SHA matches state KV SHA | status: "in-sync" |
| when file drifted | Local SHA differs from state KV SHA | status: "drifted" |
| when file missing | File doesn't exist on disk | status: "missing" |
| when no state entry | stateKV has no entry for path | status: "missing" |

### Step 6: Implement status

```go
func (p *FileProvider) Status(
	ctx context.Context,
	req StatusRequest,
) (*StatusResult, error) {
	stateKey := buildStateKey(p.hostname, req.Path)

	entry, err := p.stateKV.Get(ctx, stateKey)
	if err != nil {
		return &StatusResult{Path: req.Path, Status: "missing"}, nil
	}

	var state job.FileState
	if err := json.Unmarshal(entry.Value(), &state); err != nil {
		return nil, fmt.Errorf("failed to parse file state: %w", err)
	}

	// Check if file exists on disk
	data, err := afero.ReadFile(p.fs, req.Path)
	if err != nil {
		return &StatusResult{Path: req.Path, Status: "missing"}, nil
	}

	localSHA := computeSHA256(data)
	if localSHA == state.SHA256 {
		return &StatusResult{Path: req.Path, Status: "in-sync", SHA256: localSHA}, nil
	}

	return &StatusResult{Path: req.Path, Status: "drifted", SHA256: localSHA}, nil
}
```

### Step 7: Run tests

```bash
go test ./internal/provider/file/... -count=1 -v
```

### Step 8: Commit

```bash
git add internal/provider/file/
git commit -m "feat(file): implement deploy with SHA idempotency and status check"
```

---

## Task 8: Template Rendering

Add Go `text/template` rendering support to the file provider.

**Files:**
- Create: `internal/provider/file/template.go`
- Create: `internal/provider/file/template_public_test.go`

### Step 1: Define template context

In `template.go`:

```go
// TemplateContext is the data available to Go templates during rendering.
type TemplateContext struct {
	Facts    *job.FactsRegistration
	Vars     map[string]any
	Hostname string
}
```

### Step 2: Write failing template tests

Create `template_public_test.go`:

| Case | Template | Vars/Facts | Expected |
|------|----------|------------|----------|
| when simple var substitution | `server {{ .Vars.host }}` | `{"host":"10.0.0.1"}` | `server 10.0.0.1` |
| when fact reference | `arch: {{ .Facts.Architecture }}` | Facts with Architecture="amd64" | `arch: amd64` |
| when conditional | `{{ if eq .Facts.Architecture "arm64" }}arm{{ else }}x86{{ end }}` | Architecture="amd64" | `x86` |
| when hostname | `# {{ .Hostname }}` | hostname="web-01" | `# web-01` |
| when invalid template syntax | `{{ .Invalid` | — | error |
| when nil facts | `{{ .Hostname }}` | nil facts | uses hostname only |

### Step 3: Implement renderTemplate

```go
func (p *FileProvider) renderTemplate(
	rawTemplate []byte,
	vars map[string]any,
) ([]byte, error) {
	tmpl, err := template.New("file").Parse(string(rawTemplate))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	ctx := TemplateContext{
		Facts:    p.cachedFacts,
		Vars:     vars,
		Hostname: p.hostname,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}
```

### Step 4: Run tests

```bash
go test ./internal/provider/file/... -count=1 -v
```

### Step 5: Commit

```bash
git add internal/provider/file/template.go \
    internal/provider/file/template_public_test.go
git commit -m "feat(file): add Go text/template rendering with facts and vars"
```

---

## Task 9: Agent Wiring + Processor Dispatch

Add Object Store, file-state KV, and file provider to the agent. Add
`file` category to the processor dispatcher.

**Files:**
- Modify: `internal/agent/types.go`
- Modify: `internal/agent/agent.go` (New constructor)
- Create: `internal/agent/processor_file.go`
- Create: `internal/agent/processor_file_test.go`
- Modify: `internal/agent/processor.go`
- Modify: `internal/agent/processor_test.go`
- Modify: `cmd/agent_helpers.go`
- Modify: `cmd/api_helpers.go`

### Step 1: Update Agent struct

In `internal/agent/types.go`, add:

```go
import (
	// ... existing ...
	fileProv "github.com/retr0h/osapi/internal/provider/file"
)

type Agent struct {
	// ... existing fields ...

	// File provider for file deploy/status operations
	fileProvider fileProv.Provider

	// Object Store handle (shared primitive for future providers)
	objStore jetstream.ObjectStore

	// File-state KV for SHA tracking
	fileStateKV jetstream.KeyValue
}
```

### Step 2: Update constructor

In `internal/agent/agent.go`, add parameters to `New()`:

```go
func New(
	// ... existing params ...
	fileProvider fileProv.Provider,
	objStore jetstream.ObjectStore,
	fileStateKV jetstream.KeyValue,
) *Agent {
```

### Step 3: Create processor_file.go

```go
func (a *Agent) processFileOperation(
	jobRequest job.Request,
) (json.RawMessage, error) {
	baseOperation := strings.Split(jobRequest.Operation, ".")[0]

	switch baseOperation {
	case "deploy":
		return a.processFileDeploy(jobRequest)
	case "status":
		return a.processFileStatus(jobRequest)
	default:
		return nil, fmt.Errorf("unsupported file operation: %s", jobRequest.Operation)
	}
}

func (a *Agent) processFileDeploy(
	jobRequest job.Request,
) (json.RawMessage, error) {
	var req fileProv.DeployRequest
	if err := json.Unmarshal(jobRequest.Data, &req); err != nil {
		return nil, fmt.Errorf("failed to parse file deploy data: %w", err)
	}

	result, err := a.fileProvider.Deploy(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("file deploy failed: %w", err)
	}

	return json.Marshal(result)
}

func (a *Agent) processFileStatus(
	jobRequest job.Request,
) (json.RawMessage, error) {
	var req fileProv.StatusRequest
	if err := json.Unmarshal(jobRequest.Data, &req); err != nil {
		return nil, fmt.Errorf("failed to parse file status data: %w", err)
	}

	result, err := a.fileProvider.Status(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("file status failed: %w", err)
	}

	return json.Marshal(result)
}
```

### Step 4: Update processor.go dispatch

Add to `processJobOperation()`:

```go
case "file":
    return a.processFileOperation(jobRequest)
```

### Step 5: Write processor tests

Add test cases to `processor_test.go` for the file category, and
create `processor_file_test.go` for the file sub-dispatch.

### Step 6: Update startup wiring

In `cmd/agent_helpers.go`:

```go
// Create Object Store handle
var objStore jetstream.ObjectStore
if appConfig.NATS.Objects.Bucket != "" {
    objStoreName := job.ApplyNamespaceToInfraName(namespace, appConfig.NATS.Objects.Bucket)
    objStore, _ = nc.ObjectStore(ctx, objStoreName)
}

// Create file-state KV
var fileStateKV jetstream.KeyValue
if appConfig.NATS.FileState.Bucket != "" {
    fileStateKVConfig := cli.BuildFileStateKVConfig(namespace, appConfig.NATS.FileState)
    fileStateKV, _ = nc.CreateOrUpdateKVBucketWithConfig(ctx, fileStateKVConfig)
}

// Create file provider (after agent hostname is resolved)
fileProvider := fileProv.New(log, appFs, objStore, fileStateKV, hostname, nil)

a := agent.New(
    // ... existing args ...
    fileProvider,
    objStore,
    fileStateKV,
)
```

**Note:** The file provider's `cachedFacts` is initially nil and gets
updated when facts are collected. Add a method or field update in the
facts collection loop to keep the file provider's facts current.

### Step 7: Update all existing tests that call agent.New()

Every test that constructs an `Agent` needs the new parameters. Pass
`nil` for file provider, objStore, and fileStateKV in tests that
don't exercise file operations.

### Step 8: Run tests

```bash
go build ./...
go test ./internal/agent/... -count=1 -v
```

### Step 9: Commit

```bash
git add internal/agent/ cmd/agent_helpers.go cmd/api_helpers.go
git commit -m "feat(agent): wire file provider and Object Store into agent"
```

---

## Task 10: Job Client Methods for File Deploy/Status

Add convenience methods to the job client for triggering file operations.

**Files:**
- Modify: `internal/job/client/types.go` (JobClient interface)
- Create: `internal/job/client/file.go`
- Create: `internal/job/client/file_public_test.go`
- Modify: `internal/job/mocks/job_client.gen.go` (regenerate)

### Step 1: Add interface methods

In `internal/job/client/types.go`, add to `JobClient`:

```go
// File operations
ModifyFileDeploy(
    ctx context.Context,
    hostname string,
    objectName string,
    path string,
    contentType string,
    mode string,
    owner string,
    group string,
    vars map[string]any,
) (string, string, bool, error)

QueryFileStatus(
    ctx context.Context,
    hostname string,
    path string,
) (string, *file.StatusResult, error)
```

### Step 2: Write failing tests

Test the job creation, subject routing, and response parsing.

### Step 3: Implement methods

Follow the pattern of `ModifyNetworkDNS` and `QueryNodeStatus`:

```go
func (c *Client) ModifyFileDeploy(
    ctx context.Context,
    hostname string,
    objectName string,
    path string,
    contentType string,
    mode string,
    owner string,
    group string,
    vars map[string]any,
) (string, string, bool, error) {
    data, _ := json.Marshal(file.DeployRequest{
        ObjectName:  objectName,
        Path:        path,
        Mode:        mode,
        Owner:       owner,
        Group:       group,
        ContentType: contentType,
        Vars:        vars,
    })

    req := &job.Request{
        Type:      job.TypeModify,
        Category:  "file",
        Operation: job.OperationFileDeployExecute,
        Data:      json.RawMessage(data),
    }

    subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, hostname)
    jobID, resp, err := c.publishAndWait(ctx, subject, req)
    if err != nil {
        return "", "", false, err
    }

    changed := resp.Changed != nil && *resp.Changed
    return jobID, resp.Hostname, changed, nil
}
```

### Step 4: Regenerate mocks

```bash
go generate ./internal/job/mocks/...
```

### Step 5: Run tests

```bash
go test ./internal/job/client/... -count=1 -v
```

### Step 6: Commit

```bash
git add internal/job/client/ internal/job/mocks/
git commit -m "feat(job): add file deploy and status job client methods"
```

---

## Task 11: Node API Endpoints for File Deploy/Status

Add REST endpoints for triggering file deploy and status through the
node domain.

**Files:**
- Modify: `internal/api/node/gen/api.yaml`
- Regenerate: `internal/api/node/gen/node.gen.go`
- Create: `internal/api/node/file_deploy_post.go`
- Create: `internal/api/node/file_deploy_post_public_test.go`
- Create: `internal/api/node/file_status_post.go`
- Create: `internal/api/node/file_status_post_public_test.go`

### Step 1: Add to node OpenAPI spec

Add paths and schemas to `internal/api/node/gen/api.yaml`:

```yaml
/node/{hostname}/file/deploy:
  post:
    operationId: PostNodeFileDeploy
    summary: Deploy a file from Object Store to the host
    security:
      - BearerAuth:
          - "file:write"
    parameters:
      - $ref: "#/components/parameters/Hostname"
    requestBody:
      required: true
      content:
        application/json:
          schema:
            $ref: "#/components/schemas/FileDeployRequest"
    responses:
      "202":
        description: File deploy job accepted.
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/FileDeployResponse"
      "400":
        description: Invalid input.
      "500":
        description: Internal error.

/node/{hostname}/file/status:
  post:
    operationId: PostNodeFileStatus
    summary: Check deployment status of a file on the host
    security:
      - BearerAuth:
          - "file:read"
    parameters:
      - $ref: "#/components/parameters/Hostname"
    requestBody:
      required: true
      content:
        application/json:
          schema:
            $ref: "#/components/schemas/FileStatusRequest"
    responses:
      "200":
        description: File status.
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/FileStatusResponse"
      "400":
        description: Invalid input.
      "500":
        description: Internal error.
```

Add schemas:

```yaml
FileDeployRequest:
  type: object
  properties:
    object_name:
      type: string
      x-oapi-codegen-extra-tags:
        validate: required,min=1,max=255
    path:
      type: string
      x-oapi-codegen-extra-tags:
        validate: required,min=1
    mode:
      type: string
    owner:
      type: string
    group:
      type: string
    content_type:
      type: string
      enum: [raw, template]
      x-oapi-codegen-extra-tags:
        validate: required,oneof=raw template
    vars:
      type: object
      additionalProperties: true
  required: [object_name, path, content_type]

FileStatusRequest:
  type: object
  properties:
    path:
      type: string
      x-oapi-codegen-extra-tags:
        validate: required,min=1
  required: [path]
```

### Step 2: Regenerate

```bash
go generate ./internal/api/node/gen/...
```

### Step 3: Implement handlers

Follow the pattern of `network_dns_put_by_interface.go`. Each handler:
1. Validates hostname
2. Validates request body
3. Calls the job client method
4. Returns the response

### Step 4: Write tests

Table-driven tests with HTTP wiring and RBAC tests for each endpoint.

### Step 5: Run tests

```bash
go test ./internal/api/node/... -count=1 -v
```

### Step 6: Commit

```bash
git add internal/api/node/
git commit -m "feat(api): add node file deploy and status endpoints"
```

---

## Task 12: CLI Commands

Add CLI commands for file management and file deployment.

**Files:**
- Create: `cmd/client_file.go` — parent command
- Create: `cmd/client_file_upload.go`
- Create: `cmd/client_file_list.go`
- Create: `cmd/client_file_get.go`
- Create: `cmd/client_file_delete.go`
- Create: `cmd/client_node_file.go` — parent under node
- Create: `cmd/client_node_file_deploy.go`
- Create: `cmd/client_node_file_status.go`

### Step 1: File management commands

`osapi client file upload`:

```
--name    Name for the file in Object Store (required)
--file    Path to local file to upload (required)
```

`osapi client file list` — no extra flags

`osapi client file get --name <name>` — show metadata

`osapi client file delete --name <name>` — remove from Object Store

### Step 2: Node file commands

`osapi client node file deploy`:

```
--object        Object name in Object Store (required)
--path          Destination path on host (required)
--content-type  "raw" or "template" (default: "raw")
--mode          File mode (e.g., "0644")
--owner         File owner
--group         File group
--var           Template var (key=value, repeatable)
-T, --target    Target host (default: _any)
-j, --json      Raw JSON output
```

`osapi client node file status`:

```
--path          File path to check (required)
-T, --target    Target host (default: _any)
-j, --json      Raw JSON output
```

### Step 3: Implement commands

Follow the pattern of `cmd/client_node_command_exec.go`. Read local
file, base64 encode, call SDK upload. For deploy, call SDK deploy.
Handle all response codes in switch block.

### Step 4: Test manually

```bash
go build ./... && ./osapi client file upload --help
./osapi client node file deploy --help
```

### Step 5: Commit

```bash
git add cmd/client_file*.go cmd/client_node_file*.go
git commit -m "feat(cli): add file upload/list/get/delete and deploy/status commands"
```

---

## Task 13: SDK Integration

Update the `osapi-sdk` to support the new file endpoints.

**Files (in osapi-sdk repo):**
- Copy: `pkg/osapi/gen/file/api.yaml` (from osapi)
- Create: `pkg/osapi/file.go` — FileService
- Modify: `.gilt.yml` — add file spec overlay
- Regenerate client code

### Step 1: Add file API spec to SDK

Copy `internal/api/file/gen/api.yaml` → `pkg/osapi/gen/file/api.yaml`.

### Step 2: Update gilt overlay

Add file domain to `.gilt.yml` so `just generate` pulls the spec.

### Step 3: Create FileService

```go
type FileService struct {
    client *Client
}

func (s *FileService) Upload(ctx context.Context, name string, content []byte) (*FileInfo, error)
func (s *FileService) List(ctx context.Context) ([]FileInfo, error)
func (s *FileService) Get(ctx context.Context, name string) (*FileInfo, error)
func (s *FileService) Delete(ctx context.Context, name string) error
```

Deploy/status use the existing job system through `NodeService` or
as separate methods.

### Step 4: Regenerate and test

```bash
just generate
go test ./...
```

### Step 5: Commit and push SDK

Separate PR on osapi-sdk repo.

---

## Task 14: Orchestrator Integration

Add file operations to `osapi-orchestrator`.

**Files (in osapi-orchestrator repo):**
- Create: `pkg/orchestrator/file.go`
- Create: example `examples/file-deploy/main.go`

### Step 1: Add orchestrator steps

```go
func (o *Orchestrator) FileUpload(name, localPath string) *Step
func (o *Orchestrator) FileDeploy(target, objectName, destPath string, opts ...FileOption) *Step
func (o *Orchestrator) FileTemplate(target, objectName, destPath string, vars map[string]any, opts ...FileOption) *Step
```

`FileOption` funcs:

```go
func WithMode(mode string) FileOption
func WithOwner(owner, group string) FileOption
```

### Step 2: OnlyIfChanged integration

`FileDeploy` and `FileTemplate` return `changed: true/false` in the
result, so `OnlyIfChanged()` guards work naturally:

```go
upload := o.FileUpload("nginx.conf", "./local/nginx.conf.tmpl")
deploy := o.FileTemplate("_all", "nginx.conf", "/etc/nginx/nginx.conf",
    map[string]any{"worker_count": 4},
    WithMode("0644"),
    WithOwner("root", "root"),
).After(upload)

reload := o.CommandExec("_all", "nginx", []string{"-s", "reload"}).
    After(deploy).
    OnlyIfChanged()
```

### Step 3: Commit

Separate PR on osapi-orchestrator repo.

---

## Task 15: Documentation

Update docs for the new feature.

**Files:**
- Create: `docs/docs/sidebar/features/file-management.md`
- Create: `docs/docs/sidebar/usage/cli/client/file/file.md`
- Create: `docs/docs/sidebar/usage/cli/client/file/upload.md`
- Create: `docs/docs/sidebar/usage/cli/client/file/list.md`
- Create: `docs/docs/sidebar/usage/cli/client/file/get.md`
- Create: `docs/docs/sidebar/usage/cli/client/file/delete.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/file/file.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/file/deploy.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/file/status.md`
- Modify: `docs/docusaurus.config.ts` — add to Features dropdown
- Modify: `docs/docs/sidebar/usage/configuration.md` — add new config
- Modify: `docs/docs/sidebar/architecture/system-architecture.md` —
  add endpoints

### Step 1: Feature page

Create `file-management.md` covering:
- What it manages (file deployment with SHA idempotency)
- How it works (Object Store + file-state KV)
- Template rendering with facts
- Permissions (`file:read`, `file:write`)
- Links to CLI and API docs

### Step 2: CLI docs

One page per command with usage examples, flags table, and `--json`
output.

### Step 3: Config docs

Add `nats.objects` and `nats.file_state` sections with env vars:

| Config Key | Env Var |
|---|---|
| `nats.objects.bucket` | `OSAPI_NATS_OBJECTS_BUCKET` |
| `nats.objects.max_bytes` | `OSAPI_NATS_OBJECTS_MAX_BYTES` |
| `nats.objects.storage` | `OSAPI_NATS_OBJECTS_STORAGE` |
| `nats.objects.replicas` | `OSAPI_NATS_OBJECTS_REPLICAS` |
| `nats.file_state.bucket` | `OSAPI_NATS_FILE_STATE_BUCKET` |
| `nats.file_state.storage` | `OSAPI_NATS_FILE_STATE_STORAGE` |
| `nats.file_state.replicas` | `OSAPI_NATS_FILE_STATE_REPLICAS` |

### Step 4: Commit

```bash
git add docs/
git commit -m "docs: add file management feature documentation"
```

---

## Shared Primitive: Object Store for Future Providers

The Object Store and file-state KV infrastructure built in this plan
is designed as a **shared primitive**. The agent's `objStore` handle
is injected at startup and available to any provider. Future providers
that would consume this infrastructure:

| Provider | Operation | Usage |
|---|---|---|
| `firmware.update` | Pull binary, run flash tool | Object Store for firmware blobs |
| `package.install` | Pull `.deb`/`.rpm`, install | Object Store for packages |
| `cert.deploy` | Pull TLS cert/key | Object Store + restricted perms |
| `script.run` | Pull script, execute | Object Store for scripts |

Each provider reuses: Object Store download, SHA comparison, and state
tracking from the `file-state` KV bucket. No new infrastructure needed.

---

## Verification

After all tasks complete:

```bash
# Full test suite
just test

# Manual verification
osapi client file upload --name nginx.conf --file ./nginx.conf
osapi client file list
osapi client file get --name nginx.conf
osapi client node file deploy \
  --object nginx.conf --path /etc/nginx/nginx.conf \
  --mode 0644 --owner root --group root --target _all
osapi client node file status --path /etc/nginx/nginx.conf --target _all

# Idempotency check (second run should show changed: false)
osapi client node file deploy \
  --object nginx.conf --path /etc/nginx/nginx.conf \
  --mode 0644 --target _all
```
