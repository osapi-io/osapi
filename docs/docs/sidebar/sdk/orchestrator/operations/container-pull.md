---
sidebar_position: 23
---

# container.pull.execute

Pull a container image to the host.

## Usage

```go
task := plan.TaskFunc("pull-image",
    func(ctx context.Context, c *client.Client) (*orchestrator.Result, error) {
        resp, err := c.Container.Pull(ctx, "_any", gen.ContainerPullRequest{
            Image: "nginx:alpine",
        })
        if err != nil {
            return nil, err
        }
        r := resp.Data.Results[0]
        return &orchestrator.Result{
            Changed: true,
            Data:    map[string]any{"image_id": r.ImageID, "tag": r.Tag},
        }, nil
    },
)
```

## Parameters

| Param   | Type   | Required | Description             |
| ------- | ------ | -------- | ----------------------- |
| `image` | string | Yes      | Image reference to pull |

## Target

Accepts any valid target: `_any`, `_all`, a hostname, or a label selector
(`key:value`).

## Idempotency

**Not idempotent.** Always pulls the image. Pull is asynchronous through the job
system -- the API returns a job ID immediately and the agent pulls in the
background.

## Permissions

Requires `container:write` permission.

## Example

See
[`examples/sdk/orchestrator/features/container-targeting.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/features/container-targeting.go)
for a complete working example.
