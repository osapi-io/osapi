# Container DNS Provider Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add container detection to the platform package and a `DebianDocker` DNS provider that reads `/etc/resolv.conf` for containerized agents.

**Architecture:** A new `platform.IsContainer()` function detects Docker containers via `/.dockerenv`. The `DebianDocker` DNS provider uses `avfs.VFS` to parse `/etc/resolv.conf` for Get (ignoring the interface parameter) and returns `ErrUnsupported` for Update. A new `containerized` built-in fact key exposes the container state. Agent setup wires the container check into DNS provider selection.

**Tech Stack:** Go, avfs (memfs for tests), testify/suite, gopsutil

**Spec:** `docs/superpowers/specs/2026-03-27-container-dns-provider-design.md`

---

### Task 1: Container Detection — `platform.IsContainer()`

**Files:**
- Create: `pkg/sdk/platform/container.go`
- Create: `pkg/sdk/platform/container_public_test.go`

- [ ] **Step 1: Write the failing test**

Create `pkg/sdk/platform/container_public_test.go`:

```go
package platform_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/pkg/sdk/platform"
)

type ContainerPublicTestSuite struct {
	suite.Suite
}

func (s *ContainerPublicTestSuite) TearDownSubTest() {
	platform.ContainerCheckFn = platform.DefaultContainerCheck
}

func (s *ContainerPublicTestSuite) TestIsContainer() {
	tests := []struct {
		name    string
		checkFn func() bool
		want    bool
	}{
		{
			name: "when inside a Docker container",
			checkFn: func() bool {
				return true
			},
			want: true,
		},
		{
			name: "when not inside a container",
			checkFn: func() bool {
				return false
			},
			want: false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			platform.ContainerCheckFn = tc.checkFn

			got := platform.IsContainer()

			s.Equal(tc.want, got)
		})
	}
}

func TestContainerPublicTestSuite(t *testing.T) {
	suite.Run(t, new(ContainerPublicTestSuite))
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -run TestContainerPublicTestSuite -v ./pkg/sdk/platform/...`
Expected: FAIL — `ContainerCheckFn`, `DefaultContainerCheck`, `IsContainer` not defined.

- [ ] **Step 3: Write minimal implementation**

Create `pkg/sdk/platform/container.go`:

```go
package platform

import "os"

// DefaultContainerCheck checks for the presence of /.dockerenv
// to determine if the process is running inside a Docker container.
func DefaultContainerCheck() bool {
	_, err := os.Stat("/.dockerenv")
	return err == nil
}

// ContainerCheckFn is the function used to detect container environments.
// Override in tests to simulate different environments.
var ContainerCheckFn = DefaultContainerCheck

// IsContainer reports whether the current process is running inside
// a container (currently Docker only).
func IsContainer() bool {
	return ContainerCheckFn()
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -run TestContainerPublicTestSuite -v ./pkg/sdk/platform/...`
Expected: PASS

- [ ] **Step 5: Commit**

```
feat(platform): add IsContainer() for Docker container detection
```

---

### Task 2: DebianDocker DNS Provider — Struct and Update Stub

**Files:**
- Create: `internal/provider/network/dns/debian_docker.go`
- Create: `internal/provider/network/dns/debian_docker_update_resolv_conf_by_interface.go`
- Create: `internal/provider/network/dns/debian_docker_public_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/provider/network/dns/debian_docker_public_test.go`:

```go
package dns_test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/provider"
	"github.com/retr0h/osapi/internal/provider/network/dns"
)

type DebianDockerPublicTestSuite struct {
	suite.Suite

	logger *slog.Logger
	fs     avfs.VFS
}

func (s *DebianDockerPublicTestSuite) SetupTest() {
	s.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	s.fs = memfs.New()
}

func (s *DebianDockerPublicTestSuite) SetupSubTest() {
	s.SetupTest()
}

func (s *DebianDockerPublicTestSuite) TestUpdateResolvConfByInterface() {
	tests := []struct {
		name string
	}{
		{
			name: "returns ErrUnsupported for container",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			p := dns.NewDebianDockerProvider(s.logger, s.fs)
			result, err := p.UpdateResolvConfByInterface(
				[]string{"8.8.8.8"},
				[]string{"example.com"},
				"eth0",
			)

			s.Error(err)
			s.Nil(result)
			s.ErrorIs(err, provider.ErrUnsupported)
		})
	}
}

func TestDebianDockerPublicTestSuite(t *testing.T) {
	suite.Run(t, new(DebianDockerPublicTestSuite))
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -run TestDebianDockerPublicTestSuite -v ./internal/provider/network/dns/...`
Expected: FAIL — `NewDebianDockerProvider` not defined.

