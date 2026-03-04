# Agent Facts Collection System — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add extensible fact collection to agents, stored in a separate KV bucket, merged into existing API responses, enabling orchestrator-side host filtering.

**Architecture:** Agents gather typed system facts (architecture, kernel, CPU, FQDN, network interfaces) on a 60s interval and write them to a dedicated `agent-facts` KV bucket. The job client merges facts into `AgentInfo` when serving `ListAgents`/`GetAgent`. The `Collector` interface provides the extension point for Phase 2 pluggable collectors (cloud metadata, local facts).

**Tech Stack:** Go 1.25, NATS JetStream KV, gopsutil, oapi-codegen, testify/suite, gomock

---

### Task 1: Add Types — NetworkInterface and FactsRegistration

**Files:**
- Modify: `internal/job/types.go`
- Test: `internal/job/types_public_test.go`

**Step 1: Write the failing test**

In `internal/job/types_public_test.go`, add a test for JSON round-trip of `FactsRegistration`:

```go
func (suite *TypesPublicTestSuite) TestFactsRegistrationJSON() {
	tests := []struct {
		name     string
		input    job.FactsRegistration
		expected string
	}{
		{
			name: "when all fields are populated",
			input: job.FactsRegistration{
				Architecture:  "amd64",
				KernelVersion: "5.15.0-91-generic",
				CPUCount:      4,
				FQDN:          "web-01.example.com",
				ServiceMgr:    "systemd",
				PackageMgr:    "apt",
				Interfaces: []job.NetworkInterface{
					{
						Name: "eth0",
						IPv4: "192.168.1.10",
						MAC:  "00:11:22:33:44:55",
					},
				},
				Facts: map[string]any{
					"cloud": map[string]any{
						"region": "us-east-1",
					},
				},
			},
		},
		{
			name:  "when all fields are empty",
			input: job.FactsRegistration{},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			data, err := json.Marshal(tc.input)
			suite.Require().NoError(err)

			var result job.FactsRegistration
			err = json.Unmarshal(data, &result)
			suite.Require().NoError(err)

			suite.Equal(tc.input.Architecture, result.Architecture)
			suite.Equal(tc.input.KernelVersion, result.KernelVersion)
			suite.Equal(tc.input.CPUCount, result.CPUCount)
			suite.Equal(tc.input.FQDN, result.FQDN)
			suite.Equal(len(tc.input.Interfaces), len(result.Interfaces))
		})
	}
}
```

> If `internal/job/types_public_test.go` does not exist, check for an existing
> test file (e.g., `subjects_public_test.go`) and add the suite there, or create
> the file with `package job_test` and a `TypesPublicTestSuite`.

**Step 2: Run test to verify it fails**

Run: `go test -run TestFactsRegistrationJSON -v ./internal/job/...`
Expected: FAIL — `FactsRegistration` and `NetworkInterface` undefined

**Step 3: Write minimal implementation**

Add to `internal/job/types.go` after the `AgentInfo` struct:

```go
// NetworkInterface represents a network interface with its address.
type NetworkInterface struct {
	// Name is the interface name (e.g., "eth0").
	Name string `json:"name"`
	// IPv4 is the primary IPv4 address.
	IPv4 string `json:"ipv4,omitempty"`
	// MAC is the hardware address.
	MAC string `json:"mac,omitempty"`
}

// FactsRegistration represents an agent's facts entry in the facts KV bucket.
// This is separate from AgentRegistration (heartbeat) to allow independent
// refresh intervals and TTLs.
type FactsRegistration struct {
	// Architecture is the CPU architecture (e.g., "amd64", "arm64").
	Architecture string `json:"architecture,omitempty"`
	// KernelVersion is the OS kernel version.
	KernelVersion string `json:"kernel_version,omitempty"`
	// CPUCount is the number of logical CPUs.
	CPUCount int `json:"cpu_count,omitempty"`
	// FQDN is the fully qualified domain name.
	FQDN string `json:"fqdn,omitempty"`
	// ServiceMgr is the init system (e.g., "systemd").
	ServiceMgr string `json:"service_mgr,omitempty"`
	// PackageMgr is the package manager (e.g., "apt", "yum").
	PackageMgr string `json:"package_mgr,omitempty"`
	// Interfaces lists network interfaces with addresses.
	Interfaces []NetworkInterface `json:"interfaces,omitempty"`
	// Facts contains extended facts from pluggable collectors.
	Facts map[string]any `json:"facts,omitempty"`
}
```

