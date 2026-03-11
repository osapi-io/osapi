# Container Runtime Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development
> (if subagents available) or superpowers:executing-plans to implement this plan.
> Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add container lifecycle management (Docker), a `provider run`
subcommand for running providers inside containers, and an orchestrator DSL
layer for composing host and container operations.

**Architecture:** A `runtime.Driver` interface abstracts container runtimes with
Docker as the first implementation via the Go SDK. Container lifecycle is exposed
as a new API domain under `/node/{hostname}/container`. A hidden `provider run`
CLI subcommand executes individual provider operations with JSON I/O, enabling
the orchestrator DSL's `In(target)` to transparently run providers inside
containers via `docker exec`.

**Tech Stack:** Go, Docker Go SDK (`github.com/docker/docker/client`),
oapi-codegen, Echo, testify/suite, NATS JetStream

**Spec:** `docs/superpowers/specs/2026-03-11-container-runtime-design.md`

---

## Chunk 1: Runtime Driver Interface and Docker Implementation

### Task 1: Runtime Driver Interface Types

**Files:**
- Create: `internal/provider/container/runtime/driver.go`

- [ ] **Step 1: Write the driver interface and types**

Create the `runtime` package with the `Driver` interface and all supporting
types. Reference `internal/provider/command/types.go` for the param/result
struct pattern.

```go
// Package runtime defines the container runtime driver interface.
package runtime

import (
	"context"
	"time"
)

// Driver defines container runtime operations.
// Implementations: Docker (now), LXD/Podman (later).
type Driver interface {
	Ping(ctx context.Context) error
	Create(ctx context.Context, params CreateParams) (*Container, error)
	Start(ctx context.Context, id string) error
	Stop(ctx context.Context, id string, timeout *time.Duration) error
	Remove(ctx context.Context, id string, force bool) error
	List(ctx context.Context, params ListParams) ([]Container, error)
	Inspect(ctx context.Context, id string) (*ContainerDetail, error)
	Exec(ctx context.Context, id string, params ExecParams) (*ExecResult, error)
	Pull(ctx context.Context, image string) (*PullResult, error)
}

// CreateParams contains parameters for container creation.
type CreateParams struct {
	// Image is the container image (required).
	Image string `json:"image"`
	// Name is an optional container name.
	Name string `json:"name,omitempty"`
	// Command overrides the image's default command.
	Command []string `json:"command,omitempty"`
	// Env sets environment variables.
	Env map[string]string `json:"env,omitempty"`
	// Ports maps host ports to container ports.
	Ports []PortMapping `json:"ports,omitempty"`
	// Volumes maps host paths to container paths.
	Volumes []VolumeMapping `json:"volumes,omitempty"`
	// AutoStart starts the container after creation.
	AutoStart bool `json:"auto_start,omitempty"`
}

// PortMapping maps a host port to a container port.
type PortMapping struct {
	Host      int `json:"host"`
	Container int `json:"container"`
}

// VolumeMapping maps a host path to a container path.
type VolumeMapping struct {
	Host      string `json:"host"`
	Container string `json:"container"`
}

// ListParams contains parameters for listing containers.
type ListParams struct {
	// State filters by container state: "running", "stopped", "all".
	State string `json:"state,omitempty"`
	// Limit caps the number of results.
	Limit int `json:"limit,omitempty"`
}

// Container holds summary info for a container.
type Container struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Image   string    `json:"image"`
	State   string    `json:"state"`
	Created time.Time `json:"created"`
}

// ContainerDetail holds detailed info for a container.
type ContainerDetail struct {
	Container
	NetworkSettings *NetworkSettings  `json:"network_settings,omitempty"`
	Ports           []PortMapping     `json:"ports,omitempty"`
	Mounts          []VolumeMapping   `json:"mounts,omitempty"`
	Health          string            `json:"health,omitempty"`
}

// NetworkSettings holds container network configuration.
type NetworkSettings struct {
	IPAddress string `json:"ip_address,omitempty"`
	Gateway   string `json:"gateway,omitempty"`
}

// ExecParams contains parameters for executing a command in a container.
type ExecParams struct {
	// Command is the command and arguments.
	Command []string `json:"command"`
	// Env sets environment variables.
	Env map[string]string `json:"env,omitempty"`
	// WorkingDir sets the working directory.
	WorkingDir string `json:"working_dir,omitempty"`
}

// ExecResult contains the output of a command execution in a container.
type ExecResult struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
}

// PullResult contains the result of an image pull.
type PullResult struct {
	ImageID string `json:"image_id"`
	Tag     string `json:"tag"`
	Size    int64  `json:"size"`
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/provider/container/runtime/...`
Expected: compiles with no errors

- [ ] **Step 3: Commit**

```bash
git add internal/provider/container/runtime/driver.go
git commit -m "feat(container): add runtime driver interface and types"
```

---

### Task 2: Docker Driver Implementation

**Files:**
- Create: `internal/provider/container/runtime/docker/docker.go`
- Create: `internal/provider/container/runtime/docker/docker_public_test.go`

- [ ] **Step 1: Add Docker SDK dependency**

Run: `go get github.com/docker/docker@latest`

- [ ] **Step 2: Write failing tests for the Docker driver**

Create the test suite with table-driven tests. Since the Docker driver talks to
a real Docker socket, tests should use an interface mock or be structured so they
can run against a real Docker daemon in integration. For unit tests, mock the
Docker client interface.

```go
package docker_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/provider/container/runtime"
	"github.com/retr0h/osapi/internal/provider/container/runtime/docker"
)

type DockerDriverPublicTestSuite struct {
	suite.Suite
	ctx    context.Context
	driver runtime.Driver
}

func (s *DockerDriverPublicTestSuite) SetupTest() {
	s.ctx = context.Background()
	d, err := docker.New()
	s.Require().NoError(err)
	s.driver = d
}

func (s *DockerDriverPublicTestSuite) TestNew() {
	tests := []struct {
		name         string
		validateFunc func(d runtime.Driver)
	}{
		{
			name: "returns non-nil driver",
			validateFunc: func(d runtime.Driver) {
				s.NotNil(d)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			d, err := docker.New()
			s.Require().NoError(err)
			tt.validateFunc(d)
		})
	}
}

func TestDockerDriverPublicTestSuite(t *testing.T) {
	suite.Run(t, new(DockerDriverPublicTestSuite))
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `go test -run TestDockerDriverPublicTestSuite -v ./internal/provider/container/runtime/docker/...`
Expected: FAIL — `docker` package does not exist

- [ ] **Step 4: Write the Docker driver implementation**

Implement the `Driver` interface using `github.com/docker/docker/client`. Each
method maps to the corresponding Docker Engine API call. The constructor accepts
no arguments and creates a client from environment defaults (DOCKER_HOST or
default socket).

```go
// Package docker implements the runtime.Driver interface using the Docker Engine API.
package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"

	"github.com/retr0h/osapi/internal/provider/container/runtime"
)