- [ ] **Step 3: Write minimal implementation**

Create `internal/provider/network/dns/debian_docker.go`:

```go
package dns

import (
	"log/slog"

	"github.com/avfs/avfs"

	"github.com/retr0h/osapi/internal/provider"
)

// Compile-time check: DebianDocker must satisfy Provider and FactsSetter.
var _ Provider = (*DebianDocker)(nil)
var _ provider.FactsSetter = (*DebianDocker)(nil)

// DebianDocker implements the DNS Provider interface for Debian-family
// systems running inside Docker containers. It reads DNS configuration
// from /etc/resolv.conf directly (no systemd-resolved). Updates are
// not supported because container DNS is managed by the runtime.
type DebianDocker struct {
	provider.FactsAware

	logger *slog.Logger
	fs     avfs.VFS
}

// NewDebianDockerProvider factory to create a new DebianDocker instance.
func NewDebianDockerProvider(
	logger *slog.Logger,
	fs avfs.VFS,
) *DebianDocker {
	return &DebianDocker{
		logger: logger.With(slog.String("subsystem", "provider.dns.container")),
		fs:     fs,
	}
}
```

Create `internal/provider/network/dns/debian_docker_update_resolv_conf_by_interface.go`:

```go
package dns

import (
	"fmt"

	"github.com/retr0h/osapi/internal/provider"
)

// UpdateResolvConfByInterface returns ErrUnsupported for container
// environments. DNS configuration in containers is managed by the
// container runtime (Docker, Kubernetes), not the agent.
func (d *DebianDocker) UpdateResolvConfByInterface(
	_ []string,
	_ []string,
	_ string,
) (*UpdateResult, error) {
	return nil, fmt.Errorf("dns (container): %w", provider.ErrUnsupported)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -run TestDebianDockerPublicTestSuite -v ./internal/provider/network/dns/...`
Expected: FAIL — `GetResolvConfByInterface` not implemented yet (compile-time check). Add a temporary stub to unblock:

The compile-time `var _ Provider = (*DebianDocker)(nil)` will fail because `GetResolvConfByInterface` is not yet implemented. To unblock, add a placeholder in `debian_docker_get_resolv_conf_by_interface.go` that panics — this will be replaced in Task 3. Alternatively, remove the `var _ Provider` line and add it back in Task 3.

The simpler approach: temporarily comment out `var _ Provider = (*DebianDocker)(nil)` in `debian_docker.go`. Re-add it in Task 3 when Get is implemented. Keep only `var _ provider.FactsSetter = (*DebianDocker)(nil)`.

Run again: `go test -run TestDebianDockerPublicTestSuite -v ./internal/provider/network/dns/...`
Expected: PASS

- [ ] **Step 5: Commit**

```
feat(dns): add DebianDocker provider struct and Update stub
```

---

### Task 3: DebianDocker DNS Provider — Get (resolv.conf parsing)

**Files:**
- Create: `internal/provider/network/dns/debian_docker_get_resolv_conf_by_interface.go`
- Modify: `internal/provider/network/dns/debian_docker.go` (re-add `var _ Provider` check)
- Modify: `internal/provider/network/dns/debian_docker_public_test.go` (add Get tests)

- [ ] **Step 1: Write the failing tests**

Add to `internal/provider/network/dns/debian_docker_public_test.go` — insert this method before the `TestUpdateResolvConfByInterface` method:

