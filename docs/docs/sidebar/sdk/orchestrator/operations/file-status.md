---
sidebar_position: 5
---

# file.status.get

Check the deployment status of a file on the target agent. Reports whether the
file is in-sync, drifted, or missing compared to the expected state.

## Usage

```go
task := plan.TaskFunc("check-config",
    func(
        ctx context.Context,
        c *client.Client,
    ) (*orchestrator.Result, error) {
        resp, err := c.Node.FileStatus(
            ctx,
            "web-01",
            "/etc/nginx/nginx.conf",
        )
        if err != nil {
            return nil, err
        }

        return &orchestrator.Result{
            JobID:   resp.Data.JobID,
            Changed: resp.Data.Changed,
            Data:    orchestrator.StructToMap(resp.Data),
        }, nil
    },
)
```

## Parameters

| Param  | Type   | Required | Description                  |
| ------ | ------ | -------- | ---------------------------- |
| `path` | string | Yes      | File path to check on target |

## Target

Accepts any valid target: `_any`, `_all`, a hostname, or a label selector
(`key:value`).

## Idempotency

**Read-only.** Never modifies state. Always returns `Changed: false`.

## Permissions

Requires `file:read` permission.

## Example

See
[`examples/sdk/orchestrator/operations/file-status.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/operations/file-status.go)
for a complete working example.
