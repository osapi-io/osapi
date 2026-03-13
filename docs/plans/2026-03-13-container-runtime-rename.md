# Container → Docker Domain Rename Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to
> implement this plan task-by-task.

**Goal:** Rename the generic `container` domain to `docker`, nest under a
`container` parent in CLI/API paths, remove the shared `runtime.Driver`
interface, and add orchestrator DSL helpers.

**Architecture:** Mechanical rename across all layers (API, provider, agent,
job, CLI, SDK, permissions, docs, tests) from `container` to `docker`. API paths
change from `/node/{hostname}/container` to `/node/{hostname}/container/docker`.
CLI changes from `client container list` to `client container docker list`. The
shared `runtime.Driver` interface is removed — Docker provider owns its types
directly. New orchestrator DSL methods (`plan.DockerPull`, etc.) wrap SDK client
calls.

**Tech Stack:** Go, Echo, oapi-codegen, Cobra, testify/suite, Docker Go SDK

**Spec:** `docs/plans/2026-03-13-container-runtime-rename-design.md`

---

## Chunk 1: OpenAPI Spec + Permissions

### Task 1: Rename OpenAPI Spec

**Files:**

- Modify: `internal/api/container/gen/api.yaml`

**Step 1:** Rename the directory:

```bash
git mv internal/api/container internal/api/docker
```

**Step 2:** Edit `internal/api/docker/gen/api.yaml`:

- Change all paths from `/node/{hostname}/container` to
  `/node/{hostname}/container/docker`
- Change all security scopes from `container:read/write/execute` to
  `docker:read/write/execute`
- Rename all schema names from `Container*` to `Docker*` (e.g.,
  `ContainerCreateRequest` → `DockerCreateRequest`, `ContainerResponse` →
  `DockerResponse`)
- Rename all operation IDs from `*Container*` to `*Docker*`
- Update tag names and descriptions from "Container" to "Docker"
- Update the `ContainerId` parameter to `DockerId`

**Step 3:** Update `internal/api/docker/gen/cfg.yaml`:

- Change output filename from `container.gen.go` to `docker.gen.go`
- Update package name if needed

**Step 4:** Update `internal/api/docker/gen/generate.go`:

- Update the `go:generate` directive path if needed

**Step 5:** Regenerate:

```bash
go generate ./internal/api/docker/gen/...
```

**Step 6:** Verify generation succeeded — `docker.gen.go` should exist:

```bash
ls internal/api/docker/gen/docker.gen.go
```

**Step 7:** Delete old generated file if it still exists:

```bash
rm -f internal/api/docker/gen/container.gen.go
```

**Step 8:** Commit:

```bash
git add internal/api/docker/ internal/api/container/
git commit -m "refactor: rename container OpenAPI spec to docker"
```

---

### Task 2: Rename Permissions

**Files:**

- Modify: `internal/authtoken/permissions.go`

**Step 1:** Rename the permission constants:

- `PermContainerRead` → `PermDockerRead` with value `"docker:read"`
- `PermContainerWrite` → `PermDockerWrite` with value `"docker:write"`
- `PermContainerExecute` → `PermDockerExecute` with value `"docker:execute"`

**Step 2:** Update the `AllPermissions` slice to use the new names.

**Step 3:** Update the default role mappings (admin, write, read) to use the new
permission constants.

**Step 4:** Search for any other files referencing the old permission constants:

```bash
grep -rn 'PermContainer\|container:read\|container:write\|container:execute' \
  --include='*.go' . | grep -v '.gen.go' | grep -v '_test.go'
```

Fix all references found (likely in `internal/api/docker/` handlers and
`cmd/api_helpers.go`).

**Step 5:** Verify it compiles:

```bash
go build ./internal/authtoken/... ./internal/api/docker/...
```

**Step 6:** Run permission tests:

```bash
go test ./internal/authtoken/... -count=1
```

**Step 7:** Commit:

```bash
git add -A && git commit -m "refactor: rename container permissions to docker"
```

---

## Chunk 2: Provider Layer

### Task 3: Flatten and Rename Provider

**Files:**

- Rename: `internal/provider/container/` → `internal/provider/docker/`
- Remove: `internal/provider/container/runtime/driver.go` (shared interface)

