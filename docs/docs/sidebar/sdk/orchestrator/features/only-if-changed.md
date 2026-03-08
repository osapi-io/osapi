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
deploy := plan.Task("deploy-config", &orchestrator.Op{
    Operation: "file.deploy.execute",
    Target:    "_any",
    Params: map[string]any{
        "object_name":  "app.conf",
        "path":         "/etc/app/app.conf",
        "content_type": "text/plain",
    },
})

restart := plan.TaskFunc("restart-service",
    func(_ context.Context, _ *client.Client) (*orchestrator.Result, error) {
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
