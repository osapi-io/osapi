# Container management

osapi manages the Docker containers and images of a host with two behaviors:
[docker](../../reference/behaviors/containers/docker/docker.md) for the
containers and [image](../../reference/behaviors/containers/image/image.md) for
the images. Each operation is a job, as each osapi operation is. The controller
does not connect to Docker. The agent on the host does the work.

## What it does

| Behavior | Method | Description                              |
| -------- | ------ | ---------------------------------------- |
| docker   | list   | List the containers, with a state filter |
| docker   | get    | Get one container by its name or its ID  |
| docker   | create | Create a container from an image         |
| docker   | delete | Remove a container                       |
| docker   | start  | Start a container                        |
| docker   | stop   | Stop a container                         |
| docker   | exec   | Run a command in a container             |
| image    | list   | List the images                          |
| image    | pull   | Pull an image                            |
| image    | delete | Remove an image                          |

**list** and **get** read each container through gohai's docker collector. A
container has its ID, name, image, state, creation time, health, ports, bind
mounts and networks. Its environment shows only the names of the variables.
osapi never returns their values.

**create** makes a container from an image, with a name, a hostname, DNS
servers, an environment, port mappings and bind mounts. It does not start the
container unless `auto_start` is set. The environment goes to docker on its
standard input, not in its arguments. A container of the same name that the host
has already is not made again, and reports `changed: false`.

**start** and **stop** report `changed: false` when the container is already in
that state. **stop** waits `timeout` seconds before docker kills the container.

**delete** removes a container, also one that runs.

**exec** runs a command in a container. It returns the standard output, the
standard error and the exit code. An exit code that is not 0 is a result, not a
failure.

**pull** pulls an image. It returns the image ID and its size in bytes. A pull
that gets no new image reports `changed: false`.

## How it works

Each method runs on the agent. Reads go through gohai, which runs
`docker version`, `docker ps` and `docker inspect`. Writes run the docker CLI
with sudo. A host where `docker version` does not answer reports `skipped`.

You can target one host, broadcast to all hosts with `_all`, or target a label.
Refer to the CLI reference of
[docker](../../usage/cli/client/node/docker/docker.mdx) and
[image](../../usage/cli/client/node/image/image.mdx) for usage and examples.

## Configuration

The behaviors need no section of their own in `osapi.yaml`. The agent's user
must be able to run `docker` with sudo.

**exec is off by default.** The agent refuses each exec until you set
`agent.command.enabled: true`, the same setting as command execution. The other
methods do not need it.

## Permissions

The [docker](../../reference/behaviors/containers/docker/docker.md) and
[image](../../reference/behaviors/containers/image/image.md) reference pages
give the permission of each method.

## Related

- [docker behavior](../../reference/behaviors/containers/docker/docker.md) and
  [image behavior](../../reference/behaviors/containers/image/image.md): each
  method, its input, its output and its examples
- CLI reference: the [docker](../../usage/cli/client/node/docker/docker.mdx) and
  [image](../../usage/cli/client/node/image/image.mdx) commands
- [Job System](../platform/job-system.md): how asynchronous jobs work
- [Authentication & RBAC](../platform/authentication.md): permissions and roles
