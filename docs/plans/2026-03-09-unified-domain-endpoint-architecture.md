# Unified Domain Endpoint Architecture Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to
> implement this plan task-by-task.

**Goal:** Remove the generic `POST /job` endpoint and refactor the orchestrator
to call typed SDK client methods through domain endpoints instead of creating
raw jobs.

**Architecture:** The orchestrator's 13 `Op`-based DSL methods are converted to
`TaskFunc` calls that invoke SDK client methods directly (matching the pattern
already used by HealthCheck, FileUpload, FileChanged, AgentList, AgentGet). The
`Op` struct, `executeOp`, `pollJob`, and all broadcast polling logic are removed
from the SDK orchestrator engine. Domain endpoints handle job creation, waiting,
and broadcast collection internally via `publishAndWait`.

**Tech Stack:** Go 1.25, osapi (monorepo), osapi-orchestrator (DSL layer)

---

## Repo Layout

| Repo               | Path                    | Role                               |
| ------------------ | ----------------------- | ---------------------------------- |
| osapi              | `pkg/sdk/orchestrator/` | SDK orchestrator engine            |
| osapi              | `pkg/sdk/client/`       | SDK client (typed service methods) |
| osapi              | `internal/api/job/`     | Job API endpoints                  |
| osapi-orchestrator | `pkg/orchestrator/`     | User-facing DSL                    |

Changes span both repos. osapi changes land first (SDK engine + API), then
osapi-orchestrator (DSL layer).

---

### Task 1: Remove `POST /job` endpoint from API

Remove the job creation endpoint. List, get, delete, retry, and stats remain.

**Files:**

- Modify: `osapi/internal/api/job/gen/api.yaml` — remove `post` under `/job`
- Modify: `osapi/internal/api/job/job_create.go` — delete file
- Modify: `osapi/internal/api/job/job_create_public_test.go` — delete file
- Modify: `osapi/internal/api/handler_job.go` — remove unauthenticated
  operations entry for PostJob if present
- Modify: `osapi/internal/job/client/jobs.go` — remove `CreateJob` method
- Modify: `osapi/internal/job/client/jobs_public_test.go` — remove CreateJob
  test cases
- Modify: `osapi/pkg/sdk/client/job.go` — remove `Create` method
- Modify: `osapi/pkg/sdk/client/job_types.go` — remove `JobCreated` type if only
  used by Create

**Step 1:** Remove `post` operation from `/job` path in the OpenAPI spec.

**Step 2:** Run `just generate` to regenerate server code.

**Step 3:** Delete `job_create.go` and `job_create_public_test.go`.

**Step 4:** Remove `CreateJob` from `internal/job/client/jobs.go` and its tests.

**Step 5:** Remove `Create` from `pkg/sdk/client/job.go`. Remove `JobCreated`
type if unused after this change (check if `Retry` also uses it — if so, keep
it).

**Step 6:** Run `go build ./...` and fix any compile errors.

**Step 7:** Run `just go::unit` and fix any test failures.

**Step 8:** Commit.

```bash
git commit -m "feat(api)!: remove POST /job endpoint

All job creation now goes through typed domain endpoints.
Job list, get, delete, retry, and stats endpoints remain."
```

---

### Task 2: Remove `Op`, `executeOp`, `pollJob` from SDK orchestrator engine

Strip the generic job creation and polling machinery from the SDK orchestrator.
After this, the SDK engine only supports `TaskFunc` (and `TaskFuncWithResults`).

**Files:**

- Modify: `osapi/pkg/sdk/orchestrator/plan.go` — remove `Task(name, op)` method
  that accepts `*Op`
- Modify: `osapi/pkg/sdk/orchestrator/runner.go` — remove `executeOp`,
  `pollJob`, `countExpectedAgents`, `hostResultsFromResponses`,
  `extractHostResults`, `isCommandOp`, `parseAgentDurations`, `isTransient`,
  `IsBroadcastTarget`
- Modify: `osapi/pkg/sdk/orchestrator/result.go` — remove `Op` struct, remove
  `agentDurations` internal field from `Result`
- Modify: `osapi/pkg/sdk/orchestrator/runner_test.go` — remove tests for removed
  functions (`TestBackoffDelay`, `TestIsTransient`,
  `TestRunTaskStoresResultForAllPaths` if it uses Op)
- Modify: `osapi/pkg/sdk/orchestrator/runner_broadcast_test.go` — remove tests
  for removed functions (`TestIsBroadcastTarget`, `TestExtractHostResults`,
  `TestHostResultsFromResponses`, `TestParseAgentDurations`,
  `TestCountExpectedAgents`, `TestIsCommandOp`, `TestExecuteOp*`,
  `TestPollJob*`)