**Step 1:** Move the Docker driver implementation up and rename:

```bash
git mv internal/provider/container internal/provider/docker
```

**Step 2:** The directory structure should become:

```
internal/provider/docker/
├── provider.go
├── types.go
├── provider_public_test.go
├── mocks/
├── runtime/
│   ├── driver.go          ← DELETE this (shared interface)
│   ├── docker/
│   │   ├── docker.go
│   │   └── docker_public_test.go
│   └── mocks/
```

Move `runtime/docker/docker.go` types and implementation into the parent
package, or keep `runtime/docker/` as the actual Docker SDK driver. The provider
in `provider.go` already wraps the driver, so the structure can stay — just
update the package import paths from `internal/provider/container/...` to
`internal/provider/docker/...`.

**Step 3:** Delete the shared `runtime.Driver` interface:

```bash
rm internal/provider/docker/runtime/driver.go
```

**Step 4:** Update the Docker driver (`runtime/docker/docker.go`) to define its
own interface or use concrete types instead of the removed `runtime.Driver`. The
provider in `provider.go` should type-assert or use the concrete Docker driver
type.

**Step 5:** Update all import paths in the provider package from
`internal/provider/container` to `internal/provider/docker`.

**Step 6:** Update package declarations — `package container` → `package docker`
in `provider.go`, `types.go`, etc.

**Step 7:** Rename the `Provider` interface methods and types if they use
`Container` prefix (check `types.go`).

**Step 8:** Update mock generation directives in `mocks/generate.go`.

**Step 9:** Regenerate mocks:

```bash
go generate ./internal/provider/docker/mocks/...
go generate ./internal/provider/docker/runtime/mocks/...
```

**Step 10:** Verify it compiles:

```bash
go build ./internal/provider/docker/...
```

**Step 11:** Run tests:

```bash
go test ./internal/provider/docker/... -count=1
```

**Step 12:** Commit:

```bash
git add -A && git commit -m "refactor: rename container provider to docker"
```

---

## Chunk 3: Job Types

### Task 4: Rename Job Types and Operations

**Files:**

- Modify: `internal/job/types.go`
- Modify: `internal/job/client/modify_container.go` (rename to
  `modify_docker.go`)
- Modify: `internal/job/client/modify_container_public_test.go` (rename to
  `modify_docker_public_test.go`)

**Step 1:** In `internal/job/types.go`, rename all container operation
constants:

- `OperationContainerCreate` → `OperationDockerCreate` with value
  `"docker.create.execute"`
- `OperationContainerStart` → `OperationDockerStart` with value
  `"docker.start.execute"`
- `OperationContainerStop` → `OperationDockerStop` with value
  `"docker.stop.execute"`
- `OperationContainerRemove` → `OperationDockerRemove` with value
  `"docker.remove.execute"`
- `OperationContainerList` → `OperationDockerList` with value
  `"docker.list.get"`
- `OperationContainerInspect` → `OperationDockerInspect` with value
  `"docker.inspect.get"`
- `OperationContainerExec` → `OperationDockerExec` with value
  `"docker.exec.execute"`
- `OperationContainerPull` → `OperationDockerPull` with value
  `"docker.pull.execute"`

**Step 2:** Rename all data types:

- `ContainerCreateData` → `DockerCreateData`
- `ContainerStopData` → `DockerStopData`
- `ContainerRemoveData` → `DockerRemoveData`
- `ContainerListData` → `DockerListData`
- `ContainerExecData` → `DockerExecData`
- `ContainerPullData` → `DockerPullData`

**Step 3:** Rename the job client file:

```bash
git mv internal/job/client/modify_container.go \
       internal/job/client/modify_docker.go
git mv internal/job/client/modify_container_public_test.go \
       internal/job/client/modify_docker_public_test.go
```

**Step 4:** Update function names in the renamed files (e.g., `ModifyContainer*`
→ `ModifyDocker*` or whatever the current pattern is).

**Step 5:** Search for all remaining references to old constant/type names:

```bash
grep -rn 'OperationContainer\|ContainerCreateData\|ContainerStopData\|ContainerRemoveData\|ContainerListData\|ContainerExecData\|ContainerPullData' \
  --include='*.go' . | grep -v '.gen.go'
```

