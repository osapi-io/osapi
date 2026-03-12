---
sidebar_position: 16
---

# container.create.execute

Create a new container from a specified image.

## Usage

```go
task := plan.TaskFunc("create-container",
    func(ctx context.Context, c *client.Client) (*orchestrator.Result, error) {
        autoStart := true
        resp, err := c.Container.Create(ctx, "_any", gen.ContainerCreateRequest{
            Image:     "nginx:alpine",
            Name:      ptr("my-nginx"),
            AutoStart: &autoStart,
        })
        if err != nil {
            return nil, err
        }
        r := resp.Data.Results[0]
        return &orchestrator.Result{
            Changed: true,
            Data:    map[string]any{"id": r.ID, "name": r.Name},
        }, nil
    },
)
```

## Parameters

| Param        | Type     | Required | Description                             |
| ------------ | -------- | -------- | --------------------------------------- |
| `image`      | string   | Yes      | Container image reference               |
| `name`       | string   | No       | Optional container name                 |
| `env`        | []string | No       | Environment variables (KEY=VALUE)       |
| `ports`      | []string | No       | Port mappings (host:container)          |
| `volumes`    | []string | No       | Volume mounts (host:container)          |
| `auto_start` | bool     | No       | Start immediately after creation (true) |

## Target

Accepts any valid target: `_any`, `_all`, a hostname, or a label selector
(`key:value`).

## Idempotency

**Not idempotent.** Always creates a new container. Use guards to prevent
duplicate creation.

## Permissions

Requires `container:write` permission.

## Example

See
[`examples/sdk/orchestrator/features/container-targeting.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/features/container-targeting.go)
for a complete working example.
