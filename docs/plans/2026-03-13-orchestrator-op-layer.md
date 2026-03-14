# Orchestrator SDK Helpers Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add bridge helpers (`CollectionResult`, `StructToMap`) to the SDK orchestrator package so consumers like `osapi-orchestrator` don't need to reinvent them. Delete the misplaced `docker.go` DSL methods. Fix broken examples to use `TaskFunc` (the pattern `osapi-orchestrator` actually uses).

**Architecture:** The SDK orchestrator package provides the DAG engine (plan, task, runner) plus bridge utilities. Domain-specific operation methods (NodeHostnameGet, DockerPull, etc.) belong in consumer packages like `osapi-orchestrator`, not in the SDK.

**Tech Stack:** Go 1.25, generics, testify/suite, httptest

---

### Task 1: Add CollectionResult and StructToMap helpers

**Files:**
- Create: `pkg/sdk/orchestrator/bridge.go`
- Test: `pkg/sdk/orchestrator/bridge_public_test.go`

**Step 1: Write the failing test**

Create `bridge_public_test.go`:

```go
package orchestrator_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/pkg/sdk/client"
	"github.com/retr0h/osapi/pkg/sdk/orchestrator"
)

type BridgePublicTestSuite struct {
	suite.Suite
}

func (s *BridgePublicTestSuite) TestStructToMap() {
	tests := []struct {
		name         string
		input        any
		validateFunc func(map[string]any)
	}{
		{
			name: "converts struct with json tags",
			input: struct {
				Name    string `json:"name"`
				Changed bool   `json:"changed"`
			}{Name: "test", Changed: true},
			validateFunc: func(m map[string]any) {
				s.Equal("test", m["name"])
				s.Equal(true, m["changed"])
			},
		},
		{
			name:  "returns nil for nil input",
			input: nil,
			validateFunc: func(m map[string]any) {
				s.Nil(m)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := orchestrator.StructToMap(tt.input)
			tt.validateFunc(result)
		})
	}
}

func (s *BridgePublicTestSuite) TestCollectionResult() {
	tests := []struct {
		name         string
		jobID        string
		results      []client.HostnameResult
		toHostResult func(client.HostnameResult) orchestrator.HostResult
		validateFunc func(*orchestrator.Result)
	}{
		{
			name:  "single result extracts fields",
			jobID: "job-123",
			results: []client.HostnameResult{
				{Hostname: "web-01", Changed: false},
			},
			toHostResult: func(
				r client.HostnameResult,
			) orchestrator.HostResult {
				return orchestrator.HostResult{
					Hostname: r.Hostname,
					Changed:  r.Changed,
				}
			},
			validateFunc: func(result *orchestrator.Result) {
				s.Equal("job-123", result.JobID)
				s.False(result.Changed)
				s.Len(result.HostResults, 1)
				s.Equal("web-01", result.HostResults[0].Hostname)
				s.NotNil(result.HostResults[0].Data)
			},
		},
		{
			name:  "changed is true when any host changed",
			jobID: "job-456",
			results: []client.HostnameResult{
				{Hostname: "web-01", Changed: false},
				{Hostname: "web-02", Changed: true},
			},
			toHostResult: func(
				r client.HostnameResult,
			) orchestrator.HostResult {
				return orchestrator.HostResult{
					Hostname: r.Hostname,
					Changed:  r.Changed,
				}
			},
			validateFunc: func(result *orchestrator.Result) {
				s.True(result.Changed)
				s.Len(result.HostResults, 2)
			},
		},
		{
			name:    "empty results",
			jobID:   "job-789",
			results: []client.HostnameResult{},
			toHostResult: func(
				r client.HostnameResult,
			) orchestrator.HostResult {
				return orchestrator.HostResult{}
			},
			validateFunc: func(result *orchestrator.Result) {
				s.Equal("job-789", result.JobID)
				s.False(result.Changed)
				s.Empty(result.HostResults)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			col := client.Collection[client.HostnameResult]{
				JobID:   tt.jobID,
				Results: tt.results,
			}
			result := orchestrator.CollectionResult(
				col,
				tt.toHostResult,
			)
			tt.validateFunc(result)
		})
	}
}

func TestBridgePublicTestSuite(t *testing.T) {
	suite.Run(t, new(BridgePublicTestSuite))
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/sdk/orchestrator/... -run TestBridgePublicTestSuite -v`
Expected: FAIL — `CollectionResult` and `StructToMap` not defined

**Step 3: Write the implementation**

Create `bridge.go`:

```go
package orchestrator

import (
	"encoding/json"

	osapiclient "github.com/retr0h/osapi/pkg/sdk/client"
)

// StructToMap converts a struct to map[string]any using its JSON
// tags. Returns nil if v is nil or cannot be marshaled.
func StructToMap(
	v any,
) map[string]any {
	if v == nil {
		return nil
	}

	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}

	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil
	}

	return m
}

// CollectionResult builds a Result from a Collection response.
// It iterates all results, applies the toHostResult mapper to
// build per-host details, and auto-populates HostResult.Data
// via StructToMap when the mapper leaves it nil. Changed is true
// if any host reported a change.
func CollectionResult[T any](
	col osapiclient.Collection[T],
	toHostResult func(T) HostResult,
) *Result {
	var changed bool

	hostResults := make([]HostResult, 0, len(col.Results))

	for _, r := range col.Results {
		hr := toHostResult(r)
		if hr.Data == nil {
			hr.Data = StructToMap(r)
		}
		hostResults = append(hostResults, hr)

		if hr.Changed {
			changed = true
		}
	}

	return &Result{
		JobID:       col.JobID,
		Changed:     changed,
		HostResults: hostResults,
	}
}
```