Fix all references found.

**Step 6:** Verify it compiles:

```bash
go build ./internal/job/...
```

**Step 7:** Run tests:

```bash
go test ./internal/job/... -count=1
```

**Step 8:** Commit:

```bash
git add -A && git commit -m "refactor: rename container job types to docker"
```

---

## Chunk 4: Agent Layer

### Task 5: Rename Agent Processor and Wiring

**Files:**

- Rename: `internal/agent/processor_container.go` →
  `internal/agent/processor_docker.go`
- Rename: `internal/agent/processor_container_test.go` →
  `internal/agent/processor_docker_test.go`
- Modify: `internal/agent/types.go`
- Modify: `internal/agent/agent.go`
- Modify: `internal/agent/factory.go`
- Modify: `internal/agent/processor.go`
- Modify: `internal/agent/factory_test.go`
- Modify: `internal/agent/factory_public_test.go`

**Step 1:** Rename processor files:

```bash
git mv internal/agent/processor_container.go \
       internal/agent/processor_docker.go
git mv internal/agent/processor_container_test.go \
       internal/agent/processor_docker_test.go
```

**Step 2:** In `processor_docker.go`:

- Rename `processContainerOperation` → `processDockerOperation`
- Rename `processContainerCreate` → `processDockerCreate` (and all 8 process
  methods)
- Change `a.containerProvider` → `a.dockerProvider`
- Update import from `internal/provider/container` to `internal/provider/docker`

**Step 3:** In `processor.go`:

- Change `case "container":` → `case "docker":`
- Change `a.processContainerOperation` → `a.processDockerOperation`

**Step 4:** In `types.go`:

- Change `containerProvider containerProv.Provider` →
  `dockerProvider dockerProv.Provider`
- Update the import alias from `containerProv` to `dockerProv`

**Step 5:** In `agent.go`:

- Update parameter name from `containerProvider` to `dockerProvider`
- Update field assignment

**Step 6:** In `factory.go`:

- Rename `containerProvider` variable to `dockerProvider`
- Update import from `internal/provider/container` to `internal/provider/docker`
- Update the return value

**Step 7:** In `factory_test.go` and `factory_public_test.go`:

- Update variable names and comments

**Step 8:** In `processor_docker_test.go`:

- Update all function names and references

**Step 9:** Verify it compiles:

```bash
go build ./internal/agent/...
```

**Step 10:** Run tests:

```bash
go test ./internal/agent/... -count=1
```

**Step 11:** Commit:

```bash
git add -A && git commit -m "refactor: rename container agent wiring to docker"
```

---

## Chunk 5: API Handlers

### Task 6: Rename API Handler Files and Wiring

**Files:**

- Modify: all files in `internal/api/docker/` (already moved in Task 1)
- Rename: `internal/api/handler_container.go` → `internal/api/handler_docker.go`
- Modify: `internal/api/handler.go`
- Modify: `internal/api/types.go`
- Modify: `internal/api/handler_public_test.go`
- Modify: `cmd/api_helpers.go`

**Step 1:** In `internal/api/docker/`:

- Rename all files: `container_create.go` → `docker_create.go`, etc. (8 handler
  files + 8 test files)
- Update package declaration from `package container` to `package docker`
- Rename the `Container` struct to `Docker`
- Update `New()` to return `*Docker`
- Rename handler methods (e.g., `PostNodeContainer` → `PostNodeContainerDocker`)
- Update all gen import aliases from `containerGen` to `dockerGen`
- Update compile-time interface check
- Update references to old job operation constants
- Update references to old data types

**Step 2:** Rename and update `internal/api/docker/types.go`:

- Rename the struct and any interfaces

**Step 3:** Rename and update `internal/api/docker/convert.go`:

- Update function names and types

**Step 4:** Rename and update `internal/api/docker/validate.go`:

- Update function names

**Step 5:** Rename server wiring:

```bash
git mv internal/api/handler_container.go internal/api/handler_docker.go
```

**Step 6:** In `handler_docker.go`:

- Rename `GetContainerHandler` → `GetDockerHandler`
- Update imports from `internal/api/container` to `internal/api/docker`
- Update scope references from `container:*` to `docker:*`

**Step 7:** In `types.go`, rename the handler field and option function.

**Step 8:** In `handler.go`, update the `RegisterHandlers` call.

**Step 9:** In `handler_public_test.go`, rename `TestGetContainerHandler` →
`TestGetDockerHandler`.

**Step 10:** In `cmd/api_helpers.go`, update `GetContainerHandler` →
`GetDockerHandler`.

**Step 11:** Regenerate the combined spec and SDK client:

```bash
just generate
go generate ./pkg/sdk/client/gen/...
```

**Step 12:** Verify it compiles:

```bash
go build ./...
```

**Step 13:** Run tests:

```bash
go test ./internal/api/docker/... ./internal/api/... -count=1
```

**Step 14:** Commit:

```bash
git add -A && git commit -m "refactor: rename container API handlers to docker"
```

---

## Chunk 6: CLI

### Task 7: Restructure CLI Commands

**Files:**

- Modify: `cmd/client_container.go` (becomes parent with just `<DocCardList />`
  grouping)
- Create: `cmd/client_container_docker.go` (new `docker` subcommand)
- Rename: all `cmd/client_container_*.go` → `cmd/client_container_docker_*.go`

**Step 1:** Update `cmd/client_container.go` to be a thin parent command:

```go
var clientContainerCmd = &cobra.Command{
    Use:   "container",
    Short: "Container runtime management",
    Long:  `Manage containers using runtime-specific subcommands.`,
}

func init() {
    clientCmd.AddCommand(clientContainerCmd)
}
```

**Step 2:** Create `cmd/client_container_docker.go`:

```go
var clientContainerDockerCmd = &cobra.Command{
    Use:   "docker",
    Short: "Docker container operations",
    Long:  `Manage Docker containers on target nodes.`,
}

func init() {
    clientContainerCmd.AddCommand(clientContainerDockerCmd)
}
```

**Step 3:** Rename all subcommand files:

```bash
git mv cmd/client_container_create.go cmd/client_container_docker_create.go
git mv cmd/client_container_list.go cmd/client_container_docker_list.go
git mv cmd/client_container_inspect.go cmd/client_container_docker_inspect.go
git mv cmd/client_container_start.go cmd/client_container_docker_start.go
git mv cmd/client_container_stop.go cmd/client_container_docker_stop.go
git mv cmd/client_container_remove.go cmd/client_container_docker_remove.go
git mv cmd/client_container_exec.go cmd/client_container_docker_exec.go
git mv cmd/client_container_pull.go cmd/client_container_docker_pull.go
```

**Step 4:** In each renamed file:

- Change parent command registration from `clientContainerCmd.AddCommand(...)`
  to `clientContainerDockerCmd.AddCommand(...)`
- Rename cobra command variables from `clientContainer*Cmd` to
  `clientContainerDocker*Cmd`
- Update SDK client calls from `c.Container.*` to `c.Docker.*`
- Update generated type references from `gen.Container*` to `gen.Docker*`
- Update `gen.GetNodeContainerParams*` to `gen.GetNodeDockerParams*` (or
  whatever the regenerated names are)

**Step 5:** Verify the CLI compiles:

```bash
go build ./cmd/...
```

**Step 6:** Verify the command tree looks right:

```bash
go run main.go client container --help
go run main.go client container docker --help
```

**Step 7:** Commit:

```bash
git add -A && git commit -m "refactor: nest docker CLI under client container docker"
```

---

## Chunk 7: SDK Client

### Task 8: Rename SDK Client Service

**Files:**

- Rename: `pkg/sdk/client/container.go` → `pkg/sdk/client/docker.go`
- Rename: `pkg/sdk/client/container_types.go` → `pkg/sdk/client/docker_types.go`
- Rename: `pkg/sdk/client/container_public_test.go` →
  `pkg/sdk/client/docker_public_test.go`
- Rename: `pkg/sdk/client/container_types_test.go` →
  `pkg/sdk/client/docker_types_test.go`
- Modify: `pkg/sdk/client/osapi.go`

**Step 1:** Rename files:

