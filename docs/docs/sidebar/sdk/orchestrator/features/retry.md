---
sidebar_position: 7
---

# Retry

Automatically retry failed tasks before marking them as failed.

## Usage

```go
getLoad := plan.Task("get-load", &orchestrator.Op{
    Operation: "node.load.get",
    Target:    "_any",
})
getLoad.OnError(orchestrator.Retry(3))
```

The task will be retried up to 3 times. Use the `OnRetry` hook to observe retry
attempts:

```go
hooks := orchestrator.Hooks{
    OnRetry: func(task *orchestrator.Task, attempt int, err error) {
        fmt.Printf("[retry] %s  attempt=%d error=%q\n",
            task.Name(), attempt, err)
    },
}
```

## Example

See
[`examples/sdk/orchestrator/features/retry.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/features/retry.go)
for a complete working example.