Also add the same typed fields to `AgentInfo` (the fields that the API consumer sees):

```go
// Add these fields to the existing AgentInfo struct, after AgentVersion:

	// Architecture is the CPU architecture (e.g., "amd64", "arm64").
	Architecture string `json:"architecture,omitempty"`
	// KernelVersion is the OS kernel version.
	KernelVersion string `json:"kernel_version,omitempty"`
	// CPUCount is the number of logical CPUs.
	CPUCount int `json:"cpu_count,omitempty"`
	// FQDN is the fully qualified domain name.
	FQDN string `json:"fqdn,omitempty"`
	// ServiceMgr is the init system (e.g., "systemd").
	ServiceMgr string `json:"service_mgr,omitempty"`
	// PackageMgr is the package manager (e.g., "apt", "yum").
	PackageMgr string `json:"package_mgr,omitempty"`
	// Interfaces lists network interfaces with addresses.
	Interfaces []NetworkInterface `json:"interfaces,omitempty"`
	// Facts contains extended facts from pluggable collectors.
	Facts map[string]any `json:"facts,omitempty"`
```

**Step 4: Run test to verify it passes**

Run: `go test -run TestFactsRegistrationJSON -v ./internal/job/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/job/types.go internal/job/types_public_test.go
git commit -m "feat(job): add NetworkInterface and FactsRegistration types"
```

---

### Task 2: Add Config Types — NATSFacts and AgentFacts

**Files:**
- Modify: `internal/config/types.go`

**Step 1: Add config structs**

Add `NATSFacts` after `NATSRegistry` in `internal/config/types.go`:

```go
// NATSFacts configuration for the agent facts KV bucket.
type NATSFacts struct {
	// Bucket is the KV bucket name for agent facts entries.
	Bucket   string `mapstructure:"bucket"`
	TTL      string `mapstructure:"ttl"`     // e.g. "5m"
	Storage  string `mapstructure:"storage"` // "file" or "memory"
	Replicas int    `mapstructure:"replicas"`
}
```

Add `Facts` field to the `NATS` struct:

```go
Facts    NATSFacts    `mapstructure:"facts,omitempty"`
```

Add `AgentFacts` after `AgentConsumer`:

```go
// AgentFacts configuration for agent fact collection.
type AgentFacts struct {
	// Interval is how often facts are refreshed (e.g., "60s").
	Interval   string   `mapstructure:"interval"`
	// Collectors lists enabled fact collectors.
	Collectors []string `mapstructure:"collectors"`
}
```

Add `Facts` field to `AgentConfig`:

```go
// Facts configuration for agent fact collection.
Facts AgentFacts `mapstructure:"facts,omitempty"`
```

**Step 2: Verify build**

Run: `go build ./...`
Expected: compiles

**Step 3: Commit**

```bash
git add internal/config/types.go
git commit -m "feat(config): add NATSFacts and AgentFacts config types"
```

---

### Task 3: Add Collector Interface

**Files:**
- Create: `internal/agent/facts/types.go`
- Test: `internal/agent/facts/types_public_test.go`

**Step 1: Create the interface**

Create `internal/agent/facts/types.go`:

```go
package facts

import "context"

// Collector gathers extended facts from a specific source.
// Built-in collectors (system, hardware, network) are not Collectors —
// they are gathered directly. This interface is for pluggable extensions
// (cloud metadata, local facts) added in Phase 2.
type Collector interface {
	// Name returns the collector's namespace (e.g., "cloud", "local").
	Name() string
	// Collect gathers facts. Returns nil if not applicable (e.g., cloud
	// collector on bare metal). Errors are non-fatal.
	Collect(ctx context.Context) (map[string]any, error)
}
```