```go
func (s *DebianDockerPublicTestSuite) TestGetResolvConfByInterface() {
	tests := []struct {
		name          string
		setupFS       func(fs avfs.VFS)
		interfaceName string
		want          *dns.GetResult
		wantErr       bool
		errContains   string
	}{
		{
			name: "when resolv.conf has servers and search domains",
			setupFS: func(fs avfs.VFS) {
				_ = avfs.MkdirAll(fs, "/etc", 0o755)
				_ = fs.WriteFile("/etc/resolv.conf", []byte(
					"# Generated by Docker\n"+
						"nameserver 127.0.0.11\n"+
						"nameserver 8.8.8.8\n"+
						"search example.com local.lan\n"+
						"options ndots:0\n",
				), 0o644)
			},
			interfaceName: "eth0",
			want: &dns.GetResult{
				DNSServers:    []string{"127.0.0.11", "8.8.8.8"},
				SearchDomains: []string{"example.com", "local.lan"},
			},
		},
		{
			name: "when resolv.conf has only nameservers",
			setupFS: func(fs avfs.VFS) {
				_ = avfs.MkdirAll(fs, "/etc", 0o755)
				_ = fs.WriteFile("/etc/resolv.conf", []byte(
					"nameserver 8.8.8.8\n"+
						"nameserver 8.8.4.4\n",
				), 0o644)
			},
			interfaceName: "eth0",
			want: &dns.GetResult{
				DNSServers:    []string{"8.8.8.8", "8.8.4.4"},
				SearchDomains: []string{"."},
			},
		},
		{
			name: "when resolv.conf has IPv6 nameservers",
			setupFS: func(fs avfs.VFS) {
				_ = avfs.MkdirAll(fs, "/etc", 0o755)
				_ = fs.WriteFile("/etc/resolv.conf", []byte(
					"nameserver 2001:4860:4860::8888\n"+
						"nameserver 2001:4860:4860::8844\n",
				), 0o644)
			},
			interfaceName: "any-interface",
			want: &dns.GetResult{
				DNSServers:    []string{"2001:4860:4860::8888", "2001:4860:4860::8844"},
				SearchDomains: []string{"."},
			},
		},
		{
			name: "when resolv.conf has comments and blank lines",
			setupFS: func(fs avfs.VFS) {
				_ = avfs.MkdirAll(fs, "/etc", 0o755)
				_ = fs.WriteFile("/etc/resolv.conf", []byte(
					"# This is a comment\n"+
						"\n"+
						"nameserver 1.1.1.1\n"+
						"# Another comment\n"+
						"search test.local\n"+
						"\n",
				), 0o644)
			},
			interfaceName: "eth0",
			want: &dns.GetResult{
				DNSServers:    []string{"1.1.1.1"},
				SearchDomains: []string{"test.local"},
			},
		},
		{
			name: "when resolv.conf does not exist",
			setupFS: func(fs avfs.VFS) {
				// Don't create the file
			},
			interfaceName: "eth0",
			wantErr:       true,
			errContains:   "failed to read /etc/resolv.conf",
		},
		{
			name: "when resolv.conf is empty",
			setupFS: func(fs avfs.VFS) {
				_ = avfs.MkdirAll(fs, "/etc", 0o755)
				_ = fs.WriteFile("/etc/resolv.conf", []byte(""), 0o644)
			},
			interfaceName: "eth0",
			want: &dns.GetResult{
				DNSServers:    nil,
				SearchDomains: []string{"."},
			},
		},
		{
			name: "when interface parameter is ignored",
			setupFS: func(fs avfs.VFS) {
				_ = avfs.MkdirAll(fs, "/etc", 0o755)
				_ = fs.WriteFile("/etc/resolv.conf", []byte(
					"nameserver 10.0.0.1\n",
				), 0o644)
			},
			interfaceName: "completely-ignored",
			want: &dns.GetResult{
				DNSServers:    []string{"10.0.0.1"},
				SearchDomains: []string{"."},
			},
		},
		{
			name: "when multiple search lines uses last one",
			setupFS: func(fs avfs.VFS) {
				_ = avfs.MkdirAll(fs, "/etc", 0o755)
				_ = fs.WriteFile("/etc/resolv.conf", []byte(
					"nameserver 8.8.8.8\n"+
						"search first.com\n"+
						"search second.com third.com\n",
				), 0o644)
			},
			interfaceName: "eth0",
			want: &dns.GetResult{
				DNSServers:    []string{"8.8.8.8"},
				SearchDomains: []string{"second.com", "third.com"},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setupFS(s.fs)

			p := dns.NewDebianDockerProvider(s.logger, s.fs)
			got, err := p.GetResolvConfByInterface(tc.interfaceName)

			if tc.wantErr {
				s.Error(err)
				s.Contains(err.Error(), tc.errContains)
			} else {
				s.NoError(err)
				s.Equal(tc.want, got)
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -run TestDebianDockerPublicTestSuite/TestGetResolvConfByInterface -v ./internal/provider/network/dns/...`
Expected: FAIL — `GetResolvConfByInterface` not defined on `DebianDocker`.

- [ ] **Step 3: Write minimal implementation**

Create `internal/provider/network/dns/debian_docker_get_resolv_conf_by_interface.go`:

