---
sidebar_position: 10
---

# node.hostname.get

Get the system hostname and agent labels.

## Usage

```go
task := plan.TaskFunc("get-hostname",
    func(
        ctx context.Context,
        c *client.Client,
    ) (*orchestrator.Result, error) {
        resp, err := c.Node.Hostname(ctx, "_any")
        if err != nil {
            return nil, err
        }

        return orchestrator.CollectionResult(
            resp.Data,
            func(r client.HostnameResult) orchestrator.HostResult {
                return orchestrator.HostResult{
                    Hostname: r.Hostname,
                    Changed:  r.Changed,
                    Error:    r.Error,
                }
            },
        ), nil
    },
)
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
[`examples/sdk/orchestrator/operations/node-hostname.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/operations/node-hostname.go)
for a complete working example.

## node.hostname.update

Set the system hostname on the target node.

### Usage

```go
task := plan.TaskFunc("set-hostname",
    func(
        ctx context.Context,
        c *client.Client,
    ) (*orchestrator.Result, error) {
        resp, err := c.Node.SetHostname(ctx, "web-01", "new-hostname")
        if err != nil {
            return nil, err
        }

        return orchestrator.CollectionResult(
            resp.Data,
            func(r client.HostnameUpdateResult) orchestrator.HostResult {
                return orchestrator.HostResult{
                    Hostname: r.Hostname,
                    Changed:  r.Changed,
                    Error:    r.Error,
                }
            },
        ), nil
    },
)
```

### Parameters

| Parameter | Type   | Required | Description          |
| --------- | ------ | -------- | -------------------- |
| target    | string | yes      | Target host or group |
| name      | string | yes      | New hostname to set  |

### Idempotency

**Idempotent.** If the hostname is already set to the requested value, returns
`Changed: false` without modifying the system.

### Permissions

Requires `node:write` permission.
