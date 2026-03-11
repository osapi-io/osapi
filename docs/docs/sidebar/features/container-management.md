---
sidebar_position: 10
---

# Container Management

OSAPI can manage containers on target hosts through a pluggable runtime
driver architecture. Container operations run through the
[job system](job-system.md), so the API server never interacts with the
container runtime directly -- agents handle all execution.

## What It Does

| Operation | Description                                  |
| --------- | -------------------------------------------- |
| Create    | Create a new container from a specified image |
| List      | List containers, optionally filtered by state |
| Inspect   | Get detailed information about a container    |
| Start     | Start a stopped container                     |
| Stop      | Stop a running container                      |
| Remove    | Remove a container                            |
| Exec      | Execute a command inside a running container  |
| Pull      | Pull a container image to the host            |

**Create** builds a new container from a specified image with optional
name, environment variables, port mappings, and volume mounts. By
default, containers are started immediately after creation
(`auto_start: true`).

**List** returns containers on the target host, filtered by state
(`running`, `stopped`, or `all`). Results include container ID, name,
image, state, and creation timestamp.

**Inspect** retrieves detailed information about a specific container,
including port mappings, volume mounts, environment variables, network
settings, and health check status.

**Start** and **Stop** control the lifecycle of existing containers.
Stop accepts an optional timeout (default 10 seconds) before forcibly
killing the container.

**Remove** deletes a container from the host. Use `--force` to remove a
running container without stopping it first.

**Exec** runs a command inside a running container and returns stdout,
stderr, and the exit code. Supports optional environment variables and
a working directory override.

**Pull** downloads a container image to the host. Returns the image ID,
tag, and size.

## How It Works

Container operations follow the same request flow as all OSAPI
operations:

1. The CLI (or API client) posts a request to the API server.
2. The API server creates a job and publishes it to NATS.
3. An agent picks up the job, executes the container operation through
   the runtime driver, and writes the result back to NATS KV.
4. The API server collects the result and returns it to the client.

You can target a specific host, broadcast to all hosts with `_all`, or
route by label. See
[CLI Reference](../usage/cli/client/container/container.mdx) for usage
and examples, or the
[API Reference](/gen/api/node-container) for the REST endpoints.

### Runtime Drivers

OSAPI uses a pluggable runtime driver interface for container
operations. The agent auto-detects which runtime is available on the
host and selects the appropriate driver.

| Driver | Detection                                   | Status  |
| ------ | ------------------------------------------- | ------- |
| Docker | Checks for Docker socket at `/var/run/docker.sock` | Default |

The Docker driver communicates with the Docker daemon through its Unix
socket. No additional configuration is required -- if Docker is
installed and the agent has access to the socket, container operations
work automatically.

## Configuration

Container management uses the general job infrastructure. No
domain-specific configuration sections are required in `osapi.yaml`.

The Docker runtime driver auto-detects the Docker socket at
`/var/run/docker.sock`. Ensure the agent process has read/write access
to the socket (typically by adding the agent user to the `docker`
group).

See [Configuration](../usage/configuration.md) for NATS, agent, and
authentication settings.

## Permissions

| Endpoint                                    | Permission          |
| ------------------------------------------- | ------------------- |
| `POST /node/{hostname}/container` (create)  | `container:write`   |
| `GET /node/{hostname}/container` (list)      | `container:read`    |
| `GET /node/{hostname}/container/{id}`        | `container:read`    |
| `POST /node/{hostname}/container/{id}/start` | `container:write`   |
| `POST /node/{hostname}/container/{id}/stop`  | `container:write`   |
| `DELETE /node/{hostname}/container/{id}`     | `container:write`   |
| `POST /node/{hostname}/container/{id}/exec`  | `container:execute` |
| `POST /node/{hostname}/container/pull`       | `container:write`   |

The `admin` role includes `container:read`, `container:write`, and
`container:execute`. The `write` role includes `container:read` and
`container:write`. The `read` role includes only `container:read`.

Container exec is a privileged operation similar to command execution.
Only the `admin` role includes `container:execute` by default. Grant it
to other roles or tokens explicitly when needed:

```yaml
api:
  server:
    security:
      roles:
        container-ops:
          permissions:
            - container:read
            - container:write
            - container:execute
            - health:read
```

Or grant it directly on a token:

```bash
osapi token generate -r write -u user@example.com \
  -p container:execute
```

## Related

- [CLI Reference](../usage/cli/client/container/container.mdx) --
  container management commands
- [API Reference](/gen/api/node-container) -- REST API documentation
- [Job System](job-system.md) -- how async job processing works
- [Authentication & RBAC](authentication.md) -- permissions and roles
- [Architecture](../architecture/architecture.md) -- system design
  overview
