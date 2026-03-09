# Unified Domain Endpoint Architecture

## Problem

The system has two parallel paths for executing operations:

1. **Domain endpoints** (`GET /node/{hostname}`, `PUT /network/...`, etc.) —
   synchronous, typed, used by CLI and external consumers. Internally create a
   job via `publishAndWait()` and return the full result in one HTTP response.

2. **Generic job endpoint** (`POST /job`) — asynchronous, untyped. The
   orchestrator creates raw jobs with operation strings and `map[string]any`
   params, then polls `GET /job/{id}` for results.

Both paths create the exact same NATS job under the hood. The orchestrator
bypasses the domain endpoints entirely, duplicating job creation, polling,
broadcast handling, and result extraction. Every new operation must be wired in
both paths. The generic job endpoint also bypasses the typed validation that
domain endpoints provide.

## Decision

Remove `POST /job`. All job creation goes through domain endpoints. The
orchestrator calls typed SDK client methods directly instead of creating raw
jobs.

## Architecture

### Before

```
DSL method → Op{Operation: "node.hostname.get"} → POST /job → poll GET /job/{id}
```

### After

```
DSL method → client.Node.Hostname(ctx, target) → GET /node/{target} → result
```

The orchestrator becomes a pure DAG runner. It does not know about jobs, NATS,
or polling. Domain endpoints handle job creation, agent communication, broadcast
collection, and waiting internally.

## What Gets Removed

### From the API

- `POST /job` endpoint and OpenAPI spec
- `PostJob` handler in `internal/api/job/`
- `CreateJob` in `internal/job/client/jobs.go`
- SDK `JobService.Create()` method

### From the orchestrator

- `Op` struct (operation string + params map)
- `executeOp()` — generic job creation + polling
- `pollJob()` — polling loop with exponential backoff
- `countExpectedAgents()` — broadcast agent counting
- `hostResultsFromResponses()` — response parsing
- `extractHostResults()` — fallback host result extraction
- `isCommandOp()` — command exit code checking
- `parseAgentDurations()` — agent timing extraction

## What Stays

### Job endpoints (observability and management)

- `GET /job` — list/filter jobs
- `GET /job/{id}` — get job details (debug via job ID from domain responses)
- `DELETE /job/{id}` — delete a job
- `POST /job/{id}/retry` — retry a failed job
- `GET /job/stats` — job statistics

### Domain endpoints

All existing domain endpoints remain. They are the single path for job creation.
Every domain endpoint already returns a job ID in its response for
debugging/audit correlation.

### Orchestrator DSL

User-facing DSL methods are unchanged:

- `o.NodeHostnameGet("web-01")`
- `o.NetworkDNSUpdate("web-01", params)`
- `o.TaskFunc("name", fn)`
- Guards, retry, error strategies, hooks — all unchanged

## What Changes

### Domain endpoint responses — enrichment

Domain endpoint responses need to include agent timing/duration data. This is
currently only available through the job polling path. Per-host results for
broadcast operations need the same metadata the orchestrator currently extracts.

### SDK client — typed responses

SDK service method responses carry job ID, changed, duration, and host results
uniformly. Each response has a consistent shape the orchestrator can work with.

### Orchestrator results — typed

`Result.Data` (`map[string]any`) is replaced with typed results per operation.
Guards and `When()` predicates work with typed accessors. This is a breaking
change.

### Broadcast handling — delegated

The orchestrator no longer manages broadcast polling or expected agent counts.
Domain endpoints handle `_all` and label selector targets internally via
`publishAndWait()` and return collected per-host results.

## Key Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Remove `POST /job` | Yes | Eliminates duplicate path, forces typed validation |
| Synchronous domain endpoints | Keep `publishAndWait` | Simple consumer DX, orchestrator uses goroutines for parallelism |
| Broadcast handling | Delegate to domain endpoints | One implementation, orchestrator stays simple |
| Result types | Typed per operation | Type safety over generic maps |
| Versioning | Breaking change, no v2 | Project is early enough |
| Job ID in responses | Keep | Essential for debugging when things break |