```bash
git mv pkg/sdk/client/container.go pkg/sdk/client/docker.go
git mv pkg/sdk/client/container_types.go pkg/sdk/client/docker_types.go
git mv pkg/sdk/client/container_public_test.go \
       pkg/sdk/client/docker_public_test.go
git mv pkg/sdk/client/container_types_test.go \
       pkg/sdk/client/docker_types_test.go
```

**Step 2:** In `docker.go`:

- Rename `ContainerService` → `DockerService`
- Update all method bodies to use the regenerated `gen.Docker*` type names
- Update error message prefixes

**Step 3:** In `docker_types.go`:

- Rename all types: `ContainerResult` → `DockerResult`, `ContainerListResult` →
  `DockerListResult`, etc.
- Rename all converter functions: `containerResultCollectionFromGen` →
  `dockerResultCollectionFromGen`, etc.

**Step 4:** In `osapi.go`:

- Change field `Container *ContainerService` → `Docker *DockerService`
- Update initialization: `c.Container = &ContainerService{...}` →
  `c.Docker = &DockerService{...}`
- Update comment

**Step 5:** In test files, update all type and method references.

**Step 6:** Search for remaining `Container` references in the SDK:

```bash
grep -rn 'Container' pkg/sdk/client/ --include='*.go' | grep -v '.gen.go'
```

Fix all remaining references.

**Step 7:** Verify it compiles:

```bash
go build ./pkg/sdk/client/...
```

**Step 8:** Run tests:

```bash
go test ./pkg/sdk/client/... -count=1
```

**Step 9:** Commit:

```bash
git add -A && git commit -m "refactor: rename container SDK client to docker"
```

---

## Chunk 8: Orchestrator DSL Helpers

### Task 9: Add Orchestrator Docker Methods

**Files:**

- Create: `pkg/sdk/orchestrator/docker.go`
- Create: `pkg/sdk/orchestrator/docker_public_test.go`

**Step 1:** Write the test file `docker_public_test.go` with a test suite
covering each helper method. Use table-driven tests. Mock the client responses
or test that the correct TaskFunc is created:

```go
func (s *DockerPublicTestSuite) TestDockerPull() {
    tests := []struct {
        name         string
        target       string
        image        string
        validateFunc func(task *orchestrator.Task)
    }{
        {
            name:   "creates task with correct name",
            target: "_any",
            image:  "ubuntu:24.04",
            validateFunc: func(task *orchestrator.Task) {
                s.Equal("pull-image", task.Name())
            },
        },
    }
    // ...
}
```

Test each method: `DockerPull`, `DockerCreate`, `DockerExec`, `DockerInspect`,
`DockerStart`, `DockerStop`, `DockerRemove`, `DockerList`.

**Step 2:** Run tests to verify they fail:

```bash
go test ./pkg/sdk/orchestrator/... -count=1 -run TestDocker
```

**Step 3:** Implement `docker.go` with methods on `*Plan`:

```go
// DockerPull creates a task that pulls a Docker image on the target host.
func (p *Plan) DockerPull(
    name string,
    target string,
    image string,
) *Task {
    return p.TaskFunc(name, func(
        ctx context.Context,
        c *osapiclient.Client,
    ) (*Result, error) {
        resp, err := c.Docker.Pull(ctx, target, gen.DockerPullRequest{
            Image: image,
        })
        if err != nil {
            return nil, err
        }
        r := resp.Data.Results[0]
        return &Result{
            Changed: true,
            Data: map[string]any{
                "image_id": r.ImageID,
                "tag":      r.Tag,
                "size":     r.Size,
            },
        }, nil
    })
}
```

Follow the same pattern for all 8 operations. Each method:

- Takes the task name, target, and operation-specific params
- Returns `*Task` for chaining
- Wraps the SDK client call in a `TaskFunc`
- Sets `Changed: true` for mutations, `Changed: false` for reads (inspect, list)
- Populates `Data` with relevant result fields

Methods to implement:

