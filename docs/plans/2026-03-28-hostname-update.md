# Hostname Update Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> superpowers:subagent-driven-development (recommended) or
> superpowers:executing-plans to implement this plan task-by-task. Steps use
> checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a full-stack `hostname update` feature — from provider through API,
SDK, CLI, and docs — so operators can change a node's hostname via
`hostnamectl set-hostname`.

**Architecture:** The Debian provider calls `hostnamectl set-hostname` via
`exec.Manager`, checking the current hostname first for idempotency. Darwin,
Linux, and DebianDocker providers return `ErrUnsupported`. The API exposes
`PUT /node/{hostname}/hostname` following the DNS update pattern. The CLI adds
`client node hostname update --name <new-name>`.

**Tech Stack:** Go, exec.Manager (for hostnamectl), oapi-codegen, testify/suite,
gomock

---

### Task 1: Provider — Add SetHostname to Interface and Types

**Files:**

- Modify: `internal/provider/node/host/types.go`

- [ ] **Step 1: Add SetHostnameResult and SetHostname to the interface**

In `internal/provider/node/host/types.go`, add the result type after the
existing `Result` struct:

```go
// SetHostnameResult represents the outcome of a hostname set operation.
type SetHostnameResult struct {
	// Changed indicates whether the hostname was actually modified.
	Changed bool `json:"changed"`
}
```

Add `SetHostname` to the `Provider` interface:

```go
// SetHostname sets the system hostname.
SetHostname(name string) (*SetHostnameResult, error)
```

- [ ] **Step 2: Verify build fails**

Run: `go build ./...`
Expected: FAIL — all Provider implementations missing `SetHostname`.

- [ ] **Step 3: Commit**

```
feat(host): add SetHostname to Provider interface
```

---

### Task 2: Provider — Debian SetHostname Implementation

**Files:**

- Modify: `internal/provider/node/host/debian.go` (add `exec.Manager` field)
- Create: `internal/provider/node/host/debian_set_hostname.go`
- Create: `internal/provider/node/host/debian_set_hostname_public_test.go`
- Modify: `cmd/agent_setup.go` (pass `execManager` to `NewDebianProvider`)

- [ ] **Step 1: Add exec.Manager to Debian struct**

In `internal/provider/node/host/debian.go`, add the import and field:

```go
import (
	"os"
	"os/exec"
	"runtime"

	"github.com/shirou/gopsutil/v4/host"

	"github.com/retr0h/osapi/internal/exec"
	"github.com/retr0h/osapi/internal/provider"
)
```

Note: the import alias must avoid collision with `os/exec`. Use the full path
`iexec "github.com/retr0h/osapi/internal/exec"` if needed, or since the field
type is an interface, use the `exec.Manager` type directly. Check how the dns
package handles this — it imports `"github.com/retr0h/osapi/internal/exec"` and
the field type is `exec.Manager`.

Add `execManager` field to the `Debian` struct:

```go
type Debian struct {
	provider.FactsAware

	InfoFn      func() (*host.InfoStat, error)
	HostnameFn  func() (string, error)
	NumCPUFn    func() int
	StatFn      func(name string) (os.FileInfo, error)
	LookPathFn  func(file string) (string, error)
	execManager iexec.Manager
}
```

Update `NewDebianProvider` to accept `exec.Manager`:

```go
func NewDebianProvider(
	execManager iexec.Manager,
) *Debian {
	return &Debian{
		InfoFn:      host.Info,
		HostnameFn:  os.Hostname,
		NumCPUFn:    runtime.NumCPU,
		StatFn:      os.Stat,
		LookPathFn:  exec.LookPath,
		execManager: execManager,
	}
}
```

Update `cmd/agent_setup.go` to pass `execManager`:

```go
case "debian":
	hostProvider = nodeHost.NewDebianProvider(execManager)
```

- [ ] **Step 2: Write the failing test**

Create `internal/provider/node/host/debian_set_hostname_public_test.go`:

```go
package host_test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/exec/mocks"
	"github.com/retr0h/osapi/internal/provider/node/host"
)

type DebianSetHostnamePublicTestSuite struct {
	suite.Suite
	ctrl *gomock.Controller

	logger *slog.Logger
}

func (s *DebianSetHostnamePublicTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (s *DebianSetHostnamePublicTestSuite) SetupSubTest() {
	s.SetupTest()
}

func (s *DebianSetHostnamePublicTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *DebianSetHostnamePublicTestSuite) TestSetHostname() {
	tests := []struct {
		name         string
		setupMock    func() *mocks.MockManager
		hostname     string
		wantErr      bool
		errContains  string
		validateFunc func(*host.SetHostnameResult)
	}{
		{
			name: "when hostname changes",
			setupMock: func() *mocks.MockManager {
				mock := mocks.NewPlainMockManager(s.ctrl)
				mock.EXPECT().
					RunCmd("hostnamectl", []string{"hostname"}).
					Return("old-hostname\n", nil)
				mock.EXPECT().
					RunCmd("hostnamectl", []string{"set-hostname", "new-hostname"}).
					Return("", nil)
				return mock
			},
			hostname: "new-hostname",
			validateFunc: func(r *host.SetHostnameResult) {
				s.True(r.Changed)
			},
		},
		{
			name: "when hostname already set returns unchanged",
			setupMock: func() *mocks.MockManager {
				mock := mocks.NewPlainMockManager(s.ctrl)
				mock.EXPECT().
					RunCmd("hostnamectl", []string{"hostname"}).
					Return("same-hostname\n", nil)
				return mock
			},
			hostname: "same-hostname",
			validateFunc: func(r *host.SetHostnameResult) {
				s.False(r.Changed)
			},
		},
		{
			name: "when hostnamectl hostname errors",
			setupMock: func() *mocks.MockManager {
				mock := mocks.NewPlainMockManager(s.ctrl)
				mock.EXPECT().
					RunCmd("hostnamectl", []string{"hostname"}).
					Return("", fmt.Errorf("command not found"))
				return mock
			},
			hostname:    "new-hostname",
			wantErr:     true,
			errContains: "failed to get current hostname",
		},
		{
			name: "when hostnamectl set-hostname errors",
			setupMock: func() *mocks.MockManager {
				mock := mocks.NewPlainMockManager(s.ctrl)
				mock.EXPECT().
					RunCmd("hostnamectl", []string{"hostname"}).
					Return("old-hostname\n", nil)
				mock.EXPECT().
					RunCmd("hostnamectl", []string{"set-hostname", "new-hostname"}).
					Return("", fmt.Errorf("permission denied"))
				return mock
			},
			hostname:    "new-hostname",
			wantErr:     true,
			errContains: "failed to set hostname",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mock := tc.setupMock()

			p := host.NewDebianProvider(mock)
			result, err := p.SetHostname(tc.hostname)

			if tc.wantErr {
				s.Error(err)
				s.Contains(err.Error(), tc.errContains)
			} else {
				s.NoError(err)
				tc.validateFunc(result)
			}
		})
	}
}

func TestDebianSetHostnamePublicTestSuite(t *testing.T) {
	suite.Run(t, new(DebianSetHostnamePublicTestSuite))
}
```

Note: add `"fmt"` to imports for the test.

- [ ] **Step 3: Run test to verify it fails**

Run: `go test -run TestDebianSetHostnamePublicTestSuite -v ./internal/provider/node/host/...`
Expected: FAIL — `SetHostname` not defined.

- [ ] **Step 4: Write implementation**

Create `internal/provider/node/host/debian_set_hostname.go`:

```go
package host

import (
	"fmt"
	"strings"
)

// SetHostname sets the system hostname using hostnamectl.
// It checks the current hostname first and returns Changed: false
// if the hostname is already set to the requested value.
func (u *Debian) SetHostname(
	name string,
) (*SetHostnameResult, error) {
	current, err := u.execManager.RunCmd("hostnamectl", []string{"hostname"})
	if err != nil {
		return nil, fmt.Errorf("failed to get current hostname: %w", err)
	}

	if strings.TrimSpace(current) == name {
		return &SetHostnameResult{Changed: false}, nil
	}

	if _, err := u.execManager.RunCmd("hostnamectl", []string{"set-hostname", name}); err != nil {
		return nil, fmt.Errorf("failed to set hostname: %w", err)
	}

	return &SetHostnameResult{Changed: true}, nil
}
```

- [ ] **Step 5: Run tests**

Run: `go test -run TestDebianSetHostnamePublicTestSuite -v ./internal/provider/node/host/...`
Expected: PASS

- [ ] **Step 6: Commit**

