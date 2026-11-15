---
sidebar_position: 10
---

# Container management

osapi manages containers on target hosts through runtime drivers that you can
replace. Container operations use the [job system](job-system.md). Thus the API
server does not connect to the container runtime. The agents do all the work.

## What it does

| Operation | Description                                    |
| --------- | ---------------------------------------------- |
| Create    | Make a new container from an image             |
| List      | List containers, with an optional state filter |
| Inspect   | Get the details of a container                 |
| Start     | Start a stopped container                      |
| Stop      | Stop a running container                       |
| Remove    | Remove a container                             |
| Exec      | Run a command in a running container           |
| Pull      | Pull a container image to the host             |

**Create** makes a new container from an image. As options, it takes a name,
environment variables, port mappings and volume mounts. By default, the
container starts immediately after osapi makes it (`auto_start: true`).

**List** returns the containers on the target host, with a state filter
(`running`, `stopped` or `all`). Each result has the container ID, name, image,
state and creation time.

**Inspect** gets the details of one container: the port mappings, volume mounts,
environment variables, network settings and health check status.

**Start** and **Stop** control the lifecycle of a container. Stop accepts an
optional timeout (10 seconds by default). After the timeout, it kills the
container.

**Remove** deletes a container from the host. Use `--force` to remove a running
container. Then you do not need to stop it first.

**Exec** runs a command in a running container. It returns stdout, stderr and
the exit code. As options, it takes environment variables and a working
directory.

**Pull** downloads a container image to the host. It returns the image ID, tag
and size.

## How it works

Container operations use the same request flow as all osapi operations:

1. The CLI or an API client sends a request to the API server.
1. The API server makes a job and publishes it to NATS.
1. An agent receives the job. It runs the container operation through the
   runtime driver, and writes the result back to NATS KV.
1. The API server collects the result and returns it to the client.

You can target one host, broadcast to all hosts with `_all`, or target a label.
Refer to the [CLI Reference](../usage/cli/client/node/container/container.mdx)
for usage and examples, or to the
[API Reference](/gen/api/docker-management-api-docker-operations) for the REST
endpoints.

### Runtime drivers

osapi uses a runtime driver interface for container operations. The agent finds
the runtime on the host automatically and selects the correct driver.

| Driver | Detection                                             | Status  |
| ------ | ----------------------------------------------------- | ------- |
| Docker | Looks for the Docker socket at `/var/run/docker.sock` | Default |

The Docker driver connects to the Docker daemon through its Unix socket. You do
not need more configuration. If Docker is installed and the agent can use the
socket, container operations work automatically.

## Configuration

Container management uses the general job system. It needs no section of its own
in `osapi.yaml`.

The Docker runtime driver finds the Docker socket at `/var/run/docker.sock`
automatically. Make sure that the agent process can read and write the socket.
Usually, you add the agent user to the `docker` group.

Refer to [Configuration](../usage/configuration.md) for the NATS, agent and
authentication settings.

## Permissions

Container operations need `docker:*` permissions. The `admin` role has
`docker:read`, `docker:write` and `docker:execute`. The `write` role has
`docker:read` and `docker:write`. The `read` role has only `docker:read`. Refer
to the [API reference](/category/api) for the permission of each endpoint.

Container exec is a privileged operation, like command execution. **It is off by
default.** The agent refuses each exec job until you set
`agent.command.enabled: true`, the same setting as command execution. The other
container operations do not need it. By default, only the `admin` role has
`docker:execute`. When necessary, give it to other roles or tokens:

```yaml
controller:
  api:
    security:
      roles:
        docker-ops:
          permissions:
            - docker:read
            - docker:write
            - docker:execute
            - health:read
```

Or give it directly on a token:

```bash
osapi token generate -r write -u user@example.com \
  -p docker:execute
```

## Orchestrator

You can put container operations together as a DAG with [osapi-orchestrator][].
Refer to the
[Docker operations](https://github.com/osapi-io/osapi-orchestrator/blob/main/docs/operations/docker/)
and the
[docker.go example](https://github.com/osapi-io/osapi-orchestrator/blob/main/examples/operations/docker.go)
for a full container lifecycle: pull, create, start, exec, inspect, stop, remove
and image-remove.

[osapi-orchestrator]: https://github.com/osapi-io/osapi-orchestrator

## Related

- [CLI Reference](../usage/cli/client/node/container/container.mdx): the
  container commands
- [API Reference](/gen/api/docker-management-api-docker-operations): the REST
  API
- [Job System](job-system.md): how asynchronous jobs work
- [Authentication & RBAC](authentication.md): permissions and roles
- [Architecture](../architecture/architecture.md): an overview of the system
  design
