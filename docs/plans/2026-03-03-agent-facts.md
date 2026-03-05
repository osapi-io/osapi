# Agent Facts Collection System — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to
> implement this plan task-by-task.

**Goal:** Add extensible fact collection to agents via the provider layer,
stored in a separate KV bucket, merged into existing API responses, enabling
orchestrator-side host filtering.

**Architecture:** Extend `host.Provider` with new fact methods (architecture,
kernel, FQDN, CPU count, service manager, package manager). Create a new
`netinfo.Provider` for network interfaces. The agent gathers facts on a 60s
interval and writes them to a dedicated `agent-facts` KV bucket. The job client
merges facts into `AgentInfo` when serving `ListAgents`/`GetAgent`. No plugin
system — everything goes through providers.

**Tech Stack:** Go 1.25, NATS JetStream KV, gopsutil, oapi-codegen,
testify/suite, gomock

**Design doc:** `docs/plans/2026-03-03-agent-facts-design.md`

---

### Task 1: Add Types — NetworkInterface, FactsRegistration, AgentInfo fields

**Files:**

- Modify: `internal/job/types.go`
- Test: `internal/job/types_public_test.go` (or appropriate existing test file)

**Step 1: Write the failing test**

Add a test for JSON round-trip of `FactsRegistration` and `NetworkInterface`.
Use testify/suite table-driven pattern. Verify all fields serialize and
deserialize correctly, including the `Facts map[string]any` field.

**Step 2: Run test to verify it fails**

```bash
go test -run TestFactsRegistration -v ./internal/job/...
```

Expected: FAIL — types undefined.

**Step 3: Write minimal implementation**

Add to `internal/job/types.go`:

```go
// NetworkInterface represents a network interface with its address.
type NetworkInterface struct {
	Name string `json:"name"`
	IPv4 string `json:"ipv4,omitempty"`
	MAC  string `json:"mac,omitempty"`
}

// FactsRegistration represents an agent's facts entry in the facts KV bucket.
type FactsRegistration struct {
	Architecture  string             `json:"architecture,omitempty"`
	KernelVersion string             `json:"kernel_version,omitempty"`
	CPUCount      int                `json:"cpu_count,omitempty"`
	FQDN          string             `json:"fqdn,omitempty"`
	ServiceMgr    string             `json:"service_mgr,omitempty"`
	PackageMgr    string             `json:"package_mgr,omitempty"`
	Interfaces    []NetworkInterface `json:"interfaces,omitempty"`
	Facts         map[string]any     `json:"facts,omitempty"`
}
```

Add the same typed fields to the existing `AgentInfo` struct (after
`AgentVersion`): `Architecture`, `KernelVersion`, `CPUCount`, `FQDN`,
`ServiceMgr`, `PackageMgr`, `Interfaces`, `Facts`.

**Step 4: Run test to verify it passes**

```bash
go test -run TestFactsRegistration -v ./internal/job/...
```

**Step 5: Commit**

```
feat(job): add NetworkInterface and FactsRegistration types
```

---

### Task 2: Add Config Types — NATSFacts and AgentFacts

**Files:**

- Modify: `internal/config/types.go`

**Step 1: Add config structs**

Add `NATSFacts` after `NATSRegistry`:

```go
type NATSFacts struct {
	Bucket   string `mapstructure:"bucket"`
	TTL      string `mapstructure:"ttl"`
	Storage  string `mapstructure:"storage"`
	Replicas int    `mapstructure:"replicas"`
}
```

Add `Facts NATSFacts` field to the `NATS` struct.

Add `AgentFacts` after `AgentConsumer`:

```go
type AgentFacts struct {
	Interval string `mapstructure:"interval"`
}
```

Add `Facts AgentFacts` field to `AgentConfig`.

**Step 2: Verify build**

```bash
go build ./...
```

**Step 3: Commit**

```
feat(config): add NATSFacts and AgentFacts config types
```

---

### Task 3: Extend host.Provider with Fact Methods

**Files:**

