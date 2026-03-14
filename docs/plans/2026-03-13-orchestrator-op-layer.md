# Orchestrator SDK Bridge Helpers Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to
> implement this plan task-by-task.

**Goal:** Add bridge helpers to the SDK orchestrator package, achieve 100%
coverage, fix all examples so they compile and are complete, update docs, then
remove the misplaced docker DSL.

**Architecture:** The SDK orchestrator package provides the DAG engine (plan,
task, runner) plus bridge utilities (`CollectionResult`, `StructToMap`).
Domain-specific operation methods belong in consumer packages like
`osapi-orchestrator`, not in the SDK. Examples demonstrate the `TaskFunc`
pattern.

**Tech Stack:** Go 1.25, generics, testify/suite, httptest

---

### Task 1: Add CollectionResult and StructToMap bridge helpers

**Files:**

- Create: `pkg/sdk/orchestrator/bridge.go`
- Test: `pkg/sdk/orchestrator/bridge_public_test.go`

**Step 1: Write tests for `StructToMap` and `CollectionResult`**

Create `bridge_public_test.go` with `BridgePublicTestSuite`:

Test `StructToMap`:

- Converts struct with json tags to map
- Returns nil for nil input
- Handles nested structs
- Omits zero-value fields with `omitempty`

Test `CollectionResult`:

- Single result extracts JobID, Changed, HostResults with Data
- Multiple results — Changed is true when any host changed
- Empty results — returns result with empty HostResults
- HostResult.Data auto-populated via StructToMap when mapper leaves it nil
- HostResult.Data preserved when mapper sets it explicitly

Use `client.HostnameResult`, `client.CommandResult` etc. as test inputs since
those are the real SDK types consumers will pass.

**Step 2: Run tests to verify they fail**

Run: `go test ./pkg/sdk/orchestrator/... -run TestBridgePublicTestSuite -v`
Expected: FAIL — `CollectionResult` and `StructToMap` not defined

**Step 3: Implement `bridge.go`**

```go
package orchestrator

import (
	"encoding/json"

	osapiclient "github.com/retr0h/osapi/pkg/sdk/client"
)

// StructToMap converts a struct to map[string]any using its JSON
// tags. Returns nil if v is nil or cannot be marshaled.
func StructToMap(v any) map[string]any

// CollectionResult builds a Result from a Collection response.
// It iterates all results, applies the toHostResult mapper to
// build per-host details, and auto-populates HostResult.Data
// via StructToMap when the mapper leaves it nil. Changed is true
// if any host reported a change.
func CollectionResult[T any](
    col osapiclient.Collection[T],
    toHostResult func(T) HostResult,
) *Result
```

Mirrors `osapi-orchestrator`'s `buildResult` and `toMap` — but exported and in
the SDK where it belongs.

**Step 4: Run tests to verify they pass**

Run: `go test ./pkg/sdk/orchestrator/... -run TestBridgePublicTestSuite -v`
Expected: PASS

**Step 5: Check coverage**

Run:
`go test ./pkg/sdk/orchestrator/... -coverprofile=/tmp/bridge.out && go tool cover -func=/tmp/bridge.out | grep bridge`
Expected: 100% on bridge.go

**Step 6: Commit**

```
feat(orchestrator): add CollectionResult and StructToMap helpers
```

---

### Task 2: Delete docker.go DSL and its tests

**Files:**

- Delete: `pkg/sdk/orchestrator/docker.go`
- Delete: `pkg/sdk/orchestrator/docker_public_test.go`

**Step 1: Delete the files**

```bash
rm pkg/sdk/orchestrator/docker.go
rm pkg/sdk/orchestrator/docker_public_test.go
```

**Step 2: Run SDK tests**

Run: `go test ./pkg/sdk/orchestrator/... -count=1` Expected: PASS (engine +
bridge tests pass, docker tests gone)

**Step 3: Check what breaks**

Run: `go build ./... 2>&1` Expected: Compilation failures in
`container-targeting.go` (references `plan.DockerPull` etc.). Note the failures
— fixed in Task 3.

**Step 4: Commit**

```
refactor(orchestrator): remove docker DSL methods

Domain-specific operation methods belong in consumer packages
like osapi-orchestrator, not in the SDK engine. The SDK provides
CollectionResult and StructToMap as bridge helpers instead.
```

---

### Task 3: Fix container-targeting example

**Files:**

- Modify: `examples/sdk/orchestrator/features/container-targeting.go`

**Step 1: Rewrite to use `TaskFunc` with `CollectionResult`**

Replace all `plan.DockerPull()`, `plan.DockerCreate()`, etc. with
`plan.TaskFunc()` calls that use the SDK client directly and
`orchestrator.CollectionResult()` to build results.

Keep the same DAG structure: pre-cleanup → pull → create → exec x3 + inspect +
deliberately-fails → cleanup.

Pre-cleanup remains a `TaskFunc` that swallows errors.

Each docker operation becomes:

```go
pull := plan.TaskFunc("pull-image", func(
    ctx context.Context,
    c *client.Client,
) (*orchestrator.Result, error) {
    resp, err := c.Docker.Pull(ctx, target, gen.DockerPullRequest{
        Image: containerImage,
    })
    if err != nil {
        return nil, err
    }

    return orchestrator.CollectionResult(resp.Data,
        func(r client.DockerPullResult) orchestrator.HostResult {
            return orchestrator.HostResult{
                Hostname: r.Hostname,
                Changed:  r.Changed,
                Error:    r.Error,
            }
        },
    ), nil
})
```