- `DockerPull(name, target, image string) *Task`
- `DockerCreate(name, target string, body gen.DockerCreateRequest) *Task`
- `DockerStart(name, target, id string) *Task`
- `DockerStop(name, target, id string, body gen.DockerStopRequest) *Task`
- `DockerRemove(name, target, id string, params *gen.DeleteNodeContainerDockerByIDParams) *Task`
- `DockerExec(name, target, id string, body gen.DockerExecRequest) *Task`
- `DockerInspect(name, target, id string) *Task`
- `DockerList(name, target string, params *gen.GetNodeContainerDockerParams) *Task`

**Step 4:** Run tests to verify they pass:

```bash
go test ./pkg/sdk/orchestrator/... -count=1
```

**Step 5:** Commit:

```bash
git add -A && git commit -m "feat: add orchestrator Docker DSL helpers"
```

---

### Task 10: Rewrite Container Targeting Example

**Files:**

- Modify: `examples/sdk/orchestrator/features/container-targeting.go`

**Step 1:** Rewrite the example to use the new DSL helpers:

```go
plan := orchestrator.NewPlan(apiClient,
    orchestrator.WithHooks(hooks),
    orchestrator.OnError(orchestrator.Continue),
)

pull := plan.DockerPull("pull-image", target, containerImage)

create := plan.DockerCreate("create-container", target,
    gen.DockerCreateRequest{
        Image:     containerImage,
        Name:      ptr(containerName),
        AutoStart: &autoStart,
        Command:   &[]string{"sleep", "600"},
    },
)
create.DependsOn(pull)

plan.DockerExec("exec-hostname", target, containerName,
    gen.DockerExecRequest{Command: []string{"hostname"}},
).DependsOn(create)

plan.DockerInspect("inspect", target, containerName).DependsOn(create)

plan.DockerRemove("cleanup", target, containerName,
    &gen.DeleteNodeContainerDockerByIDParams{Force: &force},
).DependsOn(create)
```

**Step 2:** Update SDK client example too: `examples/sdk/client/container.go` —
update to use `c.Docker.*` and `gen.Docker*` types.

**Step 3:** Verify examples compile:

```bash
go build ./examples/...
```

**Step 4:** Commit:

```bash
git add -A && git commit -m "refactor: update examples to use docker DSL"
```

---

## Chunk 9: Integration Tests

### Task 11: Rename Integration Tests

**Files:**

- Rename: `test/integration/container_test.go` →
  `test/integration/docker_test.go`

**Step 1:** Rename the file:

```bash
git mv test/integration/container_test.go test/integration/docker_test.go
```

**Step 2:** Update the test to:

- Use `c.Docker.*` instead of `c.Container.*`
- Use `gen.Docker*` types instead of `gen.Container*`
- Update CLI commands from `container list` to `container docker list`
- Rename suite/test names from `Container*` to `Docker*`

**Step 3:** Verify it compiles:

```bash
go build ./test/integration/...
```

**Step 4:** Commit:

```bash
git add -A && git commit -m "refactor: rename container integration tests to docker"
```

---

## Chunk 10: Documentation

### Task 12: Update Documentation

**Files:**

- Modify: `docs/docs/sidebar/features/container-management.md` — update to
  describe Docker as first runtime, update all paths/permissions/CLI examples
- Rename: CLI docs from `docs/.../container/` to restructure under
  `docs/.../container/docker/`
- Modify: SDK orchestrator docs to reference Docker methods
- Modify: `docs/docs/sidebar/usage/configuration.md` — update permission tables
- Modify: `docs/docs/sidebar/architecture/system-architecture.md` — update
  endpoint tables
- Modify: `docs/docusaurus.config.ts` — update navbar links
- Modify: `CLAUDE.md` — update permission tables in role descriptions

**Step 1:** Update `container-management.md`:

- Update title, description
- Update all path examples from `/container` to `/container/docker`
- Update permissions table from `container:*` to `docker:*`
- Update CLI examples from `client container list` to
  `client container docker list`
- Update role descriptions

**Step 2:** Restructure CLI docs:

- Current: `docs/.../cli/client/container/container.mdx` (parent)
- New: `docs/.../cli/client/container/container.mdx` stays as parent grouping
- Create: `docs/.../cli/client/container/docker/` directory
- Move per-operation docs into the docker subdirectory
- Update all command examples

**Step 3:** Update SDK orchestrator operation docs:

