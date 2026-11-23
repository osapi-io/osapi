---
sidebar_position: 10
---

# Container management

osapi manages the Docker containers and images of a host with two behaviors:
[docker](../reference/behaviors/docker/docker.md) for the containers and
[image](../reference/behaviors/image/image.md) for the images. Each operation is
a job, as each osapi operation is. The controller does not connect to Docker.
The agent on the host does the work.

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
Refer to the [CLI reference](../usage/cli/client/node/container/container.mdx)
for usage and examples.

## Configuration

The behaviors need no section of their own in `osapi.yaml`. The agent's user
must be able to run `docker` with sudo.

**exec is off by default.** The agent refuses each exec until you set
`agent.command.enabled: true`, the same setting as command execution. The other
methods do not need it.

## Permissions

| Permission       | Methods                  | Roles                             |
| ---------------- | ------------------------ | --------------------------------- |
| `docker:read`    | docker list and get      | admin, operator, automation, read |
| `docker:write`   | docker create and delete | admin, automation                 |
| `docker:execute` | docker start and stop    | admin, operator                   |
| `docker:exec`    | docker exec              | admin                             |
| `image:read`     | image list               | admin, operator, automation, read |
| `image:write`    | image delete             | admin, automation                 |
| `image:execute`  | image pull               | admin, operator                   |

The permission of a behavior in the `container` group has the name of the
behavior, not the name of the group. Each action of a behavior has the
permission `<name>:execute`, such as `docker:execute` and `image:execute`. The
`operator` role has it. docker exec runs a command that the caller chooses. Thus
it has a permission of its own, `docker:exec`, and only the `admin` role has it.
To give it to a different role or token:

```yaml
controller:
  api:
    security:
      roles:
        container-ops:
          permissions:
            - docker:read
            - docker:write
            - docker:execute
            - docker:exec
            - image:read
            - image:write
            - image:execute
            - health:read
```

```bash
osapi token generate -r operator -u user@example.com \
  -p docker:exec
```

## Related

- [docker behavior](../reference/behaviors/docker/docker.md) and
  [image behavior](../reference/behaviors/image/image.md): each method, its
  input, its output and its examples
- [CLI reference](../usage/cli/client/node/container/container.mdx): the
  container commands
- [Job System](job-system.md): how asynchronous jobs work
- [Authentication & RBAC](authentication.md): permissions and roles
