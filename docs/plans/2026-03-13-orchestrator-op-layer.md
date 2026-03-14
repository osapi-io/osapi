# Orchestrator Op Layer Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement the `Op` struct and `plan.Task()` method so all 23 documented operations work through a single declarative API, replacing the hardcoded Docker DSL methods.

**Architecture:** Add an `Op` type and operation dispatch registry to `pkg/sdk/orchestrator/`. Each operation string (e.g. `"docker.pull.execute"`) maps to a handler that calls the corresponding SDK client method, extracts results, and returns a `*Result`. The existing `TaskFunc`/`TaskFuncWithResults` remain for custom logic.

**Tech Stack:** Go 1.25, testify/suite, httptest

---

### Task 1: Add Op struct and Task.Operation() accessor

**Files:**
- Create: `pkg/sdk/orchestrator/op.go`
- Modify: `pkg/sdk/orchestrator/task.go:28-37`

**Step 1: Create `op.go` with Op struct and structToMap helper**

```go
package orchestrator

import (
	"encoding/json"
)

// Op describes a declarative operation to execute on a target.
type Op struct {
	// Operation is the dotted operation name
	// (e.g. "docker.pull.execute", "node.hostname.get").
	Operation string

	// Target is the agent routing target (hostname, "_any", "_all",
	// or label selector like "key:value").
	Target string

	// Params holds operation-specific parameters.
	Params map[string]any
}

// structToMap converts a struct to map[string]any using its JSON tags.
func structToMap(
	v any,
) map[string]any {
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
```

**Step 2: Add `op` field and `Operation()` accessor to `task.go`**

Add `op *Op` field to Task struct (after `errorStrategy` at line 36).
Add `Operation()` method.

```go
// In Task struct, add field:
	op *Op

// New method:
// Operation returns the Op for declarative tasks, or nil for
// TaskFunc tasks.
func (t *Task) Operation() *Op {
	return t.op
}
```

**Step 3: Run tests**

Run: `go test ./pkg/sdk/orchestrator/... -count=1`
Expected: PASS (no behavior changes yet)

**Step 4: Commit**

```
feat(orchestrator): add Op struct and Task.Operation accessor
```

---

### Task 2: Implement plan.Task() with node operation handlers

**Files:**
- Create: `pkg/sdk/orchestrator/registry.go`
- Modify: `pkg/sdk/orchestrator/op.go` (add plan.Task method)

**Step 1: Create `registry.go` with handler type and node handlers**

```go
package orchestrator

import (
	"context"
	"fmt"

	osapiclient "github.com/retr0h/osapi/pkg/sdk/client"
)

// opHandler executes an operation against the SDK client.
type opHandler func(
	ctx context.Context,
	c *osapiclient.Client,
	target string,
	params map[string]any,
) (*Result, error)

// registry maps operation strings to handlers.
var registry = map[string]opHandler{
	// Node operations
	"node.hostname.get": opNodeHostname,
	"node.status.get":   opNodeStatus,
	"node.disk.get":     opNodeDisk,
	"node.memory.get":   opNodeMemory,
	"node.uptime.get":   opNodeUptime,
	"node.load.get":     opNodeLoad,
}

// collectionResult builds a Result from a Collection response,
// extracting Results[0] for single-target operations.
func collectionResult[T any](
	jobID string,
	results []T,
	changed bool,
) *Result {
	if len(results) == 0 {
		return &Result{JobID: jobID, Changed: changed}
	}

	return &Result{
		JobID:   jobID,
		Changed: changed,
		Data:    structToMap(results[0]),
	}
}

func opNodeHostname(
	ctx context.Context,
	c *osapiclient.Client,
	target string,
	_ map[string]any,
) (*Result, error) {
	resp, err := c.Node.Hostname(ctx, target)
	if err != nil {
		return nil, err
	}

	r := resp.Data
	return collectionResult(r.JobID, r.Results, false), nil
}

func opNodeStatus(
	ctx context.Context,
	c *osapiclient.Client,
	target string,
	_ map[string]any,
) (*Result, error) {
	resp, err := c.Node.Status(ctx, target)
	if err != nil {
		return nil, err
	}

	r := resp.Data
	return collectionResult(r.JobID, r.Results, false), nil
}

func opNodeDisk(
	ctx context.Context,
	c *osapiclient.Client,
	target string,
	_ map[string]any,
) (*Result, error) {
	resp, err := c.Node.Disk(ctx, target)
	if err != nil {
		return nil, err
	}

	r := resp.Data
	return collectionResult(r.JobID, r.Results, false), nil
}

func opNodeMemory(
	ctx context.Context,
	c *osapiclient.Client,
	target string,
	_ map[string]any,
) (*Result, error) {
	resp, err := c.Node.Memory(ctx, target)
	if err != nil {
		return nil, err
	}

	r := resp.Data
	return collectionResult(r.JobID, r.Results, false), nil
}

func opNodeUptime(
	ctx context.Context,
	c *osapiclient.Client,
	target string,
	_ map[string]any,
) (*Result, error) {
	resp, err := c.Node.Uptime(ctx, target)
	if err != nil {
		return nil, err
	}

	r := resp.Data
	return collectionResult(r.JobID, r.Results, false), nil
}

func opNodeLoad(
	ctx context.Context,
	c *osapiclient.Client,
	target string,
	_ map[string]any,
) (*Result, error) {
	resp, err := c.Node.Load(ctx, target)
	if err != nil {
		return nil, err
	}

	r := resp.Data
	return collectionResult(r.JobID, r.Results, false), nil
}
```

