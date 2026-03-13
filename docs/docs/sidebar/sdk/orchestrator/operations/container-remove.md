---
sidebar_position: 21
---

# docker.remove.execute

Remove a container from the host.

## Usage

```go
task := plan.TaskFunc("remove-container",
    func(ctx context.Context, c *client.Client) (*orchestrator.Result, error) {
        force := true
        _, err := c.Docker.Remove(ctx, "_any", "my-nginx",
            &gen.DeleteNodeContainerDockerByIDParams{Force: &force})
        if err != nil {
            return nil, err
        }
        return &orchestrator.Result{Changed: true}, nil
    },
)
```

## Parameters

| Param   | Type   | Required | Description                          |
| ------- | ------ | -------- | ------------------------------------ |
| `id`    | string | Yes      | Container ID (short or full) or name |
| `force` | bool   | No       | Force-remove a running container     |

## Target

Accepts any valid target: `_any`, `_all`, a hostname, or a label selector
(`key:value`).

## Idempotency

**Not idempotent.** Returns 404 if the container does not exist.

## Permissions

Requires `docker:write` permission.

## Example

See
[`examples/sdk/orchestrator/features/container-targeting.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/features/container-targeting.go)
for a complete working example.
