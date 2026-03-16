---
sidebar_position: 22
---

# docker.image-remove.execute

Remove a container image from the host.

## Usage

```go
task := plan.TaskFunc("remove-image",
    func(ctx context.Context, c *client.Client) (*orchestrator.Result, error) {
        resp, err := c.Docker.ImageRemove(ctx, "_any", "nginx:latest",
            &client.DockerImageRemoveParams{Force: true})
        if err != nil {
            return nil, err
        }

        return orchestrator.CollectionResult(resp.Data, resp.RawJSON(),
            func(r client.DockerActionResult) orchestrator.HostResult {
                return orchestrator.HostResult{
                    Hostname: r.Hostname,
                    Changed:  r.Changed,
                    Error:    r.Error,
                }
            },
        )
    },
)
```

## Parameters

| Param   | Type   | Required | Description                               |
| ------- | ------ | -------- | ----------------------------------------- |
| `image` | string | Yes      | Image name or ID (e.g., "nginx:latest")   |
| `force` | bool   | No       | Force removal even if the image is in use |

## Target

Accepts any valid target: `_any`, `_all`, a hostname, or a label selector
(`key:value`).

## Idempotency

**Idempotent.** Removing an image that does not exist returns success.

## Permissions

Requires `docker:write` permission.

## Example

See
[`examples/sdk/orchestrator/operations/docker-image-remove.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/operations/docker-image-remove.go)
for a complete working example.
