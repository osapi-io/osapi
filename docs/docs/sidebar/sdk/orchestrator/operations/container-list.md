---
sidebar_position: 17
---

# docker.list.get

List containers on the target host, optionally filtered by state.

## Usage

```go
task := plan.TaskFunc("list-containers",
    func(ctx context.Context, c *client.Client) (*orchestrator.Result, error) {
        state := gen.GetNodeContainerDockerParamsStateRunning
        resp, err := c.Docker.List(ctx, "_any", &gen.GetNodeContainerDockerParams{
            State: &state,
        })
        if err != nil {
            return nil, err
        }
        return &orchestrator.Result{Changed: false}, nil
    },
)
```

## Parameters

| Param   | Type   | Required | Description                               |
| ------- | ------ | -------- | ----------------------------------------- |
| `state` | string | No       | Filter: `running`, `stopped`, or `all`    |
| `limit` | int    | No       | Maximum containers to return (default 50) |

## Target

Accepts any valid target: `_any`, `_all`, a hostname, or a label selector
(`key:value`).

## Idempotency

**Read-only.** Never modifies state. Always returns `Changed: false`.

## Permissions

Requires `docker:read` permission.

## Example

See
[`examples/sdk/orchestrator/features/container-targeting.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/features/container-targeting.go)
for a complete working example.
