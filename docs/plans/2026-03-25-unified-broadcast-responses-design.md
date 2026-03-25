# Unified Broadcast Response Design

## Goal

Standardize all node-targeted API responses so every operation supports
broadcast (`_all`, label selectors) and returns a uniform collection
response. Every result item carries `hostname` and `error` fields.
Single-target and broadcast operations return the same response shape.
Update CLAUDE.md so future providers follow the pattern from day one.

## Problem

18 of 29 node-targeted operations lack broadcast support. Docker (9
ops), File (3 ops), Cron mutations (3 ops), and Cron get have no
broadcast path — they silently route `_all` through the single-target
path and return one random agent's response. Users targeting a fleet
see results from one host with no indication others were skipped.

Response types are also inconsistent: some return collections, some
return flat objects. Some have `hostname` on result items, some don't.

## Design

### Uniform Response Shape

Every node-targeted operation returns:

```json
{
  "job_id": "550e8400-...",
  "results": [
    {
      "hostname": "web-01",
      "error": "",
      ...domain-specific fields...
    },
    {
      "hostname": "web-02",
      "error": "operation not supported on this OS family"
    }
  ]
}
```

- Single-target (`_any`, hostname): collection with 1 result
- Broadcast (`_all`, label): collection with N results
- Failed/skipped agents: result entry with `hostname` + `error`, empty
  domain fields

### Handler Pattern

Every handler follows:

```go
func (s *Handler) PostOperation(ctx, request) {
    validate(request)
    hostname := request.Hostname

    if job.IsBroadcastTarget(hostname) {
        return s.postOperationBroadcast(ctx, hostname, ...)
    }

    // Single target path.
    resp, err := s.JobClient.SingleMethod(ctx, hostname, ...)
    // Wrap in collection with 1 result.
    return collectionResponse(resp.JobID, []ResultItem{
        {Hostname: resp.Hostname, ...fields...},
    })
}

func (s *Handler) postOperationBroadcast(ctx, target, ...) {
    jobID, results, errs, err := s.JobClient.BroadcastMethod(...)

    var items []ResultItem
    for _, r := range results {
        items = append(items, ResultItem{Hostname: r.Hostname, ...})
    }
    for host, errMsg := range errs {
        items = append(items, ResultItem{Hostname: host, Error: errMsg})
    }

    return collectionResponse(jobID, items)
}
```

### Changes Per Domain

#### Docker (9 operations)

Response types already have `hostname` and `error` on result items.
Already return collections. Need:

1. Job client: 9 `*Broadcast` methods
2. Handlers: 9 `IsBroadcastTarget` checks + broadcast functions
3. No schema changes — response shapes are correct

Operations: create, list, inspect, start, stop, remove, exec, pull,
image-remove.

#### File (3 operations: deploy, undeploy, status)

Deploy and undeploy responses are flat (not collections). Need:

1. Convert `FileDeployResponse` and `FileUndeployResponse` to
   collection pattern with `job_id` + `results[]`
2. Add `error` field to deploy/undeploy result items
3. Job client: 3 `*Broadcast` methods
4. Handlers: 3 `IsBroadcastTarget` checks + broadcast functions
5. Update SDK types and CLI output

File status already has a collection-compatible shape.

#### Cron (4 operations: get, create, update, delete)

Cron list already has broadcast. The other 4 need it:

1. Add `hostname` to `CronCreateResponse`, `CronUpdateResponse`,
   `CronDeleteResponse`
2. Convert `CronEntryResponse` (get) to collection pattern
3. Convert mutation responses to collection pattern
4. Job client: 4 `*Broadcast` methods
5. Handlers: 4 `IsBroadcastTarget` checks + broadcast functions
6. Update SDK types and CLI output

### Job Client Broadcast Pattern

Every broadcast method follows the same pattern. For query operations:

```go
func (c *Client) QueryDockerListBroadcast(
    ctx context.Context,
    target string,
    ...,
) (string, map[string]*DockerListResult, map[string]string, error) {
    // Build request, call publishAndCollect, process responses.
}
```

For modify operations:

```go
func (c *Client) ModifyDockerCreateBroadcast(
    ctx context.Context,
    target string,
    ...,
) (string, map[string]*DockerResult, map[string]string, error) {
    // Build request, call publishAndCollect, process responses.
}
```

Return signature: `(jobID, resultsByHost, errorsByHost, error)`.
Matches the pattern used by existing broadcast methods like
`QueryNodeHostnameBroadcast`.

### SDK Changes

No new types needed for Docker (already have hostname + error).

For File and Cron mutations, add `Hostname` field to existing SDK
result types where missing. The SDK `Collection[T]` type already
handles the `job_id` + `results[]` envelope.

### CLI Changes

All CLI commands should use `BuildBroadcastTable` for collection
responses. Commands that currently use `PrintKV` for single results
(cron create/update/delete, file deploy/undeploy) switch to table
output showing HOSTNAME + STATUS + domain fields.

### CLAUDE.md Update

Add to "Adding a New API Domain" section:

> **Broadcast support (MANDATORY):** Every operation that targets a
> node (`/node/{hostname}/...`) MUST support broadcast. The handler
> checks `job.IsBroadcastTarget(hostname)` and routes to a broadcast
> function. The job client has both a single-target and `*Broadcast`
> method for each operation. All responses use a collection envelope
> with `job_id` + `results[]`. Every result item includes `hostname`
> and `error` fields. Single-target returns 1 result in the
> collection; broadcast returns N results.

### Scope

| Domain | Operations | Schema changes | New broadcast methods |
| ------ | ---------- | -------------- | -------------------- |
| Docker | 9 | none | 9 |
| File | 3 | deploy+undeploy → collection | 3 |
| Cron | 4 | get/create/update/delete → collection | 4 |
| Total | 16 | 5 schemas | 16 |

### What This Does NOT Change

- Node query operations (7) — already have full broadcast
- Command exec/shell (2) — already have full broadcast
- Network DNS/ping (3) — already have full broadcast
- Cron list — already has broadcast
- File upload/list/get/delete — not node-targeted (Object Store ops)
- Health, Audit, Job, Agent endpoints — not node-targeted