**Step 2: Write a test confirming the interface is usable**

Create `internal/agent/facts/types_public_test.go`:

```go
package facts_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/agent/facts"
)

type TypesPublicTestSuite struct {
	suite.Suite
}

func TestTypesPublicTestSuite(t *testing.T) {
	suite.Run(t, new(TypesPublicTestSuite))
}

// stubCollector is a test double that implements Collector.
type stubCollector struct {
	name string
	data map[string]any
	err  error
}

func (s *stubCollector) Name() string { return s.name }

func (s *stubCollector) Collect(_ context.Context) (map[string]any, error) {
	return s.data, s.err
}

func (suite *TypesPublicTestSuite) TestCollectorInterface() {
	tests := []struct {
		name         string
		collector    facts.Collector
		expectedName string
		expectedData map[string]any
	}{
		{
			name: "when collector returns data",
			collector: &stubCollector{
				name: "cloud",
				data: map[string]any{"region": "us-east-1"},
			},
			expectedName: "cloud",
			expectedData: map[string]any{"region": "us-east-1"},
		},
		{
			name: "when collector returns nil",
			collector: &stubCollector{
				name: "cloud",
				data: nil,
			},
			expectedName: "cloud",
			expectedData: nil,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.Equal(tc.expectedName, tc.collector.Name())
			data, _ := tc.collector.Collect(context.Background())
			suite.Equal(tc.expectedData, data)
		})
	}
}
```

**Step 3: Run test**

Run: `go test -v ./internal/agent/facts/...`
Expected: PASS

**Step 4: Commit**

```bash
git add internal/agent/facts/
git commit -m "feat(agent): add Collector interface for extensible fact gathering"
```

---

### Task 4: Facts KV Bucket Infrastructure

**Files:**
- Modify: `cmd/nats_helpers.go` — create facts KV bucket
- Modify: `cmd/api_helpers.go` — add `factsKV` to `natsBundle`, pass to job client
- Modify: `internal/cli/nats.go` — add `BuildFactsKVConfig`
- Modify: `internal/job/client/client.go` — add `FactsKV` to `Options` and `factsKV` to `Client`

**Step 1: Add BuildFactsKVConfig**

In `internal/cli/nats.go`, add after `BuildRegistryKVConfig`:

```go
// BuildFactsKVConfig builds a jetstream.KeyValueConfig from facts config values.
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

In `cmd/nats_helpers.go`, add after the registry KV bucket block (after line 165):

```go
// Create facts KV bucket with configured settings
if appConfig.NATS.Facts.Bucket != "" {
	factsKVConfig := cli.BuildFactsKVConfig(namespace, appConfig.NATS.Facts)
	if _, err := nc.CreateOrUpdateKVBucketWithConfig(ctx, factsKVConfig); err != nil {
		return fmt.Errorf("create facts KV bucket %s: %w", factsKVConfig.Bucket, err)
	}
}
```

**Step 3: Add factsKV to natsBundle and job client**

In `cmd/api_helpers.go`, add `factsKV` to `natsBundle`:

```go
type natsBundle struct {
	nc         messaging.NATSClient
	jobClient  jobclient.JobClient
	jobsKV     jetstream.KeyValue
	registryKV jetstream.KeyValue
	factsKV    jetstream.KeyValue
}
```

In `connectNATSBundle`, after the registryKV creation (after line 154), add:

```go
var factsKV jetstream.KeyValue
if appConfig.NATS.Facts.Bucket != "" {
	factsKVConfig := cli.BuildFactsKVConfig(namespace, appConfig.NATS.Facts)
	factsKV, err = nc.CreateOrUpdateKVBucketWithConfig(ctx, factsKVConfig)
	if err != nil {
		cli.LogFatal(log, "failed to create facts KV bucket", err)
	}
}
```

Add `FactsKV: factsKV` to the `jobclient.Options` and `factsKV: factsKV` to the returned `natsBundle`.

**Step 4: Add FactsKV to job client**

In `internal/job/client/client.go`, add `factsKV` field to `Client`:

```go
factsKV    jetstream.KeyValue // agent-facts KV (optional)
```

Add `FactsKV` to `Options`:

```go
// FactsKV is the KV bucket for agent facts (optional).
FactsKV jetstream.KeyValue
```

In `New()`, add: `factsKV: opts.FactsKV,`

**Step 5: Add factsKV to KVInfoFn in api_helpers.go**

In `newMetricsProvider` `KVInfoFn`, add `b.factsKV` to the buckets slice:

```go
buckets := []jetstream.KeyValue{jobsKV, registryKV, factsKV, auditKV}
```

Update the function signature to accept `factsKV jetstream.KeyValue` and
pass `b.factsKV` from `setupAPIServer`.

**Step 6: Verify build**

Run: `go build ./...`
Expected: compiles

**Step 7: Commit**

```bash
git add internal/cli/nats.go cmd/nats_helpers.go cmd/api_helpers.go internal/job/client/client.go
git commit -m "feat(nats): add facts KV bucket infrastructure"
```

---

### Task 5: Facts Writer in Agent

**Files:**
- Create: `internal/agent/facts.go`
- Create: `internal/agent/facts_test.go` (internal tests for private functions)
- Modify: `internal/agent/types.go` — add `factsKV` and `factCollectors` fields
- Modify: `internal/agent/agent.go` — add `factsKV` param to `New()`, call `startFacts()` from `Start()`
- Modify: `cmd/agent_helpers.go` — pass `factsKV` to agent

**Step 1: Add fields to Agent struct**

In `internal/agent/types.go`, add after `registryKV`:

```go
// Facts KV for writing agent facts
factsKV jetstream.KeyValue

// Pluggable fact collectors (Phase 2)
factCollectors []facts.Collector
```

**Step 2: Update Agent constructor**

In `internal/agent/agent.go`, add `factsKV jetstream.KeyValue` parameter to
`New()` (after `registryKV`). Assign it: `factsKV: factsKV,`

**Step 3: Write failing test for writeFacts**

Create `internal/agent/facts_test.go`:

```go
package agent

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/job"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
)

type FactsTestSuite struct {
	suite.Suite

	mockCtrl *gomock.Controller
	mockKV   *jobmocks.MockKeyValue
	agent    *Agent
}

func TestFactsTestSuite(t *testing.T) {
	suite.Run(t, new(FactsTestSuite))
}

func (s *FactsTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockKV = jobmocks.NewMockKeyValue(s.mockCtrl)

	s.agent = &Agent{
		logger:    slog.Default(),
		appConfig: config.Config{},
		factsKV:   s.mockKV,
	}
}

func (s *FactsTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *FactsTestSuite) TestWriteFacts() {
	tests := []struct {
		name     string
		setupMock func()
		validate func()
	}{
		{
			name: "when Put succeeds writes facts",
			setupMock: func() {
				s.mockKV.EXPECT().
					Put(gomock.Any(), "facts.test_host", gomock.Any()).
					DoAndReturn(func(_ context.Context, _ string, data []byte) (uint64, error) {
						var reg job.FactsRegistration
						err := json.Unmarshal(data, &reg)
						s.Require().NoError(err)
						s.NotEmpty(reg.Architecture)
						s.Greater(reg.CPUCount, 0)
						return uint64(1), nil
					})
			},
		},
		{
			name: "when Put fails logs warning",
			setupMock: func() {
				s.mockKV.EXPECT().
					Put(gomock.Any(), "facts.test_host", gomock.Any()).
					Return(uint64(0), fmt.Errorf("put failed"))
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setupMock()
			s.agent.writeFacts(context.Background(), "test-host")
		})
	}
}
```

**Step 4: Run test to verify it fails**

Run: `go test -run TestWriteFacts -v ./internal/agent/...`
Expected: FAIL — `writeFacts` undefined

**Step 5: Implement facts.go**

Create `internal/agent/facts.go`:

```go
package agent

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"runtime"
	"time"

	"github.com/retr0h/osapi/internal/job"
)