**Step 2: Add `plan.Task()` method to `op.go`**

```go
// Task creates a declarative operation task, adds it to the plan,
// and returns it. The operation is dispatched to the appropriate
// SDK client method at execution time.
func (p *Plan) Task(
	name string,
	op *Op,
) *Task {
	handler, ok := registry[op.Operation]
	if !ok {
		// Create a task that always fails with unknown operation.
		return p.TaskFunc(name, func(
			_ context.Context,
			_ *osapiclient.Client,
		) (*Result, error) {
			return nil, fmt.Errorf("unknown operation: %s", op.Operation)
		})
	}

	t := p.TaskFunc(name, func(
		ctx context.Context,
		c *osapiclient.Client,
	) (*Result, error) {
		return handler(ctx, c, op.Target, op.Params)
	})
	t.op = op

	return t
}
```

Add `"fmt"` and `osapiclient` imports to `op.go`.

**Step 3: Build and run existing node operation examples**

Run: `go build ./pkg/sdk/orchestrator/...`
Run: `go build examples/sdk/orchestrator/operations/node-hostname.go`
Run: `go build examples/sdk/orchestrator/operations/node-disk.go`
Run: `go build examples/sdk/orchestrator/operations/node-memory.go`
Run: `go build examples/sdk/orchestrator/operations/node-load.go`
Run: `go build examples/sdk/orchestrator/operations/node-status.go`
Run: `go build examples/sdk/orchestrator/operations/node-uptime.go`
Expected: All compile successfully

**Step 4: Commit**

```
feat(orchestrator): implement plan.Task with node operation handlers
```

---

### Task 3: Add command and network operation handlers

**Files:**
- Modify: `pkg/sdk/orchestrator/registry.go`

**Step 1: Add command handlers to registry**

Add to the registry map:

```go
	// Command operations
	"command.exec.execute":  opCommandExec,
	"command.shell.execute": opCommandShell,
```

Implement:

```go
func opCommandExec(
	ctx context.Context,
	c *osapiclient.Client,
	target string,
	params map[string]any,
) (*Result, error) {
	req := osapiclient.ExecRequest{
		Target: target,
	}
	if v, ok := params["command"].(string); ok {
		req.Command = v
	}
	if v, ok := params["args"].([]string); ok {
		req.Args = v
	}
	if v, ok := params["cwd"].(string); ok {
		req.Cwd = v
	}
	if v, ok := params["timeout"].(int); ok {
		req.Timeout = v
	}

	resp, err := c.Node.Exec(ctx, req)
	if err != nil {
		return nil, err
	}

	r := resp.Data
	return collectionResult(r.JobID, r.Results, true), nil
}

func opCommandShell(
	ctx context.Context,
	c *osapiclient.Client,
	target string,
	params map[string]any,
) (*Result, error) {
	req := osapiclient.ShellRequest{
		Target: target,
	}
	if v, ok := params["command"].(string); ok {
		req.Command = v
	}
	if v, ok := params["cwd"].(string); ok {
		req.Cwd = v
	}
	if v, ok := params["timeout"].(int); ok {
		req.Timeout = v
	}

	resp, err := c.Node.Shell(ctx, req)
	if err != nil {
		return nil, err
	}

	r := resp.Data
	return collectionResult(r.JobID, r.Results, true), nil
}
```

