---
sidebar_position: 11
---

# Introspection

Inspect and debug orchestration plans before running them.

## Explain

`Explain()` returns a human-readable representation of the execution plan
showing levels, parallelism, dependencies, and guards.

```go
plan := orchestrator.NewPlan(client)

health := plan.TaskFunc("check-health", healthFn)
hostname := plan.Task("get-hostname", &orchestrator.Op{
    Operation: "node.hostname.get",
    Target:    "_any",
})
hostname.DependsOn(health)

fmt.Println(plan.Explain())
```

Output:

```
Plan: 2 tasks, 2 levels

Level 0:
  check-health [fn]

Level 1:
  get-hostname [op] <- check-health
```

Tasks are annotated with their type (`fn` for functional tasks, `op` for
declarative operations), dependency edges (`<-`), and any active guards
(`only-if-changed`, `when`).

## Levels

`Levels()` returns the levelized DAG — tasks grouped into execution levels where
all tasks in a level can run concurrently. Returns an error if the plan fails
validation.

```go
levels, err := plan.Levels()
if err != nil {
    log.Fatal(err)
}

for i, level := range levels {
    fmt.Printf("Level %d: %d tasks\n", i, len(level))
}
```

## Validate

`Validate()` checks the plan for errors without executing it. It detects
duplicate task names and dependency cycles.

```go
if err := plan.Validate(); err != nil {
    log.Fatal(err) // e.g., "duplicate task name: "foo""
                   //        "cycle detected: "a" depends on "b""
}
```

`Run()` calls `Validate()` internally, so explicit validation is only needed
when you want to catch errors before execution — for example, during plan
construction or in tests.