```
feat(host): add Debian SetHostname via hostnamectl
```

---

### Task 3: Provider — Darwin, Linux, and DebianDocker Stubs

**Files:**

- Create: `internal/provider/node/host/darwin_set_hostname.go`
- Create: `internal/provider/node/host/linux_set_hostname.go`
- Modify: existing Darwin/Linux test files to add SetHostname test

For DebianDocker: the host provider is not container-aware the same way DNS is.
The container check happens at agent_setup time — if `platform.IsContainer()`,
pass a nil `execManager` or use the existing pattern where the Debian provider
is still used but `hostnamectl` will fail naturally inside Docker. The cleaner
approach: since `hostnamectl` won't exist in Docker containers, the Debian
provider's `SetHostname` will return an error from `RunCmd`. This is acceptable —
the error message will be clear (`failed to get current hostname: exec:
"hostnamectl": executable file not found`). No separate DebianDocker host
provider is needed.

- [ ] **Step 1: Create Darwin stub**

Create `internal/provider/node/host/darwin_set_hostname.go`:

```go
package host

import (
	"fmt"

	"github.com/retr0h/osapi/internal/provider"
)

// SetHostname returns ErrUnsupported on Darwin.
// Darwin is a development platform only; mutations are not supported.
func (d *Darwin) SetHostname(
	_ string,
) (*SetHostnameResult, error) {
	return nil, fmt.Errorf("host: %w", provider.ErrUnsupported)
}
```

- [ ] **Step 2: Create Linux stub**

Create `internal/provider/node/host/linux_set_hostname.go`:

```go
package host

import (
	"fmt"

	"github.com/retr0h/osapi/internal/provider"
)

// SetHostname returns ErrUnsupported on generic Linux.
func (l *Linux) SetHostname(
	_ string,
) (*SetHostnameResult, error) {
	return nil, fmt.Errorf("host: %w", provider.ErrUnsupported)
}
```

- [ ] **Step 3: Add tests for both stubs**

Add `TestSetHostname` methods to the existing Darwin and Linux test suites
(in their respective `*_public_test.go` files). Follow the pattern in
`internal/provider/network/dns/linux_public_test.go` — single-row table
testing `ErrUnsupported`.

- [ ] **Step 4: Verify build and tests**

Run: `go build ./... && go test -v ./internal/provider/node/host/...`
Expected: PASS

- [ ] **Step 5: Commit**

```
feat(host): add Darwin and Linux SetHostname stubs (ErrUnsupported)
```

---

### Task 4: Job Operation and Agent Processor

**Files:**

- Modify: `pkg/sdk/client/operations.go` (add `OpNodeHostnameUpdate`)
- Modify: `internal/job/types.go` (add `OperationNodeHostnameUpdate`)
- Modify: `internal/agent/processor.go` (handle hostname update in processor)

- [ ] **Step 1: Add operation constant**

In `pkg/sdk/client/operations.go`, add after `OpNodeHostnameGet`:

```go
OpNodeHostnameUpdate JobOperation = "node.hostname.update"
```

In `internal/job/types.go`, add after `OperationNodeHostnameGet`:

```go
OperationNodeHostnameUpdate = client.OpNodeHostnameUpdate
```

- [ ] **Step 2: Update processor to handle hostname update**

In `internal/agent/processor.go`, change the `hostname` case to sub-dispatch:

```go
case "hostname":
	if req.Operation == job.OperationNodeHostnameUpdate {
		return setNodeHostname(hostProvider, req, logger)
	}
	return getNodeHostname(hostProvider, appConfig, logger)
```

Add the `setNodeHostname` function after `getNodeHostname`:

```go
// setNodeHostname sets the node hostname via the host provider.
func setNodeHostname(
	hostProvider nodeHost.Provider,
	req job.Request,
	logger *slog.Logger,
) (json.RawMessage, error) {
	logger.Debug("executing host.SetHostname")

	var data struct {
		Hostname string `json:"hostname"`
	}
	if err := json.Unmarshal(req.Data, &data); err != nil {
		return nil, fmt.Errorf("invalid hostname update data: %w", err)
	}

	result, err := hostProvider.SetHostname(data.Hostname)
	if err != nil {
		return nil, err
	}

	resp := map[string]interface{}{
		"hostname": data.Hostname,
		"changed":  result.Changed,
	}

	return json.Marshal(resp)
}
```