// Driver implements runtime.Driver using the Docker Engine API.
type Driver struct {
	client dockerclient.APIClient
}

// New creates a new Docker driver using default client options.
func New() (*Driver, error) {
	cli, err := dockerclient.NewClientWithOpts(
		dockerclient.FromEnv,
		dockerclient.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("create docker client: %w", err)
	}

	return &Driver{client: cli}, nil
}

// NewWithClient creates a Docker driver with an injected client (for testing).
func NewWithClient(
	client dockerclient.APIClient,
) *Driver {
	return &Driver{client: client}
}
```

Implement each method: `Create`, `Start`, `Stop`, `Remove`, `List`, `Inspect`,
`Exec`, `Pull`. Each translates runtime types to Docker API types and back.
Follow the multi-line function signature convention from CLAUDE.md.

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test -run TestDockerDriverPublicTestSuite -v ./internal/provider/container/runtime/docker/...`
Expected: PASS

- [ ] **Step 6: Verify compilation**

Run: `go build ./internal/provider/container/runtime/...`
Expected: compiles with no errors

- [ ] **Step 7: Commit**

```bash
git add internal/provider/container/runtime/docker/
git commit -m "feat(container): add Docker runtime driver implementation"
```

---

### Task 3: Container Provider Service

**Files:**
- Create: `internal/provider/container/provider.go`
- Create: `internal/provider/container/types.go`
- Create: `internal/provider/container/provider_public_test.go`

- [ ] **Step 1: Write the provider types**

The provider wraps a `runtime.Driver` and presents a domain interface that the
agent and API handler can call. Reference `internal/api/node/types.go` for the
struct pattern.

```go
// Package container provides the container management provider.
package container

import (
	"github.com/retr0h/osapi/internal/provider/container/runtime"
)

// Provider defines the container management interface.
// All methods accept context.Context for cancellation and timeout propagation,
// which is important since the Docker daemon is a remote service.
type Provider interface {
	Create(ctx context.Context, params runtime.CreateParams) (*runtime.Container, error)
	Start(ctx context.Context, id string) error
	Stop(ctx context.Context, id string, timeout *time.Duration) error
	Remove(ctx context.Context, id string, force bool) error
	List(ctx context.Context, params runtime.ListParams) ([]runtime.Container, error)
	Inspect(ctx context.Context, id string) (*runtime.ContainerDetail, error)
	Exec(ctx context.Context, id string, params runtime.ExecParams) (*runtime.ExecResult, error)
	Pull(ctx context.Context, image string) (*runtime.PullResult, error)
}
```

- [ ] **Step 2: Write failing tests**

```go
package container_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/provider/container"
	"github.com/retr0h/osapi/internal/provider/container/runtime"
)

type ProviderPublicTestSuite struct {
	suite.Suite
}

func (s *ProviderPublicTestSuite) TestNew() {
	tests := []struct {
		name         string
		validateFunc func(p container.Provider)
	}{
		{
			name: "returns non-nil provider",
			validateFunc: func(p container.Provider) {
				s.NotNil(p)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var driver runtime.Driver // nil driver for unit test
			p := container.New(driver)
			tt.validateFunc(p)
		})
	}
}

func TestProviderPublicTestSuite(t *testing.T) {
	suite.Run(t, new(ProviderPublicTestSuite))
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `go test -run TestProviderPublicTestSuite -v ./internal/provider/container/...`
Expected: FAIL — `container.New` not defined

- [ ] **Step 4: Write the provider implementation**

```go
package container

import (
	"context"
	"time"

	"github.com/retr0h/osapi/internal/provider/container/runtime"
)

// Service implements Provider by delegating to a runtime.Driver.
type Service struct {
	driver runtime.Driver
}