- Modify: `osapi/pkg/sdk/orchestrator/plan_public_test.go` — update
  `TestRunOpTask` and `TestRunOpTaskErrors` to use TaskFunc instead of Op
- Modify: `osapi/pkg/sdk/orchestrator/options.go` — check if backoff imports are
  still needed (they won't be if pollJob is removed)

**Step 1:** Remove the `Op` struct from `result.go`.

**Step 2:** Remove `Task(name string, op *Op)` from `plan.go`. Keep `TaskFunc`
and `TaskFuncWithResults`.

**Step 3:** Remove all polling and broadcast functions from `runner.go`:
`executeOp`, `pollJob`, `countExpectedAgents`, `hostResultsFromResponses`,
`extractHostResults`, `isCommandOp`, `parseAgentDurations`, `isTransient`,
`IsBroadcastTarget`.

**Step 4:** Remove `agentDurations` field from `Result`. Keep `JobID`,
`Changed`, `Data`, `Status`, `JobDuration`, `HostResults`.

**Step 5:** Clean up imports — remove `backoff`, `client`, `strings`, `errors`
if no longer needed in `runner.go`.

**Step 6:** Delete or update tests. Remove all test methods that exercise
removed code. Update `TestRunOpTask`/`TestRunOpTaskErrors` in
`plan_public_test.go` — these should be converted to use `TaskFunc` calls that
hit a test HTTP server through SDK client methods.

**Step 7:** Run `go build ./...` and
`go test ./pkg/sdk/orchestrator/... -count=1`.

**Step 8:** Commit.

```bash
git commit -m "feat(sdk)!: remove Op struct and job polling from orchestrator

The SDK orchestrator engine now only supports TaskFunc and
TaskFuncWithResults. Operation execution goes through typed
SDK client methods called from TaskFunc closures."
```

---

### Task 3: Convert orchestrator DSL operation methods to TaskFunc

Convert the 13 `Op`-based methods in osapi-orchestrator to use `TaskFunc` with
SDK client calls. This follows the existing pattern used by HealthCheck,
FileUpload, FileChanged, AgentList, and AgentGet.

**Files:**

- Modify: `osapi-orchestrator/pkg/orchestrator/ops.go` — rewrite 13 methods
- Modify: `osapi-orchestrator/pkg/orchestrator/ops_test.go` — update tests
- Modify: `osapi-orchestrator/pkg/orchestrator/ops_public_test.go` — update
  tests
- Modify: `osapi-orchestrator/pkg/orchestrator/orchestrator.go` — remove
  `newStep` if no longer used

**Before (example):**

```go
func (o *Orchestrator) NodeHostnameGet(target string) *Step {
    return o.newStep(&sdk.Op{
        Operation: opNodeHostnameGet,
        Target:    target,
    })
}
```

**After (example):**

```go
func (o *Orchestrator) NodeHostnameGet(target string) *Step {
    name := o.nextOpName("get-hostname")
    task := o.plan.TaskFunc(
        name,
        func(ctx context.Context, c *osapi.Client) (*sdk.Result, error) {
            resp, err := c.Node.Hostname(ctx, target)
            if err != nil {
                return nil, fmt.Errorf("get hostname: %w", err)
            }
            return &sdk.Result{
                JobID:   resp.Data.JobID,
                Changed: anyChanged(resp.Data.Results),
                Data:    mustRawToMap(resp.RawJSON()),
            }, nil
        },
    )
    return &Step{task: task}
}
```

**Step 1:** Add a helper `nextOpName(prefix string) string` to replace the
operation-constant-based name generation (or reuse existing name logic).

**Step 2:** Convert each method one at a time. The 13 methods and their SDK
client calls:

| DSL Method                                         | SDK Call                                                | Notes              |
| -------------------------------------------------- | ------------------------------------------------------- | ------------------ |
| `NodeHostnameGet(target)`                          | `c.Node.Hostname(ctx, target)`                          |                    |
| `NodeStatusGet(target)`                            | `c.Node.Status(ctx, target)`                            |                    |
| `NodeUptimeGet(target)`                            | `c.Node.Uptime(ctx, target)`                            |                    |
| `NodeDiskGet(target)`                              | `c.Node.Disk(ctx, target)`                              |                    |
| `NodeMemoryGet(target)`                            | `c.Node.Memory(ctx, target)`                            |                    |
| `NodeLoadGet(target)`                              | `c.Node.Load(ctx, target)`                              |                    |
| `NetworkDNSGet(target, iface)`                     | `c.Node.GetDNS(ctx, target, iface)`                     |                    |
| `NetworkDNSUpdate(target, iface, servers, search)` | `c.Node.UpdateDNS(ctx, target, iface, servers, search)` |                    |
| `NetworkPingDo(target, addr)`                      | `c.Node.Ping(ctx, target, addr)`                        |                    |
| `CommandExec(target, cmd, args)`                   | `c.Node.Exec(ctx, ExecRequest{...})`                    | Build ExecRequest  |
| `CommandShell(target, cmd)`                        | `c.Node.Shell(ctx, ShellRequest{...})`                  | Build ShellRequest |
| `FileDeploy(target, opts)`                         | `c.Node.FileDeploy(ctx, FileDeployOpts{...})`           | Map opts           |
| `FileStatusGet(target, path)`                      | `c.Node.FileStatus(ctx, target, path)`                  |                    |

**Step 3:** Remove the operation constants (`opNodeHostnameGet`, etc.) and
`newStep` method — they are no longer needed.

**Step 4:** Remove `mustRawToMap` if no longer used (check FileUpload,
AgentList, AgentGet — they still use it, so it probably stays).

**Step 5:** Update tests. The existing tests create test HTTP servers and verify
the orchestrator calls the right endpoints. Update them to expect domain
endpoint paths instead of `POST /job`.

**Step 6:** Run `go build ./...` and `go test ./pkg/orchestrator/... -count=1`.

**Step 7:** Commit.

```bash
git commit -m "feat!: use typed SDK client methods instead of generic Op

All 13 operation methods now call domain endpoints through the
SDK client. The generic Op struct and job creation path are gone."
```

---

### Task 4: Update examples

All examples in osapi-orchestrator need to work with the new API. The DSL method
signatures are unchanged, so most examples should compile without modification.
Verify each one.

**Files:**

- Check: `osapi-orchestrator/examples/features/*.go`
- Check: `osapi-orchestrator/examples/operations/*.go`

**Step 1:** Run `go build -o /dev/null` for each example file to verify
compilation.

**Step 2:** Fix any import or type changes.

**Step 3:** Commit if changes were needed.

---

### Task 5: Update documentation

**Files:**

- Modify: `osapi-orchestrator/docs/features/README.md` — update if it references
  Op or job creation
- Modify: `osapi-orchestrator/docs/gen/orchestrator.md` — regenerate
- Modify: `osapi/docs/docs/sidebar/architecture/job-architecture.md` — note that
  POST /job was removed
- Modify: `osapi/docs/docs/sidebar/architecture/system-architecture.md` — update
  endpoint tables

**Step 1:** Regenerate API docs: `just generate` in osapi, `just docs::generate`
in osapi-orchestrator (if applicable).

**Step 2:** Update architecture docs to reflect that domain endpoints are the
sole job creation path.

**Step 3:** Commit.

---

### Task 6: Clean up SDK client

After removing `JobService.Create`, verify the SDK client's job module is clean.

**Files:**

- Modify: `osapi/pkg/sdk/client/job.go` — verify Create is gone
- Modify: `osapi/pkg/sdk/client/gen/` — regenerate from updated OpenAPI spec
  (POST /job removed)

**Step 1:** Run `go generate ./pkg/sdk/client/gen/...` to regenerate the SDK
client from the updated combined spec.

**Step 2:** Verify `JobService` only has `Get`, `Delete`, `List`, `Retry`,
`Stats`.

**Step 3:** Run full test suite.

**Step 4:** Commit.

---

### Task 7: Final verification

**Step 1:** In osapi: `go build ./... && just go::unit`

**Step 2:** In osapi-orchestrator: `go build ./... && just go::unit`

**Step 3:** Run integration tests if available: `just go::unit-int`

**Step 4:** Verify all examples compile.

---

## Ordering

Tasks 1 and 2 are in osapi and can be done together on one branch. Task 3 is in
osapi-orchestrator and depends on Tasks 1-2 being published (or linked via
`replace`). Tasks 4-6 are follow-ups. Task 7 is final verification.

## What Does NOT Change

- Domain endpoint handlers (`internal/api/node/`, `internal/api/network/`, etc.)
- `publishAndWait` and broadcast handling in `internal/job/client/query.go`
- Job observability endpoints (GET, DELETE, list, retry, stats)
- CLI commands (they call domain endpoints, not POST /job)
- DSL method signatures (users' DAG code is unchanged)
- Guards, retry, error strategies, hooks
- TaskFunc and TaskFuncWithResults