- [ ] **Step 3: Run tests**

Run: `go build ./... && go test -v ./internal/agent/...`
Expected: PASS

- [ ] **Step 4: Commit**

```
feat(agent): add hostname update operation to node processor
```

---

### Task 5: OpenAPI Spec — PUT /node/{hostname}/hostname

**Files:**

- Modify: `internal/controller/api/node/gen/api.yaml`

- [ ] **Step 1: Add PUT endpoint and schemas**

In `internal/controller/api/node/gen/api.yaml`, add `put:` under the existing
`/node/{hostname}/hostname` path (after the `get:` block, before the next path):

```yaml
    put:
      summary: Update node hostname
      description: Set the system hostname on the target node.
      tags:
        - node_operations
      operationId: PutNodeHostname
      security:
        - BearerAuth:
            - node:write
      parameters:
        - $ref: '#/components/parameters/Hostname'
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/HostnameUpdateRequest'
      responses:
        '202':
          description: Hostname update accepted.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HostnameUpdateCollectionResponse'
        '400':
          description: Invalid input.
          content:
            application/json:
              schema:
                $ref: '../../common/gen/api.yaml#/components/schemas/ErrorResponse'
        '401':
          description: Unauthorized - API key required
          content:
            application/json:
              schema:
                $ref: '../../common/gen/api.yaml#/components/schemas/ErrorResponse'
        '403':
          description: Forbidden - Insufficient permissions
          content:
            application/json:
              schema:
                $ref: '../../common/gen/api.yaml#/components/schemas/ErrorResponse'
        '500':
          description: Error updating hostname.
          content:
            application/json:
              schema:
                $ref: '../../common/gen/api.yaml#/components/schemas/ErrorResponse'
```

Add the request and response schemas to `components/schemas`:

```yaml
    HostnameUpdateRequest:
      type: object
      properties:
        hostname:
          type: string
          x-oapi-codegen-extra-tags:
            validate: required,min=1,max=253
          description: The new hostname to set.
          example: "web-01"
      required:
        - hostname

    HostnameUpdateResultItem:
      type: object
      properties:
        hostname:
          type: string
          description: The hostname of the agent.
        status:
          type: string
          enum: [ok, failed]
        changed:
          type: boolean
          description: Whether the hostname was actually modified.
        error:
          type: string
      required:
        - hostname
        - status

    HostnameUpdateCollectionResponse:
      type: object
      properties:
        job_id:
          type: string
          format: uuid
          description: The job ID used to process this request.
          example: "550e8400-e29b-41d4-a716-446655440000"
        results:
          type: array
          items:
            $ref: '#/components/schemas/HostnameUpdateResultItem'
      required:
        - results
```

- [ ] **Step 2: Regenerate code**

Run: `just generate`

- [ ] **Step 3: Verify generated code compiles**

Run: `go build ./...`
Expected: FAIL — new `PutNodeHostname` method required by `StrictServerInterface`
but not implemented yet.

- [ ] **Step 4: Commit**

```
feat(api): add PUT /node/{hostname}/hostname OpenAPI spec
```

---

### Task 6: API Handler — PutNodeHostname

**Files:**

- Create: `internal/controller/api/node/node_hostname_put.go`
- Create: `internal/controller/api/node/node_hostname_put_public_test.go`

- [ ] **Step 1: Write the handler**

Create `internal/controller/api/node/node_hostname_put.go` following the
`network_dns_put_by_interface.go` pattern exactly. The handler:

1. Validates the target hostname via `validateHostname()`
2. Validates the request body via `validation.Struct(request.Body)`
3. Checks `job.IsBroadcastTarget()` and routes accordingly
4. Single target: calls `s.JobClient.Modify()` with category `"node"` and
   operation `job.OperationNodeHostnameUpdate`
5. Broadcast: calls `s.JobClient.ModifyBroadcast()` with same
6. Returns 202 with `HostnameUpdateCollectionResponse`

Data passed to the job:

```go
data := map[string]any{
	"hostname": request.Body.Hostname,
}
```

- [ ] **Step 2: Write tests**

Create `internal/controller/api/node/node_hostname_put_public_test.go` with
table-driven tests covering: success (single target), success (broadcast),
validation error (empty hostname), bad target hostname, and job client error.
Follow the existing test patterns in
`internal/controller/api/node/node_hostname_get_public_test.go`.

