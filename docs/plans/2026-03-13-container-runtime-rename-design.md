# Container Runtime: Docker-Specific Domain Design

## Problem

The current container domain uses a generic `container` name with an
auto-detecting runtime driver. This is wrong — the user decides which
runtime to use, not the agent. Docker, LXD, and Podman are fundamentally
different systems with different concepts, options, and behaviors. A shared
abstraction would be lowest-common-denominator or leak everywhere.

## Decision

Rename the `container` domain to `docker`. Each future runtime (LXD,
Podman) becomes its own independent domain — no shared interface, no
shared types, no abstraction tax.

The CLI groups runtimes under a `container` parent command for
discoverability. API paths mirror this with `/container/docker/`.

## Architecture

### Naming Convention

| Layer | Current | New |
|-------|---------|-----|
| API paths | `/node/{hostname}/container` | `/node/{hostname}/container/docker` |
| Permissions | `container:read/write/execute` | `docker:read/write/execute` |
| CLI | `client container list` | `client container docker list` |
| SDK | `client.Container.Pull()` | `client.Docker.Pull()` |
| Job category | `container` | `docker` |
| Provider pkg | `internal/provider/container/` | `internal/provider/docker/` |
| API pkg | `internal/api/container/` | `internal/api/docker/` |

### API Endpoints

| Method | Path | Operation | Permission |
|--------|------|-----------|------------|
| `POST` | `/node/{hostname}/container/docker` | Create | `docker:write` |
| `GET` | `/node/{hostname}/container/docker` | List | `docker:read` |
| `GET` | `/node/{hostname}/container/docker/{id}` | Inspect | `docker:read` |
| `POST` | `/node/{hostname}/container/docker/{id}/start` | Start | `docker:write` |
| `POST` | `/node/{hostname}/container/docker/{id}/stop` | Stop | `docker:write` |
| `DELETE` | `/node/{hostname}/container/docker/{id}` | Remove | `docker:write` |
| `POST` | `/node/{hostname}/container/docker/{id}/exec` | Exec | `docker:execute` |
| `POST` | `/node/{hostname}/container/docker/pull` | Pull | `docker:write` |

### CLI

```
osapi client container docker list [--target HOST] [--state STATE] [--limit N]
osapi client container docker create --target HOST --image IMAGE [--name NAME] ...
osapi client container docker inspect --target HOST --id ID
osapi client container docker start --target HOST --id ID
osapi client container docker stop --target HOST --id ID [--timeout SECONDS]
osapi client container docker remove --target HOST --id ID [--force]
osapi client container docker exec --target HOST --id ID --command CMD...
osapi client container docker pull --target HOST --image IMAGE
```

The `container` command is a parent with `<DocCardList />` for grouping.
Each runtime is a subcommand. Future runtimes add `client container lxd`,
`client container podman`, etc.

### Role Updates

| Role | Permissions |
|------|------------|
| `admin` | `docker:read`, `docker:write`, `docker:execute` |
| `write` | `docker:read`, `docker:write` |
| `read` | `docker:read` |

### Package Layout

```
internal/provider/docker/
├── docker.go                  # Provider struct, New()
├── types.go                   # CreateParams, Container, etc.
├── docker_test.go             # Tests
└── (no runtime/ subdirectory)

internal/api/docker/
├── gen/
│   ├── api.yaml               # OpenAPI spec
│   ├── cfg.yaml               # oapi-codegen config
│   └── generate.go            # go:generate directive
├── types.go                   # Domain struct, interfaces
├── docker.go                  # New(), interface check
├── docker_create.go           # Create handler
├── docker_list.go             # List handler
├── docker_inspect.go          # Inspect handler
├── docker_start.go            # Start handler
├── docker_stop.go             # Stop handler
├── docker_remove.go           # Remove handler
├── docker_exec.go             # Exec handler
├── docker_pull.go             # Pull handler
└── *_public_test.go           # Tests

internal/api/
├── handler_docker.go          # GetDockerHandler() method
└── handler.go                 # +RegisterHandlers() wiring

cmd/
├── client_container.go                # parent: `container` subcommand
├── client_container_docker.go         # parent: `docker` subcommand
├── client_container_docker_create.go
├── client_container_docker_list.go
├── client_container_docker_inspect.go
├── client_container_docker_start.go
├── client_container_docker_stop.go
├── client_container_docker_remove.go
├── client_container_docker_exec.go
└── client_container_docker_pull.go

pkg/sdk/client/
├── docker.go                  # DockerService
└── docker_types.go            # DockerResult, etc.

internal/agent/
├── processor_docker.go        # docker case + dispatch
└── types.go                   # dockerProvider field
```

### No Shared Runtime Interface

The `runtime.Driver` interface in
`internal/provider/container/runtime/driver.go` is removed. The Docker
provider defines its own types directly. When LXD is added, it gets its
own provider package (`internal/provider/lxd/`) with its own types —
LXD concepts (instances, profiles, projects) don't map to Docker
concepts (images, containers, layers).

Each runtime is fully independent:
- Own API domain, paths, and OpenAPI schemas
- Own CLI subcommands under `client container <runtime>`
- Own SDK service (`client.Docker`, `client.Lxd`)
- Own permissions (`docker:read`, `lxd:read`)
- Own provider package with own types
- Own orchestrator helpers

### Orchestrator DSL

Convenience methods on `*Plan` in `pkg/sdk/orchestrator/` wrap
`client.Docker.*` calls so users don't write boilerplate TaskFunc
bodies:

```go
plan.DockerPull("pull-image", target, "ubuntu:24.04")
plan.DockerCreate("create-app", target, gen.DockerCreateRequest{...})
plan.DockerExec("run-cmd", target, "my-app", gen.DockerExecRequest{...})
plan.DockerInspect("check", target, "my-app")
plan.DockerStart("start", target, "my-app")
plan.DockerStop("stop", target, "my-app", gen.DockerStopRequest{...})
plan.DockerRemove("cleanup", target, "my-app", &gen.DeleteNodeDockerByIDParams{...})
```

Each returns `*Task` for chaining (`DependsOn`, `OnlyIfChanged`, etc.).
Future runtimes add `plan.LxdLaunch(...)`, `plan.LxdExec(...)` — no
shared interface.

### Documentation

- `docs/docs/sidebar/features/container-management.md` — update to
  describe Docker as the first runtime, explain the per-runtime model
- CLI docs: restructure under `container/docker/`
- SDK orchestrator docs: update container-targeting to use
  `plan.DockerPull` etc.

## What This Changes

This is a mechanical rename + restructure of the existing fully-built
container domain. No behavior changes. The scope is:

1. Rename ~40+ files across all layers (API, CLI, SDK, agent, provider,
   job types, tests, docs)
2. Remove `internal/provider/container/runtime/driver.go` shared
   interface
3. Flatten `internal/provider/container/runtime/docker/` →
   `internal/provider/docker/`
4. Add `container` parent CLI command
5. Add orchestrator DSL helpers
6. Update all docs

## Key Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| User chooses runtime | Yes | Agent shouldn't guess; Docker/LXD/Podman are different |
| Separate domains per runtime | Yes | No useful shared abstraction across runtimes |
| CLI nesting | `container docker` | Groups runtimes for discoverability |
| API path nesting | `/container/docker` | Mirrors CLI structure |
| No shared interface | Yes | LXD concepts don't map to Docker concepts |
| Flat provider packages | Yes | No shared parent code to justify nesting |
| Orchestrator helpers | Methods on Plan | Eliminates TaskFunc boilerplate |