- Modify: `internal/provider/node/host/types.go` — add methods to interface
- Modify: `internal/provider/node/host/ubuntu.go` — implement for Ubuntu
- Modify: `internal/provider/node/host/mocks/types.gen.go` — update mock
  defaults
- Test: `internal/provider/node/host/ubuntu_public_test.go` or similar

**Step 1: Write failing tests**

Add table-driven tests for each new method: `GetArchitecture`,
`GetKernelVersion`, `GetFQDN`, `GetCPUCount`, `GetServiceManager`,
`GetPackageManager`. Test success cases and that errors don't panic.

**Step 2: Run tests to verify they fail**

```bash
go test -run TestGetArchitecture -v ./internal/provider/node/host/...
```

**Step 3: Add methods to Provider interface**

In `internal/provider/node/host/types.go`:

```go
type Provider interface {
	GetUptime() (time.Duration, error)
	GetHostname() (string, error)
	GetOSInfo() (*OSInfo, error)
	GetArchitecture() (string, error)
	GetKernelVersion() (string, error)
	GetFQDN() (string, error)
	GetCPUCount() (int, error)
	GetServiceManager() (string, error)
	GetPackageManager() (string, error)
}
```

**Step 4: Implement in Ubuntu provider**

In `internal/provider/node/host/ubuntu.go`:

- `GetArchitecture()` → `runtime.GOARCH`
- `GetKernelVersion()` → `host.KernelVersion()` from gopsutil
- `GetFQDN()` → `os.Hostname()` (FQDN lookup optional)
- `GetCPUCount()` → `runtime.NumCPU()`
- `GetServiceManager()` → check `/run/systemd/system` existence → `"systemd"`
- `GetPackageManager()` → check executable existence (`apt`, `yum`, `dnf`)

Wrap gopsutil/stdlib calls in package-level function variables for testability,
following the existing pattern (e.g., `hostInfoFn`).

**Step 5: Regenerate mocks**

```bash
go generate ./internal/provider/node/host/...
```

Update `mocks/types.gen.go` to add defaults for new methods in
`NewDefaultMockProvider`:

```go
mock.EXPECT().GetArchitecture().Return("amd64", nil).AnyTimes()
mock.EXPECT().GetKernelVersion().Return("5.15.0-91-generic", nil).AnyTimes()
mock.EXPECT().GetFQDN().Return("default-hostname.local", nil).AnyTimes()
mock.EXPECT().GetCPUCount().Return(4, nil).AnyTimes()
mock.EXPECT().GetServiceManager().Return("systemd", nil).AnyTimes()
mock.EXPECT().GetPackageManager().Return("apt", nil).AnyTimes()
```

**Step 6: Run all tests**

```bash
go test -v ./internal/provider/node/host/...
go build ./...
```

**Step 7: Commit**

```
feat(provider): extend host.Provider with fact methods
```

---

### Task 4: Create netinfo.Provider for Network Interfaces

**Files:**

- Create: `internal/provider/network/netinfo/types.go`
- Create: `internal/provider/network/netinfo/netinfo.go`
- Create: `internal/provider/network/netinfo/mocks/` (generate)
- Test: `internal/provider/network/netinfo/netinfo_public_test.go`

**Step 1: Write failing test**

Test `GetInterfaces()` returns non-loopback, up interfaces with name, IPv4, and
MAC. Use table-driven pattern. Mock `net.Interfaces` via a package-level
function variable.

**Step 2: Define the interface**

In `types.go`:

```go
package netinfo

import "github.com/retr0h/osapi/internal/job"

type Provider interface {
	GetInterfaces() ([]job.NetworkInterface, error)
}
```

**Step 3: Implement**

In `netinfo.go`:

```go
package netinfo

import (
	"net"

	"github.com/retr0h/osapi/internal/job"
)

type Netinfo struct{}

func New() *Netinfo { return &Netinfo{} }

var netInterfacesFn = net.Interfaces

func (n *Netinfo) GetInterfaces() ([]job.NetworkInterface, error) {
	ifaces, err := netInterfacesFn()
	if err != nil {
		return nil, err
	}

	var result []job.NetworkInterface
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		ni := job.NetworkInterface{
			Name: iface.Name,
			MAC:  iface.HardwareAddr.String(),
		}

		addrs, err := iface.Addrs()
		if err == nil {
			for _, addr := range addrs {
				if ipNet, ok := addr.(*net.IPNet); ok && ipNet.IP.To4() != nil {
					ni.IPv4 = ipNet.IP.String()
					break
				}
			}
		}

		result = append(result, ni)
	}

	return result, nil
}
```

