---
sidebar_position: 15
---

# node.load.get

Get load averages (1-minute, 5-minute, and 15-minute).

## Usage

```go
task := plan.Task("get-load", &orchestrator.Op{
    Operation: "node.load.get",
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
[`examples/sdk/orchestrator/operations/node-load.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/operations/node-load.go)
for a complete working example.
