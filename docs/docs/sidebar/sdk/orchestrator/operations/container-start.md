---
sidebar_position: 19
---

# docker.start.execute

Start a stopped container.

## Usage

```go
task := plan.TaskFunc("start-container",
    func(ctx context.Context, c *client.Client) (*orchestrator.Result, error) {
        _, err := c.Docker.Start(ctx, "_any", "my-nginx")
        if err != nil {
            return nil, err
        }
        return &orchestrator.Result{Changed: true}, nil
    },
)
```

## Parameters

| Param | Type   | Required | Description                          |
| ----- | ------ | -------- | ------------------------------------ |
| `id`  | string | Yes      | Container ID (short or full) or name |

## Target

Accepts any valid target: `_any`, `_all`, a hostname, or a label selector
(`key:value`).

## Idempotency

**Not idempotent.** Returns 409 if the container is already running. Use guards
to check state first.

## Permissions

Requires `docker:write` permission.

## Example

See
[`examples/sdk/orchestrator/features/container-targeting.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/features/container-targeting.go)
for a complete working example.
