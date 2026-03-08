---
sidebar_position: 14
---

# node.uptime.get

Get system uptime information.

## Usage

```go
task := plan.Task("get-uptime", &orchestrator.Op{
    Operation: "node.uptime.get",
    Target:    "_any",
})
```

## Parameters

None.

## Target

Accepts any valid target: `_any`, `_all`, a hostname, or a label selector
(`key:value`).

## Idempotency

**Read-only.** Never modifies state. Always returns `Changed: false`.

## Permissions

Requires `node:read` permission.

## Example

See
[`examples/sdk/orchestrator/operations/node-uptime.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/operations/node-uptime.go)
for a complete working example.