**Step 4: Generate mocks and add defaults**

```bash
# Add generate.go with //go:generate directive
go generate ./internal/provider/network/netinfo/...
```

Create `mocks/types.gen.go` with `NewDefaultMockProvider` returning a stub
interface list.

**Step 5: Run tests**

```bash
go test -v ./internal/provider/network/netinfo/...
```

**Step 6: Commit**

```
feat(provider): add netinfo.Provider for network interface facts
```

---

### Task 5: Facts KV Bucket Infrastructure

**Files:**

- Modify: `internal/cli/nats.go` — add `BuildFactsKVConfig`
- Modify: `cmd/nats_helpers.go` — create facts KV in `setupJetStream`
- Modify: `cmd/api_helpers.go` — add `factsKV` to `natsBundle`, pass to job
  client and metrics provider
- Modify: `internal/job/client/client.go` — add `FactsKV` to `Options` and
  `factsKV` to `Client`

**Step 1: Add BuildFactsKVConfig**

In `internal/cli/nats.go`, add after `BuildRegistryKVConfig` (follow the exact
same pattern):

```go
func BuildFactsKVConfig(
	namespace string,
	factsCfg config.NATSFacts,
) jetstream.KeyValueConfig {
	factsBucket := job.ApplyNamespaceToInfraName(namespace, factsCfg.Bucket)
	factsTTL, _ := time.ParseDuration(factsCfg.TTL)

	return jetstream.KeyValueConfig{
		Bucket:   factsBucket,
		TTL:      factsTTL,
		Storage:  ParseJetstreamStorageType(factsCfg.Storage),
		Replicas: factsCfg.Replicas,
	}
}
```

**Step 2: Create facts KV in setupJetStream**

In `cmd/nats_helpers.go`, add after the registry KV block (line ~165):

```go
if appConfig.NATS.Facts.Bucket != "" {
	factsKVConfig := cli.BuildFactsKVConfig(namespace, appConfig.NATS.Facts)
	if _, err := nc.CreateOrUpdateKVBucketWithConfig(ctx, factsKVConfig); err != nil {
		return fmt.Errorf("create facts KV bucket %s: %w", factsKVConfig.Bucket, err)
	}
}
```

**Step 3: Wire into natsBundle and job client**

Add `factsKV jetstream.KeyValue` to `natsBundle` struct.

In `connectNATSBundle`, create the facts KV bucket (only if configured) and pass
it as `FactsKV` in `jobclient.Options`.

Add `factsKV` to the returned `natsBundle`.

In `newMetricsProvider`, add `b.factsKV` to the `KVInfoFn` buckets slice.

**Step 4: Add to job client**

In `internal/job/client/client.go`:

- Add `FactsKV jetstream.KeyValue` to `Options`
- Add `factsKV jetstream.KeyValue` to `Client` struct
- Assign in `New()`: `factsKV: opts.FactsKV,`

**Step 5: Verify build**

```bash
go build ./...
```

**Step 6: Commit**

```
feat(nats): add facts KV bucket infrastructure
```

---

### Task 6: Facts Writer in Agent

**Files:**

- Create: `internal/agent/facts.go`
- Create: `internal/agent/facts_test.go` (internal tests)
- Modify: `internal/agent/types.go` — add `factsKV` and `netinfoProvider`
- Modify: `internal/agent/agent.go` — add params to `New()`, call `startFacts()`
  in `Start()`
- Modify: `internal/agent/factory.go` — create netinfo provider
- Modify: `cmd/agent_helpers.go` — pass `factsKV` and netinfo provider

**Step 1: Add fields to Agent struct**

In `internal/agent/types.go`, add:

```go
factsKV         jetstream.KeyValue
netinfoProvider netinfo.Provider
```

