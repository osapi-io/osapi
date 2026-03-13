---
sidebar_position: 18
---

# docker.inspect.get

Retrieve detailed information about a specific container.

## Usage

```go
task := plan.TaskFunc("inspect-container",
    func(ctx context.Context, c *client.Client) (*orchestrator.Result, error) {
        resp, err := c.Docker.Inspect(ctx, "_any", "my-nginx")
        if err != nil {
            return nil, err
        }
        r := resp.Data.Results[0]
        return &orchestrator.Result{
            Changed: false,
            Data:    map[string]any{"state": r.State, "image": r.Image},
        }, nil
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

**Read-only.** Never modifies state. Always returns `Changed: false`.

## Permissions

Requires `docker:read` permission.

## Example

See
[`examples/sdk/orchestrator/features/container-targeting.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/features/container-targeting.go)
for a complete working example.
