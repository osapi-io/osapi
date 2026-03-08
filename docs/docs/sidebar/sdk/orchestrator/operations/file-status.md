---
sidebar_position: 5
---

# file.status.get

Check the deployment status of a file on the target agent. Reports whether the
file is in-sync, drifted, or missing compared to the expected state.

## Usage

```go
task := plan.Task("check-config", &orchestrator.Op{
    Operation: "file.status.get",
    Target:    "web-01",
    Params: map[string]any{
        "path": "/etc/nginx/nginx.conf",
    },
})
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