**Step 2: Update New() and factory**

In `internal/agent/agent.go`, add `netinfoProvider netinfo.Provider` and
`factsKV jetstream.KeyValue` parameters. Assign them.

In `internal/agent/factory.go`, add `netinfo.New()` to the provider factory
return values. Update `CreateProviders()` signature.

**Step 3: Write failing test for writeFacts**

Create `internal/agent/facts_test.go` (internal, `package agent`). Use
`FactsTestSuite` with gomock. Mock the `factsKV.Put()` call. Verify the written
data contains architecture, cpu_count, interfaces. Follow the existing
`heartbeat_test.go` pattern exactly.

Test cases:

- `"when Put succeeds writes facts"` — verify JSON contains expected fields
- `"when Put fails logs warning"` — verify no panic
- `"when marshal fails logs warning"` — override `marshalJSON` variable

**Step 4: Run test to verify it fails**

```bash
go test -run TestWriteFacts -v ./internal/agent/...
```

**Step 5: Implement facts.go**

Create `internal/agent/facts.go`:

```go
package agent

// factsInterval controls the fact refresh period.
var factsInterval = 60 * time.Second

func (a *Agent) startFacts(ctx context.Context, hostname string) {
	if a.factsKV == nil {
		return
	}
	a.writeFacts(ctx, hostname)
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		ticker := time.NewTicker(factsInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				a.writeFacts(ctx, hostname)
			}
		}
	}()
}

func (a *Agent) writeFacts(ctx context.Context, hostname string) {
	reg := job.FactsRegistration{}

	// Call providers — errors are non-fatal
	if arch, err := a.hostProvider.GetArchitecture(); err == nil {
		reg.Architecture = arch
	}
	if kv, err := a.hostProvider.GetKernelVersion(); err == nil {
		reg.KernelVersion = kv
	}
	if fqdn, err := a.hostProvider.GetFQDN(); err == nil {
		reg.FQDN = fqdn
	}
	if count, err := a.hostProvider.GetCPUCount(); err == nil {
		reg.CPUCount = count
	}
	if mgr, err := a.hostProvider.GetServiceManager(); err == nil {
		reg.ServiceMgr = mgr
	}
	if mgr, err := a.hostProvider.GetPackageManager(); err == nil {
		reg.PackageMgr = mgr
	}
	if ifaces, err := a.netinfoProvider.GetInterfaces(); err == nil {
		reg.Interfaces = ifaces
	}

	data, err := marshalJSON(reg)
	if err != nil {
		a.logger.Warn("failed to marshal facts", ...)
		return
	}

	key := factsKey(hostname)
	if _, err := a.factsKV.Put(ctx, key, data); err != nil {
		a.logger.Warn("failed to write facts", ...)
	}
}

func factsKey(hostname string) string {
	return "facts." + job.SanitizeHostname(hostname)
}
```

**Step 6: Wire into Start()**

In `internal/agent/server.go`, after `a.startHeartbeat(a.ctx, hostname)`:

```go
a.startFacts(a.ctx, hostname)
```

**Step 7: Update cmd/agent_helpers.go**

Pass `b.factsKV` and the netinfo provider to `agent.New()`.

**Step 8: Run tests**

```bash
go test -v ./internal/agent/...
go build ./...
```

**Step 9: Commit**

```
feat(agent): add facts writer with provider-based collection
```

---

### Task 7: Merge Facts into ListAgents and GetAgent

**Files:**

- Modify: `internal/job/client/query.go` — add `mergeFacts` helper
- Test: `internal/job/client/query_public_test.go`

**Step 1: Write failing test**

Add test cases for facts merging. Test:

- Facts KV has data → fields appear in AgentInfo
- Facts KV is nil → graceful degradation (fields empty)
- Facts KV Get returns error → graceful degradation

Follow existing test patterns in `query_public_test.go`.

**Step 2: Run test to verify it fails**

```bash
go test -run TestListAgentsWithFacts -v ./internal/job/client/...
```

**Step 3: Implement mergeFacts**

Add to `internal/job/client/query.go`:

