---
title: Standardize server/worker CLI lifecycle and context handling
status: done
created: 2026-02-15
updated: 2026-02-18
---

## Objective

Standardize how the CLI powers all server-type commands (API server, job
worker, and any future NATS-based servers) so they operate consistently with
context creation, signal handling, start/stop, and graceful termination.

Currently the API server and job worker use different lifecycle patterns:

- **API server** (`cmd/api_server_start.go`): Non-blocking `Start()`, manual
  `<-ctx.Done()` block, explicit `Stop(shutdownCtx)` with 10s timeout, manual
  NATS connection cleanup in the command.
- **Job worker** (`cmd/job_worker_start.go`): Blocking `Start(ctx)` that runs
  until context cancellation, no explicit Stop method, internal WaitGroup
  cleanup, no shutdown timeout.

## Current Inconsistencies

1. **Start semantics**: API server `Start()` is non-blocking (spawns goroutine),
   worker `Start(ctx)` is blocking (runs until ctx.Done).
2. **Shutdown method**: API server requires explicit `Stop(ctx)` call, worker
   relies solely on context cancellation.
3. **Timeout handling**: API server creates a 10s shutdown timeout context,
   worker has no shutdown timeout at all.
4. **Resource cleanup**: API server manually closes NATS connection in the
   command, worker handles cleanup internally.
5. **Signal flow**: Different patterns make it harder to reason about shutdown
   ordering.

## Desired State

All server-type commands should follow a single, consistent pattern:

- CLI creates context from `cmd.Context()` (already provided by root.go signal
  handler)
- All servers/workers implement the same lifecycle interface
- Graceful shutdown with configurable timeout applied uniformly
- Resource cleanup (NATS connections, KV buckets) handled consistently
- A shared `RunServer` or similar helper in `cmd/` that encapsulates the
  start-block-stop pattern so each command just wires up its dependencies

### Proposed Interface

```go
// Lifecycle represents a long-running server or worker.
type Lifecycle interface {
    Start() error
    Stop(ctx context.Context) error
}
```

Or alternatively, the blocking context pattern:

```go
// Runner represents a long-running process driven by context.
type Runner interface {
    Run(ctx context.Context) error  // blocks until ctx.Done, handles cleanup
}
```

### Proposed Helper

```go
func runServer(ctx context.Context, server Lifecycle, cleanupFns ...func()) {
    if err := server.Start(); err != nil {
        logFatal("failed to start", err)
    }
    <-ctx.Done()
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    server.Stop(shutdownCtx)
    for _, fn := range cleanupFns {
        fn()
    }
}
```

## Scope

- `cmd/api_server_start.go`
- `cmd/job_worker_start.go`
- `internal/api/server.go` (ServerManager interface + implementation)
- `internal/job/worker/server.go` (Worker lifecycle)
- Any future server commands

## Notes

- Root context and signal handling in `cmd/root.go` is already good — signals
  trigger `cancel()` which propagates to all commands via `ExecuteContext(ctx)`.
- Chose non-blocking `Start()` / `Stop(ctx)` over blocking `Run(ctx)` because
  Echo and the embedded NATS server are inherently non-blocking.
- Shutdown timeout is hardcoded at 10s in `runServer`.

## Outcome

All three server commands now use the same `Lifecycle` interface and shared
`runServer` helper in `cmd/lifecycle.go`. The worker was refactored from
blocking `Start(ctx)` to non-blocking `Start()` + `Stop(ctx)` with a
`WaitGroup`-based graceful shutdown. NATS connection cleanup was added to
the worker command (previously missing). The `ServerManager` interface was
replaced by the shared `Lifecycle` interface.
