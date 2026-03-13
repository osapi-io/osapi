---
sidebar_position: 10
---

# Container Management

OSAPI can manage containers on target hosts through a pluggable runtime driver
architecture. Container operations run through the [job system](job-system.md),
so the API server never interacts with the container runtime directly -- agents
handle all execution.

## What It Does

| Operation | Description                                   |
| --------- | --------------------------------------------- |
| Create    | Create a new container from a specified image |
| List      | List containers, optionally filtered by state |
| Inspect   | Get detailed information about a container    |
| Start     | Start a stopped container                     |
| Stop      | Stop a running container                      |
| Remove    | Remove a container                            |
| Exec      | Execute a command inside a running container  |
| Pull      | Pull a container image to the host            |

**Create** builds a new container from a specified image with optional name,
environment variables, port mappings, and volume mounts. By default, containers
are started immediately after creation (`auto_start: true`).

**List** returns containers on the target host, filtered by state (`running`,
`stopped`, or `all`). Results include container ID, name, image, state, and
creation timestamp.

**Inspect** retrieves detailed information about a specific container, including
port mappings, volume mounts, environment variables, network settings, and
health check status.

**Start** and **Stop** control the lifecycle of existing containers. Stop
accepts an optional timeout (default 10 seconds) before forcibly killing the
container.

**Remove** deletes a container from the host. Use `--force` to remove a running
container without stopping it first.

**Exec** runs a command inside a running container and returns stdout, stderr,
and the exit code. Supports optional environment variables and a working
directory override.

**Pull** downloads a container image to the host. Returns the image ID, tag, and
size.

## How It Works

Container operations follow the same request flow as all OSAPI operations:

1. The CLI (or API client) posts a request to the API server.
2. The API server creates a job and publishes it to NATS.
3. An agent picks up the job, executes the container operation through the
   runtime driver, and writes the result back to NATS KV.
4. The API server collects the result and returns it to the client.

You can target a specific host, broadcast to all hosts with `_all`, or route by
label. See [CLI Reference](../usage/cli/client/container/container.mdx) for
usage and examples, or the
[API Reference](/gen/api/docker-management-api-docker-operations) for the REST
endpoints.

### Runtime Drivers

OSAPI uses a pluggable runtime driver interface for container operations. The
agent auto-detects which runtime is available on the host and selects the
appropriate driver.

| Driver | Detection                                          | Status  |
| ------ | -------------------------------------------------- | ------- |
| Docker | Checks for Docker socket at `/var/run/docker.sock` | Default |

The Docker driver communicates with the Docker daemon through its Unix socket.
No additional configuration is required -- if Docker is installed and the agent
has access to the socket, container operations work automatically.

## Configuration

Container management uses the general job infrastructure. No domain-specific
configuration sections are required in `osapi.yaml`.

The Docker runtime driver auto-detects the Docker socket at
`/var/run/docker.sock`. Ensure the agent process has read/write access to the
socket (typically by adding the agent user to the `docker` group).

See [Configuration](../usage/configuration.md) for NATS, agent, and
authentication settings.

## Permissions

| Endpoint                                            | Permission       |
| --------------------------------------------------- | ---------------- |
| `POST /node/{hostname}/container/docker` (create)   | `docker:write`   |
| `GET /node/{hostname}/container/docker` (list)      | `docker:read`    |
| `GET /node/{hostname}/container/docker/{id}`        | `docker:read`    |
| `POST /node/{hostname}/container/docker/{id}/start` | `docker:write`   |
| `POST /node/{hostname}/container/docker/{id}/stop`  | `docker:write`   |
| `DELETE /node/{hostname}/container/docker/{id}`     | `docker:write`   |
| `POST /node/{hostname}/container/docker/{id}/exec`  | `docker:execute` |
| `POST /node/{hostname}/container/docker/pull`       | `docker:write`   |

The `admin` role includes `docker:read`, `docker:write`, and
`docker:execute`. The `write` role includes `docker:read` and
`docker:write`. The `read` role includes only `docker:read`.

Container exec is a privileged operation similar to command execution. Only the
`admin` role includes `docker:execute` by default. Grant it to other roles or
tokens explicitly when needed:

```yaml
api:
  server:
    security:
      roles:
        docker-ops:
          permissions:
            - docker:read
            - docker:write
            - docker:execute
            - health:read
```

Or grant it directly on a token:

```bash
osapi token generate -r write -u user@example.com \
  -p docker:execute
```

## Orchestrator

The [orchestrator](../sdk/orchestrator/orchestrator.md) SDK can compose container
operations as a DAG using `TaskFunc`. Pull, create, exec, inspect, and cleanup
steps chain together with dependencies and guards:

```go
plan := orchestrator.NewPlan(client, orchestrator.OnError(orchestrator.Continue))

pull := plan.TaskFunc("pull-image",
    func(ctx context.Context, c *client.Client) (*orchestrator.Result, error) {
        _, err := c.Docker.Pull(ctx, "_any", gen.DockerPullRequest{
            Image: "nginx:alpine",
        })
        if err != nil {
            return nil, err
        }
        return &orchestrator.Result{Changed: true}, nil
    },
)

create := plan.TaskFunc("create-container",
    func(ctx context.Context, c *client.Client) (*orchestrator.Result, error) {
        autoStart := true
        _, err := c.Docker.Create(ctx, "_any", gen.DockerCreateRequest{
            Image:     "nginx:alpine",
            Name:      ptr("web-server"),
            AutoStart: &autoStart,
        })
        if err != nil {
            return nil, err
        }
        return &orchestrator.Result{Changed: true}, nil
    },
)
create.DependsOn(pull)

exec := plan.TaskFunc("check-config",
    func(ctx context.Context, c *client.Client) (*orchestrator.Result, error) {
        _, err := c.Docker.Exec(ctx, "_any", "web-server",
            gen.DockerExecRequest{Command: []string{"nginx", "-t"}})
        if err != nil {
            return nil, err
        }
        return &orchestrator.Result{Changed: true}, nil
    },
)
exec.DependsOn(create)
```

See
[`examples/sdk/orchestrator/features/container-targeting.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/features/container-targeting.go)
for a complete working example.

## Related

- [CLI Reference](../usage/cli/client/container/container.mdx) -- container
  management commands
- [API Reference](/gen/api/container-management-api-container-operations) --
  REST API documentation
- [Job System](job-system.md) -- how async job processing works
- [Authentication & RBAC](authentication.md) -- permissions and roles
- [Architecture](../architecture/architecture.md) -- system design overview