Include `TestPutNodeHostnameHTTP` and `TestPutNodeHostnameRBACHTTP` methods
for wiring tests through the full Echo middleware stack.

- [ ] **Step 3: Run tests**

Run: `go test -v ./internal/controller/api/node/...`
Expected: PASS

- [ ] **Step 4: Commit**

```
feat(api): add PutNodeHostname handler with broadcast support
```

---

### Task 7: Server Wiring and Permissions

**Files:**

- Modify: `internal/controller/api/handler.go` (if handler registration needs
  updating — check if the node handler auto-registers all methods)
- Modify: `internal/config/permissions.go` or equivalent — add `node:write`
  permission to the `admin` and `write` roles

- [ ] **Step 1: Check if handler registration is automatic**

The node handler already implements `StrictServerInterface`. After regenerating,
the new `PutNodeHostname` method is automatically included in the handler. Check
if any explicit route registration is needed.

- [ ] **Step 2: Add node:write permission to roles**

Check where permissions are defined (likely in the security/auth middleware
configuration or in `internal/config/`). Add `node:write` to the `admin` and
`write` roles. This is the permission declared in the OpenAPI spec's
`BearerAuth` security for the PUT endpoint.

- [ ] **Step 3: Verify build and tests**

Run: `go build ./... && go test -v ./internal/controller/api/...`
Expected: PASS

- [ ] **Step 4: Commit**

```
feat(auth): add node:write permission for hostname update
```

---

### Task 8: SDK — SetHostname Method

**Files:**

- Modify: `pkg/sdk/client/node.go` (add `SetHostname` method)
- Modify: `pkg/sdk/client/node_types.go` (add `HostnameUpdateResult` type)
- Modify: `pkg/sdk/client/gen/` (regenerate SDK client from combined spec)

- [ ] **Step 1: Regenerate SDK client**

Run: `go generate ./pkg/sdk/client/gen/...`

- [ ] **Step 2: Add SDK types**

In `pkg/sdk/client/node_types.go`, add:

```go
// HostnameUpdateResult represents a hostname update result from a single agent.
type HostnameUpdateResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
	Changed  bool   `json:"changed"`
}
```

Add the `gen→SDK` conversion function:

```go
func hostnameUpdateCollectionFromGen(
	r *gen.HostnameUpdateCollectionResponse,
) Collection[HostnameUpdateResult] {
	// ... follow existing pattern from hostnameCollectionFromGen
}
```

- [ ] **Step 3: Add SetHostname method**

In `pkg/sdk/client/node.go`, add:

```go
// SetHostname updates the hostname on the target node.
func (s *NodeService) SetHostname(
	ctx context.Context,
	target string,
	name string,
) (*Response[Collection[HostnameUpdateResult]], error) {
	body := gen.HostnameUpdateRequest{
		Hostname: name,
	}

	resp, err := s.client.PutNodeHostnameWithResponse(ctx, target, body)
	if err != nil {
		return nil, fmt.Errorf("set hostname: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON400,
		resp.JSON401,
		resp.JSON403,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON202 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(
		hostnameUpdateCollectionFromGen(resp.JSON202),
		resp.Body,
	), nil
}
```

- [ ] **Step 4: Run tests**

Run: `go build ./... && go test -v ./pkg/sdk/client/...`
Expected: PASS

- [ ] **Step 5: Commit**

```
feat(sdk): add SetHostname method to NodeService
```

---

### Task 9: CLI — client node hostname update

**Files:**

- Create: `cmd/client_node_hostname_update.go`

- [ ] **Step 1: Create the CLI command**

Create `cmd/client_node_hostname_update.go` following the pattern of
`cmd/client_node_network_dns_update.go`:

```go
var clientNodeHostnameUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the node's hostname",
	Long:  `Set a new hostname on the target node using hostnamectl.`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")
		name, _ := cmd.Flags().GetString("name")

		resp, err := sdkClient.Node.SetHostname(ctx, host, name)
		if err != nil {
			cli.HandleError(err, logger)
			return
		}

		if jsonOutput {
			fmt.Println(string(resp.RawJSON()))
			return
		}

		if resp.Data.JobID != "" {
			fmt.Println()
			cli.PrintKV("Job ID", resp.Data.JobID)
		}

		results := make([]cli.ResultRow, 0, len(resp.Data.Results))
		for _, r := range resp.Data.Results {
			var errPtr *string
			if r.Error != "" {
				errPtr = &r.Error
			}
			results = append(results, cli.ResultRow{
				Hostname: r.Hostname,
				Error:    errPtr,
				Fields:   []string{fmt.Sprintf("%t", r.Changed)},
			})
		}
		headers, rows := cli.BuildBroadcastTable(results, []string{"CHANGED"})
		cli.PrintCompactTable([]cli.Section{{Headers: headers, Rows: rows}})
	},
}

func init() {
	clientNodeHostnameCmd.AddCommand(clientNodeHostnameUpdateCmd)
	clientNodeHostnameUpdateCmd.Flags().String("name", "", "New hostname to set (required)")
	_ = clientNodeHostnameUpdateCmd.MarkFlagRequired("name")
}
```

- [ ] **Step 2: Verify it works**

Run: `go build ./... && go run main.go client node hostname update --help`
Expected: Shows help with `--name` and `--target` flags.

- [ ] **Step 3: Commit**

```
feat(cli): add client node hostname update command
```

---

### Task 10: Documentation Updates

**Files:**

- Modify: `docs/docs/sidebar/usage/cli/client/node/hostname.md`
- Modify: `docs/docs/sidebar/sdk/client/node.md`
- Modify: `docs/docs/sidebar/sdk/orchestrator/operations/node-hostname.md`
- Modify: `examples/sdk/client/node.go`
- Modify: `docs/docs/sidebar/usage/configuration.md` (add `node:write` to
  permissions table)

- [ ] **Step 1: Update CLI docs**

In `docs/docs/sidebar/usage/cli/client/node/hostname.md`, add the update section
after the existing get examples:

```markdown
## Update

Set the hostname on the target node:

\`\`\`bash
$ osapi client node hostname update --name web-01

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  CHANGED
  web-01    true
\`\`\`

When targeting all hosts:

\`\`\`bash
$ osapi client node hostname update --name web-01 --target _all
\`\`\`

### Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `--name`       | New hostname to set (required)                           |         |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
```

- [ ] **Step 2: Update SDK docs**

In `docs/docs/sidebar/sdk/client/node.md`, add `SetHostname` to the Node Info
methods table and add a usage example.

In `docs/docs/sidebar/sdk/orchestrator/operations/node-hostname.md`, add a
section for `node.hostname.update`.

- [ ] **Step 3: Update SDK example**

In `examples/sdk/client/node.go`, add a `SetHostname` example after the
Hostname get block:

```go
// Set hostname (uncomment to run — this mutates the system)
// setResp, err := c.Node.SetHostname(ctx, "web-01", "new-hostname")
// if err != nil {
//     log.Fatalf("set hostname: %v", err)
// }
// fmt.Printf("Set hostname changed: %t\n", setResp.Data.Results[0].Changed)
```

- [ ] **Step 4: Update permissions docs**

In `docs/docs/sidebar/usage/configuration.md`, add `node:write` to the `admin`
and `write` role permission lists.

- [ ] **Step 5: Commit**

```
docs: add hostname update to CLI, SDK, and configuration docs
```

---

### Task 11: Integration Test

**Files:**

- Modify: `test/integration/node_test.go`

- [ ] **Step 1: Add hostname update integration test**

Add a test case to the existing node test suite. Guard with `skipWrite(s.T())`
since this is a mutation:

```go
{
	name: "updates hostname",
	args: []string{"client", "node", "hostname", "update", "--name", currentHostname, "--json"},
	validateFunc: func(stdout string, exitCode int) {
		skipWrite(s.T())
		s.Require().Equal(0, exitCode)
		// Parse JSON response and verify changed field
	},
},
```

Use the current hostname to ensure idempotency (Changed: false).

- [ ] **Step 2: Commit**

```
test(integration): add hostname update integration test
```

---

### Task 12: Format and Final Verification

- [ ] **Step 1: Format**

Run: `just go::fmt`

- [ ] **Step 2: Lint**

Run: `just go::vet`
Expected: 0 issues

- [ ] **Step 3: Full test suite**

Run: `just go::unit`
Expected: PASS

- [ ] **Step 4: Commit any formatting changes**

```
style: format hostname update files
```