```go
func (c *Client) mergeFacts(ctx context.Context, info *job.AgentInfo) {
	if c.factsKV == nil {
		return
	}

	key := "facts." + job.SanitizeHostname(info.Hostname)
	entry, err := c.factsKV.Get(ctx, key)
	if err != nil {
		return
	}

	var facts job.FactsRegistration
	if err := json.Unmarshal(entry.Value(), &facts); err != nil {
		return
	}

	info.Architecture = facts.Architecture
	info.KernelVersion = facts.KernelVersion
	info.CPUCount = facts.CPUCount
	info.FQDN = facts.FQDN
	info.ServiceMgr = facts.ServiceMgr
	info.PackageMgr = facts.PackageMgr
	info.Interfaces = facts.Interfaces
	info.Facts = facts.Facts
}
```

Call `c.mergeFacts(ctx, &info)` in both `ListAgents` (after
`agentInfoFromRegistration`) and `GetAgent` (after building info).

**Step 4: Run tests**

```bash
go test -v ./internal/job/client/...
```

**Step 5: Commit**

```
feat(job): merge facts KV data into ListAgents and GetAgent
```

---

### Task 8: OpenAPI Spec and API Handler

**Files:**

- Modify: `internal/api/agent/gen/api.yaml`
- Run: `go generate ./internal/api/agent/gen/...`
- Modify: `internal/api/agent/agent_list.go` — update `buildAgentInfo`
- Test: `internal/api/agent/agent_list_public_test.go` (or existing test)

**Step 1: Extend OpenAPI spec**

Add to `AgentInfo` properties in `api.yaml`:

```yaml
architecture:
  type: string
  description: CPU architecture.
  example: 'amd64'
kernel_version:
  type: string
  description: OS kernel version.
  example: '5.15.0-91-generic'
cpu_count:
  type: integer
  description: Number of logical CPUs.
  example: 4
fqdn:
  type: string
  description: Fully qualified domain name.
  example: 'web-01.example.com'
service_mgr:
  type: string
  description: Init system.
  example: 'systemd'
package_mgr:
  type: string
  description: Package manager.
  example: 'apt'
interfaces:
  type: array
  items:
    $ref: '#/components/schemas/NetworkInterfaceResponse'
facts:
  type: object
  additionalProperties: true
  description: Extended facts from additional providers.
```

Add `NetworkInterfaceResponse` schema:

```yaml
NetworkInterfaceResponse:
  type: object
  properties:
    name:
      type: string
      example: 'eth0'
    ipv4:
      type: string
      example: '192.168.1.10'
    mac:
      type: string
      example: '00:11:22:33:44:55'
  required:
    - name
```

**Step 2: Regenerate**

```bash
go generate ./internal/api/agent/gen/...
```

**Step 3: Update buildAgentInfo**

In `internal/api/agent/agent_list.go`, add mappings for new fields after the
existing memory block. Map each non-zero/non-empty field. Map `Interfaces` as
`[]gen.NetworkInterfaceResponse`.

Check the generated field names in `agent.gen.go` and match them exactly.

**Step 4: Run tests**

```bash
go test -v ./internal/api/agent/...
go build ./...
```

**Step 5: Commit**

```
feat(api): expose agent facts in AgentInfo responses
```

---

### Task 9: Default Config

**Files:**

- Modify: `osapi.yaml`

**Step 1: Add defaults**

Add `nats.facts` section after `nats.registry`:

```yaml
facts:
  bucket: 'agent-facts'
  ttl: '5m'
  storage: 'file'
  replicas: 1
```

Add `agent.facts` section after `agent.labels`:

```yaml
facts:
  interval: '60s'
```

**Step 2: Verify config loads**

```bash
go build ./...
```

**Step 3: Commit**

```
chore: add default facts config to osapi.yaml
```

---

### Task 10: Update Documentation — Configuration Reference

**Files:**

- Modify: `docs/docs/sidebar/usage/configuration.md`

**Step 1: Add environment variable mappings**

Add to the env var table:

