---
sidebar_position: 20
---

# container.stop.execute

Stop a running container.

## Usage

```go
task := plan.TaskFunc("stop-container",
    func(ctx context.Context, c *client.Client) (*orchestrator.Result, error) {
        timeout := 30
        _, err := c.Container.Stop(ctx, "_any", "my-nginx", gen.ContainerStopRequest{
            Timeout: &timeout,
        })
        if err != nil {
            return nil, err
        }
        return &orchestrator.Result{Changed: true}, nil
    },
)
```

## Parameters

| Param     | Type   | Required | Description                            |
| --------- | ------ | -------- | -------------------------------------- |
| `id`      | string | Yes      | Container ID (short or full) or name   |
| `timeout` | int    | No       | Seconds before force kill (default 10) |

## Target

Accepts any valid target: `_any`, `_all`, a hostname, or a label selector
(`key:value`).

## Idempotency

**Not idempotent.** Returns 409 if the container is already stopped.

## Permissions

Requires `container:write` permission.

## Example

See
[`examples/sdk/orchestrator/features/container-targeting.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/features/container-targeting.go)
for a complete working example.
