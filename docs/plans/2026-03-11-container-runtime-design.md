# Container Runtime and Provider Execution Context

## Problem

OSAPI providers execute operations directly on the host OS. As we build new
providers (user management, cron, etc.), there is no way to develop, test, or
run them against an isolated environment. Additionally, there is no way to
manage containers on a host or compose workflows that span both the host and
containers running on it.

We need:

1. Container lifecycle management (Docker first, LXD/Podman later) as a new API
   domain
2. A mechanism to run existing providers inside a container without rewriting
   them
3. An orchestrator DSL layer to compose host and container operations in a
   single plan

## Decision

Add a `container` API domain with a pluggable runtime driver interface. Docker
is the first implementation using the Go SDK. Introduce a `provider run` CLI
subcommand that executes a single provider operation as a standalone process
with JSON I/O. The orchestrator DSL uses `docker exec` + `provider run` to
transparently run providers inside containers.

## Architecture

### Runtime Driver Interface

A `runtime.Driver` interface in `internal/provider/container/runtime/` abstracts
container runtime operations. The Docker implementation uses
`github.com/docker/docker/client` to talk to the Docker socket.

```go
type Driver interface {
    Create(ctx context.Context, params CreateParams) (*Container, error)
    Start(ctx context.Context, id string) error
    Stop(ctx context.Context, id string, timeout *time.Duration) error
    Remove(ctx context.Context, id string, force bool) error
    List(ctx context.Context, params ListParams) ([]Container, error)
    Inspect(ctx context.Context, id string) (*ContainerDetail, error)
    Exec(ctx context.Context, id string, params ExecParams) (*ExecResult, error)
    Pull(ctx context.Context, image string) (*PullResult, error)
}
```

**Types:**

- `CreateParams` — image (required), name, env vars, port mappings, volumes,
  command override, auto-start flag
- `ListParams` — optional filters: state (running/stopped/all), name prefix,
  image, limit
- `Container` — ID, name, image, state (running/stopped/created), created
  timestamp
- `ContainerDetail` — everything in `Container` plus network settings, port
  mappings, mounts, resource limits, health status
- `ExecParams` — command (string slice), env vars, working directory
- `ExecResult` — stdout, stderr, exit code (mirrors `command.Result` pattern)
- `PullResult` — image ID, tag, size

The Docker implementation lives in
`internal/provider/container/runtime/docker/`. Future drivers (LXD, Podman)
implement the same interface.

### API Domain

The `container` domain nests under `/node/{hostname}`, consistent with how disk,
memory, and DNS are scoped to a node.

| Method   | Path                                    | Operation | Permission          |
| -------- | --------------------------------------- | --------- | ------------------- |
| `POST`   | `/node/{hostname}/container`            | Create    | `container:write`   |
| `GET`    | `/node/{hostname}/container`            | List      | `container:read`    |
| `GET`    | `/node/{hostname}/container/{id}`       | Inspect   | `container:read`    |
| `POST`   | `/node/{hostname}/container/{id}/start` | Start     | `container:write`   |
| `POST`   | `/node/{hostname}/container/{id}/stop`  | Stop      | `container:write`   |
| `DELETE` | `/node/{hostname}/container/{id}`       | Remove    | `container:write`   |
| `POST`   | `/node/{hostname}/container/{id}/exec`  | Exec      | `container:execute` |
| `POST`   | `/node/{hostname}/container/pull`       | Pull      | `container:write`   |

**Permissions:** `container:read`, `container:write`, and `container:execute`.
The `execute` permission is separate from lifecycle management, matching the
precedent set by `command:execute`.

**Role updates:**

- `admin` gains `container:read`, `container:write`, `container:execute`
- `write` gains `container:read`, `container:write`
- `read` gains `container:read`

**Path parameter `{id}`:** The `{id}` parameter accepts a Docker container ID
(hex string or short prefix) or container name. Unlike job and audit IDs which
use `format: uuid`, this parameter uses `type: string` with a `pattern` regex in
the OpenAPI spec to validate the allowed character set
(`[a-zA-Z0-9][a-zA-Z0-9_.-]*`). A custom validator tag is not needed — the
Docker SDK resolves both formats and returns a typed error if the container is
not found.

**Error responses:**

- `400` — validation failures on Create (missing image), Exec (missing command),
  Pull (missing image), and List (invalid filter values)
- `404` — Inspect, Start, Stop, Remove, Exec when the container ID/name does not
  resolve to an existing container
- `409` — Start on an already-running container, Stop on an already-stopped
  container
- `500` — Docker daemon errors, socket unreachable

**Request bodies:**

Create:

```json
{
  "image": "ubuntu:24.04",
  "name": "my-container",
  "command": ["/bin/bash"],
  "env": { "FOO": "bar" },
  "ports": [{ "host": 8080, "container": 80 }],
  "volumes": [{ "host": "/data", "container": "/mnt/data" }],
  "auto_start": true
}
```

Exec:

```json
{
  "command": ["useradd", "testuser"],
  "env": { "HOME": "/home/testuser" },
  "working_dir": "/root"
}
```

Stop (optional body):

```json
{
  "timeout": 10
}
```

Remove uses a query parameter:
`DELETE /node/{hostname}/container/{id}?force=true`. No request body. This is
consistent with the existing `DELETE` endpoints in the codebase which carry no
body.

**Pull is asynchronous.** `POST /node/{hostname}/container/pull` creates a job
and returns a job ID immediately. The pull proceeds in the background on the
agent. Clients poll `GET /job/{id}` for completion, consistent with how all
other state-changing operations work through the job system. Large image pulls
can take minutes; blocking the HTTP response would be unreliable.

**List query parameters:**

| Parameter | Type   | Description                                         |
| --------- | ------ | --------------------------------------------------- |
| `state`   | string | Filter by state: `running`, `stopped`, `all`        |
| `limit`   | int    | Maximum number of containers to return (default 50) |

### Agent Wiring

Container operations route through the existing job system. The job category is
`container`, and the operation field matches the endpoint (create, start, stop,
remove, list, inspect, exec, pull).

- `internal/agent/types.go` — add `containerProvider` field
- `internal/agent/factory.go` — create Docker driver and container service.
  Conditional on Docker socket availability: if the socket is not reachable, the
  provider is `nil` and container jobs return a descriptive error ("container
  runtime not available"). No startup failure.
- `internal/agent/processor.go` — add `container` case to category switch
- `internal/agent/processor_container.go` — dispatch by operation

### Server Wiring

Following the existing handler pattern:

- `internal/api/handler_container.go` — add `GetContainerHandler()` method on
  `Server`. Wraps the handler with `NewStrictHandler` + `scopeMiddleware`. No
  unauthenticated operations — all container endpoints require auth.
- `internal/api/handler.go` — call `GetContainerHandler()` in
  `RegisterHandlers()` and append results
- `internal/api/handler_public_test.go` — add `TestGetContainerHandler`
- `cmd/api_helpers.go` — add `GetContainerHandler()` to the `ServerManager`
  interface and call it in `registerAPIHandlers()`
- `cmd/api_server_start.go` — initialize the container handler with the Docker
  driver and pass it to `api.New()`

The `Server` struct does not store handler references as fields. Handlers are
constructed via `GetXxxHandler()` methods and returned as closures, consistent
with all existing domains.

### Provider Run Subcommand

A new `osapi provider run` CLI subcommand executes a single provider operation
as a standalone process with JSON I/O. This is the bridge that allows providers
to run inside containers.

```
osapi provider run <provider> <operation> --data '<json>'
```

This subcommand is hidden from `--help` output. It is a machine interface
consumed by the orchestrator's Docker exec layer, not a user-facing command.

**Behavior:**

1. Parse provider name, operation, and JSON data from flags
2. Look up the provider in a runtime registry
3. Deserialize JSON into the typed parameter struct
4. Instantiate the provider using the local platform factory
5. Call the operation method
6. Serialize result to JSON on stdout
7. Exit 0 on success, non-zero on failure (error message as JSON on stderr)

**Provider registry:**

```go
type Registration struct {
    Name       string
    Operations map[string]OperationSpec
}

type OperationSpec struct {
    NewParams func() any
    Run       func(ctx context.Context, params any) (any, error)
}
```

Each provider registers its operations and parameter types. The `provider run`
command looks up the provider and operation, creates the param type, unmarshals
JSON into it, calls `Run`, and marshals the result.

The SDK already has typed methods and parameter structs. The `provider run`
subcommand only needs JSON-in/JSON-out because the SDK handles type safety on
the caller side.

### Orchestrator DSL

The orchestrator DSL adds container targeting through `Docker()` and `In()`
methods on the `Plan` instance:

```go
p := orchestrator.NewPlan(client)
web := p.Docker("web-server", "ubuntu:24.04")

create := p.TaskFunc("create container", func(ctx context.Context, c *client.Client) (*orchestrator.Result, error) {
    return c.Container.Create(ctx, target, container.CreateParams{
        Image: "ubuntu:24.04",
        Name:  "web-server",
    })
})

p.In(web).TaskFunc("add user", func(ctx context.Context, c *client.Client) (*orchestrator.Result, error) {
    // c here is container-scoped — calls route through docker exec
    return c.User.Create(ctx, user.CreateParams{
        Username: "deploy",
        Shell:    "/bin/bash",
    })
}).DependsOn(create)
```

**How `In(target)` works:**

1. `p.Docker(name, image)` returns a `RuntimeTarget` handle that knows it is a
   Docker container with that name and image
2. At container creation, the orchestrator volume-mounts the host's `osapi`
   binary into the container
3. `p.In(target)` returns a scoped plan context where SDK client method calls
   are intercepted
4. Instead of HTTP requests to the API server, the scoped client serializes
   params to JSON and executes
   `docker exec <container> /osapi provider run <provider> <operation> --data '<json>'`
   through the Docker driver's `Exec` method
5. JSON stdout is deserialized back into the typed result struct

The developer works with the same typed SDK methods. The transport changes from
HTTP to Docker exec + provider run, but the interface is identical.

**`RuntimeTarget` interface** (for future LXD/Podman support):

```go
type RuntimeTarget interface {
    Name() string
    Runtime() string  // "docker", "lxd", "podman"
    ExecProvider(ctx context.Context, provider, operation string, data []byte) ([]byte, error)
}
```

`p.Docker()` and (future) `p.LXD()` return different implementations of the same
interface. `p.In()` accepts any `RuntimeTarget`.

### Configuration

No new configuration sections are needed in `osapi.yaml`. The Docker driver
connects to the Docker socket at its default path (`/var/run/docker.sock` on
Linux, the default Docker Desktop socket on macOS). If Docker is not available,
the provider is nil and container operations fail gracefully.

Future configuration (if needed) could add a `container` section for socket path
overrides, but this is out of scope for the initial implementation.

### Package Layout

```
internal/provider/container/
├── runtime/
│   ├── driver.go              # Driver interface + types
│   └── docker/
│       ├── docker.go          # Docker SDK implementation
│       └── docker_test.go
├── provider.go                # Service struct wrapping Driver
└── types.go                   # Domain types

internal/api/container/
├── gen/
│   ├── api.yaml               # OpenAPI spec
│   ├── cfg.yaml               # oapi-codegen config
│   └── generate.go            # go:generate directive
├── types.go                   # Domain struct, interfaces
├── container.go               # New(), interface check
├── container_create.go        # Create handler
├── container_list.go          # List handler
├── container_inspect.go       # Inspect handler
├── container_start.go         # Start handler
├── container_stop.go          # Stop handler
├── container_remove.go        # Remove handler
├── container_exec.go          # Exec handler
├── container_pull.go          # Pull handler
└── *_public_test.go           # Tests (unit + HTTP wiring + RBAC)

internal/api/
├── handler_container.go       # GetContainerHandler() method
├── handler.go                 # +RegisterHandlers() wiring
└── handler_public_test.go     # +TestGetContainerHandler

cmd/
├── provider_run.go            # provider run subcommand (hidden)
├── client_container.go        # parent command
├── client_container_create.go # CLI per endpoint
├── client_container_list.go
├── client_container_inspect.go
├── client_container_start.go
├── client_container_stop.go
├── client_container_remove.go
├── client_container_exec.go
└── client_container_pull.go

pkg/sdk/
├── client/container.go        # SDK service wrapper
└── orchestrator/
    ├── runtime_target.go      # RuntimeTarget interface
    ├── docker_target.go       # Docker implementation
    └── plan_in.go             # In() scoped context
```

### Documentation

- `docs/docs/sidebar/features/container-management.md` — feature page
- `docs/docs/sidebar/usage/cli/client/container/container.md` — parent CLI page
  with `<DocCardList />`
- `docs/docs/sidebar/usage/cli/client/container/{operation}.md` — one page per
  CLI subcommand
- `docs/docusaurus.config.ts` — add to Features navbar dropdown
- `docs/docs/sidebar/usage/configuration.md` — note that no new config sections
  are needed (Docker socket auto-detected)
- `docs/docs/sidebar/architecture/system-architecture.md` — add container
  endpoints to the endpoint tables

### Verification

```bash
just generate        # regenerate specs + code
go build ./...       # compiles
just go::unit        # tests pass
just go::vet         # lint passes
```

## Key Design Decisions

| Decision                      | Choice                          | Rationale                                                                 |
| ----------------------------- | ------------------------------- | ------------------------------------------------------------------------- |
| Runtime driver interface      | `runtime.Driver`                | Pluggable for Docker now, LXD/Podman later                                |
| Docker interaction            | Go SDK, not CLI                 | Typed responses, proper error handling, no output parsing                 |
| API nesting                   | Under `/node/{hostname}`        | Containers run on a node, consistent with existing API conventions        |
| Separate `execute` permission | `container:execute`             | Running commands in containers is a distinct privilege from lifecycle ops |
| Graceful absence              | Nil provider, descriptive error | Agents without Docker still work for all other providers                  |
| Provider run subcommand       | JSON-in/JSON-out, hidden        | Machine interface consumed by SDK; type safety lives in SDK               |
| Provider registry             | Runtime registration            | No code generation needed; new providers just register themselves         |
| Volume-mount binary           | Mount host osapi into container | Any base image works; no custom OSAPI container image required            |
| DSL via `In(target)`          | Scoped client, same SDK types   | Developers use identical API; only transport changes                      |
| `{id}` parameter              | String with pattern, not UUID   | Docker IDs are hex strings/names, not UUIDs                               |
| Remove force flag             | Query parameter, no body        | Consistent with existing DELETE endpoints                                 |
| Pull behavior                 | Async via job system            | Large pulls can take minutes; blocking HTTP is unreliable                 |
