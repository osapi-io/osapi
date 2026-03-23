---
sidebar_position: 3
---

# DockerService

Docker container lifecycle management — create, list, inspect, start, stop,
remove, exec, and pull operations.

## Methods

| Method                                 | Description                      |
| -------------------------------------- | -------------------------------- |
| `Create(ctx, hostname, opts)`          | Create a new container           |
| `List(ctx, hostname, params)`          | List containers                  |
| `Inspect(ctx, hostname, id)`           | Get detailed container info      |
| `Start(ctx, hostname, id)`             | Start a stopped container        |
| `Stop(ctx, hostname, id, opts)`        | Stop a running container         |
| `Remove(ctx, hostname, id, p)`         | Remove a container               |
| `Exec(ctx, hostname, id, opts)`        | Execute a command in a container |
| `Pull(ctx, hostname, opts)`            | Pull a container image           |
| `ImageRemove(ctx, hostname, image, p)` | Remove a container image         |

## Request Types

The Docker service uses SDK-defined request types. Consumers never need to
import `gen`.

| Type                      | Fields                                               |
| ------------------------- | ---------------------------------------------------- |
| `DockerCreateOpts`        | Image, Name, Command, Env, Ports, Volumes, AutoStart |
| `DockerStopOpts`          | Timeout                                              |
| `DockerListParams`        | State, Limit                                         |
| `DockerRemoveParams`      | Force                                                |
| `DockerPullOpts`          | Image                                                |
| `DockerExecOpts`          | Command                                              |
| `DockerImageRemoveParams` | Force                                                |

## Usage

```go
import "github.com/retr0h/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// Pull an image
resp, err := c.Docker.Pull(ctx, "_any", client.DockerPullOpts{
    Image: "nginx:latest",
})

// Create a container
autoStart := true
resp, err := c.Docker.Create(ctx, "_any", client.DockerCreateOpts{
    Image:     "nginx:latest",
    Name:      "web",
    Ports:     []string{"8080:80"},
    AutoStart: &autoStart,
})

// List running containers
resp, err := c.Docker.List(ctx, "_any", &client.DockerListParams{
    State: "running",
})

// Execute a command
resp, err := c.Docker.Exec(ctx, "_any", "web", client.DockerExecOpts{
    Command: []string{"hostname"},
})

// Stop with timeout
resp, err := c.Docker.Stop(ctx, "_any", "web", client.DockerStopOpts{
    Timeout: 30,
})

// Force remove container
resp, err := c.Docker.Remove(ctx, "_any", "web", &client.DockerRemoveParams{
    Force: true,
})

// Remove an image
resp, err := c.Docker.ImageRemove(ctx, "_any", "nginx:latest",
    &client.DockerImageRemoveParams{Force: true},
)
```

## Examples

- [`examples/sdk/client/container.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/client/container.go)

## Permissions

| Operation   | Permission       |
| ----------- | ---------------- |
| Create      | `docker:write`   |
| List        | `docker:read`    |
| Inspect     | `docker:read`    |
| Start       | `docker:write`   |
| Stop        | `docker:write`   |
| Remove      | `docker:write`   |
| Exec        | `docker:execute` |
| Pull        | `docker:write`   |
| ImageRemove | `docker:write`   |
