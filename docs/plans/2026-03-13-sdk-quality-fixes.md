# SDK Quality Fixes Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development
> to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix 9 SDK quality issues identified by code review — add JSON tags to
result types, wrap Docker gen types, add Collection.First(), fix
CollectionResult to populate Result.Data, fix error wrapping, then update
osapi-orchestrator to use SDK types directly.

**Architecture:** Fix the SDK client types first (JSON tags, Docker wrappers,
Collection.First), then fix the bridge helper (CollectionResult), then update
osapi-orchestrator to consume the improvements (delete duplicated types, use SDK
helpers).

**Tech Stack:** Go 1.25, generics, testify/suite

---

## Chunk 1: osapi SDK fixes

### Task 1: Add JSON tags to all SDK client result types

**Files:**

- Modify: `pkg/sdk/client/node_types.go`
- Modify: `pkg/sdk/client/docker_types.go`
- Modify: `pkg/sdk/client/file_types.go`
- Modify: `pkg/sdk/client/audit_types.go`
- Modify: `pkg/sdk/client/health_types.go`
- Modify: `pkg/sdk/client/job_types.go`
- Modify: `pkg/sdk/client/agent_types.go`

- [ ] **Step 1: Add JSON tags to node_types.go**

Add `json:"..."` tags to all exported struct fields in these types:
`Collection`, `Disk`, `HostnameResult`, `NodeStatus`, `DiskResult`,
`MemoryResult`, `LoadResult`, `OSInfoResult`, `UptimeResult`, `DNSConfig`,
`DNSUpdateResult`, `PingResult`, `CommandResult`, `LoadAverage`, `Memory`,
`OSInfo`.

Use snake_case keys matching the API response format. For example:

```go
type Collection[T any] struct {
    Results []T    `json:"results"`
    JobID   string `json:"job_id"`
}

type HostnameResult struct {
    Hostname string            `json:"hostname"`
    Error    string            `json:"error,omitempty"`
    Changed  bool              `json:"changed"`
    Labels   map[string]string `json:"labels,omitempty"`
}
```

Apply to all types — every exported field gets a `json` tag.

- [ ] **Step 2: Add JSON tags to docker_types.go**

Same pattern for: `DockerResult`, `DockerListResult`, `DockerSummaryItem`,
`DockerDetailResult`, `DockerActionResult`, `DockerExecResult`,
`DockerPullResult`.

- [ ] **Step 3: Add JSON tags to remaining types files**

Add tags to all result/model types in `file_types.go`, `audit_types.go`,
`health_types.go`, `job_types.go`, `agent_types.go`.

- [ ] **Step 4: Run tests**

Run: `go test ./pkg/sdk/client/... -count=1` Expected: PASS — JSON tags don't
break existing behavior.

- [ ] **Step 5: Commit**

```
feat(sdk): add JSON tags to all client result types
```

---

### Task 2: Add Collection[T].First() method

**Files:**

- Modify: `pkg/sdk/client/node_types.go`
- Test: `pkg/sdk/client/node_types_test.go` (or appropriate test file)

- [ ] **Step 1: Write the failing test**

Add to the existing client test suite:

```go
func (s *SuiteType) TestCollectionFirst() {
    tests := []struct {
        name         string
        col          client.Collection[client.HostnameResult]
        validateFunc func(client.HostnameResult, bool)
    }{
        {
            name: "returns first result and true",
            col: client.Collection[client.HostnameResult]{
                Results: []client.HostnameResult{
                    {Hostname: "web-01"},
                    {Hostname: "web-02"},
                },
                JobID: "job-1",
            },
            validateFunc: func(r client.HostnameResult, ok bool) {
                s.True(ok)
                s.Equal("web-01", r.Hostname)
            },
        },
        {
            name: "returns zero value and false when empty",
            col: client.Collection[client.HostnameResult]{
                Results: []client.HostnameResult{},
            },
            validateFunc: func(r client.HostnameResult, ok bool) {
                s.False(ok)
                s.Equal("", r.Hostname)
            },
        },
    }

    for _, tt := range tests {
        s.Run(tt.name, func() {
            r, ok := tt.col.First()
            tt.validateFunc(r, ok)
        })
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/sdk/client/... -run TestCollectionFirst -v` Expected: FAIL —
`First` not defined

- [ ] **Step 3: Implement First()**

Add to `node_types.go` after `Collection` definition:

```go
// First returns the first result and true, or the zero value
// and false if the collection is empty.
func (c Collection[T]) First() (T, bool) {
    if len(c.Results) == 0 {
        var zero T
        return zero, false
    }

    return c.Results[0], true
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./pkg/sdk/client/... -run TestCollectionFirst -v` Expected: PASS

- [ ] **Step 5: Commit**

```
feat(sdk): add Collection[T].First() method
```