Note: This takes `Collection[T]` directly (not `*Response[Collection[T]]`)
so callers pass `resp.Data` — cleaner than requiring the full response
wrapper. The caller can still access `resp.RawJSON()` separately if needed.

Also confirm `Collection` is exported from the client package. Check:

```go
// In pkg/sdk/client/response.go — Collection should be exported
type Collection[T any] struct {
	Results []T
	JobID   string
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/sdk/orchestrator/... -run TestBridgePublicTestSuite -v`
Expected: PASS

**Step 5: Commit**

```
feat(orchestrator): add CollectionResult and StructToMap helpers
```

---

### Task 2: Delete docker.go and its tests

**Files:**
- Delete: `pkg/sdk/orchestrator/docker.go`
- Delete: `pkg/sdk/orchestrator/docker_public_test.go`

**Step 1: Delete the files**

```bash
rm pkg/sdk/orchestrator/docker.go
rm pkg/sdk/orchestrator/docker_public_test.go
```

**Step 2: Run tests to verify nothing else depends on them**

Run: `go build ./...`
Expected: Compilation failure in examples that reference `DockerPull` etc.

Note which files fail — these will be fixed in Task 3.

Run: `go test ./pkg/sdk/orchestrator/... -count=1`
Expected: PASS (docker tests are gone, engine tests still pass)

**Step 3: Commit**

```
refactor(orchestrator): remove docker DSL methods

Domain-specific operation methods belong in consumer packages
like osapi-orchestrator, not in the SDK engine. The SDK provides
CollectionResult and StructToMap as bridge helpers instead.
```

---

### Task 3: Fix container-targeting example to use TaskFunc

**Files:**
- Modify: `examples/sdk/orchestrator/features/container-targeting.go`

**Step 1: Rewrite to use `plan.TaskFunc()` pattern**

Replace `plan.DockerPull()`, `plan.DockerCreate()`, etc. with
`plan.TaskFunc()` calls that use the SDK client directly — the same
pattern `osapi-orchestrator` uses. Import `gen` for request types.

Each operation becomes:

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

Keep the same DAG structure (pre-cleanup, pull, create, exec x3,
inspect, deliberately-fails, cleanup).

**Step 2: Build and verify**

Run: `go build examples/sdk/orchestrator/features/container-targeting.go`
Expected: Compiles successfully

**Step 3: Commit**

```
refactor: update container-targeting to use TaskFunc with bridge helpers
```

---

### Task 4: Fix broken operation examples

**Files:**
- Modify: all 14 files in `examples/sdk/orchestrator/operations/` that use `plan.Task(&Op{...})`
- Modify: all feature examples in `examples/sdk/orchestrator/features/` that use `plan.Task(&Op{...})`

**Step 1: Rewrite operation examples to use `plan.TaskFunc()`**

Each operation example currently does:

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

Do the same for all 14 operation examples and all feature examples
that reference `orchestrator.Op`.

**Step 2: Add 8 docker operation examples**

Create one file per docker operation in
`examples/sdk/orchestrator/operations/`:
- `docker-pull.go`
- `docker-create.go`
- `docker-list.go`
- `docker-inspect.go`
- `docker-start.go`
- `docker-stop.go`
- `docker-remove.go`
- `docker-exec.go`

Follow the same pattern as the node/command/file examples.

**Step 3: Build every example individually**

```bash
for f in examples/sdk/orchestrator/operations/*.go; do
    go build "$f" 2>&1 || echo "FAIL: $f"
done
for f in examples/sdk/orchestrator/features/*.go; do
    go build "$f" 2>&1 || echo "FAIL: $f"
done
```

Expected: ALL files compile. Zero failures.

**Step 4: Commit**

```
fix: rewrite orchestrator examples to use TaskFunc
```

---

### Task 5: Update orchestrator docs

**Files:**
- Modify: `docs/docs/sidebar/sdk/orchestrator/orchestrator.md`
- Modify: `docs/docs/sidebar/sdk/orchestrator/operations/docker-*.md` (8 files)
- Modify: all operation doc pages that show `plan.Task(&Op{...})` pattern

**Step 1: Update operation docs**

Each operation doc currently shows the `plan.Task(&Op{...})` pattern.
Update to show the `plan.TaskFunc()` pattern with `CollectionResult`.

Also update the orchestrator overview to document `CollectionResult`
and `StructToMap` as SDK-provided bridge helpers.

**Step 2: Build docs**

Run: `cd docs && bun run build`
Expected: Build succeeds, no broken links

**Step 3: Commit**

```
docs: update orchestrator docs for TaskFunc pattern
```

---

### Task 6: Final verification

**Step 1: Full test suite**

Run: `go test ./... -count=1`
Expected: All packages pass

**Step 2: Lint and format**

Run formatter and linter:
```bash
find . -type f -name '*.go' -not -name '*.gen.go' -not -name '*.pb.go' \
  -not -path './.worktrees/*' -not -path './.claude/*' \
  | xargs go tool github.com/segmentio/golines \
    --base-formatter="go tool mvdan.cc/gofumpt" -w
go tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint run \
  --config .golangci.yml
```
Expected: 0 issues

**Step 3: Coverage**

Run: `go test ./pkg/sdk/orchestrator/... -coverprofile=/tmp/orch.out && go tool cover -func=/tmp/orch.out | grep bridge`
Expected: 100% on bridge.go

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

Run: `cd docs && bun run build`
Expected: No broken links

**Step 6: Commit any fixes**

```
chore: final verification cleanup
```