```go
package dns

import (
	"bufio"
	"fmt"
	"strings"
)

const resolvConfPath = "/etc/resolv.conf"

// GetResolvConfByInterface reads DNS configuration from /etc/resolv.conf.
// The interfaceName parameter is accepted but ignored — containers have
// a single global DNS configuration managed by the container runtime.
func (d *DebianDocker) GetResolvConfByInterface(
	_ string,
) (*GetResult, error) {
	f, err := d.fs.Open(resolvConfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", resolvConfPath, err)
	}
	defer f.Close()

	result := &GetResult{}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		switch fields[0] {
		case "nameserver":
			result.DNSServers = append(result.DNSServers, fields[1])
		case "search":
			result.SearchDomains = fields[1:]
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", resolvConfPath, err)
	}

	if len(result.SearchDomains) == 0 {
		result.SearchDomains = []string{"."}
	}

	return result, nil
}
```

Re-add the Provider compile-time check in `debian_docker.go`:

```go
var _ Provider = (*DebianDocker)(nil)
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -run TestDebianDockerPublicTestSuite -v ./internal/provider/network/dns/...`
Expected: PASS (both Get and Update tests)

- [ ] **Step 5: Commit**

```
feat(dns): add DebianDocker Get implementation for resolv.conf parsing
```

---

### Task 4: Add `containerized` Built-in Fact Key

**Files:**
- Modify: `internal/facts/keys.go`
- Modify: `internal/facts/keys_public_test.go`
- Modify: `internal/controller/api/facts/facts_keys_get.go`
- Modify: `internal/controller/api/facts/facts_keys_get_public_test.go`

- [ ] **Step 1: Update keys.go — add the constant**

In `internal/facts/keys.go`, add `KeyContainerized` to the const block:

```go
const (
	KeyInterfacePrimary = "interface.primary"
	KeyHostname         = "hostname"
	KeyArch             = "arch"
	KeyKernel           = "kernel"
	KeyFQDN             = "fqdn"
	KeyContainerized    = "containerized"
)
```

Add `KeyContainerized` to `BuiltInKeys()`:

```go
func BuiltInKeys() []string {
	return []string{
		KeyInterfacePrimary,
		KeyHostname,
		KeyArch,
		KeyKernel,
		KeyFQDN,
		KeyContainerized,
	}
}
```

Add `KeyContainerized` to the `IsKnownKey` switch:

```go
case KeyInterfacePrimary, KeyHostname, KeyArch, KeyKernel, KeyFQDN, KeyContainerized:
	return true
```

- [ ] **Step 2: Update the keys test**

In `internal/facts/keys_public_test.go`, update `TestBuiltInKeys`:

```go
{
	name: "when called returns all six built-in keys",
	validateFunc: func(keys []string) {
		s.Len(keys, 6)
		s.Contains(keys, facts.KeyInterfacePrimary)
		s.Contains(keys, facts.KeyHostname)
		s.Contains(keys, facts.KeyArch)
		s.Contains(keys, facts.KeyKernel)
		s.Contains(keys, facts.KeyFQDN)
		s.Contains(keys, facts.KeyContainerized)
	},
},
```

Add a test row to `TestIsKnownKey`:

```go
{
	name:   "when containerized",
	key:    facts.KeyContainerized,
	wantOK: true,
},
```

- [ ] **Step 3: Update the facts keys API handler**

In `internal/controller/api/facts/facts_keys_get.go`, add the description:

```go
var builtInDescriptions = map[string]string{
	factskeys.KeyInterfacePrimary: "Primary network interface name",
	factskeys.KeyHostname:         "Agent hostname",
	factskeys.KeyArch:             "CPU architecture",
	factskeys.KeyKernel:           "Kernel version",
	factskeys.KeyFQDN:             "Fully qualified domain name",
	factskeys.KeyContainerized:    "Whether the agent is running inside a container",
}
```

- [ ] **Step 4: Update the facts keys API test**

In `internal/controller/api/facts/facts_keys_get_public_test.go`, update the test that checks the count of keys returned (if there is a count assertion, update from 5 to 6).

- [ ] **Step 5: Run tests**

Run: `go test -v ./internal/facts/... ./internal/controller/api/facts/...`
Expected: PASS

- [ ] **Step 6: Commit**

```
feat(facts): add containerized built-in fact key
```

---

### Task 5: Add `Containerized` to `FactsRegistration` and Fact Resolution

**Files:**
- Modify: `internal/job/types.go`
- Modify: `internal/agent/factref.go`
- Modify: `internal/agent/factref_public_test.go`

- [ ] **Step 1: Add field to FactsRegistration**

In `internal/job/types.go`, add `Containerized` to the `FactsRegistration` struct:

