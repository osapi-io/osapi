---
sidebar_position: 7
---

# Retry

Automatically retry failed tasks before marking them as failed. Retries can be
immediate or use exponential backoff to avoid overwhelming a recovering service.

## Usage

### Immediate Retry

Retry up to 3 times with no delay between attempts:

```go
getLoad := plan.Task("get-load", &orchestrator.Op{
    Operation: "node.load.get",
    Target:    "_any",
})
getLoad.OnError(orchestrator.Retry(3))
```

### Retry with Exponential Backoff

Add exponential backoff between retry attempts using `WithRetryBackoff`. The
delay doubles on each attempt, clamped to the max interval:

```go
// Retry 3 times: ~1s, ~2s, ~4s between attempts.
getLoad.OnError(orchestrator.Retry(3,
    orchestrator.WithRetryBackoff(1*time.Second, 30*time.Second),
))
```

### Plan-Level Retry with Backoff

Set as the default strategy for all tasks in the plan:

```go
plan := orchestrator.NewPlan(client,
    orchestrator.OnError(orchestrator.Retry(3,
        orchestrator.WithRetryBackoff(500*time.Millisecond, 10*time.Second),
    )),
)
```

## OnRetry Hook

Use the `OnRetry` hook to observe retry attempts:

```go
hooks := orchestrator.Hooks{
    OnRetry: func(task *orchestrator.Task, attempt int, err error) {
        fmt.Printf("[retry] %s  attempt=%d error=%q\n",
            task.Name(), attempt, err)
    },
}
```

## Transient Poll Errors

Job polling automatically retries transient HTTP errors (404, 500) with
exponential backoff. This handles the race where the agent hasn't written
results yet when the SDK first polls. Non-transient errors (401, 403, network
failures) fail immediately.

## Example

See
[`examples/sdk/orchestrator/features/retry.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/features/retry.go)
for a complete working example.