// factsInterval is the interval between fact refreshes.
var factsInterval = 60 * time.Second

// Package-level functions for testability.
var (
	runtimeGOARCH  = func() string { return runtime.GOARCH }
	runtimeNumCPU  = func() int { return runtime.NumCPU() }
	netInterfaces  = net.Interfaces
	osHostname     = getHostname
)

// startFacts writes initial facts, spawns a goroutine that refreshes on a
// ticker, and stops on ctx.Done().
func (a *Agent) startFacts(
	ctx context.Context,
	hostname string,
) {
	if a.factsKV == nil {
		return
	}

	a.writeFacts(ctx, hostname)

	a.logger.Info(
		"facts writer started",
		slog.String("hostname", hostname),
		slog.String("interval", factsInterval.String()),
	)

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

// writeFacts gathers system facts and writes them to the facts KV bucket.
func (a *Agent) writeFacts(
	ctx context.Context,
	hostname string,
) {
	reg := job.FactsRegistration{
		Architecture: runtimeGOARCH(),
		CPUCount:     runtimeNumCPU(),
	}

	// Kernel version from host provider
	if a.hostProvider != nil {
		if info, err := a.hostProvider.GetOSInfo(); err == nil && info != nil {
			// Use KernelVersion if available from gopsutil
			// OSInfo is already in heartbeat; kernel comes from the same source
		}
	}

	// FQDN
	if fqdn, err := osHostname(); err == nil {
		reg.FQDN = fqdn
	}

	// Network interfaces
	reg.Interfaces = gatherInterfaces()

	// Service manager detection
	reg.ServiceMgr = detectServiceMgr()

	// Package manager detection
	reg.PackageMgr = detectPackageMgr()

	// Pluggable collectors (Phase 2)
	if len(a.factCollectors) > 0 {
		reg.Facts = make(map[string]any)
		for _, c := range a.factCollectors {
			if data, err := c.Collect(ctx); err == nil && data != nil {
				reg.Facts[c.Name()] = data
			}
		}
	}

	data, err := json.Marshal(reg)
	if err != nil {
		a.logger.Warn(
			"failed to marshal facts",
			slog.String("hostname", hostname),
			slog.String("error", err.Error()),
		)
		return
	}

	key := factsKey(hostname)
	if _, err := a.factsKV.Put(ctx, key, data); err != nil {
		a.logger.Warn(
			"failed to write facts",
			slog.String("hostname", hostname),
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
	}
}

// factsKey returns the KV key for an agent's facts entry.
func factsKey(
	hostname string,
) string {
	return "facts." + job.SanitizeHostname(hostname)
}

// gatherInterfaces returns a list of non-loopback network interfaces.
func gatherInterfaces() []job.NetworkInterface {
	ifaces, err := netInterfaces()
	if err != nil {
		return nil
	}

	var result []job.NetworkInterface
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if iface.Flags&net.FlagUp == 0 {
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

	return result
}

// getHostname returns the FQDN of the current host.
func getHostname() (string, error) {
	return osHostnameFunc()
}

var osHostnameFunc = func() (string, error) {
	return "", nil // Platform-specific; overridden in tests
}

// detectServiceMgr detects the init system.
func detectServiceMgr() string {
	// Check if systemd is running by looking for its PID 1 comm
	// This is a best-effort detection; returns "" if unknown
	return "" // Platform-specific implementation
}

// detectPackageMgr detects the package manager.
func detectPackageMgr() string {
	return "" // Platform-specific implementation
}
```

> **Note:** The exact implementations of `detectServiceMgr`,
> `detectPackageMgr`, `getHostname`, and kernel version depend on the
> platform. Start with stubs that return empty strings. Implement the
> Linux-specific logic in `facts_linux.go` with build tags. For now,
> the tests verify the structure and KV write mechanics.

**Step 6: Wire into Agent Start()**

In `internal/agent/server.go`, after the `a.startHeartbeat(a.ctx, hostname)`
call, add:

```go
a.startFacts(a.ctx, hostname)
```

**Step 7: Update agent_helpers.go**

In `cmd/agent_helpers.go`, pass `b.factsKV` to `agent.New()`.

**Step 8: Run tests**

Run: `go test -run TestWriteFacts -v ./internal/agent/...`
Expected: PASS

Run: `go build ./...`
Expected: compiles

**Step 9: Commit**

```bash
git add internal/agent/facts.go internal/agent/facts_test.go \
       internal/agent/types.go internal/agent/agent.go \
       internal/agent/server.go cmd/agent_helpers.go
git commit -m "feat(agent): add facts writer with system fact collection"
```

---

### Task 6: Merge Facts into ListAgents and GetAgent

**Files:**
- Modify: `internal/job/client/query.go` — merge facts KV data into AgentInfo
- Test: `internal/job/client/query_public_test.go` — test facts merging

**Step 1: Write the failing test**

Add a test case to the existing `TestListAgents` or add `TestListAgentsWithFacts`
in `internal/job/client/query_public_test.go`:

```go
func (suite *QueryPublicTestSuite) TestListAgentsWithFacts() {
	tests := []struct {
		name              string
		setupRegistryKV   func(kv *jobmocks.MockKeyValue)
		setupFactsKV      func(kv *jobmocks.MockKeyValue)
		expectedArch      string
		expectedCPUCount  int
	}{
		{
			name: "when facts KV has data merges into agent info",
			setupRegistryKV: func(kv *jobmocks.MockKeyValue) {
				reg := job.AgentRegistration{
					Hostname: "server1",
				}
				data, _ := json.Marshal(reg)
				entry := jobmocks.NewMockKeyValueEntry(data)
				kv.EXPECT().Keys(gomock.Any()).Return([]string{"agents.server1"}, nil)
				kv.EXPECT().Get(gomock.Any(), "agents.server1").Return(entry, nil)
			},
			setupFactsKV: func(kv *jobmocks.MockKeyValue) {
				facts := job.FactsRegistration{
					Architecture: "amd64",
					CPUCount:     8,
				}
				data, _ := json.Marshal(facts)
				entry := jobmocks.NewMockKeyValueEntry(data)
				kv.EXPECT().Get(gomock.Any(), "facts.server1").Return(entry, nil)
			},
			expectedArch:     "amd64",
			expectedCPUCount: 8,
		},
		{
			name: "when facts KV is nil degrades gracefully",
			setupRegistryKV: func(kv *jobmocks.MockKeyValue) {
				reg := job.AgentRegistration{
					Hostname: "server1",
				}
				data, _ := json.Marshal(reg)
				entry := jobmocks.NewMockKeyValueEntry(data)
				kv.EXPECT().Keys(gomock.Any()).Return([]string{"agents.server1"}, nil)
				kv.EXPECT().Get(gomock.Any(), "agents.server1").Return(entry, nil)
			},
			// factsKV is nil — no setupFactsKV
			expectedArch:     "",
			expectedCPUCount: 0,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			// Build client with registry and optional facts KV
			// Verify agent info has merged facts
		})
	}
}
```

> Adapt this to match the existing test patterns in `query_public_test.go`.
> Look at how `TestListAgents` is structured and follow the same mock setup.

**Step 2: Run test to verify it fails**

Run: `go test -run TestListAgentsWithFacts -v ./internal/job/client/...`
Expected: FAIL — facts not merged

**Step 3: Implement facts merging**

In `internal/job/client/query.go`, modify `ListAgents`:

```go
func (c *Client) ListAgents(
	ctx context.Context,
) ([]job.AgentInfo, error) {
	if c.registryKV == nil {
		return nil, fmt.Errorf("agent registry not configured")
	}

	keys, err := c.registryKV.Keys(ctx)
	if err != nil {
		if err.Error() == "nats: no keys found" {
			return []job.AgentInfo{}, nil
		}
		return nil, fmt.Errorf("failed to list registry keys: %w", err)
	}

	agents := make([]job.AgentInfo, 0, len(keys))
	for _, key := range keys {
		entry, err := c.registryKV.Get(ctx, key)
		if err != nil {
			continue
		}

		var reg job.AgentRegistration
		if err := json.Unmarshal(entry.Value(), &reg); err != nil {
			continue
		}

		info := agentInfoFromRegistration(&reg)
		c.mergeFacts(ctx, &info)
		agents = append(agents, info)
	}

	return agents, nil
}
```

Add `mergeFacts` helper:

```go
// mergeFacts reads facts from the facts KV and overlays them onto AgentInfo.
func (c *Client) mergeFacts(
	ctx context.Context,
	info *job.AgentInfo,
) {
	if c.factsKV == nil {
		return
	}

	key := "facts." + job.SanitizeHostname(info.Hostname)
	entry, err := c.factsKV.Get(ctx, key)
	if err != nil {
		return // facts not available — not an error
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

Also update `GetAgent` to call `c.mergeFacts(ctx, &info)` after building
the AgentInfo from registration.

**Step 4: Run tests**

Run: `go test -run TestListAgentsWithFacts -v ./internal/job/client/...`
Expected: PASS

Run: `go test -v ./internal/job/client/...`
Expected: all tests PASS

**Step 5: Commit**

```bash
git add internal/job/client/query.go internal/job/client/query_public_test.go
git commit -m "feat(job): merge facts KV data into ListAgents and GetAgent"
```

---

### Task 7: OpenAPI Spec and API Handler

**Files:**
- Modify: `internal/api/agent/gen/api.yaml` — extend AgentInfo schema
- Run: `go generate ./internal/api/agent/gen/...`
- Modify: `internal/api/agent/agent_list.go` — update `buildAgentInfo`

**Step 1: Extend OpenAPI spec**

In `internal/api/agent/gen/api.yaml`, add new properties to `AgentInfo`:

```yaml
    AgentInfo:
      type: object
      properties:
        # ... existing properties ...
        architecture:
          type: string
          description: CPU architecture (e.g., "amd64", "arm64").
          example: "amd64"
        kernel_version:
          type: string
          description: OS kernel version.
          example: "5.15.0-91-generic"
        cpu_count:
          type: integer
          description: Number of logical CPUs.
          example: 4
        fqdn:
          type: string
          description: Fully qualified domain name.
          example: "web-01.example.com"
        service_mgr:
          type: string
          description: Init system (e.g., "systemd").
          example: "systemd"
        package_mgr:
          type: string
          description: Package manager (e.g., "apt", "yum").
          example: "apt"
        interfaces:
          type: array
          items:
            $ref: '#/components/schemas/NetworkInterfaceResponse'
          description: Network interfaces with addresses.
        facts:
          type: object
          additionalProperties: true
          description: Extended facts from pluggable collectors.
```

Add `NetworkInterfaceResponse` schema:

```yaml
    NetworkInterfaceResponse:
      type: object
      description: A network interface with its address.
      properties:
        name:
          type: string
          description: Interface name.
          example: "eth0"
        ipv4:
          type: string
          description: Primary IPv4 address.
          example: "192.168.1.10"
        mac:
          type: string
          description: Hardware (MAC) address.
          example: "00:11:22:33:44:55"
      required:
        - name
```

**Step 2: Regenerate**

Run: `go generate ./internal/api/agent/gen/...`
Expected: `agent.gen.go` regenerated with new fields

**Step 3: Update buildAgentInfo**

In `internal/api/agent/agent_list.go`, add mappings after the existing
`MemoryStats` block:

```go
if a.Architecture != "" {
	info.Architecture = &a.Architecture
}

if a.KernelVersion != "" {
	info.KernelVersion = &a.KernelVersion
}

if a.CPUCount > 0 {
	cpuCount := a.CPUCount
	info.CpuCount = &cpuCount
}

if a.FQDN != "" {
	info.Fqdn = &a.FQDN
}

if a.ServiceMgr != "" {
	info.ServiceMgr = &a.ServiceMgr
}

if a.PackageMgr != "" {
	info.PackageMgr = &a.PackageMgr
}

if len(a.Interfaces) > 0 {
	ifaces := make([]gen.NetworkInterfaceResponse, 0, len(a.Interfaces))
	for _, iface := range a.Interfaces {
		ni := gen.NetworkInterfaceResponse{Name: iface.Name}
		if iface.IPv4 != "" {
			ni.Ipv4 = &iface.IPv4
		}
		if iface.MAC != "" {
			ni.Mac = &iface.MAC
		}
		ifaces = append(ifaces, ni)
	}
	info.Interfaces = &ifaces
}

if len(a.Facts) > 0 {
	facts := gen.AgentInfo_Facts{AdditionalProperties: a.Facts}
	info.Facts = &facts
}
```

> **Note:** The exact generated field names may differ. Check the generated
> `agent.gen.go` after running `go generate` and match the field names.

**Step 4: Run tests**

Run: `go test -v ./internal/api/agent/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/api/agent/gen/api.yaml internal/api/agent/gen/agent.gen.go \
       internal/api/agent/agent_list.go
git commit -m "feat(api): expose agent facts in AgentInfo responses"
```

---

### Task 8: Default Config and Documentation

**Files:**
- Modify: `osapi.yaml` — add `nats.facts` and `agent.facts` sections
- Modify: `docs/docs/sidebar/usage/configuration.md` — document new config

**Step 1: Add defaults to osapi.yaml**

```yaml
nats:
  # ... existing sections ...

  # ── Facts KV bucket ──────────────────────────────────────
  facts:
    # KV bucket for agent facts entries.
    bucket: 'agent-facts'
    # TTL for facts entries (Go duration). Agents refresh
    # every 60s; the TTL acts as a staleness timeout.
    ttl: '5m'
    # Storage backend: "file" or "memory".
    storage: 'file'
    # Number of KV replicas.
    replicas: 1

agent:
  # ... existing sections ...

  # Fact collection settings.
  facts:
    # How often to refresh facts (Go duration).
    interval: '60s'
    # Enabled fact collectors. Built-in: system, hardware, network.
    # Phase 2: cloud, local.
    collectors:
      - system
      - hardware
      - network
```

**Step 2: Update configuration docs**

Add env vars table, section reference, and full reference entries for the
new config sections in `docs/docs/sidebar/usage/configuration.md`.

**Step 3: Commit**

```bash
git add osapi.yaml docs/docs/sidebar/usage/configuration.md
git commit -m "docs: add facts KV and agent facts configuration"
```

---

### Task 9: Update agentInfoFromRegistration Helper

**Files:**
- Modify: `internal/job/client/query.go` — extend `agentInfoFromRegistration`

**Step 1: Update the helper**

The existing `agentInfoFromRegistration` maps `AgentRegistration` → `AgentInfo`.
Since `AgentInfo` now has new fields but `AgentRegistration` does not (facts
come from the facts KV), no changes are needed to this function. The merge
happens in `mergeFacts`.

> Verify that `agentInfoFromRegistration` compiles with the new `AgentInfo`
> fields (they have zero values by default).

**Step 2: Verify build and all tests**

Run: `go build ./...`
Run: `just go::unit`
Run: `just go::vet`
Expected: all pass

**Step 3: Commit (if any changes were needed)**

```bash
git commit -m "chore: verify agentInfoFromRegistration with new fields"
```

---

### Task 10: Final Verification

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

All must pass. Fix any issues found.

---

## Out of Scope (Phase 2+)

- Cloud metadata collector (AWS/GCP/Azure)
- Local facts collector (`/etc/osapi/facts.d/`)
- `agent.facts.collectors` config wiring (currently collectors are hardcoded)
- Linux-specific implementations of `detectServiceMgr`, `detectPackageMgr`
- Kernel version from gopsutil `host.KernelVersion()`
- Orchestrator DSL extensions (`Discover`, `WhenFact`, `GroupByFact`)
- SDK sync and regeneration
