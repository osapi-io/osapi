---
sidebar_position: 5
---

# Only If Changed

Skip a task unless at least one of its dependencies reported a change.

## Usage

Call `OnlyIfChanged()` on a task to make it conditional on upstream changes. If
every dependency completed with `Changed: false`, the task is skipped and the
`OnSkip` hook fires with reason `"no dependencies changed"`.

```go
deploy := plan.TaskFunc("deploy-config",
    func(
        ctx context.Context,
        c *client.Client,
    ) (*orchestrator.Result, error) {
        resp, err := c.Node.FileDeploy(ctx, client.FileDeployOpts{
            Target:      "_any",
            ObjectName:  "app.conf",
            Path:        "/etc/app/app.conf",
            ContentType: "text/plain",
        })
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

restart := plan.TaskFunc("restart-service",
    func(
        _ context.Context,
        _ *client.Client,
    ) (*orchestrator.Result, error) {
        fmt.Println("Restarting service...")
        return &orchestrator.Result{Changed: true}, nil
    },
)
restart.DependsOn(deploy)
restart.OnlyIfChanged()
```

In this example, `restart-service` only runs if the deploy step actually changed
the file on disk. If the file was already up to date, the restart is skipped.

## How It Works

The orchestrator checks the `Changed` field of every dependency's `Result`. If
at least one dependency has `Changed: true`, the task runs normally. If all
dependencies have `Changed: false`, the task is skipped with status
`StatusSkipped`.

This is equivalent to a `When` guard that checks dependency results, but
provided as a convenience for the common "only act on changes" pattern.

## Example

See
[`examples/sdk/orchestrator/features/only-if-changed.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/features/only-if-changed.go)
for a complete working example.