- Rename `container-create.md` → `docker-create.md`, etc.
- Update code examples to use `plan.DockerCreate(...)` etc.

**Step 4:** Update configuration.md permission tables.

**Step 5:** Update system-architecture.md endpoint tables.

**Step 6:** Regenerate API docs:

```bash
just docs::generate-api
```

**Step 7:** Commit:

```bash
git add -A && git commit -m "docs: update documentation for docker domain rename"
```

---

## Chunk 11: Verify

### Task 13: Full Verification

**Step 1:** Regenerate everything:

```bash
just generate
```

**Step 2:** Build:

```bash
go build ./...
```

**Step 3:** Run all unit tests:

```bash
just go::unit
```

**Step 4:** Run lint:

```bash
just go::vet
```

**Step 5:** Search for any remaining `container` references that should be
`docker` (excluding the `container` parent command which is intentional):

```bash
grep -rn 'container:read\|container:write\|container:execute' \
  --include='*.go' . | grep -v '.gen.go'
grep -rn 'ContainerService\|ContainerResult\|ContainerCreateData' \
  --include='*.go' . | grep -v '.gen.go'
grep -rn 'OperationContainer' --include='*.go' .
grep -rn 'containerProvider' --include='*.go' .
grep -rn 'processContainer' --include='*.go' .
```

All should return empty.

**Step 6:** Verify CLI works:

```bash
go run main.go client container --help
go run main.go client container docker --help
```

**Step 7:** Commit any final fixes:

```bash
git add -A && git commit -m "chore: final verification cleanup"
```

---

## Files Modified

| Repo  | File                                                            | Change                                 |
| ----- | --------------------------------------------------------------- | -------------------------------------- |
| osapi | `internal/api/container/` → `internal/api/docker/`              | Full directory rename + content update |
| osapi | `internal/api/handler_container.go` → `handler_docker.go`       | Rename + update                        |
| osapi | `internal/api/handler.go`                                       | Update registration                    |
| osapi | `internal/api/types.go`                                         | Rename handler field                   |
| osapi | `internal/api/handler_public_test.go`                           | Rename test                            |
| osapi | `internal/provider/container/` → `internal/provider/docker/`    | Full directory rename                  |
| osapi | `internal/provider/container/runtime/driver.go`                 | DELETE                                 |
| osapi | `internal/agent/processor_container.go` → `processor_docker.go` | Rename + update                        |
| osapi | `internal/agent/types.go`                                       | Rename field                           |
| osapi | `internal/agent/agent.go`                                       | Rename parameter                       |
| osapi | `internal/agent/factory.go`                                     | Rename variable                        |
| osapi | `internal/agent/processor.go`                                   | Update case                            |
| osapi | `internal/job/types.go`                                         | Rename 8 constants + 6 types           |
| osapi | `internal/job/client/modify_container.go` → `modify_docker.go`  | Rename + update                        |
| osapi | `internal/authtoken/permissions.go`                             | Rename 3 constants                     |
| osapi | `cmd/client_container.go`                                       | Simplify to parent                     |
| osapi | `cmd/client_container_docker.go`                                | NEW parent for docker                  |
| osapi | `cmd/client_container_*.go` → `client_container_docker_*.go`    | Rename 8 files                         |
| osapi | `cmd/api_helpers.go`                                            | Update handler call                    |
| osapi | `pkg/sdk/client/container.go` → `docker.go`                     | Rename + update                        |
| osapi | `pkg/sdk/client/container_types.go` → `docker_types.go`         | Rename + update                        |
| osapi | `pkg/sdk/client/osapi.go`                                       | Rename field                           |
| osapi | `pkg/sdk/orchestrator/docker.go`                                | NEW — DSL helpers                      |
| osapi | `pkg/sdk/orchestrator/docker_public_test.go`                    | NEW — tests                            |
| osapi | `examples/sdk/orchestrator/features/container-targeting.go`     | Rewrite with DSL                       |
| osapi | `examples/sdk/client/container.go`                              | Update SDK calls                       |
| osapi | `test/integration/container_test.go` → `docker_test.go`         | Rename + update                        |
| osapi | `docs/` (multiple)                                              | Update paths, permissions, examples    |