**Step 2: Add network handlers to registry**

Add to the registry map:

```go
	// Network operations
	"network.dns.get":    opNetworkDNSGet,
	"network.dns.update": opNetworkDNSUpdate,
	"network.ping.do":    opNetworkPing,
```

Implement:

```go
func opNetworkDNSGet(
	ctx context.Context,
	c *osapiclient.Client,
	target string,
	params map[string]any,
) (*Result, error) {
	interfaceName, _ := params["interface_name"].(string)

	resp, err := c.Node.GetDNS(ctx, target, interfaceName)
	if err != nil {
		return nil, err
	}

	r := resp.Data
	return collectionResult(r.JobID, r.Results, false), nil
}

func opNetworkDNSUpdate(
	ctx context.Context,
	c *osapiclient.Client,
	target string,
	params map[string]any,
) (*Result, error) {
	interfaceName, _ := params["interface_name"].(string)

	var addresses []string
	if v, ok := params["addresses"]; ok {
		switch a := v.(type) {
		case []string:
			addresses = a
		case []any:
			for _, item := range a {
				if s, ok := item.(string); ok {
					addresses = append(addresses, s)
				}
			}
		}
	}

	var searchDomains []string
	if v, ok := params["search_domains"]; ok {
		switch a := v.(type) {
		case []string:
			searchDomains = a
		case []any:
			for _, item := range a {
				if s, ok := item.(string); ok {
					searchDomains = append(searchDomains, s)
				}
			}
		}
	}

	resp, err := c.Node.UpdateDNS(ctx, target, interfaceName, addresses, searchDomains)
	if err != nil {
		return nil, err
	}

	r := resp.Data
	if len(r.Results) > 0 {
		return collectionResult(r.JobID, r.Results, r.Results[0].Changed), nil
	}

	return collectionResult(r.JobID, r.Results, false), nil
}

func opNetworkPing(
	ctx context.Context,
	c *osapiclient.Client,
	target string,
	params map[string]any,
) (*Result, error) {
	address, _ := params["address"].(string)

	resp, err := c.Node.Ping(ctx, target, address)
	if err != nil {
		return nil, err
	}

	r := resp.Data
	return collectionResult(r.JobID, r.Results, false), nil
}
```

**Step 3: Build examples**

Run: `go build examples/sdk/orchestrator/operations/command-exec.go`
Run: `go build examples/sdk/orchestrator/operations/command-shell.go`
Run: `go build examples/sdk/orchestrator/operations/network-dns-get.go`
Run: `go build examples/sdk/orchestrator/operations/network-dns-update.go`
Run: `go build examples/sdk/orchestrator/operations/network-ping.go`
Expected: All compile successfully

**Step 4: Commit**

```
feat(orchestrator): add command and network operation handlers
```

---

### Task 4: Add file and docker operation handlers

**Files:**
- Modify: `pkg/sdk/orchestrator/registry.go`

**Step 1: Add file handlers to registry**

Add to the registry map:

```go
	// File operations
	"file.deploy.execute": opFileDeploy,
	"file.status.get":     opFileStatus,
```

Note: `file.upload` is excluded — it requires `io.Reader` which
doesn't fit the `Params map[string]any` pattern. The `file-upload.go`
example correctly uses `TaskFunc` for this.

Implement:

