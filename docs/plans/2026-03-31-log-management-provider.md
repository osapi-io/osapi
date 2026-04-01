# Log Management Provider Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add read-only log viewing to OSAPI via `journalctl --output=json`, with optional filtering by lines, time range, and priority.

**Architecture:** Direct provider at `internal/provider/node/log/` using `exec.Manager` to run `journalctl`. Two API endpoints: query all journal entries and query by unit name. Read-only with `log:read` permission added to all built-in roles.

**Tech Stack:** Go, exec.Manager, journalctl JSON output, oapi-codegen strict-server

---

## File Structure

### Provider Layer
- Create: `internal/provider/node/log/types.go` — Provider interface + domain types
- Create: `internal/provider/node/log/debian.go` — journalctl implementation
- Create: `internal/provider/node/log/debian_query.go` — shared query logic
- Create: `internal/provider/node/log/darwin.go` — macOS stub
- Create: `internal/provider/node/log/linux.go` — generic Linux stub
- Create: `internal/provider/node/log/mocks/generate.go` — mockgen directive
- Test: `internal/provider/node/log/debian_public_test.go`
- Test: `internal/provider/node/log/darwin_public_test.go`
- Test: `internal/provider/node/log/linux_public_test.go`

### Agent Layer
- Create: `internal/agent/processor_log.go` — log operation dispatcher
- Modify: `internal/agent/processor.go` — add `log` case + logProvider param
- Modify: `cmd/agent_setup.go` — create log provider factory, wire into registry
- Test: `internal/agent/processor_log_public_test.go`

### API Layer
- Create: `internal/controller/api/node/log/gen/api.yaml` — OpenAPI spec
- Create: `internal/controller/api/node/log/gen/cfg.yaml` — oapi-codegen config
- Create: `internal/controller/api/node/log/gen/generate.go` — go:generate
- Create: `internal/controller/api/node/log/types.go` — handler struct
- Create: `internal/controller/api/node/log/log.go` — New(), compile-time check
- Create: `internal/controller/api/node/log/validate.go` — validateHostname
- Create: `internal/controller/api/node/log/log_query_get.go` — query handler
- Create: `internal/controller/api/node/log/log_unit_get.go` — query unit handler
- Create: `internal/controller/api/node/log/handler.go` — Handler() registration
- Modify: `cmd/controller_setup.go` — register log handler
- Test: `internal/controller/api/node/log/log_query_get_public_test.go`
- Test: `internal/controller/api/node/log/log_unit_get_public_test.go`
- Test: `internal/controller/api/node/log/handler_public_test.go`

### Operations & Permissions
- Modify: `pkg/sdk/client/operations.go` — add log operation constants
- Modify: `internal/job/types.go` — add log operation aliases
- Modify: `pkg/sdk/client/permissions.go` — add `PermLogRead`
- Modify: `internal/authtoken/permissions.go` — add `PermLogRead` to all roles

### SDK Layer
- Create: `pkg/sdk/client/log.go` — LogService methods
- Create: `pkg/sdk/client/log_types.go` — SDK result types + conversions
- Modify: `pkg/sdk/client/osapi.go` — add Log field
- Test: `pkg/sdk/client/log_public_test.go`
- Test: `pkg/sdk/client/log_types_public_test.go`

### CLI Layer
- Create: `cmd/client_node_log.go` — parent command
- Create: `cmd/client_node_log_query.go` — query subcommand
- Create: `cmd/client_node_log_unit.go` — query-unit subcommand

### Documentation
- Create: `docs/docs/sidebar/features/log-management.md` — feature page
- Create: `docs/docs/sidebar/usage/cli/client/node/log/log.md` — CLI landing
- Create: `docs/docs/sidebar/usage/cli/client/node/log/query.md` — query CLI doc
- Create: `docs/docs/sidebar/usage/cli/client/node/log/unit.md` — unit CLI doc
- Create: `docs/docs/sidebar/sdk/client/operations/log.md` — SDK doc
- Create: `examples/sdk/client/log.go` — SDK example
- Modify: `docs/docs/sidebar/features/features.md` — add log to table
- Modify: `docs/docs/sidebar/features/authentication.md` — add log:read to roles
- Modify: `docs/docs/sidebar/usage/configuration.md` — add log:read to permissions
- Modify: `docs/docs/sidebar/architecture/architecture.md` — add log feature link
- Modify: `docs/docs/sidebar/architecture/api-guidelines.md` — add log endpoints
- Modify: `docs/docusaurus.config.ts` — add SDK dropdown + features dropdown

### Integration Test
- Create: `test/integration/log_test.go` — smoke test

---

### Task 1: Provider Interface and Types

**Files:**
- Create: `internal/provider/node/log/types.go`

- [ ] **Step 1: Create provider interface and types**

```go
// Copyright (c) 2026 John Dewey

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

// Package log provides log viewing operations.
package log

import (
	"context"
)

// Provider implements log viewing operations.
type Provider interface {
	// Query returns journal entries with optional filtering.
	Query(ctx context.Context, opts QueryOpts) ([]Entry, error)
	// QueryUnit returns journal entries for a specific systemd unit.
	QueryUnit(ctx context.Context, unit string, opts QueryOpts) ([]Entry, error)
}

// QueryOpts contains optional filters for log queries.
type QueryOpts struct {
	Lines    int    `json:"lines,omitempty"`
	Since    string `json:"since,omitempty"`
	Priority string `json:"priority,omitempty"`
}

// Entry represents a single journal entry.
type Entry struct {
	Timestamp string `json:"timestamp"`
	Unit      string `json:"unit,omitempty"`
	Priority  string `json:"priority"`
	Message   string `json:"message"`
	PID       int    `json:"pid,omitempty"`
	Hostname  string `json:"hostname,omitempty"`
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/provider/node/log/...`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add internal/provider/node/log/types.go
git commit -m "feat(log): add provider interface and types"
```

---

### Task 2: Platform Stubs (Darwin + Linux)

**Files:**
- Create: `internal/provider/node/log/darwin.go`
- Create: `internal/provider/node/log/linux.go`
- Test: `internal/provider/node/log/darwin_public_test.go`
- Test: `internal/provider/node/log/linux_public_test.go`

- [ ] **Step 1: Write darwin stub tests**

```go
// Copyright (c) 2026 John Dewey

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package log_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	logProv "github.com/retr0h/osapi/internal/provider/node/log"
	"github.com/retr0h/osapi/internal/provider"
)

type DarwinPublicTestSuite struct {
	suite.Suite

	provider *logProv.Darwin
}

func (s *DarwinPublicTestSuite) SetupTest() {
	s.provider = logProv.NewDarwinProvider()
}

func (s *DarwinPublicTestSuite) TestQuery() {
	_, err := s.provider.Query(context.Background(), logProv.QueryOpts{})

	assert.ErrorIs(s.T(), err, provider.ErrUnsupported)
}

func (s *DarwinPublicTestSuite) TestQueryUnit() {
	_, err := s.provider.QueryUnit(context.Background(), "nginx.service", logProv.QueryOpts{})

	assert.ErrorIs(s.T(), err, provider.ErrUnsupported)
}

func TestDarwinPublicTestSuite(t *testing.T) {
	suite.Run(t, new(DarwinPublicTestSuite))
}
```

- [ ] **Step 2: Write linux stub tests**

```go
// Copyright (c) 2026 John Dewey

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package log_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	logProv "github.com/retr0h/osapi/internal/provider/node/log"
	"github.com/retr0h/osapi/internal/provider"
)

type LinuxPublicTestSuite struct {
	suite.Suite

	provider *logProv.Linux
}

func (s *LinuxPublicTestSuite) SetupTest() {
	s.provider = logProv.NewLinuxProvider()
}

func (s *LinuxPublicTestSuite) TestQuery() {
	_, err := s.provider.Query(context.Background(), logProv.QueryOpts{})

	assert.ErrorIs(s.T(), err, provider.ErrUnsupported)
}