// New creates a new container provider service.
func New(
	driver runtime.Driver,
) *Service {
	return &Service{driver: driver}
}
```

Implement each method, delegating to `s.driver` with a background context.
Each method is a thin pass-through that creates a context and calls the driver.

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test -run TestProviderPublicTestSuite -v ./internal/provider/container/...`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/provider/container/
git commit -m "feat(container): add container provider service"
```

---

## Chunk 2: Job System Integration

### Task 4: Job Types and Operation Constants

**Files:**
- Modify: `internal/job/types.go`
- Modify: `internal/job/subjects.go`

- [ ] **Step 1: Add container data types to job/types.go**

Add data structs for container operations, following the `CommandExecData`
pattern at `internal/job/types.go:229-239`.

```go
// ContainerCreateData represents data for container creation.
type ContainerCreateData struct {
	Image     string            `json:"image"`
	Name      string            `json:"name,omitempty"`
	Command   []string          `json:"command,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
	Ports     []PortMapping     `json:"ports,omitempty"`
	Volumes   []VolumeMapping   `json:"volumes,omitempty"`
	AutoStart bool              `json:"auto_start,omitempty"`
}

// PortMapping maps a host port to a container port (job layer).
// Intentionally duplicated from runtime.PortMapping to keep the job
// layer decoupled from the provider layer. Both have the same shape.
type PortMapping struct {
	Host      int `json:"host"`
	Container int `json:"container"`
}

// VolumeMapping maps a host path to a container path (job layer).
// Intentionally duplicated from runtime.VolumeMapping for the same reason.
type VolumeMapping struct {
	Host      string `json:"host"`
	Container string `json:"container"`
}

// ContainerStopData represents data for stopping a container.
type ContainerStopData struct {
	Timeout *int `json:"timeout,omitempty"`
}

// ContainerRemoveData represents data for removing a container.
type ContainerRemoveData struct {
	Force bool `json:"force,omitempty"`
}

// ContainerListData represents data for listing containers.
type ContainerListData struct {
	State string `json:"state,omitempty"`
	Limit int    `json:"limit,omitempty"`
}

// ContainerExecData represents data for executing a command in a container.
type ContainerExecData struct {
	Command    []string          `json:"command"`
	Env        map[string]string `json:"env,omitempty"`
	WorkingDir string            `json:"working_dir,omitempty"`
}

// ContainerPullData represents data for pulling an image.
type ContainerPullData struct {
	Image string `json:"image"`
}
```

- [ ] **Step 2: Add container operation constants**

Find the existing operation constants (e.g., `OperationCommandExecExecute`) and
add container equivalents:

```go
// Container operation types
const (
	OperationContainerCreate  = "container.create"
	OperationContainerStart   = "container.start"
	OperationContainerStop    = "container.stop"
	OperationContainerRemove  = "container.remove"
	OperationContainerList    = "container.list"
	OperationContainerInspect = "container.inspect"
	OperationContainerExec    = "container.exec"
	OperationContainerPull    = "container.pull"
)
```

- [ ] **Step 3: Add `SubjectCategoryContainer` constant to subjects.go**

```go
const (
	SubjectCategoryContainer = "container"
)
```

- [ ] **Step 4: Verify compilation**

Run: `go build ./internal/job/...`
Expected: compiles

- [ ] **Step 5: Commit**

```bash
git add internal/job/types.go internal/job/subjects.go
git commit -m "feat(container): add container job types and operation constants"
```

---

### Task 5: Job Client Container Methods

**Files:**
- Create: `internal/job/client/modify_container.go`
- Create: `internal/job/client/modify_container_public_test.go`
- Modify: `internal/job/client/types.go` (add methods to `JobClient` interface)

- [ ] **Step 1: Add container methods to the JobClient interface**

Open `internal/job/client/types.go` and add methods for each container
operation. Follow the existing pattern from `ModifyCommandExec`.

- [ ] **Step 2: Write failing tests**

Create `internal/job/client/modify_container_public_test.go` following the
same table-driven suite pattern as
`internal/job/client/modify_command_public_test.go`. Test the `Create` method
first as the representative case — success, job failure, and publish error.

- [ ] **Step 3: Run test to verify it fails**

Run: `go test -run TestModifyContainerPublicTestSuite -v ./internal/job/client/...`
Expected: FAIL — methods not implemented

- [ ] **Step 4: Implement the job client container methods**

Create `internal/job/client/modify_container.go`. Each method marshals the
appropriate data struct, builds a `job.Request` with category `"container"`
and the correct operation constant, then calls `publishAndWait`. Follow
the pattern from `internal/job/client/modify_command.go:33-62`.

Implement: `ModifyContainerCreate`, `ModifyContainerStart`,
`ModifyContainerStop`, `ModifyContainerRemove`, `QueryContainerList`,
`QueryContainerInspect`, `ModifyContainerExec`, `ModifyContainerPull`.

Note: `List` and `Inspect` are query operations (`job.TypeQuery`), while
all others are modify operations (`job.TypeModify`).

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test -run TestModifyContainerPublicTestSuite -v ./internal/job/client/...`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/job/client/modify_container.go \
       internal/job/client/modify_container_public_test.go \
       internal/job/client/types.go
git commit -m "feat(container): add container job client methods"
```

---

### Task 6: Agent Processor Container Dispatch

**Files:**
- Modify: `internal/agent/types.go`
- Modify: `internal/agent/factory.go`
- Modify: `internal/agent/processor.go`
- Create: `internal/agent/processor_container.go`
- Create: `internal/agent/processor_container_test.go`

- [ ] **Step 1: Add `containerProvider` field to Agent struct**

In `internal/agent/types.go`, add:

```go
// Container provider
containerProvider containerProv.Provider
```

Add the import for the container provider package.

- [ ] **Step 2: Update the factory to create container provider**

In `internal/agent/factory.go`, add Docker driver creation. The factory should
check Docker socket availability and return `nil` if unavailable:

```go
// Create container provider (conditional on Docker availability)
var containerProvider containerProv.Provider
dockerDriver, err := docker.New()
if err == nil {
    if pingErr := dockerDriver.Ping(context.Background()); pingErr == nil {
        containerProvider = containerProv.New(dockerDriver)
    } else {
        f.logger.Info("Docker not available, container operations disabled",
            slog.String("error", pingErr.Error()))
    }
} else {
    f.logger.Info("Docker client creation failed, container operations disabled",
        slog.String("error", err.Error()))
}
```

Update `CreateProviders` return signature to include the container provider
as the 9th return value.

**IMPORTANT: This changes the return signature.** The following callers must
also be updated in this step:

- `internal/agent/agent.go` — the `New()` constructor must accept a
  `containerProv.Provider` parameter and assign it to the struct field
- `cmd/agent_helpers.go` or wherever `CreateProviders()` is called — update
  the call site to capture the 9th return value and pass it to `agent.New()`
- All existing test files that call `agent.New()` — pass `nil` as the
  container provider parameter. Affected files include:
  - `internal/agent/processor_command_test.go`
  - `internal/agent/processor_file_test.go`
  - `internal/agent/processor_test.go`
  - Any other files in `internal/agent/` that construct an `Agent`

Search for all usages with: `grep -r "agent.New(" internal/ cmd/`

- [ ] **Step 3: Add container case to processor dispatch**

In `internal/agent/processor.go`, add to the category switch in
`processJobOperation`:

```go
case "container":
    return a.processContainerOperation(jobRequest)
```

- [ ] **Step 4: Write failing test for container processor**

Create `internal/agent/processor_container_test.go` with `package agent`
(internal test — matches existing `processor_command_test.go` pattern since
`processContainerOperation` is unexported). Use a table-driven test for
`processContainerOperation`. Test the nil-provider case (returns
"container runtime not available" error) and a successful dispatch case
with a mock provider.

- [ ] **Step 5: Run test to verify it fails**

Run: `go test -run TestProcessContainerOperation -v ./internal/agent/...`
Expected: FAIL — `processContainerOperation` not defined

- [ ] **Step 6: Implement container processor**

Create `internal/agent/processor_container.go`:

Note: The existing processor dispatch (`processJobOperation`) does not pass
context to sub-processors. Since the container provider is the first to
require `context.Context`, the processor chain needs context propagation.
Either thread `context.Context` through from `processJobOperation` or use
`context.Background()` as a starting point (matching the existing pattern
where processors don't receive context). The preferred approach is to add
context to `processContainerOperation` and update the dispatch call:

In `processor.go`, change the container case to:
```go
case "container":
    return a.processContainerOperation(ctx, jobRequest)
```

Where `ctx` is derived from the handler's context (check how
`handleJobMessage` creates/receives its context).

```go
package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/provider/container/runtime"
)

func (a *Agent) processContainerOperation(
	ctx context.Context,
	jobRequest job.Request,
) (json.RawMessage, error) {
	if a.containerProvider == nil {
		return nil, fmt.Errorf("container runtime not available")
	}

	switch jobRequest.Operation {
	case job.OperationContainerCreate:
		var data job.ContainerCreateData
		if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
			return nil, fmt.Errorf("unmarshal create data: %w", err)
		}
		result, err := a.containerProvider.Create(ctx, runtime.CreateParams{
			Image:     data.Image,
			Name:      data.Name,
			Command:   data.Command,
			Env:       data.Env,
			AutoStart: data.AutoStart,
			// Map ports and volumes from job types to runtime types
		})
		if err != nil {
			return nil, err
		}
		return json.Marshal(result)

	case job.OperationContainerStart:
		// jobRequest.Data contains the container ID as a string
		var id string
		if err := json.Unmarshal(jobRequest.Data, &id); err != nil {
			return nil, fmt.Errorf("unmarshal start data: %w", err)
		}
		return nil, a.containerProvider.Start(ctx, id)

	// ... remaining operations follow the same pattern, always passing ctx

	default:
		return nil, fmt.Errorf("unsupported container operation: %s", jobRequest.Operation)
	}
}
```

Implement all eight operation cases.

- [ ] **Step 7: Run tests to verify they pass**

Run: `go test -run TestProcessContainerOperation -v ./internal/agent/...`
Expected: PASS

- [ ] **Step 8: Run full test suite**

Run: `just go::unit`
Expected: all tests pass (no regressions from agent changes)

- [ ] **Step 9: Commit**

```bash
git add internal/agent/types.go internal/agent/factory.go \
       internal/agent/processor.go internal/agent/processor_container.go \
       internal/agent/processor_container_test.go
git commit -m "feat(container): add agent container processor dispatch"
```

---

## Chunk 3: OpenAPI Spec and API Handlers

### Task 7: OpenAPI Specification

**Files:**
- Create: `internal/api/container/gen/api.yaml`
- Create: `internal/api/container/gen/cfg.yaml`
- Create: `internal/api/container/gen/generate.go`

- [ ] **Step 1: Write the OpenAPI spec**

Create `internal/api/container/gen/api.yaml`. Model it after
`internal/api/node/gen/api.yaml`. Key differences:

- All paths under `/node/{hostname}/container`
- `{id}` parameter uses `type: string` with `pattern: ^[a-zA-Z0-9][a-zA-Z0-9_.-]*$`
  (not `format: uuid`)
- Security scopes: `container:read`, `container:write`, `container:execute`
- Request body validation via `x-oapi-codegen-extra-tags`
- Error responses: 400, 401, 403, 404, 409, 500 per the spec

Define these paths:
- `POST /node/{hostname}/container` — CreateContainer
- `GET /node/{hostname}/container` — ListContainers
- `GET /node/{hostname}/container/{id}` — InspectContainer
- `POST /node/{hostname}/container/{id}/start` — StartContainer
- `POST /node/{hostname}/container/{id}/stop` — StopContainer
- `DELETE /node/{hostname}/container/{id}` — RemoveContainer (force as query param)
- `POST /node/{hostname}/container/{id}/exec` — ExecContainer
- `POST /node/{hostname}/container/pull` — PullImage

Define schemas: `ContainerCreateRequest`, `ContainerExecRequest`,
`ContainerStopRequest`, `ContainerPullRequest`,
`ContainerResponse`, `ContainerDetailResponse`,
`ContainerListResponse`, `ContainerExecResponse`,
`ContainerResultCollectionResponse`.

Use `x-oapi-codegen-extra-tags` for validation:
```yaml
properties:
  image:
    type: string
    x-oapi-codegen-extra-tags:
      validate: required,min=1
  command:
    type: array
    items:
      type: string
    x-oapi-codegen-extra-tags:
      validate: required,min=1
```

- [ ] **Step 2: Write the codegen config**

Create `internal/api/container/gen/cfg.yaml`:

```yaml
---
package: gen
output: container.gen.go
generate:
  models: true
  echo-server: true
  strict-server: true
import-mapping:
  ../../common/gen/api.yaml: github.com/retr0h/osapi/internal/api/common/gen
output-options:
  skip-prune: true
```

- [ ] **Step 3: Write the generate directive**

Create `internal/api/container/gen/generate.go`:

```go
// Package gen contains generated code for the container API.
package gen

//go:generate go tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config cfg.yaml api.yaml
```

- [ ] **Step 4: Run code generation**

Run: `go generate ./internal/api/container/gen/...`
Expected: `container.gen.go` is generated with no errors

- [ ] **Step 5: Verify compilation**

Run: `go build ./internal/api/container/...`
Expected: compiles

- [ ] **Step 6: Commit**

```bash
git add internal/api/container/gen/
git commit -m "feat(container): add OpenAPI spec and code generation"
```

---

### Task 8: Container API Handler

**Files:**
- Create: `internal/api/container/types.go`
- Create: `internal/api/container/container.go`
- Create: `internal/api/container/container_create.go`
- Create: `internal/api/container/container_create_public_test.go`
- Create: `internal/api/container/container_list.go`
- Create: `internal/api/container/container_list_public_test.go`
- Create: `internal/api/container/container_inspect.go`
- Create: `internal/api/container/container_inspect_public_test.go`
- Create: `internal/api/container/container_start.go`
- Create: `internal/api/container/container_start_public_test.go`
- Create: `internal/api/container/container_stop.go`
- Create: `internal/api/container/container_stop_public_test.go`
- Create: `internal/api/container/container_remove.go`
- Create: `internal/api/container/container_remove_public_test.go`
- Create: `internal/api/container/container_exec.go`
- Create: `internal/api/container/container_exec_public_test.go`
- Create: `internal/api/container/container_pull.go`
- Create: `internal/api/container/container_pull_public_test.go`

This task implements one handler at a time, TDD-style. The `create` handler
is shown in detail as the template; remaining handlers follow the same pattern.

- [ ] **Step 1: Write types.go**

```go
package container

import (
	"log/slog"

	"github.com/retr0h/osapi/internal/job/client"
)

// Container implementation of the Container APIs operations.
type Container struct {
	JobClient client.JobClient
	logger    *slog.Logger
}
```

- [ ] **Step 2: Write container.go**

```go
// Package container provides container management API handlers.
package container

import (
	"log/slog"

	"github.com/retr0h/osapi/internal/api/container/gen"
	"github.com/retr0h/osapi/internal/job/client"
)

var _ gen.StrictServerInterface = (*Container)(nil)

// New factory to create a new instance.
func New(
	logger *slog.Logger,
	jobClient client.JobClient,
) *Container {
	return &Container{
		JobClient: jobClient,
		logger:    logger,
	}
}
```

- [ ] **Step 3: Write failing test for CreateContainer**

Create `container_create_public_test.go`. Follow the pattern from
`internal/api/node/command_exec_post_public_test.go`:
- Test validation failure (missing image → 400)
- Test success (202 with job ID and container info)
- Test error (500 from job client)
- TestCreateContainerHTTP (raw HTTP through Echo stack)
- TestCreateContainerRBACHTTP (401/403/200)

- [ ] **Step 4: Run test to verify it fails**

Run: `go test -run TestContainerCreatePublicTestSuite -v ./internal/api/container/...`
Expected: FAIL

- [ ] **Step 5: Implement CreateContainer handler**

Follow `internal/api/node/command_exec_post.go` as the template:
- Validate hostname with `validateHostname`
- Validate request body with `validation.Struct`
- Call `s.JobClient.ModifyContainerCreate`
- Return 202 with job ID and result

- [ ] **Step 6: Run test to verify it passes**

Run: `go test -run TestContainerCreatePublicTestSuite -v ./internal/api/container/...`
Expected: PASS

- [ ] **Step 7: Commit create handler**

```bash
git add internal/api/container/types.go internal/api/container/container.go \
       internal/api/container/container_create.go \
       internal/api/container/container_create_public_test.go
git commit -m "feat(container): add create container handler"
```

- [ ] **Step 8: Implement remaining handlers (TDD cycle each)**

For each of: `list`, `inspect`, `start`, `stop`, `remove`, `exec`, `pull`:
1. Write the failing test file
2. Run test → FAIL
3. Write the handler
4. Run test → PASS
5. Verify coverage:
   `go test -coverprofile=cover.out ./internal/api/container/ && go tool cover -func=cover.out | grep -v '100.0%'`
   Expected: no uncovered lines in the new handler file
6. Commit

Commit message pattern: `feat(container): add {operation} container handler`

- [ ] **Step 9: Run full test suite**

Run: `just go::unit`
Expected: all tests pass

- [ ] **Step 10: Commit if any remaining files**

---

## Chunk 4: Server Wiring and Permissions

### Task 9: Permissions

**Files:**
- Modify: `internal/authtoken/permissions.go`

- [ ] **Step 1: Add container permission constants**

Add to `internal/authtoken/permissions.go`:

```go
PermContainerRead    Permission = "container:read"
PermContainerWrite   Permission = "container:write"
PermContainerExecute Permission = "container:execute"
```

- [ ] **Step 2: Add to AllPermissions slice**

Append the three new permissions.

- [ ] **Step 3: Update DefaultRolePermissions**

- `admin`: add `PermContainerRead`, `PermContainerWrite`, `PermContainerExecute`
- `write`: add `PermContainerRead`, `PermContainerWrite`
- `read`: add `PermContainerRead`

- [ ] **Step 4: Run existing auth tests**

Run: `go test -v ./internal/authtoken/...`
Expected: PASS (or update any tests that assert on the full permission set)

- [ ] **Step 5: Commit**

```bash
git add internal/authtoken/permissions.go
git commit -m "feat(container): add container permissions and role mappings"
```

---

### Task 10: Server Handler Wiring

**Files:**
- Create: `internal/api/handler_container.go`
- Modify: `internal/api/handler_public_test.go`
- Modify: `cmd/api_helpers.go`
- Modify: `cmd/api_server_start.go`

Note: `internal/api/handler.go` does NOT need modification. It only contains
the `RegisterHandlers()` pass-through method. Handler wiring happens in
`cmd/api_helpers.go:registerAPIHandlers()`.

- [ ] **Step 1: Write failing test for GetContainerHandler**

Add to `internal/api/handler_public_test.go`, following the
`TestGetHealthHandler` pattern:

```go
func (s *HandlerPublicTestSuite) TestGetContainerHandler() {
	tests := []struct {
		name     string
		validate func([]func(e *echo.Echo))
	}{
		{
			name: "returns container handler functions",
			validate: func(handlers []func(e *echo.Echo)) {
				s.NotEmpty(handlers)
			},
		},
		{
			name: "closure registers routes and middleware executes",
			validate: func(handlers []func(e *echo.Echo)) {
				e := echo.New()
				for _, h := range handlers {
					h(e)
				}
				s.NotEmpty(e.Routes())
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			handlers := s.server.GetContainerHandler(s.mockJobClient)
			tt.validate(handlers)
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -run TestGetContainerHandler -v ./internal/api/...`
Expected: FAIL — `GetContainerHandler` not defined

- [ ] **Step 3: Implement GetContainerHandler**

Create `internal/api/handler_container.go`. Follow
`internal/api/handler_node.go` exactly — no unauthenticated operations:

```go
package api

import (
	"github.com/labstack/echo/v4"
	strictecho "github.com/oapi-codegen/runtime/strictmiddleware/echo"

	"github.com/retr0h/osapi/internal/api/container"
	containerGen "github.com/retr0h/osapi/internal/api/container/gen"
	"github.com/retr0h/osapi/internal/authtoken"
	"github.com/retr0h/osapi/internal/job/client"
)

// GetContainerHandler returns container handler for registration.
func (s *Server) GetContainerHandler(
	jobClient client.JobClient,
) []func(e *echo.Echo) {
	var tokenManager TokenValidator = authtoken.New(s.logger)

	containerHandler := container.New(s.logger, jobClient)

	strictHandler := containerGen.NewStrictHandler(
		containerHandler,
		[]containerGen.StrictMiddlewareFunc{
			func(handler strictecho.StrictEchoHandlerFunc, _ string) strictecho.StrictEchoHandlerFunc {
				return scopeMiddleware(
					handler,
					tokenManager,
					s.appConfig.API.Server.Security.SigningKey,
					containerGen.BearerAuthScopes,
					s.customRoles,
				)
			},
		},
	)

	return []func(e *echo.Echo){
		func(e *echo.Echo) {
			containerGen.RegisterHandlers(e, strictHandler)
		},
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -run TestGetContainerHandler -v ./internal/api/...`
Expected: PASS

- [ ] **Step 5: Add to ServerManager interface**

In `cmd/api_helpers.go`, add to the `ServerManager` interface:

```go
// GetContainerHandler returns container handler for registration.
GetContainerHandler(jobClient jobclient.JobClient) []func(e *echo.Echo)
```

- [ ] **Step 6: Wire into registerAPIHandlers**

In `cmd/api_helpers.go`, add to `registerAPIHandlers`:

```go
handlers = append(handlers, sm.GetContainerHandler(jc)...)
```

- [ ] **Step 7: Wire startup dependencies**

In `cmd/api_server_start.go`, ensure the container handler is passed. Since
`GetContainerHandler` takes only a `jobClient` (same as `GetNodeHandler`),
no new dependencies are needed — it's already wired through the existing
`registerAPIHandlers` call after step 6.

- [ ] **Step 8: Regenerate combined spec**

Run: `just generate`
Expected: combined spec at `internal/api/gen/api.yaml` includes container paths

- [ ] **Step 9: Verify full build**

Run: `go build ./...`
Expected: compiles

- [ ] **Step 10: Run full test suite**

Run: `just go::unit`
Expected: all tests pass

- [ ] **Step 11: Commit**

```bash
git add internal/api/handler_container.go internal/api/handler_public_test.go \
       cmd/api_helpers.go cmd/api_server_start.go
git commit -m "feat(container): wire container handler into API server"
```

---

## Chunk 5: SDK Client and CLI Commands

### Task 11: SDK Client Container Service

**Files:**
- Create: `pkg/sdk/client/container.go`
- Modify: `pkg/sdk/client/client.go` (add Container field)

- [ ] **Step 1: Regenerate SDK client from combined spec**

Run: `go generate ./pkg/sdk/client/gen/...`
Expected: SDK client picks up new container endpoints

- [ ] **Step 2: Write the ContainerService**

Create `pkg/sdk/client/container.go`. Follow `pkg/sdk/client/health.go` for
the service wrapper pattern. Each method calls the generated client method,
handles response codes (200/202, 400, 401, 403, 404, 409, 500), and returns
typed results.

Methods: `Create`, `List`, `Inspect`, `Start`, `Stop`, `Remove`, `Exec`,
`Pull`.

- [ ] **Step 3: Add Container field to Client**

In `pkg/sdk/client/client.go`, add:

```go
Container *ContainerService
```

Initialize it in the constructor.

- [ ] **Step 4: Verify compilation**

Run: `go build ./pkg/sdk/...`
Expected: compiles

- [ ] **Step 5: Commit**

```bash
git add pkg/sdk/client/container.go pkg/sdk/client/client.go \
       pkg/sdk/client/gen/
git commit -m "feat(container): add SDK container service"
```

---

### Task 12: CLI Commands

**Files:**
- Create: `cmd/client_container.go`
- Create: `cmd/client_container_create.go`
- Create: `cmd/client_container_list.go`
- Create: `cmd/client_container_inspect.go`
- Create: `cmd/client_container_start.go`
- Create: `cmd/client_container_stop.go`
- Create: `cmd/client_container_remove.go`
- Create: `cmd/client_container_exec.go`
- Create: `cmd/client_container_pull.go`

- [ ] **Step 1: Write parent command**

Create `cmd/client_container.go`. Follow `cmd/client_health.go`:

```go
package cmd

import (
	"github.com/spf13/cobra"
)

var clientContainerCmd = &cobra.Command{
	Use:   "container",
	Short: "Container management operations",
	Long:  `Manage containers on target nodes.`,
}

func init() {
	clientCmd.AddCommand(clientContainerCmd)
}
```

- [ ] **Step 2: Write create subcommand**

Create `cmd/client_container_create.go`. Follow
`cmd/client_health_status.go` for the pattern. Use flags:
`--target` (required), `--image` (required), `--name`, `--env`,
`--port`, `--volume`, `--auto-start`, `--json`.

Handle all response codes in the switch block: 202, 400
(`handleUnknownError`), 401/403 (`handleAuthError`), 500
(`handleUnknownError`).

- [ ] **Step 3: Write remaining subcommands**

For each operation: `list`, `inspect`, `start`, `stop`, `remove`, `exec`,
`pull`. Each subcommand:
- Registers under `clientContainerCmd`
- Uses flags (e.g., `--id` for operations on a specific container,
  `--target` for node targeting)
- Supports `--json` for raw output
- Handles all API response codes

Commit message pattern per batch:
```bash
git commit -m "feat(container): add container CLI commands"
```

- [ ] **Step 4: Verify build**

Run: `go build ./...`
Expected: compiles

- [ ] **Step 5: Verify help output**

Run: `go run main.go client container --help`
Expected: shows subcommands (create, list, inspect, start, stop, remove,
exec, pull)

- [ ] **Step 6: Commit**

```bash
git add cmd/client_container*.go
git commit -m "feat(container): add container CLI commands"
```

---

## Chunk 6: Provider Run Subcommand

### Task 13: Provider Registry

**Files:**
- Create: `internal/provider/registry/registry.go`
- Create: `internal/provider/registry/registry_public_test.go`

- [ ] **Step 1: Write failing tests**

Test: register a provider, look it up, look up nonexistent provider.

```go
package registry_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/provider/registry"
)

type RegistryPublicTestSuite struct {
	suite.Suite
}

func (s *RegistryPublicTestSuite) TestLookup() {
	tests := []struct {
		name         string
		provider     string
		operation    string
		register     bool
		expectFound  bool
	}{
		{
			name:        "registered provider found",
			provider:    "host",
			operation:   "hostname",
			register:    true,
			expectFound: true,
		},
		{
			name:        "unregistered provider not found",
			provider:    "nonexistent",
			operation:   "get",
			register:    false,
			expectFound: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			r := registry.New()
			if tt.register {
				r.Register(registry.Registration{
					Name: tt.provider,
					Operations: map[string]registry.OperationSpec{
						tt.operation: {
							NewParams: func() any { return nil },
							Run: func(_ context.Context, _ any) (any, error) {
								return "result", nil
							},
						},
					},
				})
			}
			spec, ok := r.Lookup(tt.provider, tt.operation)
			s.Equal(tt.expectFound, ok)
			if tt.expectFound {
				s.NotNil(spec)
			}
		})
	}
}

func TestRegistryPublicTestSuite(t *testing.T) {
	suite.Run(t, new(RegistryPublicTestSuite))
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -run TestRegistryPublicTestSuite -v ./internal/provider/registry/...`
Expected: FAIL

- [ ] **Step 3: Implement the registry**

```go
// Package registry provides a runtime registry for provider operations.
package registry

import "context"

// OperationSpec defines how to create params and run an operation.
type OperationSpec struct {
	NewParams func() any
	Run       func(ctx context.Context, params any) (any, error)
}

// Registration describes a provider and its operations.
type Registration struct {
	Name       string
	Operations map[string]OperationSpec
}

// Registry holds provider registrations.
type Registry struct {
	providers map[string]Registration
}

// New creates a new empty registry.
func New() *Registry {
	return &Registry{
		providers: make(map[string]Registration),
	}
}

// Register adds a provider registration.
func (r *Registry) Register(
	reg Registration,
) {
	r.providers[reg.Name] = reg
}

// Lookup finds an operation spec by provider and operation name.
func (r *Registry) Lookup(
	provider string,
	operation string,
) (*OperationSpec, bool) {
	reg, ok := r.providers[provider]
	if !ok {
		return nil, false
	}

	spec, ok := reg.Operations[operation]
	if !ok {
		return nil, false
	}

	return &spec, true
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -run TestRegistryPublicTestSuite -v ./internal/provider/registry/...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/provider/registry/
git commit -m "feat(container): add provider runtime registry"
```

---

### Task 14: Provider Run CLI Subcommand

**Files:**
- Create: `cmd/provider_run.go`

- [ ] **Step 1: Write the provider run command**

Create `cmd/provider_run.go`. This is a hidden command
(`Hidden: true` on the Cobra command). It:

1. Takes positional args: `provider` and `operation`
2. Takes `--data` flag for JSON input
3. Builds a registry, registers all known providers
4. Looks up the provider/operation
5. Unmarshals JSON into params via `spec.NewParams()`
6. Calls `spec.Run()`
7. Marshals result to JSON on stdout
8. Exits with code 0 on success, 1 on failure

```go
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/provider/registry"
)

var providerRunData string

var providerRunCmd = &cobra.Command{
	Use:    "run [provider] [operation]",
	Short:  "Run a provider operation (internal)",
	Hidden: true,
	Args:   cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		providerName := args[0]
		operationName := args[1]

		reg := buildProviderRegistry()
		spec, ok := reg.Lookup(providerName, operationName)
		if !ok {
			errJSON, _ := json.Marshal(map[string]string{
				"error": fmt.Sprintf("unknown provider/operation: %s/%s", providerName, operationName),
			})
			fmt.Fprintln(os.Stderr, string(errJSON))
			os.Exit(1)
		}

		params := spec.NewParams()
		if providerRunData != "" && params != nil {
			if err := json.Unmarshal([]byte(providerRunData), params); err != nil {
				errJSON, _ := json.Marshal(map[string]string{"error": err.Error()})
				fmt.Fprintln(os.Stderr, string(errJSON))
				os.Exit(1)
			}
		}

		result, err := spec.Run(context.Background(), params)
		if err != nil {
			errJSON, _ := json.Marshal(map[string]string{"error": err.Error()})
			fmt.Fprintln(os.Stderr, string(errJSON))
			os.Exit(1)
		}

		output, _ := json.Marshal(result)
		fmt.Println(string(output))
	},
}

var providerCmd = &cobra.Command{
	Use:    "provider",
	Short:  "Provider operations (internal)",
	Hidden: true,
}

func init() {
	providerRunCmd.Flags().StringVar(&providerRunData, "data", "", "JSON input data")
	providerCmd.AddCommand(providerRunCmd)
	rootCmd.AddCommand(providerCmd)
}
```

- [ ] **Step 2: Write buildProviderRegistry function**

This function creates a `registry.Registry` and registers all known providers
using the platform factory. Initially register just the `host` provider as a
proof of concept. More providers get registered as they are built.

- [ ] **Step 3: Verify build**

Run: `go build ./...`
Expected: compiles

- [ ] **Step 4: Verify hidden from help**

Run: `go run main.go --help`
Expected: `provider` does NOT appear in the command list

- [ ] **Step 5: Commit**

```bash
git add cmd/provider_run.go
git commit -m "feat(container): add hidden provider run subcommand"
```

---

## Chunk 7: Orchestrator DSL

### Task 15: RuntimeTarget Interface

**Files:**
- Create: `pkg/sdk/orchestrator/runtime_target.go`
- Create: `pkg/sdk/orchestrator/runtime_target_public_test.go`

- [ ] **Step 1: Write the RuntimeTarget interface**

```go
// Package orchestrator provides DAG-based task orchestration.
package orchestrator

import "context"

// RuntimeTarget represents a container runtime target that can execute
// provider operations. Implementations exist for Docker (now) and
// LXD/Podman (later).
type RuntimeTarget interface {
	// Name returns the target name (container name).
	Name() string
	// Runtime returns the runtime type ("docker", "lxd", "podman").
	Runtime() string
	// ExecProvider executes a provider operation inside the target.
	ExecProvider(ctx context.Context, provider, operation string, data []byte) ([]byte, error)
}
```

- [ ] **Step 2: Commit**

```bash
git add pkg/sdk/orchestrator/runtime_target.go
git commit -m "feat(container): add RuntimeTarget interface"
```

---

### Task 16: Docker RuntimeTarget Implementation

**Files:**
- Create: `pkg/sdk/orchestrator/docker_target.go`
- Create: `pkg/sdk/orchestrator/docker_target_public_test.go`

- [ ] **Step 1: Write failing test**

Test that `DockerTarget` implements `RuntimeTarget`, returns correct name and
runtime type, and that `ExecProvider` constructs the correct exec call.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -run TestDockerTargetPublicTestSuite -v ./pkg/sdk/orchestrator/...`
Expected: FAIL

- [ ] **Step 3: Implement DockerTarget**

Note: The orchestrator lives in `pkg/sdk/` which should not import `internal/`
packages directly. Instead of importing `runtime.Driver`, we inject an
`ExecFn` function type that the caller wires to the Docker driver.

```go
package orchestrator

import (
	"context"
	"fmt"
)

// ExecFn executes a command inside a container and returns stdout/stderr/exit code.
// This is injected by the caller, typically wired to runtime.Driver.Exec.
type ExecFn func(ctx context.Context, containerID string, command []string) (stdout, stderr string, exitCode int, err error)

// DockerTarget implements RuntimeTarget for Docker containers.
type DockerTarget struct {
	name   string
	image  string
	execFn ExecFn
}

// NewDockerTarget creates a new Docker runtime target.
func NewDockerTarget(
	name string,
	image string,
	execFn ExecFn,
) *DockerTarget {
	return &DockerTarget{
		name:   name,
		image:  image,
		execFn: execFn,
	}
}

// Name returns the container name.
func (t *DockerTarget) Name() string {
	return t.name
}

// Runtime returns "docker".
func (t *DockerTarget) Runtime() string {
	return "docker"
}

// Image returns the container image.
func (t *DockerTarget) Image() string {
	return t.image
}

// ExecProvider runs a provider operation inside this container via docker exec.
func (t *DockerTarget) ExecProvider(
	ctx context.Context,
	provider string,
	operation string,
	data []byte,
) ([]byte, error) {
	cmd := []string{"/osapi", "provider", "run", provider, operation}
	if len(data) > 0 {
		cmd = append(cmd, "--data", string(data))
	}

	stdout, stderr, exitCode, err := t.execFn(ctx, t.name, cmd)
	if err != nil {
		return nil, fmt.Errorf("exec provider in container %s: %w", t.name, err)
	}

	if exitCode != 0 {
		return nil, fmt.Errorf("provider %s/%s failed (exit %d): %s",
			provider, operation, exitCode, stderr)
	}

	return []byte(stdout), nil
}
```

The caller (e.g., `cmd/` or the user's orchestrator code) wires the `ExecFn`
by closing over the Docker driver:

```go
execFn := func(ctx context.Context, id string, cmd []string) (string, string, int, error) {
    result, err := dockerDriver.Exec(ctx, id, runtime.ExecParams{Command: cmd})
    if err != nil {
        return "", "", -1, err
    }
    return result.Stdout, result.Stderr, result.ExitCode, nil
}
target := orchestrator.NewDockerTarget("web", "ubuntu:24.04", execFn)
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -run TestDockerTargetPublicTestSuite -v ./pkg/sdk/orchestrator/...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/sdk/orchestrator/docker_target.go \
       pkg/sdk/orchestrator/docker_target_public_test.go
git commit -m "feat(container): add Docker RuntimeTarget implementation"
```

---

### Task 17: Plan.Docker() and Plan.In() Methods

**Files:**
- Create: `pkg/sdk/orchestrator/plan_in.go`
- Create: `pkg/sdk/orchestrator/plan_in_public_test.go`

- [ ] **Step 1: Write failing tests**

Test that `p.Docker()` returns a `*DockerTarget` with correct name/image.
Test that `p.In()` returns a `ScopedPlan` and that `TaskFunc` on the scoped
plan adds a task to the parent plan.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -run TestPlanInPublicTestSuite -v ./pkg/sdk/orchestrator/...`
Expected: FAIL

- [ ] **Step 3: Implement Docker() and In()**

```go
package orchestrator

// Docker creates a DockerTarget bound to this plan's container exec function.
// Panics if no ExecFn was provided via WithDockerExecFn option.
func (p *Plan) Docker(
	name string,
	image string,
) *DockerTarget {
	if p.dockerExecFn == nil {
		panic("orchestrator: Plan.Docker() called without WithDockerExecFn option")
	}
	return NewDockerTarget(name, image, p.dockerExecFn)
}

// In returns a ScopedPlan that routes provider operations through the
// given RuntimeTarget (e.g., a Docker container).
type ScopedPlan struct {
	plan   *Plan
	target RuntimeTarget
}

// In creates a scoped plan context for the given runtime target.
func (p *Plan) In(
	target RuntimeTarget,
) *ScopedPlan {
	return &ScopedPlan{
		plan:   p,
		target: target,
	}
}

// TaskFunc creates a task on the parent plan that executes within the
// scoped runtime target context.
func (sp *ScopedPlan) TaskFunc(
	name string,
	fn TaskFn,
) *Task {
	// Wrap the function to inject the runtime target context
	return sp.plan.TaskFunc(name, fn)
}

// TaskFuncWithResults creates a task with results on the parent plan.
func (sp *ScopedPlan) TaskFuncWithResults(
	name string,
	fn TaskFnWithResults,
) *Task {
	return sp.plan.TaskFuncWithResults(name, fn)
}
```

Note: The full client-interception layer (where SDK calls automatically route
through `docker exec` + `provider run`) is more complex. This initial
implementation provides the `In()` scoping mechanism. The interception layer
that replaces the HTTP transport with Docker exec is the next evolution —
for now, task functions inside `In()` can manually use `target.ExecProvider()`
to run providers in the container.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -run TestPlanInPublicTestSuite -v ./pkg/sdk/orchestrator/...`
Expected: PASS

- [ ] **Step 5: Add dockerExecFn field to Plan**

Modify `pkg/sdk/orchestrator/plan.go` to add an optional `dockerExecFn ExecFn`
field to `PlanConfig` and a `WithDockerExecFn(fn ExecFn)` plan option. The
`Plan.Docker()` method reads from `p.config.dockerExecFn`.

- [ ] **Step 6: Run full orchestrator tests**

Run: `go test -v ./pkg/sdk/orchestrator/...`
Expected: all tests pass

- [ ] **Step 7: Commit**

```bash
git add pkg/sdk/orchestrator/plan_in.go \
       pkg/sdk/orchestrator/plan_in_public_test.go \
       pkg/sdk/orchestrator/plan.go
git commit -m "feat(container): add Plan.Docker() and Plan.In() DSL methods"
```

---

## Chunk 8: Documentation and Verification

### Task 18: Documentation

**Files:**
- Create: `docs/docs/sidebar/features/container-management.md`
- Create: `docs/docs/sidebar/usage/cli/client/container/container.md`
- Create: `docs/docs/sidebar/usage/cli/client/container/create.md`
- Create: `docs/docs/sidebar/usage/cli/client/container/list.md`
- Create: `docs/docs/sidebar/usage/cli/client/container/inspect.md`
- Create: `docs/docs/sidebar/usage/cli/client/container/start.md`
- Create: `docs/docs/sidebar/usage/cli/client/container/stop.md`
- Create: `docs/docs/sidebar/usage/cli/client/container/remove.md`
- Create: `docs/docs/sidebar/usage/cli/client/container/exec.md`
- Create: `docs/docs/sidebar/usage/cli/client/container/pull.md`
- Modify: `docs/docusaurus.config.ts`
- Modify: `docs/docs/sidebar/usage/configuration.md`
- Modify: `docs/docs/sidebar/architecture/system-architecture.md`

- [ ] **Step 1: Write feature page**

Create `docs/docs/sidebar/features/container-management.md`. Follow the
template from existing feature pages in the `features/` directory.

- [ ] **Step 2: Write CLI documentation pages**

Create the parent page with `<DocCardList />` and one page per CLI subcommand
with usage examples and `--json` output.

- [ ] **Step 3: Update docusaurus.config.ts**

Add "Container Management" to the Features navbar dropdown.

- [ ] **Step 4: Update configuration.md**

Add a note that no new configuration sections are needed for container
management (Docker socket is auto-detected). Also update the Permissions
table to include `container:read`, `container:write`, `container:execute`
and update the role mappings table to show which roles get which container
permissions.

- [ ] **Step 5: Update system-architecture.md**

Add container endpoints to the endpoint tables.

- [ ] **Step 6: Check docs formatting**

Run: `just docs::fmt-check`
Expected: passes (or fix formatting)

- [ ] **Step 7: Commit**

```bash
git add docs/
git commit -m "docs: add container management documentation"
```

---

### Task 19: Integration Test Smoke Suite

**Files:**
- Create: `test/integration/container_test.go`

Note: Integration tests require a running Docker daemon. They are guarded by
`//go:build integration` and run with `just go::unit-int`. This task creates
a minimal smoke test. Write tests (mutations) must be guarded by
`skipWrite(s.T())`.

- [ ] **Step 1: Create integration test file**

Follow existing patterns in `test/integration/`. The test should:
- Create a container (`POST /node/{hostname}/container`)
- List containers and verify it appears
- Inspect the container
- Exec a command inside it
- Stop and remove the container
- Pull a known small image (e.g., `alpine:latest`)

- [ ] **Step 2: Commit**

```bash
git add test/integration/container_test.go
git commit -m "test(container): add integration test smoke suite"
```

---

### Task 20: Final Verification

- [ ] **Step 1: Regenerate all specs and code**

Run: `just generate`
Expected: no errors, all generated files up to date

- [ ] **Step 2: Build**

Run: `go build ./...`
Expected: compiles

- [ ] **Step 3: Run all unit tests**

Run: `just go::unit`
Expected: all tests pass

- [ ] **Step 4: Run linter**

Run: `just go::vet`
Expected: no lint errors

- [ ] **Step 5: Run full test suite with coverage**

Run: `just test`
Expected: all checks pass (lint + unit + coverage)

- [ ] **Step 6: Verify new packages have 100% coverage**

Run coverage and check that every new package has 100% line coverage:

```bash
go test -race -coverprofile=.coverage/cover.out -v ./...
grep -v -f .coverignore .coverage/cover.out > .coverage/cover.tmp && mv .coverage/cover.tmp .coverage/cover.out
go tool cover -func=.coverage/cover.out | grep -E 'container|registry' | grep -v '100.0%'
```

Expected: **no output** (all container and registry packages at 100%).

If any lines are uncovered, add tests before proceeding. The `.coverignore`
already excludes `/cmd/`, `/gen/`, `main.go`, and `/mocks/`, so handler
tests in `internal/api/container/` and provider tests in
`internal/provider/container/` are what matter.

Also verify that overall project coverage did not decrease by comparing with
the Codecov baseline. Run:

```bash
go tool cover -func=.coverage/cover.out | tail -1
```

This shows the total coverage percentage. It should be at or above the
pre-existing level.

- [ ] **Step 7: Commit any formatting fixes**

If `just go::fmt` produces changes:

```bash
just go::fmt
git add -u
git commit -m "style: format container code"
```