```go
func opFileDeploy(
	ctx context.Context,
	c *osapiclient.Client,
	target string,
	params map[string]any,
) (*Result, error) {
	req := osapiclient.FileDeployOpts{
		Target: target,
	}
	if v, ok := params["object_name"].(string); ok {
		req.ObjectName = v
	}
	if v, ok := params["path"].(string); ok {
		req.Path = v
	}
	if v, ok := params["content_type"].(string); ok {
		req.ContentType = v
	}
	if v, ok := params["mode"].(string); ok {
		req.Mode = v
	}
	if v, ok := params["owner"].(string); ok {
		req.Owner = v
	}
	if v, ok := params["group"].(string); ok {
		req.Group = v
	}
	if v, ok := params["vars"].(map[string]any); ok {
		req.Vars = v
	}

	resp, err := c.Node.FileDeploy(ctx, req)
	if err != nil {
		return nil, err
	}

	r := resp.Data
	return &Result{
		JobID:   r.JobID,
		Changed: r.Changed,
		Data:    structToMap(r),
	}, nil
}

func opFileStatus(
	ctx context.Context,
	c *osapiclient.Client,
	target string,
	params map[string]any,
) (*Result, error) {
	path, _ := params["path"].(string)

	resp, err := c.Node.FileStatus(ctx, target, path)
	if err != nil {
		return nil, err
	}

	r := resp.Data
	return &Result{
		JobID:   r.JobID,
		Changed: r.Changed,
		Data:    structToMap(r),
	}, nil
}
```

**Step 2: Add docker handlers to registry**

Add to the registry map:

```go
	// Docker operations
	"docker.create.execute":  opDockerCreate,
	"docker.list.get":        opDockerList,
	"docker.inspect.get":     opDockerInspect,
	"docker.start.execute":   opDockerStart,
	"docker.stop.execute":    opDockerStop,
	"docker.remove.execute":  opDockerRemove,
	"docker.exec.execute":    opDockerExec,
	"docker.pull.execute":    opDockerPull,
```

Implement (each follows the same collection pattern):

```go
func opDockerCreate(
	ctx context.Context,
	c *osapiclient.Client,
	target string,
	params map[string]any,
) (*Result, error) {
	body := gen.DockerCreateRequest{}
	if v, ok := params["image"].(string); ok {
		body.Image = v
	}
	if v, ok := params["name"].(string); ok {
		body.Name = &v
	}
	if v, ok := params["auto_start"].(bool); ok {
		body.AutoStart = &v
	}
	if v, ok := params["command"].([]string); ok {
		body.Command = &v
	}

	resp, err := c.Docker.Create(ctx, target, body)
	if err != nil {
		return nil, err
	}

	r := resp.Data
	if len(r.Results) > 0 {
		return collectionResult(r.JobID, r.Results, r.Results[0].Changed), nil
	}

	return collectionResult(r.JobID, r.Results, false), nil
}

func opDockerList(
	ctx context.Context,
	c *osapiclient.Client,
	target string,
	params map[string]any,
) (*Result, error) {
	var p *gen.GetNodeContainerDockerParams
	if len(params) > 0 {
		p = &gen.GetNodeContainerDockerParams{}
		if v, ok := params["state"].(string); ok {
			p.State = (*gen.GetNodeContainerDockerParamsState)(&v)
		}
	}

	resp, err := c.Docker.List(ctx, target, p)
	if err != nil {
		return nil, err
	}

	r := resp.Data
	return collectionResult(r.JobID, r.Results, false), nil
}

func opDockerInspect(
	ctx context.Context,
	c *osapiclient.Client,
	target string,
	params map[string]any,
) (*Result, error) {
	id, _ := params["id"].(string)

	resp, err := c.Docker.Inspect(ctx, target, id)
	if err != nil {
		return nil, err
	}

	r := resp.Data
	return collectionResult(r.JobID, r.Results, false), nil
}

func opDockerStart(
	ctx context.Context,
	c *osapiclient.Client,
	target string,
	params map[string]any,
) (*Result, error) {
	id, _ := params["id"].(string)

	resp, err := c.Docker.Start(ctx, target, id)
	if err != nil {
		return nil, err
	}

	r := resp.Data
	if len(r.Results) > 0 {
		return collectionResult(r.JobID, r.Results, r.Results[0].Changed), nil
	}

	return collectionResult(r.JobID, r.Results, true), nil
}

func opDockerStop(
	ctx context.Context,
	c *osapiclient.Client,
	target string,
	params map[string]any,
) (*Result, error) {
	id, _ := params["id"].(string)
	body := gen.DockerStopRequest{}
	if v, ok := params["timeout"].(int); ok {
		body.Timeout = &v
	}

	resp, err := c.Docker.Stop(ctx, target, id, body)
	if err != nil {
		return nil, err
	}

	r := resp.Data
	if len(r.Results) > 0 {
		return collectionResult(r.JobID, r.Results, r.Results[0].Changed), nil
	}

	return collectionResult(r.JobID, r.Results, true), nil
}

func opDockerRemove(
	ctx context.Context,
	c *osapiclient.Client,
	target string,
	params map[string]any,
) (*Result, error) {
	id, _ := params["id"].(string)

	var p *gen.DeleteNodeContainerDockerByIDParams
	if v, ok := params["force"].(bool); ok {
		p = &gen.DeleteNodeContainerDockerByIDParams{Force: &v}
	}

	resp, err := c.Docker.Remove(ctx, target, id, p)
	if err != nil {
		return nil, err
	}

	r := resp.Data
	if len(r.Results) > 0 {
		return collectionResult(r.JobID, r.Results, r.Results[0].Changed), nil
	}

	return collectionResult(r.JobID, r.Results, true), nil
}

func opDockerExec(
	ctx context.Context,
	c *osapiclient.Client,
	target string,
	params map[string]any,
) (*Result, error) {
	id, _ := params["id"].(string)
	body := gen.DockerExecRequest{}

	if v, ok := params["command"].([]string); ok {
		body.Command = v
	}

	resp, err := c.Docker.Exec(ctx, target, id, body)
	if err != nil {
		return nil, err
	}

	r := resp.Data
	if len(r.Results) > 0 {
		return collectionResult(r.JobID, r.Results, r.Results[0].Changed), nil
	}

	return collectionResult(r.JobID, r.Results, true), nil
}

func opDockerPull(
	ctx context.Context,
	c *osapiclient.Client,
	target string,
	params map[string]any,
) (*Result, error) {
	image, _ := params["image"].(string)
	body := gen.DockerPullRequest{Image: image}

	resp, err := c.Docker.Pull(ctx, target, body)
	if err != nil {
		return nil, err
	}

	r := resp.Data
	if len(r.Results) > 0 {
		return collectionResult(r.JobID, r.Results, r.Results[0].Changed), nil
	}

	return collectionResult(r.JobID, r.Results, true), nil
}
```

