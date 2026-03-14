---
sidebar_position: 3
---

# command.shell.execute

Execute a shell command string on the target node. The command is passed to
`/bin/sh -c`.

## Usage

```go
task := plan.TaskFunc("check-disk-space",
    func(
        ctx context.Context,
        c *client.Client,
    ) (*orchestrator.Result, error) {
        resp, err := c.Node.Shell(ctx, client.ShellRequest{
            Target:  "_any",
            Command: "df -h / | tail -1 | awk '{print $5}'",
        })
        if err != nil {
            return nil, err
        }

        return orchestrator.CollectionResult(
            resp.Data,
            func(r client.CommandResult) orchestrator.HostResult {
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

| Param     | Type   | Required | Description                   |
| --------- | ------ | -------- | ----------------------------- |
| `command` | string | Yes      | The full shell command string |

## Target

Accepts any valid target: `_any`, `_all`, a hostname, or a label selector
(`key:value`).

## Idempotency

**Not idempotent.** Always returns `Changed: true`. Use guards (`OnlyIfChanged`,
`When`) to control execution.

## Permissions

Requires `command:execute` permission.

## Example

See
[`examples/sdk/orchestrator/operations/command-shell.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/operations/command-shell.go)
for a complete working example.