---

### Task 3: Wrap Docker gen types with SDK-defined request types

**Files:**

- Modify: `pkg/sdk/client/docker_types.go`
- Modify: `pkg/sdk/client/docker.go`
- Modify: `pkg/sdk/client/docker_public_test.go`

- [ ] **Step 1: Define SDK Docker request types in docker_types.go**

Add after the result types:

```go
// DockerCreateOpts contains options for creating a container.
type DockerCreateOpts struct {
    // Image is the container image reference (required).
    Image string
    // Name is an optional container name.
    Name string
    // Command overrides the image's default command.
    Command []string
    // Env is environment variables in KEY=VALUE format.
    Env []string
    // Ports is port mappings in host_port:container_port format.
    Ports []string
    // Volumes is volume mounts in host_path:container_path format.
    Volumes []string
    // AutoStart starts the container after creation (default true).
    AutoStart *bool
}

// DockerStopOpts contains options for stopping a container.
type DockerStopOpts struct {
    // Timeout is seconds to wait before killing. Zero uses default.
    Timeout int
}

// DockerListParams contains parameters for listing containers.
type DockerListParams struct {
    // State filters by state: "running", "stopped", "all".
    State string
    // Limit caps the number of results.
    Limit int
}

// DockerRemoveParams contains parameters for removing a container.
type DockerRemoveParams struct {
    // Force forces removal of a running container.
    Force bool
}
```

- [ ] **Step 2: Update Docker service methods to use SDK types**

Change method signatures in `docker.go`:

`Create`: Change `body gen.DockerCreateRequest` to `opts DockerCreateOpts`.
Inside, build `gen.DockerCreateRequest` from the opts fields, converting zero
values to nil pointers.

`Stop`: Change `body gen.DockerStopRequest` to `opts DockerStopOpts`. Build
`gen.DockerStopRequest` from opts.

`List`: Change `params *gen.GetNodeContainerDockerParams` to
`params *DockerListParams`. Build gen params from SDK params.

`Remove`: Change `params *gen.DeleteNodeContainerDockerByIDParams` to
`params *DockerRemoveParams`. Build gen params from SDK params.

- [ ] **Step 3: Update tests**

Update `docker_public_test.go` to use the new SDK types instead of gen types.
Also update any examples that reference the old signatures.

- [ ] **Step 4: Update all callers**

Search for `gen.DockerCreateRequest`, `gen.DockerStopRequest`,
`gen.GetNodeContainerDockerParams`, `gen.DeleteNodeContainerDockerByIDParams`
in:

- `examples/sdk/client/container.go`
- `examples/sdk/orchestrator/features/container-targeting.go`
- `examples/sdk/orchestrator/operations/docker-*.go`

Replace with the new SDK types.

- [ ] **Step 5: Build and test**

Run: `go build ./...` Run: `go test ./pkg/sdk/client/... -count=1` Run: Build
each docker example:
`go build examples/sdk/orchestrator/operations/docker-pull.go` etc. Expected:
All compile, all tests pass

- [ ] **Step 6: Commit**

```
refactor(sdk): wrap Docker gen types with SDK-defined request types
```

---

### Task 4: Fix CollectionResult to populate Result.Data

**Files:**

- Modify: `pkg/sdk/orchestrator/bridge.go`
- Modify: `pkg/sdk/orchestrator/bridge_public_test.go`
- Modify: `pkg/sdk/orchestrator/bridge_test.go`

- [ ] **Step 1: Update CollectionResult signature**

Add `rawJSON []byte` parameter:

```go
func CollectionResult[T any](
    col client.Collection[T],
    rawJSON []byte,
    toHostResult func(T) HostResult,
) *Result
```

When `rawJSON` is non-nil, unmarshal into `Result.Data`. Use `jsonUnmarshalFn`
(already injectable for testing).

- [ ] **Step 2: Update tests**

Update all test cases in `bridge_public_test.go` to pass `nil` for `rawJSON`
(existing behavior preserved). Add new test cases:

- rawJSON populated: pass valid JSON, verify Result.Data is set
- rawJSON nil: verify Result.Data is nil (existing behavior)
- rawJSON invalid: verify Result.Data is nil (graceful degradation)

- [ ] **Step 3: Update all callers**

Update all example files that call `CollectionResult` to pass `resp.RawJSON()`
as the second argument.

- [ ] **Step 4: Run tests**

Run: `go test ./pkg/sdk/orchestrator/... -count=1` Run: Build all examples
Expected: PASS, all compile

- [ ] **Step 5: Commit**

```
feat(orchestrator): populate Result.Data from raw JSON in CollectionResult
```

---

### Task 5: Fix AuditService.Get UUID error wrapping

**Files:**

- Modify: `pkg/sdk/client/audit.go:78-81`