**Step 3: Build examples**

Run: `go build examples/sdk/orchestrator/operations/file-deploy.go`
Run: `go build examples/sdk/orchestrator/operations/file-status.go`
Expected: Compile successfully

**Step 4: Commit**

```
feat(orchestrator): add file and docker operation handlers
```

---

### Task 5: Delete docker.go and update tests

**Files:**
- Delete: `pkg/sdk/orchestrator/docker.go`
- Delete: `pkg/sdk/orchestrator/docker_public_test.go`
- Create: `pkg/sdk/orchestrator/op_public_test.go`

**Step 1: Delete `docker.go` and `docker_public_test.go`**

```bash
rm pkg/sdk/orchestrator/docker.go
rm pkg/sdk/orchestrator/docker_public_test.go
```

**Step 2: Create `op_public_test.go` with tests for plan.Task()**

Test structure:
- `OpPublicTestSuite` with testify/suite
- One method per operation category testing success, error, and
  unknown operation
- Use `httptest.Server` to mock API responses (same pattern as
  the deleted docker tests)
- Test `task.Operation()` returns the `*Op`
- Test `structToMap()` helper
- Test unknown operation returns error

Cover at minimum:
- `TestTask` — creates task, verifies name, Operation() accessor
- `TestTaskUnknownOperation` — unknown op string fails at runtime
- `TestNodeHostname` — success with httptest mock
- `TestCommandExec` — success with params extraction
- `TestDockerPull` — success (replaces deleted docker test)
- `TestDockerCreate` — success with complex params
- `TestStructToMap` — helper function

Use valid UUIDs for `job_id` in all mock JSON responses.

**Step 3: Run tests**

Run: `go test ./pkg/sdk/orchestrator/... -count=1 -v`
Expected: PASS

**Step 4: Run coverage**

Run: `go test ./pkg/sdk/orchestrator/... -coverprofile=/tmp/op-cov.out && go tool cover -func=/tmp/op-cov.out | grep -E 'op\.go|registry\.go'`
Expected: High coverage on op.go and registry.go

**Step 5: Commit**

```
refactor(orchestrator): replace docker DSL with Op-based dispatch
```

---

### Task 6: Update container-targeting example and add docker operation examples