```go
type FactsRegistration struct {
	Architecture     string             `json:"architecture,omitempty"`
	KernelVersion    string             `json:"kernel_version,omitempty"`
	CPUCount         int                `json:"cpu_count,omitempty"`
	FQDN             string             `json:"fqdn,omitempty"`
	ServiceMgr       string             `json:"service_mgr,omitempty"`
	PackageMgr       string             `json:"package_mgr,omitempty"`
	Containerized    bool               `json:"containerized"`
	Interfaces       []NetworkInterface `json:"interfaces,omitempty"`
	PrimaryInterface string             `json:"primary_interface,omitempty"`
	Routes           []Route            `json:"routes,omitempty"`
	Facts            map[string]any     `json:"facts,omitempty"`
}
```

Note: `Containerized bool` does NOT use `omitempty` — the field must always be present (false is meaningful).

- [ ] **Step 2: Add fact resolution for containerized**

In `internal/agent/factref.go`, add a case to the `lookupFact` switch:

```go
case facts.KeyContainerized:
	if f.Containerized {
		return "true", nil
	}
	return "false", nil
```

Insert after the `facts.KeyFQDN` case.

- [ ] **Step 3: Add tests for @fact.containerized resolution**

In `internal/agent/factref_public_test.go`, add two test rows to the `TestResolveFacts` table:

```go
{
	name: "when containerized is true",
	params: map[string]any{
		"in_container": "@fact.containerized",
	},
	facts: &job.FactsRegistration{
		Containerized: true,
	},
	hostname: "web-01",
	validateFunc: func(result map[string]any) {
		s.Equal("true", result["in_container"])
	},
},
{
	name: "when containerized is false",
	params: map[string]any{
		"in_container": "@fact.containerized",
	},
	facts: &job.FactsRegistration{
		Containerized: false,
	},
	hostname: "web-01",
	validateFunc: func(result map[string]any) {
		s.Equal("false", result["in_container"])
	},
},
```

- [ ] **Step 4: Run tests**

Run: `go test -v ./internal/agent/... ./internal/job/...`
Expected: PASS

- [ ] **Step 5: Commit**

```
feat(facts): add Containerized field to FactsRegistration and fact resolver
```

---

### Task 6: Wire Container Detection into Facts Collection

**Files:**
- Modify: `internal/agent/facts.go`
- Modify: `internal/agent/facts_public_test.go` (if test covers writeFacts fields)

- [ ] **Step 1: Add platform.IsContainer() call to writeFacts**

In `internal/agent/facts.go`, add the import:

```go
"github.com/retr0h/osapi/pkg/sdk/platform"
```

In the `writeFacts` method, add after `reg := job.FactsRegistration{}`:

```go
reg.Containerized = platform.IsContainer()
```

- [ ] **Step 2: Run tests**

Run: `go test -v ./internal/agent/...`
Expected: PASS (existing tests should still pass — `IsContainer()` returns false on the test host, matching the default zero value)

- [ ] **Step 3: Commit**

```
feat(agent): collect containerized fact during facts refresh
```

---

### Task 7: Wire DebianDocker DNS Provider into Agent Setup

**Files:**
- Modify: `cmd/agent_setup.go`

- [ ] **Step 1: Update the DNS provider switch**

In `cmd/agent_setup.go`, update the DNS provider selection:

```go
// --- Network providers ---
var dnsProvider dns.Provider
switch plat {
case "debian":
	if platform.IsContainer() {
		dnsProvider = dns.NewDebianDockerProvider(log, appFs)
	} else {
		dnsProvider = dns.NewDebianProvider(log, execManager)
	}
case "darwin":
	dnsProvider = dns.NewDarwinProvider(log, execManager)
default:
	dnsProvider = dns.NewLinuxProvider()
}
```

No new imports needed — `dns` and `platform` are already imported.

- [ ] **Step 2: Verify build**

Run: `go build ./...`
Expected: Compiles with no errors.

- [ ] **Step 3: Run all tests**

Run: `just go::unit`
Expected: PASS

- [ ] **Step 4: Run lint**

Run: `just go::vet`
Expected: Clean

- [ ] **Step 5: Commit**

```
feat(agent): wire DebianDocker DNS provider for containerized agents
```

---

### Task 8: Format and Final Verification

- [ ] **Step 1: Format code**

Run: `just go::fmt`

- [ ] **Step 2: Run full test suite**

Run: `just test`
Expected: PASS (lint + unit + coverage)

- [ ] **Step 3: Commit any formatting changes**

If `just go::fmt` produced changes:

```
style: format new container DNS provider files
```