**Step 2: Build**

Run: `go build examples/sdk/orchestrator/features/container-targeting.go`
Expected: Compiles successfully

**Step 3: Commit**

```
refactor: update container-targeting to use TaskFunc with bridge helpers
```

---

### Task 4: Fix all broken operation examples and add docker examples

**Files:**

- Modify: 13 files in `examples/sdk/orchestrator/operations/` that use
  `plan.Task(&Op{...})` (all except `file-upload.go` which already uses
  `TaskFunc`)
- Modify: feature examples that use `plan.Task(&Op{...})`: `basic.go`,
  `broadcast.go`, `error-strategy.go`, `file-deploy-workflow.go`, `guards.go`,
  `hooks.go`, `only-if-changed.go`, `parallel.go`, `result-decode.go`,
  `task-func-results.go`, `task-func.go`
- Create: 8 docker operation examples: `docker-pull.go`, `docker-create.go`,
  `docker-list.go`, `docker-inspect.go`, `docker-start.go`, `docker-stop.go`,
  `docker-remove.go`, `docker-exec.go`

**Step 1: Rewrite operation examples to use `TaskFunc`**

Each currently does:

```go
plan.Task("get-hostname", &orchestrator.Op{
    Operation: "node.hostname.get",
    Target:    "_any",
})
```

Replace with:

```go
plan.TaskFunc("get-hostname", func(
    ctx context.Context,
    c *client.Client,
) (*orchestrator.Result, error) {
    resp, err := c.Node.Hostname(ctx, "_any")
    if err != nil {
        return nil, err
    }

    return orchestrator.CollectionResult(resp.Data,
        func(r client.HostnameResult) orchestrator.HostResult {
            return orchestrator.HostResult{
                Hostname: r.Hostname,
                Changed:  r.Changed,
                Error:    r.Error,
            }
        },
    ), nil
})
```

Apply this pattern to all 13 operation files and all feature files.

For operations with params (command exec, DNS update, file deploy, etc.), unpack
from the example's local variables into the SDK request types directly — no
`Params map[string]any` needed.

**Step 2: Create 8 docker operation examples**

Follow the exact same pattern as node/command examples. One file per docker
operation in `examples/sdk/orchestrator/operations/`.

**Step 3: Build every example individually**

```bash
for f in examples/sdk/orchestrator/operations/*.go; do
    go build "$f" 2>&1 || echo "FAIL: $f"
done
for f in examples/sdk/orchestrator/features/*.go; do
    go build "$f" 2>&1 || echo "FAIL: $f"
done
```

Expected: ALL files compile. Zero failures. This is a hard gate — do not proceed
until every example compiles.

**Step 4: Commit**

```
fix: rewrite all orchestrator examples to use TaskFunc
```

---

### Task 5: Update orchestrator docs

**Files:**

- Modify: `docs/docs/sidebar/sdk/orchestrator/orchestrator.md`
- Modify: all operation doc pages in
  `docs/docs/sidebar/sdk/orchestrator/operations/`
- Modify: `docs/docs/sidebar/sdk/orchestrator/features/container-targeting.md`

**Step 1: Update orchestrator overview**

Add documentation for `CollectionResult` and `StructToMap` as SDK-provided
bridge helpers. Update the usage examples to show `TaskFunc` pattern instead of
`plan.Task(&Op{...})`.

**Step 2: Update operation doc pages**

Each page currently shows the `plan.Task(&Op{...})` pattern. Update to show
`plan.TaskFunc()` with `CollectionResult`. Match the code in the corresponding
example file exactly.

**Step 3: Update container-targeting feature doc**

Update code examples to match the rewritten `container-targeting.go`.

**Step 4: Build docs**

Run: `cd docs && bun run build` Expected: Build succeeds, no broken links

**Step 5: Commit**

```
docs: update orchestrator docs for TaskFunc pattern
```

---

### Task 6: Final verification

**Step 1: Full test suite**

Run: `go test ./... -count=1` Expected: All packages pass

**Step 2: SDK coverage check**

Run:
`go test ./pkg/sdk/... -coverprofile=/tmp/sdk.out && go tool cover -func=/tmp/sdk.out | grep -v gen | grep -v '100.0%'`
Expected: All SDK packages at 100% (excluding gen)

**Step 3: Lint and format**

```bash
find . -type f -name '*.go' -not -name '*.gen.go' -not -name '*.pb.go' \
  -not -path './.worktrees/*' -not -path './.claude/*' \
  | xargs go tool github.com/segmentio/golines \
    --base-formatter="go tool mvdan.cc/gofumpt" -w
go tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint run \
  --config .golangci.yml
```

Expected: 0 issues

**Step 4: Verify all examples compile**

```bash
for f in examples/sdk/orchestrator/operations/*.go; do
    go build "$f" 2>&1 || echo "FAIL: $f"
done
for f in examples/sdk/orchestrator/features/*.go; do
    go build "$f" 2>&1 || echo "FAIL: $f"
done
```

Expected: Zero failures

**Step 5: Build docs**

Run: `cd docs && bun run build` Expected: No broken links

**Step 6: Commit any remaining fixes**

```
chore: final verification cleanup
```
