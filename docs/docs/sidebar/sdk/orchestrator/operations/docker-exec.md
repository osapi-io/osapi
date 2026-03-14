---
sidebar_position: 22
---

# docker.exec.execute

Execute a command inside a running container.

## Usage

```go
task := plan.TaskFunc("exec-in-container",
    func(ctx context.Context, c *client.Client) (*orchestrator.Result, error) {
        resp, err := c.Docker.Exec(ctx, "_any", "my-nginx", gen.DockerExecRequest{
            Command: []string{"nginx", "-t"},
        })
        if err != nil {
            return nil, err
        }
        r := resp.Data.Results[0]
        return &orchestrator.Result{
            Changed: true,
            Data: map[string]any{
                "exit_code": r.ExitCode,
                "stdout":    r.Stdout,
            },
        }, nil
    },
)
```

## Parameters

| Param         | Type     | Required | Description                            |
| ------------- | -------- | -------- | -------------------------------------- |
| `id`          | string   | Yes      | Container ID (short or full) or name   |
| `command`     | []string | Yes      | Command and arguments to execute       |
| `env`         | []string | No       | Environment variables (KEY=VALUE)      |
| `working_dir` | string   | No       | Working directory inside the container |

## Target

Accepts any valid target: `_any`, `_all`, a hostname, or a label selector
(`key:value`).

## Idempotency

**Not idempotent.** Always returns `Changed: true`. Use guards (`OnlyIfChanged`,
`When`) to control execution.

## Permissions

Requires `docker:execute` permission.

## Example

See
[`examples/sdk/orchestrator/features/container-targeting.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/features/container-targeting.go)
for a complete working example.
