---
sidebar_position: 14
---

# Container Targeting

Orchestrate container lifecycle operations -- pull, create, exec, inspect,
stop, and remove -- as a DAG of `TaskFunc` steps using the standard SDK client.

## Setup

Container operations use the same `client.Container` service as any other SDK
call. No special plan options are needed:

```go
plan := orchestrator.NewPlan(client, orchestrator.OnError(orchestrator.Continue))
```

## Building the DAG

Chain container operations with `DependsOn` to enforce ordering. Independent
operations at the same level run in parallel:

```go
pull := plan.TaskFunc("pull-image",
    func(ctx context.Context, c *client.Client) (*orchestrator.Result, error) {
        _, err := c.Container.Pull(ctx, "_any", gen.ContainerPullRequest{
            Image: "ubuntu:24.04",
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
        resp, err := c.Docker.Create(ctx, "_any", gen.DockerCreateRequest{
            Image:     "ubuntu:24.04",
            Name:      ptr("my-app"),
            AutoStart: &autoStart,
            Command:   &[]string{"sleep", "600"},
        })
        if err != nil {
            return nil, err
        }
        r := resp.Data.Results[0]
        return &orchestrator.Result{
            Changed: true,
            Data:    map[string]any{"id": r.ID},
        }, nil
    },
)
create.DependsOn(pull)
```

## Exec

Execute commands inside running containers. Multiple exec tasks that depend on
the same create step run in parallel:

```go
execHostname := plan.TaskFunc("exec-hostname",
    func(ctx context.Context, c *client.Client) (*orchestrator.Result, error) {
        resp, err := c.Container.Exec(ctx, "_any", "my-app",
            gen.ContainerExecRequest{Command: []string{"hostname"}})
        if err != nil {
            return nil, err
        }
        r := resp.Data.Results[0]
        return &orchestrator.Result{
            Changed: true,
            Data:    map[string]any{"stdout": r.Stdout},
        }, nil
    },
)
execHostname.DependsOn(create)

execUname := plan.TaskFunc("exec-uname",
    func(ctx context.Context, c *client.Client) (*orchestrator.Result, error) {
        resp, err := c.Container.Exec(ctx, "_any", "my-app",
            gen.ContainerExecRequest{Command: []string{"uname", "-a"}})
        if err != nil {
            return nil, err
        }
        r := resp.Data.Results[0]
        return &orchestrator.Result{
            Changed: true,
            Data:    map[string]any{"stdout": r.Stdout},
        }, nil
    },
)
execUname.DependsOn(create)
```

## Cleanup

Use a cleanup task that depends on all operational tasks to ensure the container
is removed even when some tasks fail (with `OnError(Continue)`):

```go
cleanup := plan.TaskFunc("cleanup",
    func(ctx context.Context, c *client.Client) (*orchestrator.Result, error) {
        force := true
        _, err := c.Docker.Remove(ctx, "_any", "my-app",
            &gen.DeleteNodeContainerDockerByIDParams{Force: &force})
        if err != nil {
            return nil, err
        }
        return &orchestrator.Result{Changed: true}, nil
    },
)
cleanup.DependsOn(execHostname, execUname)
```

## Example

See
[`examples/sdk/orchestrator/features/container-targeting.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/features/container-targeting.go)
for a complete working example with hooks, error handling, and a deliberately
failing task.
