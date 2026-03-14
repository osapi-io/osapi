---
sidebar_position: 2
---

# command.exec.execute

Execute a command directly on the target node.

## Usage

```go
task := plan.TaskFunc("install-nginx",
    func(
        ctx context.Context,
        c *client.Client,
    ) (*orchestrator.Result, error) {
        resp, err := c.Node.Exec(ctx, client.ExecRequest{
            Target:  "_all",
            Command: "apt",
            Args:    []string{"install", "-y", "nginx"},
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

| Param     | Type     | Required | Description            |
| --------- | -------- | -------- | ---------------------- |
| `command` | string   | Yes      | The command to execute |
| `args`    | []string | No       | Command arguments      |

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
[`examples/sdk/orchestrator/operations/command-exec.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/operations/command-exec.go)
for a complete working example.
