---
sidebar_position: 1
---

# Docker

Manage the Docker containers and images of a host: list, get, create, start,
stop, remove, exec, pull and image remove. The methods call the
[docker](../../../reference/behaviors/docker/docker.md) and
[image](../../../reference/behaviors/image/image.md) behaviors.

## Methods

| Method                                     | Description                  |
| ------------------------------------------ | ---------------------------- |
| `List(ctx, hostname, state)`               | List the containers          |
| `Get(ctx, hostname, name)`                 | Get one container            |
| `Create(ctx, hostname, opts)`              | Create a container           |
| `Remove(ctx, hostname, name)`              | Remove a container           |
| `Start(ctx, hostname, name)`               | Start a container            |
| `Stop(ctx, hostname, name, timeout)`       | Stop a container             |
| `Exec(ctx, hostname, name, opts)`          | Run a command in a container |
| `ListImages(ctx, hostname)`                | List the images              |
| `Pull(ctx, hostname, image)`               | Pull an image                |
| `RemoveImage(ctx, hostname, image, force)` | Remove an image              |

`state` is `running`, `stopped`, `all`, or empty for all. `name` is the name of
a container, its ID, or the start of its ID.

## Request types

The service uses request types of the SDK. You do not import `gen`.

| Type               | Fields                                                              |
| ------------------ | ------------------------------------------------------------------- |
| `DockerCreateOpts` | Name, Image, Hostname, DNS, Command, Env, Ports, Volumes, AutoStart |
| `DockerExecOpts`   | Command, Env, WorkingDir                                            |

## Result types

| Type                      | Fields                                                                     |
| ------------------------- | -------------------------------------------------------------------------- |
| `DockerResult`            | ID, Name, Image, State, Created, Health, Ports, Mounts, Networks, EnvNames |
| `DockerActionResult`      | Name, ID, Changed                                                          |
| `DockerExecResult`        | Stdout, Stderr, ExitCode, Changed                                          |
| `DockerImageResult`       | ID, Repository, Tag, Size                                                  |
| `DockerPullResult`        | Image, ImageID, Size, Changed                                              |
| `DockerImageRemoveResult` | Image, Changed                                                             |

Each result also has Hostname, Status and Error. `EnvNames` has only the names
of the environment variables. osapi never returns their values.

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// Pull an image.
resp, err := c.Docker.Pull(ctx, "_any", "alpine:3.20")

// Create a container, and start it.
resp, err := c.Docker.Create(ctx, "_any", client.DockerCreateOpts{
    Name:      "web",
    Image:     "alpine:3.20",
    Command:   []string{"sleep", "300"},
    Ports:     []string{"8080:80"},
    AutoStart: true,
})

// List the running containers.
resp, err := c.Docker.List(ctx, "_any", "running")

// Run a command.
resp, err := c.Docker.Exec(ctx, "_any", "web", client.DockerExecOpts{
    Command: []string{"hostname"},
})

// Stop, and wait 30 seconds before docker kills it.
resp, err := c.Docker.Stop(ctx, "_any", "web", 30)

// Remove the container, also when it runs.
resp, err := c.Docker.Remove(ctx, "_any", "web")

// Remove the image.
resp, err := c.Docker.RemoveImage(ctx, "_any", "alpine:3.20", false)
```

## Examples

- [`examples/sdk/client/docker.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/docker.go)

## Permissions

| Operation                   | Permission          |
| --------------------------- | ------------------- |
| List, Get, ListImages       | `container:read`    |
| Create, Remove, RemoveImage | `container:write`   |
| Start, Stop, Pull           | `container:execute` |
| Exec                        | `container:exec`    |
