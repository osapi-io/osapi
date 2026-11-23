---
sidebar_position: 1
---

# Docker

Manage the Docker containers and images of a host: list, get, create, start,
stop, remove, exec, pull and image remove. The methods call the
[docker](../../../reference/behaviors/docker/docker.md) and
[image](../../../reference/behaviors/image/image.md) behaviors.

## Methods

| Method                                                                      | Returns                                           | Permission       | Description                     |
| --------------------------------------------------------------------------- | ------------------------------------------------- | ---------------- | ------------------------------- |
| `c.ContainerDocker.List(ctx, hostname, params ContainerDockerListParams)`   | `Response[Collection[ContainerDockerEntry]]`      | `docker:read`    | List the containers of the host |
| `c.ContainerDocker.Create(ctx, hostname, input ContainerDockerCreateInput)` | `Response[Collection[ContainerDockerResult]]`     | `docker:write`   | Create a container              |
| `c.ContainerDocker.Get(ctx, hostname, name)`                                | `Response[Collection[ContainerDockerEntry]]`      | `docker:read`    | Get one container               |
| `c.ContainerDocker.Delete(ctx, hostname, name)`                             | `Response[Collection[ContainerDockerResult]]`     | `docker:write`   | Remove a container              |
| `c.ContainerDocker.Exec(ctx, hostname, input ContainerDockerExecInput)`     | `Response[Collection[ContainerDockerExecResult]]` | `docker:exec`    | Run a command in a container    |
| `c.ContainerDocker.Start(ctx, hostname, name)`                              | `Response[Collection[ContainerDockerResult]]`     | `docker:execute` | Start a container               |
| `c.ContainerDocker.Stop(ctx, hostname, input ContainerDockerStopInput)`     | `Response[Collection[ContainerDockerResult]]`     | `docker:execute` | Stop a container                |
| `c.ContainerImage.List(ctx, hostname)`                                      | `Response[Collection[ContainerImageEntry]]`       | `image:read`     | List the images of the host     |
| `c.ContainerImage.Pull(ctx, hostname, input ContainerImagePullInput)`       | `Response[Collection[ContainerImagePullResult]]`  | `image:execute`  | Pull an image                   |
| `c.ContainerImage.Delete(ctx, hostname, input ContainerImageDeleteInput)`   | `Response[Collection[ContainerImageResult]]`      | `image:write`    | Remove an image                 |

## Types

The fields of each type are the properties of its schema in the API's spec.

### ContainerDockerListParams

| Field   | Type     | Description                                                              |
| ------- | -------- | ------------------------------------------------------------------------ |
| `State` | `string` | Only the containers in this state: running, stopped or all, the default. |

### ContainerDockerEntry

| Field      | Type       | Description                                                                            |
| ---------- | ---------- | -------------------------------------------------------------------------------------- |
| `Created`  | `string`   | When it was created, RFC 3339.                                                         |
| `EnvNames` | `[]string` | The names of its environment variables; osapi never reads their values.                |
| `Error`    | `string`   | Error message if the agent failed.                                                     |
| `Health`   | `string`   | The state of its health check: healthy, unhealthy or starting; empty without one.      |
| `Hostname` | `string`   | Hostname of the agent that produced this result.                                       |
| `ID`       | `string`   | The container's full ID.                                                               |
| `Image`    | `string`   | The image it runs.                                                                     |
| `Mounts`   | `[]string` | The mounts, source:destination, with :ro when read only.                               |
| `Name`     | `string`   | The container's name.                                                                  |
| `Networks` | `[]string` | The networks, with the container's address and gateway on each.                        |
| `Ports`    | `[]string` | The published ports, host:port->container port.                                        |
| `State`    | `string`   | Running, exited, created, paused or restarting.                                        |
| `Status`   | `string`   | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |
| `Summary`  | `string`   | What docker says of its state, such as Up 2 minutes.                                   |

### ContainerDockerCreateInput

| Field               | Type                | Description                                                                            |
| ------------------- | ------------------- | -------------------------------------------------------------------------------------- |
| `AutoStart`         | `bool`              | Start the container once it is created.                                                |
| `Command`           | `[]string`          | The command, in place of the image's own.                                              |
| `ContainerHostname` | `string`            | The hostname inside the container.                                                     |
| `DNS`               | `[]string`          | The DNS servers of the container.                                                      |
| `Env`               | `map[string]string` | The environment variables of the container.                                            |
| `Image`             | `string`            | The image to run, such as alpine:3.20; docker pulls it when the host does not have it. |
| `Name`              | `string`            | The container's name.                                                                  |
| `Ports`             | `[]string`          | The published ports, each host:container, such as 8080:80.                             |
| `Volumes`           | `[]string`          | The bind mounts, each host:container, such as /srv/data:/data.                         |

### ContainerDockerResult

| Field      | Type     | Description                                                                            |
| ---------- | -------- | -------------------------------------------------------------------------------------- |
| `Changed`  | `bool`   | Whether the host changed.                                                              |
| `Error`    | `string` | Error message if the agent failed.                                                     |
| `Hostname` | `string` | Hostname of the agent that produced this result.                                       |
| `ID`       | `string` | The container's full ID, for create.                                                   |
| `Name`     | `string` | The container.                                                                         |
| `Status`   | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |

### ContainerDockerExecInput

| Field        | Type                | Description                                       |
| ------------ | ------------------- | ------------------------------------------------- |
| `Name`       | `string`            | The name of the path: the container's name or ID. |
| `Command`    | `[]string`          | The command and its arguments.                    |
| `Env`        | `map[string]string` | Environment variables for the command.            |
| `WorkingDir` | `string`            | The directory the command runs in.                |

### ContainerDockerExecResult

| Field      | Type     | Description                                                                            |
| ---------- | -------- | -------------------------------------------------------------------------------------- |
| `Changed`  | `bool`   | Always true: osapi cannot tell what the command changed.                               |
| `Error`    | `string` | Error message if the agent failed.                                                     |
| `ExitCode` | `int`    | The command's exit code.                                                               |
| `Hostname` | `string` | Hostname of the agent that produced this result.                                       |
| `Status`   | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |
| `Stderr`   | `string` | What the command wrote to standard error.                                              |
| `Stdout`   | `string` | What the command wrote to standard output.                                             |

### ContainerDockerStopInput

| Field     | Type     | Description                                                      |
| --------- | -------- | ---------------------------------------------------------------- |
| `Name`    | `string` | The name of the path: the container's name or ID.                |
| `Timeout` | `int`    | Seconds to wait before docker kills it; docker's default when 0. |

### ContainerImageEntry

| Field        | Type     | Description                                                                            |
| ------------ | -------- | -------------------------------------------------------------------------------------- |
| `Error`      | `string` | Error message if the agent failed.                                                     |
| `Hostname`   | `string` | Hostname of the agent that produced this result.                                       |
| `ID`         | `string` | The image's ID.                                                                        |
| `Repository` | `string` | Its repository, such as alpine.                                                        |
| `Size`       | `string` | Its size, as docker prints it, such as 7.8MB.                                          |
| `Status`     | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |
| `Tag`        | `string` | Its tag, such as 3.20.                                                                 |

### ContainerImagePullInput

| Field   | Type     | Description                     |
| ------- | -------- | ------------------------------- |
| `Image` | `string` | The image, such as alpine:3.20. |

### ContainerImagePullResult

| Field      | Type     | Description                                                                            |
| ---------- | -------- | -------------------------------------------------------------------------------------- |
| `Changed`  | `bool`   | Whether the pull got a new image: false when the host had it already.                  |
| `Error`    | `string` | Error message if the agent failed.                                                     |
| `Hostname` | `string` | Hostname of the agent that produced this result.                                       |
| `Image`    | `string` | The image.                                                                             |
| `ImageID`  | `string` | The ID of the image that the host has now.                                             |
| `Size`     | `int64`  | Its size in bytes.                                                                     |
| `Status`   | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |

### ContainerImageDeleteInput

| Field   | Type     | Description                                                   |
| ------- | -------- | ------------------------------------------------------------- |
| `Image` | `string` | The image of the path: the image, by its reference or its ID. |
| `Force` | `bool`   | Remove it also when a container uses it.                      |

### ContainerImageResult

| Field      | Type     | Description                                                                            |
| ---------- | -------- | -------------------------------------------------------------------------------------- |
| `Changed`  | `bool`   | Whether the host changed.                                                              |
| `Error`    | `string` | Error message if the agent failed.                                                     |
| `Hostname` | `string` | Hostname of the agent that produced this result.                                       |
| `Image`    | `string` | The image.                                                                             |
| `Status`   | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// Pull an image.
resp, err := c.ContainerImage.Pull(ctx, "_any",
    client.ContainerImagePullInput{Image: "alpine:3.20"})

// Create a container, and start it.
resp, err := c.ContainerDocker.Create(ctx, "_any", client.ContainerDockerCreateInput{
    Name:      "web",
    Image:     "alpine:3.20",
    Command:   []string{"sleep", "300"},
    Ports:     []string{"8080:80"},
    AutoStart: true,
})

// List the running containers.
resp, err := c.ContainerDocker.List(ctx, "_any",
    client.ContainerDockerListParams{State: "running"})

// Run a command.
resp, err := c.ContainerDocker.Exec(ctx, "_any", client.ContainerDockerExecInput{
    Name: "web",
    Command: []string{"hostname"},
})

// Stop, and wait 30 seconds before docker kills it.
resp, err := c.ContainerDocker.Stop(ctx, "_any", client.ContainerDockerStopInput{
    Name: "web",
    Timeout: 30,
})

// Remove the container, also when it runs.
resp, err := c.ContainerDocker.Delete(ctx, "_any", "web")

// Remove the image.
resp, err := c.ContainerImage.Delete(ctx, "_any", client.ContainerImageDeleteInput{Image: "alpine:3.20"})
```

## Examples

- [`examples/sdk/client/docker.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/docker.go)

## Permissions

The Permission column of [Methods](#methods) gives the permission of each
method.