**Files:**
- Modify: `examples/sdk/orchestrator/features/container-targeting.go`
- Create: `examples/sdk/orchestrator/operations/docker-pull.go`
- Create: `examples/sdk/orchestrator/operations/docker-create.go`
- Create: `examples/sdk/orchestrator/operations/docker-list.go`
- Create: `examples/sdk/orchestrator/operations/docker-inspect.go`
- Create: `examples/sdk/orchestrator/operations/docker-start.go`
- Create: `examples/sdk/orchestrator/operations/docker-stop.go`
- Create: `examples/sdk/orchestrator/operations/docker-remove.go`
- Create: `examples/sdk/orchestrator/operations/docker-exec.go`

**Step 1: Rewrite `container-targeting.go` to use `plan.Task(&Op{...})`**

Replace all `plan.DockerPull()`, `plan.DockerCreate()`, etc. with
`plan.Task(name, &orchestrator.Op{...})`. Keep the same DAG structure
(pre-cleanup, pull, create, exec x3, inspect, deliberately-fails,
cleanup).

Pre-cleanup remains a `TaskFunc` since it swallows errors.

**Step 2: Create docker operation examples**

Follow the exact same pattern as existing operation examples
(node-hostname.go, command-exec.go, etc.):
- Same boilerplate (env vars, hooks, plan setup)
- Single `plan.Task()` call with `&orchestrator.Op{}`
- Print report

Example `docker-pull.go`:

```go
plan.Task("pull-image", &orchestrator.Op{
    Operation: "docker.pull.execute",
    Target:    "_any",
    Params: map[string]any{
        "image": "alpine:latest",
    },
})
```

**Step 3: Build every example individually**

Run each file through `go build` to verify compilation:

```bash
for f in examples/sdk/orchestrator/operations/*.go; do
    go build "$f" && echo "OK: $f" || echo "FAIL: $f"
done
for f in examples/sdk/orchestrator/features/*.go; do
    go build "$f" && echo "OK: $f" || echo "FAIL: $f"
done
```

Expected: ALL files compile successfully. Zero failures.

**Step 4: Commit**

```
refactor(orchestrator): update examples to use Op-based API
```

---

### Task 7: Update orchestrator operation docs

**Files:**
- Modify: `docs/docs/sidebar/sdk/orchestrator/orchestrator.md`
- Modify: `docs/docs/sidebar/sdk/orchestrator/operations/docker-*.md` (8 files)

**Step 1: Update docker operation docs**

Each docker operation doc currently shows the old `c.Docker.*` SDK
pattern. Update to show the `plan.Task(&Op{...})` pattern matching
the new examples.

**Step 2: Remove `file.upload` from operations table**

`file.upload` doesn't fit the `Op` pattern (requires `io.Reader`).
Remove it from the operations table in `orchestrator.md` and update
the `file-upload.md` doc to clarify it uses `TaskFunc`.

**Step 3: Build docs**

Run: `cd docs && bun run build`
Expected: Build succeeds with no broken links

**Step 4: Commit**

```
docs: update orchestrator docs for Op-based API
```

---

### Task 8: Final verification

**Step 1: Full test suite**

Run: `go test ./... -count=1`
Expected: All packages pass

**Step 2: Lint and format**

Run: `find . -type f -name '*.go' -not -name '*.gen.go' -not -name '*.pb.go' -not -path './.worktrees/*' -not -path './.claude/*' | xargs go tool github.com/segmentio/golines --base-formatter="go tool mvdan.cc/gofumpt" -w`
Run: `go tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint run --config .golangci.yml`
Expected: 0 issues

**Step 3: Coverage check**

Run: `go test ./pkg/sdk/orchestrator/... -coverprofile=/tmp/orch.out && go tool cover -func=/tmp/orch.out | grep -E 'op\.go|registry\.go'`
Expected: 100% on op.go, high coverage on registry.go

**Step 4: Verify all examples compile**

```bash
for f in examples/sdk/orchestrator/operations/*.go; do
    go build "$f" 2>&1 || echo "FAIL: $f"
done
for f in examples/sdk/orchestrator/features/*.go; do
    go build "$f" 2>&1 || echo "FAIL: $f"
done
```

Expected: Zero compilation failures

**Step 5: Build docs**

Run: `cd docs && bun run build`
Expected: No broken links

**Step 6: Commit any fixes**

```
chore: final verification cleanup
```
