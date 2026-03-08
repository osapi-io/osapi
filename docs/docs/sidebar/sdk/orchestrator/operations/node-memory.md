---
sidebar_position: 13
---

# node.memory.get

Get memory statistics including total, available, used, and swap.

## Usage

```go
task := plan.Task("get-memory", &orchestrator.Op{
    Operation: "node.memory.get",
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
[`examples/sdk/orchestrator/operations/node-memory.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/operations/node-memory.go)
for a complete working example.