- [ ] **Step 1: Fix the error wrapping**

Change:

```go
parsedID, err := uuid.Parse(id)
if err != nil {
    return nil, err
}
```

To:

```go
parsedID, err := uuid.Parse(id)
if err != nil {
    return nil, fmt.Errorf("invalid audit ID: %w", err)
}
```

- [ ] **Step 2: Update test if one exists**

Check if there's a test for invalid audit ID. If so, update the expected error
message.

- [ ] **Step 3: Run tests**

Run: `go test ./pkg/sdk/client/... -count=1` Expected: PASS

- [ ] **Step 4: Commit**

```
fix(sdk): wrap audit UUID parse error with context
```

---

### Task 6: Final SDK verification

- [ ] **Step 1: Full test suite**

Run: `go test ./... -count=1` Expected: All pass

- [ ] **Step 2: Lint and format**

```bash
find . -type f -name '*.go' -not -name '*.gen.go' -not -name '*.pb.go' \
  -not -path './.worktrees/*' -not -path './.claude/*' \
  | xargs go tool github.com/segmentio/golines \
    --base-formatter="go tool mvdan.cc/gofumpt" -w
go tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint run \
  --config .golangci.yml
```

Expected: 0 issues

- [ ] **Step 3: Coverage**

Run:
`go test ./pkg/sdk/... -coverprofile=/tmp/sdk.out && go tool cover -func=/tmp/sdk.out | grep -v gen | grep -v '100.0%'`
Expected: All SDK packages at 100% (excluding gen)

- [ ] **Step 4: Build all examples**

```bash
for f in examples/sdk/orchestrator/operations/*.go; do
    go build "$f" 2>&1 || echo "FAIL: $f"
done
for f in examples/sdk/orchestrator/features/*.go; do
    go build "$f" 2>&1 || echo "FAIL: $f"
done
```

Expected: Zero failures

- [ ] **Step 5: Commit any fixes**

```
chore: SDK quality fixes verification
```

---

## Chunk 2: osapi-orchestrator fixes

These changes are in the separate repo at `~/git/osapi-io/osapi-orchestrator/`.

### Task 7: Update osapi-orchestrator to use SDK types directly

**Files:**

- Delete: `pkg/orchestrator/result_types.go`
- Modify: `pkg/orchestrator/ops.go`
- Modify: `pkg/orchestrator/result.go`
- Modify: any test files that reference deleted types

**Prerequisites:** Tasks 1-6 must be complete and the updated osapi SDK must be
available (update `go.mod` to point at the new version or use a `replace`
directive).

- [ ] **Step 1: Update go.mod to use latest SDK**

Either `go get github.com/retr0h/osapi@latest` or add a `replace` directive
pointing to the local checkout.

- [ ] **Step 2: Delete result_types.go**

Remove the file entirely. All types it defines have equivalents in the SDK
`client` package.

- [ ] **Step 3: Update ops.go imports and types**

Replace all local type references with SDK `client.*` types:

- `HostnameResult` → `client.HostnameResult`
- `CommandResult` → `client.CommandResult`
- `FileDeployOpts` → `client.FileDeployOpts`
- etc.

Replace `buildResult` calls with `orchestrator.CollectionResult`:

```go
return orchestrator.CollectionResult(resp.Data, resp.RawJSON(),
    func(r client.HostnameResult) orchestrator.HostResult {
        return orchestrator.HostResult{
            Hostname: r.Hostname,
            Changed:  r.Changed,
            Error:    r.Error,
        }
    },
), nil
```

Delete `buildResult`, `toMap`, `mustRawToMap` helper functions.

- [ ] **Step 4: Fix mustRawToMap callers that aren't Collection**

For non-collection operations (FileDeploy, FileStatus, FileUpload, FileChanged,
AgentList, AgentGet), replace:

```go
Data: mustRawToMap(resp.RawJSON()),
```

With:

```go
Data: orchestrator.StructToMap(resp.Data),
```

This works now because SDK types have JSON tags (Task 1).

- [ ] **Step 5: Update result.go**

Delete duplicated `Summary()` — delegate to `sdk.Report.Summary()`.

- [ ] **Step 6: Fix HealthCheck target parameter**

Remove the unused `target string` parameter from `HealthCheck()`. Update all
callers.

- [ ] **Step 7: Update tests**

Fix all test files that reference deleted types or changed signatures. Run full
test suite.

- [ ] **Step 8: Build and test**

Run: `go test ./... -count=1` Run: `go build ./...` Expected: All pass, all
compile

- [ ] **Step 9: Commit**

```
refactor: use SDK types directly, remove duplicated types

Delete result_types.go (~200 lines of duplicated SDK types).
Replace buildResult/toMap/mustRawToMap with SDK bridge helpers.
Use client.* types directly throughout ops.go.
```
