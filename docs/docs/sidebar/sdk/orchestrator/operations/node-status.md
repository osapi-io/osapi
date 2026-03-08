---
sidebar_position: 11
---

# node.status.get

Get comprehensive node status including hostname, OS information, uptime, disk
usage, memory statistics, and load averages.

## Usage

```go
task := plan.Task("get-status", &orchestrator.Op{
    Operation: "node.status.get",
    Target:    "web-01",
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
[`examples/sdk/orchestrator/operations/node-status.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/operations/node-status.go)
for a complete working example.