| `nats.facts.bucket` | `OSAPI_NATS_FACTS_BUCKET` | | `nats.facts.ttl` |
`OSAPI_NATS_FACTS_TTL` | | `nats.facts.storage` | `OSAPI_NATS_FACTS_STORAGE` | |
`nats.facts.replicas` | `OSAPI_NATS_FACTS_REPLICAS` | | `agent.facts.interval` |
`OSAPI_AGENT_FACTS_INTERVAL` |

**Step 2: Add section references**

Add `nats.facts` section reference table (Bucket, TTL, Storage, Replicas). Add
`agent.facts` section reference table (Interval).

**Step 3: Update full YAML reference**

Add the `nats.facts` and `agent.facts` blocks to the full reference YAML with
inline comments.

**Step 4: Commit**

```
docs: add facts configuration reference
```

---

### Task 11: Update Documentation — Feature and Architecture Pages

**Files:**

- Modify: `docs/docs/sidebar/features/node-management.md`
- Modify: `docs/docs/sidebar/architecture/system-architecture.md`
- Modify: `docs/docs/sidebar/architecture/job-architecture.md`

**Step 1: Update node-management.md**

- In "Agent vs. Node" section, add that agents now expose typed system facts
  (architecture, kernel, FQDN, CPU count, network interfaces) in addition to the
  basic heartbeat metrics.
- Clarify: facts are gathered every 60s via providers, stored in a separate
  `agent-facts` KV bucket with a 5-minute TTL.
- Add a "System Facts" row to the "What It Manages" table.

**Step 2: Update system-architecture.md**

- Add `agent-facts` KV bucket to the NATS JetStream section alongside
  `agent-registry`.
- Update the component map table to mention facts in the Agent/Provider layer
  description.

**Step 3: Update job-architecture.md**

- Add a brief section on facts collection:
  - Facts are collected independently from the job system.
  - 60-second interval, separate KV bucket.
  - Providers gather system facts (architecture, kernel, network interfaces,
    etc.).
  - API merges registry + facts KV into a single AgentInfo response.

**Step 4: Commit**

```
docs: update feature and architecture pages with facts
```

---

### Task 12: Update Documentation — CLI Pages

**Files:**

- Modify: `docs/docs/sidebar/usage/cli/client/agent/list.md`
- Modify: `docs/docs/sidebar/usage/cli/client/agent/get.md`
- Modify: `docs/docs/sidebar/usage/cli/client/health/status.md`

**Step 1: Update agent list.md**

Update the example output to show any new facts-derived columns if the CLI is
updated to display them (e.g., ARCH column). If no CLI column changes are
planned for Phase 1, add a note that `--json` output includes full facts data.

**Step 2: Update agent get.md**

Add facts fields to the example output and field description table:

| Architecture | CPU architecture (e.g., amd64) | | Kernel | OS kernel version |
| FQDN | Fully qualified domain name | | CPUs | Number of logical CPUs | |
Service Mgr | Init system (e.g., systemd) | | Package Mgr | Package manager
(e.g., apt) | | Interfaces | Network interfaces with IPv4 and MAC |

Update the example output block to show these new fields.

**Step 3: Update health status.md**

Add `agent-facts` to the KV buckets section in the example output (e.g.,
`Bucket: agent-facts (2 keys, 1.5 KB)`).

**Step 4: Commit**

```
docs: update CLI docs with agent facts output
```

---

### Task 13: Final Verification

**Step 1: Build**

```bash
go build ./...
```

**Step 2: Unit tests**

```bash
just go::unit
```

**Step 3: Lint**

```bash
just go::vet
```

**Step 4: Format**

```bash
just go::fmt
```

**Step 5: Docs format**

```bash
just docs::fmt-check
```

All must pass. Fix any issues found.

---

## Out of Scope (Phase 2+)

- Cloud metadata provider (AWS/GCP/Azure metadata endpoints)
- Local facts provider (`/etc/osapi/facts.d/` JSON/YAML files)
- CLI column changes for `agent list` (facts available via `--json`)
- Orchestrator DSL extensions (`Discover`, `WhenFact`, `GroupByFact`)
- SDK sync and regeneration
- `Facts map[string]any` population (reserved for Phase 2 providers)