func (s *LinuxPublicTestSuite) TestQueryUnit() {
	_, err := s.provider.QueryUnit(context.Background(), "nginx.service", logProv.QueryOpts{})

	assert.ErrorIs(s.T(), err, provider.ErrUnsupported)
}

func TestLinuxPublicTestSuite(t *testing.T) {
	suite.Run(t, new(LinuxPublicTestSuite))
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `go test -v ./internal/provider/node/log/...`
Expected: FAIL — Darwin and Linux types don't exist yet

- [ ] **Step 4: Implement darwin stub**

```go
// Copyright (c) 2026 John Dewey

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package log

import (
	"context"
	"fmt"

	"github.com/retr0h/osapi/internal/provider"
)

// Darwin implements the Provider interface for Darwin (macOS).
// All methods return ErrUnsupported as log viewing is not available on macOS.
type Darwin struct{}

// NewDarwinProvider factory to create a new Darwin instance.
func NewDarwinProvider() *Darwin {
	return &Darwin{}
}

// Query returns ErrUnsupported on Darwin.
func (d *Darwin) Query(
	_ context.Context,
	_ QueryOpts,
) ([]Entry, error) {
	return nil, fmt.Errorf("log: %w", provider.ErrUnsupported)
}

// QueryUnit returns ErrUnsupported on Darwin.
func (d *Darwin) QueryUnit(
	_ context.Context,
	_ string,
	_ QueryOpts,
) ([]Entry, error) {
	return nil, fmt.Errorf("log: %w", provider.ErrUnsupported)
}
```

- [ ] **Step 5: Implement linux stub**

```go
// Copyright (c) 2026 John Dewey

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package log

import (
	"context"
	"fmt"

	"github.com/retr0h/osapi/internal/provider"
)

// Linux implements the Provider interface for generic Linux.
// All methods return ErrUnsupported as this is a generic Linux stub.
type Linux struct{}

// NewLinuxProvider factory to create a new Linux instance.
func NewLinuxProvider() *Linux {
	return &Linux{}
}

// Query returns ErrUnsupported on generic Linux.
func (l *Linux) Query(
	_ context.Context,
	_ QueryOpts,
) ([]Entry, error) {
	return nil, fmt.Errorf("log: %w", provider.ErrUnsupported)
}

// QueryUnit returns ErrUnsupported on generic Linux.
func (l *Linux) QueryUnit(
	_ context.Context,
	_ string,
	_ QueryOpts,
) ([]Entry, error) {
	return nil, fmt.Errorf("log: %w", provider.ErrUnsupported)
}
```

- [ ] **Step 6: Run tests to verify they pass**

Run: `go test -v ./internal/provider/node/log/...`
Expected: PASS — all 4 tests pass

- [ ] **Step 7: Commit**

```bash
git add internal/provider/node/log/darwin.go internal/provider/node/log/linux.go \
  internal/provider/node/log/darwin_public_test.go internal/provider/node/log/linux_public_test.go
git commit -m "feat(log): add darwin and linux provider stubs"
```

---

### Task 3: Debian Provider Implementation

**Files:**
- Create: `internal/provider/node/log/debian.go`
- Create: `internal/provider/node/log/debian_query.go`
- Create: `internal/provider/node/log/mocks/generate.go`
- Test: `internal/provider/node/log/debian_public_test.go`

- [ ] **Step 1: Create mock generator**

```go
// Copyright (c) 2026 John Dewey

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

// Package mocks provides mock implementations for testing.
package mocks

//go:generate go tool github.com/golang/mock/mockgen -source=../types.go -destination=provider.gen.go -package=mocks
```

- [ ] **Step 2: Generate mocks**

Run: `go generate ./internal/provider/node/log/mocks/...`
Expected: generates `provider.gen.go`

- [ ] **Step 3: Write debian provider tests**

```go
// Copyright (c) 2026 John Dewey

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package log_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/exec"
	execMocks "github.com/retr0h/osapi/internal/exec/mocks"
	logProv "github.com/retr0h/osapi/internal/provider/node/log"
)

type DebianPublicTestSuite struct {
	suite.Suite

	mockCtrl        *gomock.Controller
	mockExecManager *execMocks.MockManager
	provider        *logProv.Debian
	ctx             context.Context
	logger          *slog.Logger
}

func (s *DebianPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockExecManager = execMocks.NewMockManager(s.mockCtrl)
	s.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	s.provider = logProv.NewDebianProvider(s.logger, s.mockExecManager)
	s.ctx = context.Background()
}

func (s *DebianPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *DebianPublicTestSuite) TestQuery() {
	journalLine1 := `{"__REALTIME_TIMESTAMP":"1711929045123456","SYSLOG_IDENTIFIER":"nginx","PRIORITY":"6","MESSAGE":"Started nginx","_PID":"1234","_HOSTNAME":"web-01"}`
	journalLine2 := `{"__REALTIME_TIMESTAMP":"1711929046000000","SYSLOG_IDENTIFIER":"sshd","PRIORITY":"4","MESSAGE":"Connection closed","_PID":"5678","_HOSTNAME":"web-01"}`
	journalOutput := journalLine1 + "\n" + journalLine2 + "\n"

	tests := []struct {
		name         string
		opts         logProv.QueryOpts
		setupMock    func()
		validateFunc func(entries []logProv.Entry, err error)
	}{
		{
			name: "default query with no options",
			opts: logProv.QueryOpts{},
			setupMock: func() {
				s.mockExecManager.EXPECT().
					RunCmd("journalctl", []string{"--output=json", "-n", "100"}).
					Return(journalOutput, nil)
			},
			validateFunc: func(entries []logProv.Entry, err error) {
				assert.NoError(s.T(), err)
				assert.Len(s.T(), entries, 2)
				assert.Equal(s.T(), "nginx", entries[0].Unit)
				assert.Equal(s.T(), "info", entries[0].Priority)
				assert.Equal(s.T(), "Started nginx", entries[0].Message)
				assert.Equal(s.T(), 1234, entries[0].PID)
				assert.Equal(s.T(), "web-01", entries[0].Hostname)
				assert.Equal(s.T(), "sshd", entries[1].Unit)
				assert.Equal(s.T(), "warning", entries[1].Priority)
			},
		},
		{
			name: "query with all options",
			opts: logProv.QueryOpts{
				Lines:    50,
				Since:    "1 hour ago",
				Priority: "err",
			},
			setupMock: func() {
				s.mockExecManager.EXPECT().
					RunCmd("journalctl", []string{
						"--output=json",
						"-n", "50",
						"--since=1 hour ago",
						"--priority=err",
					}).
					Return(journalOutput, nil)
			},
			validateFunc: func(entries []logProv.Entry, err error) {
				assert.NoError(s.T(), err)
				assert.Len(s.T(), entries, 2)
			},
		},
		{
			name: "query with custom lines",
			opts: logProv.QueryOpts{Lines: 10},
			setupMock: func() {
				s.mockExecManager.EXPECT().
					RunCmd("journalctl", []string{"--output=json", "-n", "10"}).
					Return(journalOutput, nil)
			},
			validateFunc: func(entries []logProv.Entry, err error) {
				assert.NoError(s.T(), err)
				assert.Len(s.T(), entries, 2)
			},
		},
		{
			name: "exec error",
			opts: logProv.QueryOpts{},
			setupMock: func() {
				s.mockExecManager.EXPECT().
					RunCmd("journalctl", []string{"--output=json", "-n", "100"}).
					Return("", fmt.Errorf("exec failed"))
			},
			validateFunc: func(entries []logProv.Entry, err error) {
				assert.Error(s.T(), err)
				assert.Nil(s.T(), entries)
				assert.Contains(s.T(), err.Error(), "log: query")
			},
		},
		{
			name: "empty output",
			opts: logProv.QueryOpts{},
			setupMock: func() {
				s.mockExecManager.EXPECT().
					RunCmd("journalctl", []string{"--output=json", "-n", "100"}).
					Return("", nil)
			},
			validateFunc: func(entries []logProv.Entry, err error) {
				assert.NoError(s.T(), err)
				assert.Empty(s.T(), entries)
			},
		},
		{
			name: "malformed JSON line skipped",
			opts: logProv.QueryOpts{},
			setupMock: func() {
				output := "not json\n" + journalLine1 + "\n"
				s.mockExecManager.EXPECT().
					RunCmd("journalctl", []string{"--output=json", "-n", "100"}).
					Return(output, nil)
			},
			validateFunc: func(entries []logProv.Entry, err error) {
				assert.NoError(s.T(), err)
				assert.Len(s.T(), entries, 1)
				assert.Equal(s.T(), "nginx", entries[0].Unit)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setupMock()
			entries, err := s.provider.Query(s.ctx, tc.opts)
			tc.validateFunc(entries, err)
		})
	}
}

func (s *DebianPublicTestSuite) TestQueryUnit() {
	journalLine := `{"__REALTIME_TIMESTAMP":"1711929045123456","SYSLOG_IDENTIFIER":"nginx","PRIORITY":"6","MESSAGE":"Started nginx","_PID":"1234","_HOSTNAME":"web-01"}`

	tests := []struct {
		name         string
		unit         string
		opts         logProv.QueryOpts
		setupMock    func()
		validateFunc func(entries []logProv.Entry, err error)
	}{
		{
			name: "query unit with defaults",
			unit: "nginx.service",
			opts: logProv.QueryOpts{},
			setupMock: func() {
				s.mockExecManager.EXPECT().
					RunCmd("journalctl", []string{
						"--output=json",
						"-u", "nginx.service",
						"-n", "100",
					}).
					Return(journalLine+"\n", nil)
			},
			validateFunc: func(entries []logProv.Entry, err error) {
				assert.NoError(s.T(), err)
				assert.Len(s.T(), entries, 1)
				assert.Equal(s.T(), "nginx", entries[0].Unit)
			},
		},
		{
			name: "query unit with all options",
			unit: "sshd.service",
			opts: logProv.QueryOpts{
				Lines:    25,
				Since:    "2026-03-31",
				Priority: "warning",
			},
			setupMock: func() {
				s.mockExecManager.EXPECT().
					RunCmd("journalctl", []string{
						"--output=json",
						"-u", "sshd.service",
						"-n", "25",
						"--since=2026-03-31",
						"--priority=warning",
					}).
					Return(journalLine+"\n", nil)
			},
			validateFunc: func(entries []logProv.Entry, err error) {
				assert.NoError(s.T(), err)
				assert.Len(s.T(), entries, 1)
			},
		},
		{
			name: "exec error",
			unit: "nginx.service",
			opts: logProv.QueryOpts{},
			setupMock: func() {
				s.mockExecManager.EXPECT().
					RunCmd("journalctl", gomock.Any()).
					Return("", fmt.Errorf("exec failed"))
			},
			validateFunc: func(entries []logProv.Entry, err error) {
				assert.Error(s.T(), err)
				assert.Nil(s.T(), entries)
				assert.Contains(s.T(), err.Error(), "log: query unit")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setupMock()
			entries, err := s.provider.QueryUnit(s.ctx, tc.unit, tc.opts)
			tc.validateFunc(entries, err)
		})
	}
}

func TestDebianPublicTestSuite(t *testing.T) {
	suite.Run(t, new(DebianPublicTestSuite))
}
```

- [ ] **Step 4: Run tests to verify they fail**

Run: `go test -v ./internal/provider/node/log/...`
Expected: FAIL — Debian type doesn't exist yet

- [ ] **Step 5: Implement debian.go**

```go
// Copyright (c) 2026 John Dewey

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package log

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/retr0h/osapi/internal/exec"
	"github.com/retr0h/osapi/internal/provider"
)

// Compile-time checks.
var (
	_ Provider             = (*Debian)(nil)
	_ provider.FactsSetter = (*Debian)(nil)
)

// Debian implements the Provider interface for Debian-family systems.
type Debian struct {
	provider.FactsAware
	logger      *slog.Logger
	execManager exec.Manager
}

// NewDebianProvider factory to create a new Debian instance.
func NewDebianProvider(
	logger *slog.Logger,
	execManager exec.Manager,
) *Debian {
	return &Debian{
		logger:      logger.With(slog.String("subsystem", "provider.log")),
		execManager: execManager,
	}
}

// Query returns journal entries with optional filtering.
func (d *Debian) Query(
	_ context.Context,
	opts QueryOpts,
) ([]Entry, error) {
	d.logger.Debug("executing log.Query")

	args := buildArgs(opts)

	output, err := d.execManager.RunCmd("journalctl", args)
	if err != nil {
		return nil, fmt.Errorf("log: query: %w", err)
	}

	return parseJournalOutput(output, d.logger), nil
}

// QueryUnit returns journal entries for a specific systemd unit.
func (d *Debian) QueryUnit(
	_ context.Context,
	unit string,
	opts QueryOpts,
) ([]Entry, error) {
	d.logger.Debug("executing log.QueryUnit",
		slog.String("unit", unit),
	)

	args := buildUnitArgs(unit, opts)

	output, err := d.execManager.RunCmd("journalctl", args)
	if err != nil {
		return nil, fmt.Errorf("log: query unit: %w", err)
	}

	return parseJournalOutput(output, d.logger), nil
}
```

- [ ] **Step 6: Implement debian_query.go**

```go
// Copyright (c) 2026 John Dewey

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package log

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

// journalEntry represents the raw JSON output from journalctl --output=json.
type journalEntry struct {
	RealtimeTimestamp string `json:"__REALTIME_TIMESTAMP"`
	SyslogIdentifier string `json:"SYSLOG_IDENTIFIER"`
	Priority         string `json:"PRIORITY"`
	Message          string `json:"MESSAGE"`
	PID              string `json:"_PID"`
	Hostname         string `json:"_HOSTNAME"`
}

// priorityNames maps journalctl numeric priorities to human-readable names.
var priorityNames = map[string]string{
	"0": "emerg",
	"1": "alert",
	"2": "crit",
	"3": "err",
	"4": "warning",
	"5": "notice",
	"6": "info",
	"7": "debug",
}

// buildArgs constructs the journalctl command arguments for a general query.
func buildArgs(
	opts QueryOpts,
) []string {
	args := []string{"--output=json"}

	lines := opts.Lines
	if lines <= 0 {
		lines = 100
	}
	args = append(args, "-n", fmt.Sprintf("%d", lines))

	if opts.Since != "" {
		args = append(args, "--since="+opts.Since)
	}

	if opts.Priority != "" {
		args = append(args, "--priority="+opts.Priority)
	}

	return args
}

// buildUnitArgs constructs the journalctl command arguments for a unit query.
func buildUnitArgs(
	unit string,
	opts QueryOpts,
) []string {
	args := []string{"--output=json", "-u", unit}

	lines := opts.Lines
	if lines <= 0 {
		lines = 100
	}
	args = append(args, "-n", fmt.Sprintf("%d", lines))

	if opts.Since != "" {
		args = append(args, "--since="+opts.Since)
	}

	if opts.Priority != "" {
		args = append(args, "--priority="+opts.Priority)
	}

	return args
}

// parseJournalOutput parses the JSON lines output from journalctl.
func parseJournalOutput(
	output string,
	logger *slog.Logger,
) []Entry {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var entries []Entry

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var je journalEntry
		if err := json.Unmarshal([]byte(line), &je); err != nil {
			logger.Debug("skipping malformed journal line",
				slog.String("error", err.Error()),
			)
			continue
		}

		entries = append(entries, journalEntryToEntry(je))
	}

	return entries
}

// journalEntryToEntry converts a raw journal entry to the provider Entry type.
func journalEntryToEntry(
	je journalEntry,
) Entry {
	ts := parseTimestamp(je.RealtimeTimestamp)
	priority := priorityNames[je.Priority]
	if priority == "" {
		priority = je.Priority
	}

	var pid int
	if je.PID != "" {
		pid, _ = strconv.Atoi(je.PID)
	}

	return Entry{
		Timestamp: ts,
		Unit:      je.SyslogIdentifier,
		Priority:  priority,
		Message:   je.Message,
		PID:       pid,
		Hostname:  je.Hostname,
	}
}

// parseTimestamp converts a journalctl __REALTIME_TIMESTAMP (microseconds
// since epoch) to RFC3339 format.
func parseTimestamp(
	usec string,
) string {
	us, err := strconv.ParseInt(usec, 10, 64)
	if err != nil {
		return usec
	}

	t := time.UnixMicro(us).UTC()

	return t.Format(time.RFC3339Nano)
}
```

- [ ] **Step 7: Run tests to verify they pass**

Run: `go test -v ./internal/provider/node/log/...`
Expected: PASS — all tests pass

- [ ] **Step 8: Commit**

```bash
git add internal/provider/node/log/debian.go internal/provider/node/log/debian_query.go \
  internal/provider/node/log/debian_public_test.go \
  internal/provider/node/log/mocks/
git commit -m "feat(log): add debian provider with journalctl parsing"
```

---

### Task 4: Operations, Permissions, and Agent Wiring

**Files:**
- Modify: `pkg/sdk/client/operations.go` — add log operations
- Modify: `internal/job/types.go` — add log operation aliases
- Modify: `pkg/sdk/client/permissions.go` — add `PermLogRead`
- Modify: `internal/authtoken/permissions.go` — add `PermLogRead` to all roles
- Create: `internal/agent/processor_log.go` — log dispatcher
- Modify: `internal/agent/processor.go` — add `log` case + logProvider param
- Modify: `cmd/agent_setup.go` — create log provider, wire into registry
- Test: `internal/agent/processor_log_public_test.go`

- [ ] **Step 1: Add operation constants to SDK**

Add to `pkg/sdk/client/operations.go` after the Package operations block:

```go
// Log operations.
const (
	OpLogQuery     JobOperation = "node.log.query"
	OpLogQueryUnit JobOperation = "node.log.queryUnit"
)
```

- [ ] **Step 2: Add operation aliases to job types**

Add to `internal/job/types.go` after the Package operations block:

```go
// Log operations.
const (
	OperationLogQuery     = client.OpLogQuery
	OperationLogQueryUnit = client.OpLogQueryUnit
)
```

- [ ] **Step 3: Add permission constant to SDK**

Add to `pkg/sdk/client/permissions.go` after the Package permissions:

```go
	PermLogRead Permission = "log:read"
```

- [ ] **Step 4: Add permission to authtoken**

Add to `internal/authtoken/permissions.go`:

1. Add constant after PackageWrite:
```go
	PermLogRead = client.PermLogRead
```

2. Add to `AllPermissions` slice:
```go
	PermLogRead,
```

3. Add to admin role after `PermPackageWrite`:
```go
		PermLogRead,
```

4. Add to write role after `PermPackageWrite`:
```go
		PermLogRead,
```

5. Add to read role after `PermPackageRead`:
```go
		PermLogRead,
```

- [ ] **Step 5: Write agent processor tests**

Create `internal/agent/processor_log_public_test.go`. The tests exercise `processLogOperation` via the node processor's `log` case. Follow the same pattern as `processor_process_public_test.go` — table-driven tests with gomock for the log provider mock. Test cases:
- `log.query` with default opts (empty data)
- `log.query` with all options (lines, since, priority)
- `log.queryUnit` with unit name
- unsupported sub-operation (`log.invalid`)
- invalid operation format (`log` with no sub-op)
- nil log provider

The tests construct a `job.Request` with `Operation: "log.query"` (note: the node processor strips the base operation from the dotted format `"hostname.get"` → `"hostname"`, but for log the operation string passed to `processLogOperation` is already `"log.query"`, `"log.queryUnit"` etc. The node processor matches on `baseOperation` which is `"log"`, then delegates to `processLogOperation` which splits on `.` to get the sub-op).

- [ ] **Step 6: Run tests to verify they fail**

Run: `go test -run TestProcessorLogPublicTestSuite -v ./internal/agent/...`
Expected: FAIL — `processLogOperation` doesn't exist

- [ ] **Step 7: Implement processor_log.go**

Create `internal/agent/processor_log.go`:

```go
// Copyright (c) 2026 John Dewey

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/retr0h/osapi/internal/job"
	logProv "github.com/retr0h/osapi/internal/provider/node/log"
)

// processLogOperation dispatches log sub-operations.
func processLogOperation(
	logProvider logProv.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	if logProvider == nil {
		return nil, fmt.Errorf("log provider not available")
	}

	parts := strings.Split(jobRequest.Operation, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid log operation: %s", jobRequest.Operation)
	}
	subOp := parts[1]

	ctx := context.Background()

	switch subOp {
	case "query":
		return processLogQuery(ctx, logProvider, logger, jobRequest)
	case "queryUnit":
		return processLogQueryUnit(ctx, logProvider, logger, jobRequest)
	default:
		return nil, fmt.Errorf("unsupported log operation: %s", jobRequest.Operation)
	}
}

// processLogQuery handles the log.query operation.
func processLogQuery(
	ctx context.Context,
	logProvider logProv.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	logger.Debug("executing log.Query")

	var opts logProv.QueryOpts
	if jobRequest.Data != nil {
		_ = json.Unmarshal(jobRequest.Data, &opts)
	}

	result, err := logProvider.Query(ctx, opts)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processLogQueryUnit handles the log.queryUnit operation.
func processLogQueryUnit(
	ctx context.Context,
	logProvider logProv.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	logger.Debug("executing log.QueryUnit")

	var data struct {
		Unit     string `json:"unit"`
		logProv.QueryOpts
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal log query unit data: %w", err)
	}

	result, err := logProvider.QueryUnit(ctx, data.Unit, data.QueryOpts)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}
```

- [ ] **Step 8: Wire log into NewNodeProcessor**

Add `logProvider logProv.Provider` parameter to `NewNodeProcessor` in `internal/agent/processor.go`. Add the import:

```go
logProv "github.com/retr0h/osapi/internal/provider/node/log"
```

Add the case in the switch:

```go
		case "log":
			return processLogOperation(logProvider, logger, req)
```

- [ ] **Step 9: Create log provider factory in agent_setup.go**

Add import:
```go
logProv "github.com/retr0h/osapi/internal/provider/node/log"
```

Add factory function:
```go
// createLogProvider creates a platform-specific log provider. On Debian, the
// log provider reads journal entries via journalctl. In containers, journalctl
// is not available — returns ErrUnsupported. On other platforms, all operations
// return ErrUnsupported.
func createLogProvider(
	log *slog.Logger,
	execManager exec.Manager,
) logProv.Provider {
	plat := platform.Detect()

	switch plat {
	case "debian":
		if platform.IsContainer() {
			log.Info("running in container, log operations disabled")
			return logProv.NewLinuxProvider()
		}
		return logProv.NewDebianProvider(log, execManager)
	case "darwin":
		return logProv.NewDarwinProvider()
	default:
		return logProv.NewLinuxProvider()
	}
}
```

Add to `setupAgent` after `packageProvider`:
```go
	// --- Log provider ---
	logProvider := createLogProvider(log, execManager)
```

Add `logProvider` to the `NewNodeProcessor` call and the providers list in `registry.Register("node", ...)`.

- [ ] **Step 10: Run tests**

Run: `go test -v ./internal/agent/... && go build ./...`
Expected: PASS

- [ ] **Step 11: Commit**

```bash
git add pkg/sdk/client/operations.go internal/job/types.go \
  pkg/sdk/client/permissions.go internal/authtoken/permissions.go \
  internal/agent/processor_log.go internal/agent/processor_log_public_test.go \
  internal/agent/processor.go cmd/agent_setup.go
git commit -m "feat(log): add operations, permissions, and agent wiring"
```

---

### Task 5: OpenAPI Spec and Code Generation

**Files:**
- Create: `internal/controller/api/node/log/gen/api.yaml`
- Create: `internal/controller/api/node/log/gen/cfg.yaml`
- Create: `internal/controller/api/node/log/gen/generate.go`

- [ ] **Step 1: Create OpenAPI spec**

Create `internal/controller/api/node/log/gen/api.yaml`:

```yaml
# Copyright (c) 2026 John Dewey
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to
# deal in the Software without restriction, including without limitation the
# rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
# sell copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
# FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
# DEALINGS IN THE SOFTWARE.

---
openapi: 3.0.0
info:
  title: Log Management API
  version: 1.0.0
tags:
  - name: log_operations
    x-displayName: Node/Log
    description: Log viewing operations on a target node.

paths:
  /node/{hostname}/log:
    get:
      summary: Query journal entries
      description: >
        Query systemd journal entries on the target node with optional
        filtering by lines, time range, and priority.
      tags:
        - log_operations
      operationId: GetNodeLog
      security:
        - BearerAuth:
            - log:read
      parameters:
        - $ref: '#/components/parameters/Hostname'
        - name: lines
          in: query
          required: false
          description: Number of entries to return (default 100).
          x-oapi-codegen-extra-tags:
            validate: omitempty,min=1,max=10000
          schema:
            type: integer
            default: 100
            minimum: 1
            maximum: 10000
        - name: since
          in: query
          required: false
          description: >
            Time filter in journalctl format (e.g., "1 hour ago",
            "2026-03-31").
          schema:
            type: string
        - name: priority
          in: query
          required: false
          description: >
            Minimum priority level (0-7 or name: emerg, alert, crit,
            err, warning, notice, info, debug).
          schema:
            type: string
      responses:
        '200':
          description: Journal entries.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/LogCollectionResponse'
        '401':
          description: Unauthorized - API key required
          content:
            application/json:
              schema:
                $ref: '../../../common/gen/api.yaml#/components/schemas/ErrorResponse'
        '403':
          description: Forbidden - Insufficient permissions
          content:
            application/json:
              schema:
                $ref: '../../../common/gen/api.yaml#/components/schemas/ErrorResponse'
        '500':
          description: Error querying journal.
          content:
            application/json:
              schema:
                $ref: '../../../common/gen/api.yaml#/components/schemas/ErrorResponse'

  /node/{hostname}/log/unit/{name}:
    get:
      summary: Query journal entries for a unit
      description: >
        Query systemd journal entries for a specific unit on the target
        node with optional filtering.
      tags:
        - log_operations
      operationId: GetNodeLogUnit
      security:
        - BearerAuth:
            - log:read
      parameters:
        - $ref: '#/components/parameters/Hostname'
        - $ref: '#/components/parameters/UnitName'
        - name: lines
          in: query
          required: false
          description: Number of entries to return (default 100).
          x-oapi-codegen-extra-tags:
            validate: omitempty,min=1,max=10000
          schema:
            type: integer
            default: 100
            minimum: 1
            maximum: 10000
        - name: since
          in: query
          required: false
          description: >
            Time filter in journalctl format (e.g., "1 hour ago",
            "2026-03-31").
          schema:
            type: string
        - name: priority
          in: query
          required: false
          description: >
            Minimum priority level (0-7 or name: emerg, alert, crit,
            err, warning, notice, info, debug).
          schema:
            type: string
      responses:
        '200':
          description: Journal entries for the unit.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/LogCollectionResponse'
        '401':
          description: Unauthorized - API key required
          content:
            application/json:
              schema:
                $ref: '../../../common/gen/api.yaml#/components/schemas/ErrorResponse'
        '403':
          description: Forbidden - Insufficient permissions
          content:
            application/json:
              schema:
                $ref: '../../../common/gen/api.yaml#/components/schemas/ErrorResponse'
        '500':
          description: Error querying journal.
          content:
            application/json:
              schema:
                $ref: '../../../common/gen/api.yaml#/components/schemas/ErrorResponse'

# -- Reusable components --

components:
  parameters:
    Hostname:
      name: hostname
      in: path
      required: true
      description: >
        Target agent hostname, reserved routing value (_any, _all),
        or label selector (key:value).
      # NOTE: x-oapi-codegen-extra-tags on path params do not generate
      # validate tags in strict-server mode. Validation is handled
      # manually in handlers via validateHostname().
      x-oapi-codegen-extra-tags:
        validate: required,min=1,valid_target
      schema:
        type: string
        minLength: 1

    UnitName:
      name: name
      in: path
      required: true
      description: >
        Systemd unit name (e.g., nginx.service, sshd.service).
      # NOTE: x-oapi-codegen-extra-tags on path params do not generate
      # validate tags in strict-server mode. Validation is handled
      # manually in the handler.
      x-oapi-codegen-extra-tags:
        validate: required,min=1
      schema:
        type: string
        minLength: 1

  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT

  schemas:
    ErrorResponse:
      $ref: '../../../common/gen/api.yaml#/components/schemas/ErrorResponse'

    # -- Response schemas --

    LogEntryInfo:
      type: object
      description: A single journal entry.
      properties:
        timestamp:
          type: string
          description: Entry timestamp in RFC3339 format.
          example: "2026-03-31T22:30:45.123456Z"
        unit:
          type: string
          description: Systemd unit or syslog identifier.
          example: "nginx.service"
        priority:
          type: string
          description: Priority level name.
          example: "info"
        message:
          type: string
          description: Log message.
          example: "Started nginx"
        pid:
          type: integer
          description: Process ID that generated the entry.
          example: 1234
        hostname:
          type: string
          description: Hostname where the entry originated.
          example: "web-01"

    LogResultEntry:
      type: object
      description: Log query result for a single agent.
      properties:
        hostname:
          type: string
          description: The hostname of the agent.
        status:
          type: string
          enum: [ok, failed, skipped]
          description: The status of the operation for this host.
        entries:
          type: array
          description: Journal entries from this agent.
          items:
            $ref: '#/components/schemas/LogEntryInfo'
        error:
          type: string
          description: Error message if the agent failed.
      required:
        - hostname
        - status

    # -- Collection responses --

    LogCollectionResponse:
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
            $ref: '#/components/schemas/LogResultEntry'
      required:
        - results
```

- [ ] **Step 2: Create oapi-codegen config**

Create `internal/controller/api/node/log/gen/cfg.yaml`:

```yaml
# Copyright (c) 2026 John Dewey
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to
# deal in the Software without restriction, including without limitation the
# rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
# sell copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
# FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
# DEALINGS IN THE SOFTWARE.

---
package: gen
output: log.gen.go
generate:
  models: true
  echo-server: true
  strict-server: true
import-mapping:
  ../../../common/gen/api.yaml: github.com/retr0h/osapi/internal/controller/api/common/gen
output-options:
  # to make sure that all types are generated
  skip-prune: true
```

- [ ] **Step 3: Create generate.go**

Create `internal/controller/api/node/log/gen/generate.go`:

```go
// Copyright (c) 2026 John Dewey

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

// Package gen contains generated code for the log API.
package gen

//go:generate go tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config cfg.yaml api.yaml
```

- [ ] **Step 4: Generate code**

Run: `go generate ./internal/controller/api/node/log/gen/...`
Expected: generates `log.gen.go`

- [ ] **Step 5: Regenerate combined spec**

Run: `just generate`
Expected: combined spec updated, all code regenerates

- [ ] **Step 6: Commit**

```bash
git add internal/controller/api/node/log/gen/
git commit -m "feat(log): add OpenAPI spec and generated code"
```

---

### Task 6: API Handler Implementation

**Files:**
- Create: `internal/controller/api/node/log/types.go`
- Create: `internal/controller/api/node/log/log.go`
- Create: `internal/controller/api/node/log/validate.go`
- Create: `internal/controller/api/node/log/log_query_get.go`
- Create: `internal/controller/api/node/log/log_unit_get.go`
- Create: `internal/controller/api/node/log/handler.go`
- Modify: `cmd/controller_setup.go`
- Test: `internal/controller/api/node/log/log_query_get_public_test.go`
- Test: `internal/controller/api/node/log/log_unit_get_public_test.go`
- Test: `internal/controller/api/node/log/handler_public_test.go`

- [ ] **Step 1: Create handler types, factory, and validate**

Create `internal/controller/api/node/log/types.go`:
```go
package log

import (
	"log/slog"

	"github.com/retr0h/osapi/internal/job/client"
)

// Log implementation of the Log APIs operations.
type Log struct {
	// JobClient provides job-based operations for log management.
	JobClient client.JobClient
	logger    *slog.Logger
}
```

Create `internal/controller/api/node/log/log.go`:
```go
package log

import (
	"log/slog"

	"github.com/retr0h/osapi/internal/controller/api/node/log/gen"
	"github.com/retr0h/osapi/internal/job/client"
)

// ensure that we've conformed to the `StrictServerInterface` with a compile-time check
var _ gen.StrictServerInterface = (*Log)(nil)

// New factory to create a new instance.
func New(
	logger *slog.Logger,
	jobClient client.JobClient,
) *Log {
	return &Log{
		JobClient: jobClient,
		logger:    logger.With(slog.String("subsystem", "api.log")),
	}
}
```

Create `internal/controller/api/node/log/validate.go`:
```go
package log

import "github.com/retr0h/osapi/internal/validation"

// validateHostname validates a hostname path parameter using the shared
// validator. Returns the error message and false if invalid.
//
// This exists because oapi-codegen does not generate validate tags on
// path parameters in strict-server mode (upstream limitation).
func validateHostname(
	hostname string,
) (string, bool) {
	return validation.Var(hostname, "required,min=1,valid_target")
}
```

(All files need full license headers — follow existing patterns.)

- [ ] **Step 2: Write handler tests for GetNodeLog**

Create `internal/controller/api/node/log/log_query_get_public_test.go` — follow the same pattern as `process_list_get_public_test.go`. Test cases:
- success (single target)
- skipped (single target)
- broadcast success
- broadcast with failed/skipped hosts
- validation error (invalid hostname)
- job client error
- TestGetNodeLogHTTP (raw HTTP through middleware)
- TestGetNodeLogRBACHTTP (auth: 401, 403, 200)

The test should mock `s.mockJobClient.EXPECT().Query(...)` with category `"node"` and operation `job.OperationLogQuery`. The handler must pass query params (`lines`, `since`, `priority`) as JSON data.

- [ ] **Step 3: Write handler tests for GetNodeLogUnit**

Create `internal/controller/api/node/log/log_unit_get_public_test.go` — same pattern. Test cases:
- success (single target)
- skipped (single target)
- broadcast success
- validation error
- job client error
- TestGetNodeLogUnitHTTP
- TestGetNodeLogUnitRBACHTTP

The handler must pass `unit` (from path param) and query params as JSON data with operation `job.OperationLogQueryUnit`.

- [ ] **Step 4: Implement GetNodeLog handler**

Create `internal/controller/api/node/log/log_query_get.go`:

```go
package log

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/node/log/gen"
	"github.com/retr0h/osapi/internal/job"
	logProv "github.com/retr0h/osapi/internal/provider/node/log"
)

// GetNodeLog queries journal entries on a target node.
func (s *Log) GetNodeLog(
	ctx context.Context,
	request gen.GetNodeLogRequestObject,
) (gen.GetNodeLogResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeLog500JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	// Defense in depth: current query fields use omitempty so validation
	// always passes, but guards against future field additions.
	if errMsg, ok := validation.Struct(request.Params); !ok {
		return gen.GetNodeLog500JSONResponse{Error: &errMsg}, nil
	}

	s.logger.Debug("log query",
		slog.String("target", hostname),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	opts := logProv.QueryOpts{}
	if request.Params.Lines != nil {
		opts.Lines = *request.Params.Lines
	}
	if request.Params.Since != nil {
		opts.Since = *request.Params.Since
	}
	if request.Params.Priority != nil {
		opts.Priority = *request.Params.Priority
	}

	data, _ := json.Marshal(opts)

	if job.IsBroadcastTarget(hostname) {
		return s.getNodeLogBroadcast(ctx, hostname, data)
	}

	jobID, resp, err := s.JobClient.Query(ctx, hostname, "node", job.OperationLogQuery, data)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeLog500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		e := resp.Error
		jobUUID := uuid.MustParse(jobID)
		return gen.GetNodeLog200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.LogResultEntry{
				{
					Hostname: resp.Hostname,
					Status:   gen.LogResultEntryStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	entries := logEntriesFromResponse(resp)
	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeLog200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.LogResultEntry{
			{
				Hostname: resp.Hostname,
				Status:   gen.LogResultEntryStatusOk,
				Entries:  &entries,
			},
		},
	}, nil
}

// logEntriesFromResponse extracts LogEntryInfo slice from a job response.
func logEntriesFromResponse(
	resp *job.Response,
) []gen.LogEntryInfo {
	var provEntries []logProv.Entry
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &provEntries)
	}

	result := make([]gen.LogEntryInfo, 0, len(provEntries))
	for _, e := range provEntries {
		result = append(result, logEntryToGen(e))
	}

	return result
}

// logEntryToGen converts a provider Entry to a gen.LogEntryInfo.
func logEntryToGen(
	e logProv.Entry,
) gen.LogEntryInfo {
	ts := e.Timestamp
	unit := e.Unit
	priority := e.Priority
	message := e.Message
	pid := e.PID
	hostname := e.Hostname

	return gen.LogEntryInfo{
		Timestamp: &ts,
		Unit:      stringPtrOrNil(unit),
		Priority:  &priority,
		Message:   &message,
		Pid:       intPtrOrNil(pid),
		Hostname:  stringPtrOrNil(hostname),
	}
}

func stringPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func intPtrOrNil(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}

// getNodeLogBroadcast handles broadcast targets for log query.
func (s *Log) getNodeLogBroadcast(
	ctx context.Context,
	target string,
	data json.RawMessage,
) (gen.GetNodeLogResponseObject, error) {
	jobID, responses, err := s.JobClient.QueryBroadcast(
		ctx,
		target,
		"node",
		job.OperationLogQuery,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeLog500JSONResponse{Error: &errMsg}, nil
	}

	var items []gen.LogResultEntry
	for host, resp := range responses {
		item := gen.LogResultEntry{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.LogResultEntryStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.LogResultEntryStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.LogResultEntryStatusOk
			entries := logEntriesFromResponse(resp)
			item.Entries = &entries
		}
		items = append(items, item)
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.GetNodeLog200JSONResponse{
		JobId:   &jobUUID,
		Results: items,
	}, nil
}
```

Note: The import for `validation` is `"github.com/retr0h/osapi/internal/validation"`. All files need full license headers.

- [ ] **Step 5: Implement GetNodeLogUnit handler**

Create `internal/controller/api/node/log/log_unit_get.go` — same pattern as `log_query_get.go` but adds unit from `request.Name` path param. Passes `{"unit":"...","lines":...,"since":"...","priority":"..."}` as job data. Uses `job.OperationLogQueryUnit`.

- [ ] **Step 6: Implement handler.go**

Create `internal/controller/api/node/log/handler.go` — same pattern as `internal/controller/api/node/process/handler.go`:

```go
package log

import (
	"log/slog"

	"github.com/labstack/echo/v4"
	strictecho "github.com/oapi-codegen/runtime/strictmiddleware/echo"

	"github.com/retr0h/osapi/internal/authtoken"
	"github.com/retr0h/osapi/internal/controller/api"
	gen "github.com/retr0h/osapi/internal/controller/api/node/log/gen"
	"github.com/retr0h/osapi/internal/job/client"
)

// Handler returns Log route registration functions.
func Handler(
	logger *slog.Logger,
	jobClient client.JobClient,
	signingKey string,
	customRoles map[string][]string,
) []func(e *echo.Echo) {
	var tokenManager api.TokenValidator = authtoken.New(logger)

	logHandler := New(logger, jobClient)

	strictHandler := gen.NewStrictHandler(
		logHandler,
		[]gen.StrictMiddlewareFunc{
			func(handler strictecho.StrictEchoHandlerFunc, _ string) strictecho.StrictEchoHandlerFunc {
				return api.ScopeMiddleware(
					handler,
					tokenManager,
					signingKey,
					gen.BearerAuthScopes,
					customRoles,
				)
			},
		},
	)

	return []func(e *echo.Echo){
		func(e *echo.Echo) {
			gen.RegisterHandlers(e, strictHandler)
		},
	}
}
```

- [ ] **Step 7: Register handler in controller_setup.go**

Add import:
```go
logAPI "github.com/retr0h/osapi/internal/controller/api/node/log"
```

Add after the `packageAPI.Handler(...)` line:
```go
	handlers = append(handlers, logAPI.Handler(log, jc, signingKey, customRoles)...)
```

- [ ] **Step 8: Write handler_public_test.go**

Test route registration and middleware execution (same pattern as other handler tests).

- [ ] **Step 9: Run tests**

Run: `go test -v ./internal/controller/api/node/log/... && go build ./...`
Expected: PASS

- [ ] **Step 10: Commit**

```bash
git add internal/controller/api/node/log/ cmd/controller_setup.go
git commit -m "feat(log): add API handlers with broadcast support"
```

---

### Task 7: SDK Service

**Files:**
- Create: `pkg/sdk/client/log.go`
- Create: `pkg/sdk/client/log_types.go`
- Modify: `pkg/sdk/client/osapi.go`
- Test: `pkg/sdk/client/log_public_test.go`
- Test: `pkg/sdk/client/log_types_public_test.go`

- [ ] **Step 1: Write SDK service tests**

Create `pkg/sdk/client/log_public_test.go` — test with `httptest.Server`. Test cases for `Query` and `QueryUnit`:
- success (200)
- auth error (401, 403)
- server error (500)
- nil response body
- transport error

- [ ] **Step 2: Write SDK types tests**

Create `pkg/sdk/client/log_types_public_test.go` — test conversion functions:
- `logCollectionFromGen` with full data
- `logCollectionFromGen` with error entries
- `logEntryInfoFromGen` field mapping
- Nil/empty fields

- [ ] **Step 3: Implement log_types.go**

```go
package client

import (
	"github.com/retr0h/osapi/pkg/sdk/client/gen"
)

// LogEntryResult represents the result of a log query for one host.
type LogEntryResult struct {
	Hostname string     `json:"hostname"`
	Status   string     `json:"status"`
	Entries  []LogEntry `json:"entries,omitempty"`
	Error    string     `json:"error,omitempty"`
}

// LogEntry represents a single journal entry.
type LogEntry struct {
	Timestamp string `json:"timestamp,omitempty"`
	Unit      string `json:"unit,omitempty"`
	Priority  string `json:"priority,omitempty"`
	Message   string `json:"message,omitempty"`
	PID       int    `json:"pid,omitempty"`
	Hostname  string `json:"hostname,omitempty"`
}

// LogQueryOpts contains options for log query operations.
type LogQueryOpts struct {
	Lines    *int
	Since    *string
	Priority *string
}

// logCollectionFromGen converts a gen.LogCollectionResponse
// to a Collection[LogEntryResult].
func logCollectionFromGen(
	g *gen.LogCollectionResponse,
) Collection[LogEntryResult] {
	results := make([]LogEntryResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, logEntryResultFromGen(r))
	}

	return Collection[LogEntryResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// logEntryResultFromGen converts a gen.LogResultEntry to a LogEntryResult.
func logEntryResultFromGen(
	r gen.LogResultEntry,
) LogEntryResult {
	result := LogEntryResult{
		Hostname: r.Hostname,
		Status:   string(r.Status),
		Error:    derefString(r.Error),
	}

	if r.Entries != nil {
		entries := make([]LogEntry, 0, len(*r.Entries))
		for _, e := range *r.Entries {
			entries = append(entries, logEntryInfoFromGen(e))
		}
		result.Entries = entries
	}

	return result
}

// logEntryInfoFromGen converts a gen.LogEntryInfo to a LogEntry.
func logEntryInfoFromGen(
	e gen.LogEntryInfo,
) LogEntry {
	return LogEntry{
		Timestamp: derefString(e.Timestamp),
		Unit:      derefString(e.Unit),
		Priority:  derefString(e.Priority),
		Message:   derefString(e.Message),
		PID:       derefInt(e.Pid),
		Hostname:  derefString(e.Hostname),
	}
}
```

- [ ] **Step 4: Implement log.go**

```go
package client

import (
	"context"
	"fmt"

	"github.com/retr0h/osapi/pkg/sdk/client/gen"
)

// LogService provides log viewing operations.
type LogService struct {
	client *gen.ClientWithResponses
}

// Query returns journal entries from the target host.
func (s *LogService) Query(
	ctx context.Context,
	hostname string,
	opts LogQueryOpts,
) (*Response[Collection[LogEntryResult]], error) {
	params := &gen.GetNodeLogParams{
		Lines:    opts.Lines,
		Since:    opts.Since,
		Priority: opts.Priority,
	}

	resp, err := s.client.GetNodeLogWithResponse(ctx, hostname, params)
	if err != nil {
		return nil, fmt.Errorf("log query: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON401,
		resp.JSON403,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(logCollectionFromGen(resp.JSON200), resp.Body), nil
}

// QueryUnit returns journal entries for a specific unit on the target host.
func (s *LogService) QueryUnit(
	ctx context.Context,
	hostname string,
	unit string,
	opts LogQueryOpts,
) (*Response[Collection[LogEntryResult]], error) {
	params := &gen.GetNodeLogUnitParams{
		Lines:    opts.Lines,
		Since:    opts.Since,
		Priority: opts.Priority,
	}

	resp, err := s.client.GetNodeLogUnitWithResponse(ctx, hostname, unit, params)
	if err != nil {
		return nil, fmt.Errorf("log query unit: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON401,
		resp.JSON403,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(logCollectionFromGen(resp.JSON200), resp.Body), nil
}
```

- [ ] **Step 5: Wire LogService in osapi.go**

Add field to Client struct:
```go
	// Log provides log viewing operations (query journal entries).
	Log *LogService
```

Add initialization in `New()`:
```go
	c.Log = &LogService{client: httpClient}
```

- [ ] **Step 6: Regenerate SDK client**

Run: `go generate ./pkg/sdk/client/gen/...`
Expected: SDK client picks up log endpoints

- [ ] **Step 7: Run tests**

Run: `go test -v ./pkg/sdk/client/...`
Expected: PASS

- [ ] **Step 8: Commit**

```bash
git add pkg/sdk/client/log.go pkg/sdk/client/log_types.go \
  pkg/sdk/client/log_public_test.go pkg/sdk/client/log_types_public_test.go \
  pkg/sdk/client/osapi.go pkg/sdk/client/gen/
git commit -m "feat(log): add SDK service with tests"
```

---

### Task 8: CLI Commands

**Files:**
- Create: `cmd/client_node_log.go`
- Create: `cmd/client_node_log_query.go`
- Create: `cmd/client_node_log_unit.go`

- [ ] **Step 1: Create parent command**

Create `cmd/client_node_log.go`:

```go
package cmd

import (
	"github.com/spf13/cobra"
)

// clientNodeLogCmd represents the clientNodeLog command.
var clientNodeLogCmd = &cobra.Command{
	Use:   "log",
	Short: "View journal logs",
}

func init() {
	clientNodeCmd.AddCommand(clientNodeLogCmd)
}
```

- [ ] **Step 2: Create query subcommand**

Create `cmd/client_node_log_query.go`:

```go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/cli"
	"github.com/retr0h/osapi/pkg/sdk/client"
)

// clientNodeLogQueryCmd represents the log query command.
var clientNodeLogQueryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query journal entries",
	Long:  `Query systemd journal entries on the target node.`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")
		lines, _ := cmd.Flags().GetInt("lines")
		since, _ := cmd.Flags().GetString("since")
		priority, _ := cmd.Flags().GetString("priority")

		opts := client.LogQueryOpts{}
		if cmd.Flags().Changed("lines") {
			opts.Lines = &lines
		}
		if since != "" {
			opts.Since = &since
		}
		if priority != "" {
			opts.Priority = &priority
		}

		resp, err := sdkClient.Log.Query(ctx, host, opts)
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
			fmt.Println()
		}

		results := make([]cli.ResultRow, 0)
		for _, r := range resp.Data.Results {
			if r.Error != "" {
				var errPtr *string
				e := r.Error
				errPtr = &e
				results = append(results, cli.ResultRow{
					Hostname: r.Hostname,
					Status:   r.Status,
					Error:    errPtr,
				})

				continue
			}

			for _, entry := range r.Entries {
				message := entry.Message
				if len(message) > 80 {
					message = message[:77] + "..."
				}

				results = append(results, cli.ResultRow{
					Hostname: r.Hostname,
					Status:   r.Status,
					Fields: []string{
						entry.Timestamp,
						entry.Priority,
						entry.Unit,
						message,
					},
				})
			}
		}
		headers, rows := cli.BuildBroadcastTable(
			results,
			[]string{"TIMESTAMP", "PRIORITY", "UNIT", "MESSAGE"},
		)
		cli.PrintCompactTable([]cli.Section{{Headers: headers, Rows: rows}})
	},
}

func init() {
	clientNodeLogCmd.AddCommand(clientNodeLogQueryCmd)

	clientNodeLogQueryCmd.PersistentFlags().
		Int("lines", 100, "Number of entries to return")
	clientNodeLogQueryCmd.PersistentFlags().
		String("since", "", "Time filter (e.g., \"1 hour ago\")")
	clientNodeLogQueryCmd.PersistentFlags().
		String("priority", "", "Minimum priority (emerg..debug or 0-7)")
}
```

- [ ] **Step 3: Create query-unit subcommand**

Create `cmd/client_node_log_unit.go`:

```go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/cli"
	"github.com/retr0h/osapi/pkg/sdk/client"
)

// clientNodeLogUnitCmd represents the log unit command.
var clientNodeLogUnitCmd = &cobra.Command{
	Use:   "unit",
	Short: "Query journal entries for a unit",
	Long:  `Query systemd journal entries for a specific unit on the target node.`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")
		unit, _ := cmd.Flags().GetString("name")
		lines, _ := cmd.Flags().GetInt("lines")
		since, _ := cmd.Flags().GetString("since")
		priority, _ := cmd.Flags().GetString("priority")

		opts := client.LogQueryOpts{}
		if cmd.Flags().Changed("lines") {
			opts.Lines = &lines
		}
		if since != "" {
			opts.Since = &since
		}
		if priority != "" {
			opts.Priority = &priority
		}

		resp, err := sdkClient.Log.QueryUnit(ctx, host, unit, opts)
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
			fmt.Println()
		}

		results := make([]cli.ResultRow, 0)
		for _, r := range resp.Data.Results {
			if r.Error != "" {
				var errPtr *string
				e := r.Error
				errPtr = &e
				results = append(results, cli.ResultRow{
					Hostname: r.Hostname,
					Status:   r.Status,
					Error:    errPtr,
				})

				continue
			}

			for _, entry := range r.Entries {
				message := entry.Message
				if len(message) > 80 {
					message = message[:77] + "..."
				}

				results = append(results, cli.ResultRow{
					Hostname: r.Hostname,
					Status:   r.Status,
					Fields: []string{
						entry.Timestamp,
						entry.Priority,
						entry.Unit,
						message,
					},
				})
			}
		}
		headers, rows := cli.BuildBroadcastTable(
			results,
			[]string{"TIMESTAMP", "PRIORITY", "UNIT", "MESSAGE"},
		)
		cli.PrintCompactTable([]cli.Section{{Headers: headers, Rows: rows}})
	},
}

func init() {
	clientNodeLogCmd.AddCommand(clientNodeLogUnitCmd)

	clientNodeLogUnitCmd.PersistentFlags().
		String("name", "", "Systemd unit name (required)")
	clientNodeLogUnitCmd.PersistentFlags().
		Int("lines", 100, "Number of entries to return")
	clientNodeLogUnitCmd.PersistentFlags().
		String("since", "", "Time filter (e.g., \"1 hour ago\")")
	clientNodeLogUnitCmd.PersistentFlags().
		String("priority", "", "Minimum priority (emerg..debug or 0-7)")

	_ = clientNodeLogUnitCmd.MarkPersistentFlagRequired("name")
}
```

- [ ] **Step 4: Build and verify**

Run: `go build ./... && go run main.go client node log --help`
Expected: shows `query` and `unit` subcommands

- [ ] **Step 5: Commit**

```bash
git add cmd/client_node_log.go cmd/client_node_log_query.go cmd/client_node_log_unit.go
git commit -m "feat(log): add CLI commands for journal log viewing"
```

---

### Task 9: Documentation and SDK Example

**Files:**
- Create: `docs/docs/sidebar/features/log-management.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/log/log.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/log/query.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/log/unit.md`
- Create: `docs/docs/sidebar/sdk/client/operations/log.md`
- Create: `examples/sdk/client/log.go`
- Modify: `docs/docs/sidebar/features/features.md`
- Modify: `docs/docs/sidebar/features/authentication.md`
- Modify: `docs/docs/sidebar/usage/configuration.md`
- Modify: `docs/docs/sidebar/architecture/architecture.md`
- Modify: `docs/docs/sidebar/architecture/api-guidelines.md`
- Modify: `docs/docusaurus.config.ts`

- [ ] **Step 1: Create feature page**

Create `docs/docs/sidebar/features/log-management.md` following the process-management.md template. Include:
- How It Works (Query, QueryUnit)
- Operations table
- CLI Usage examples
- Broadcast Support section
- Supported Platforms table (Debian: Full, Darwin: Skipped, Linux: Skipped)
- Container Behavior (skipped — journalctl requires systemd)
- Permissions table (log:read for all operations)
- Related links

- [ ] **Step 2: Create CLI doc pages**

Create landing page `docs/docs/sidebar/usage/cli/client/node/log/log.md`:
```markdown
---
sidebar_position: 1
---

# Log

<DocCardList />
```

Create `query.md` and `unit.md` pages with usage examples, flags, and output samples.

- [ ] **Step 3: Create SDK doc page**

Create `docs/docs/sidebar/sdk/client/operations/log.md` following existing SDK doc patterns. Document `Query` and `QueryUnit` methods with code examples.

- [ ] **Step 4: Create SDK example**

Create `examples/sdk/client/log.go` — demonstrate `Query` and `QueryUnit` with error handling and result printing. Under ~100 lines.

- [ ] **Step 5: Update cross-references**

Update these files to add log management:
- `features/features.md` — add row to features table
- `features/authentication.md` — add `log:read` to all three role tables
- `usage/configuration.md` — add `log:read` to permissions comments and role tables
- `architecture/architecture.md` — add log feature link
- `architecture/api-guidelines.md` — add log endpoint rows to path pattern table
- `docusaurus.config.ts` — add to Features dropdown and SDK dropdown

- [ ] **Step 6: Commit**

```bash
git add docs/ examples/sdk/client/log.go
git commit -m "docs: add log management feature docs, SDK example, and cross-references"
```

---

### Task 10: Integration Test

**Files:**
- Create: `test/integration/log_test.go`

- [ ] **Step 1: Write integration test**

Create `test/integration/log_test.go` with `//go:build integration` tag. Follow the pattern of existing integration tests. Test:
- `osapi client node log query --target _any --json` → verify JSON output
- `osapi client node log unit --target _any --name sshd.service --json` → verify JSON output or graceful error

- [ ] **Step 2: Commit**

```bash
git add test/integration/log_test.go
git commit -m "test(log): add integration test"
```

---

### Task 11: Final Verification

- [ ] **Step 1: Run full test suite**

```bash
just generate
go build ./...
just go::unit
just go::vet
```

Expected: all pass, lint clean

- [ ] **Step 2: Commit any fixes**

If `just generate` produces diffs (combined spec, formatting), commit them:

```bash
git add -A
git commit -m "chore(log): regenerate specs and fix formatting"
```
