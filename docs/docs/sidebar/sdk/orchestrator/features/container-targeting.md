---
sidebar_position: 14
---

# Container Targeting

Create containers and run provider operations inside them using the same typed
SDK methods. The transport changes from HTTP to `docker exec` + `provider run`,
but the interface is identical.

## Setup

Provide a `WithDockerExecFn` option when creating the plan. The exec function
calls the Docker SDK's `ContainerExecCreate` / `ContainerExecAttach` APIs to run
commands inside containers.

```go
plan := orchestrator.NewPlan(client, orchestrator.WithDockerExecFn(execFn))
```

## Docker Target

`Plan.Docker()` creates a `DockerTarget` bound to a container name and image. It
implements the `RuntimeTarget` interface:

```go
web := plan.Docker("web-server", "nginx:alpine")
```

`RuntimeTarget` is pluggable — Docker is the first implementation. Future
runtimes (LXD, Podman) implement the same interface.

## Scoped Plans

`Plan.In()` returns a `ScopedPlan` that routes provider operations through the
target's `ExecProvider` method:

```go
plan.In(web).TaskFunc("run-inside",
    func(ctx context.Context, c *client.Client) (*orchestrator.Result, error) {
        // This executes inside the container via:
        //   docker exec web-server /osapi provider run <provider> <op> --data '<json>'
        return &orchestrator.Result{Changed: true}, nil
    },
)
```

The `ScopedPlan` supports `TaskFunc` and `TaskFuncWithResults`, with the same
dependency, guard, and error strategy features as the parent plan.

## Full Example

A typical workflow creates a container, runs operations inside it, then cleans
up:

```go
plan := orchestrator.NewPlan(client, orchestrator.WithDockerExecFn(execFn))
web := plan.Docker("my-app", "ubuntu:24.04")

pull := plan.TaskFunc("pull-image",
    func(ctx context.Context, c *client.Client) (*orchestrator.Result, error) {
        resp, err := c.Container.Pull(ctx, "_any", gen.ContainerPullRequest{
            Image: "ubuntu:24.04",
        })
        if err != nil {
            return nil, err
        }
        return &orchestrator.Result{Changed: true, Data: map[string]any{
            "image_id": resp.Data.Results[0].ImageID,
        }}, nil
    },
)

create := plan.TaskFunc("create-container",
    func(ctx context.Context, c *client.Client) (*orchestrator.Result, error) {
        autoStart := true
        resp, err := c.Container.Create(ctx, "_any", gen.ContainerCreateRequest{
            Image:     "ubuntu:24.04",
            Name:      ptr("my-app"),
            AutoStart: &autoStart,
        })
        if err != nil {
            return nil, err
        }
        return &orchestrator.Result{Changed: true, Data: map[string]any{
            "container_id": resp.Data.Results[0].ID,
        }}, nil
    },
)
create.DependsOn(pull)

// Run a command inside the container
checkOS := plan.In(web).TaskFunc("check-os",
    func(ctx context.Context, c *client.Client) (*orchestrator.Result, error) {
        resp, err := c.Container.Exec(ctx, "_any", "my-app", gen.ContainerExecRequest{
            Command: []string{"cat", "/etc/os-release"},
        })
        if err != nil {
            return nil, err
        }
        return &orchestrator.Result{
            Changed: false,
            Data:    map[string]any{"stdout": resp.Data.Results[0].Stdout},
        }, nil
    },
)
checkOS.DependsOn(create)

cleanup := plan.TaskFunc("remove-container",
    func(ctx context.Context, c *client.Client) (*orchestrator.Result, error) {
        force := true
        _, err := c.Container.Remove(ctx, "_any", "my-app",
            &gen.DeleteNodeContainerByIDParams{Force: &force})
        if err != nil {
            return nil, err
        }
        return &orchestrator.Result{Changed: true}, nil
    },
)
cleanup.DependsOn(checkOS)

report, err := plan.Run(context.Background())
```

## RuntimeTarget Interface

```go
type RuntimeTarget interface {
    Name() string
    Runtime() string  // "docker", "lxd", "podman"
    ExecProvider(ctx context.Context, provider, operation string, data []byte) ([]byte, error)
}
```

`DockerTarget` implements this by running
`docker exec <container> /osapi provider run <provider> <operation> --data '<json>'`.
The host's `osapi` binary is volume-mounted into the container at creation time.

## Example

See
[`examples/sdk/orchestrator/features/container-targeting.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/features/container-targeting.go)
for a complete working example.
