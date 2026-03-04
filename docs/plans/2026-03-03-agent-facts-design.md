# Design: Agent Fact Collection System

## Problem

OSAPI agents register with basic metadata via heartbeat (OS info, uptime, load,
memory), but there's no extensible fact collection system. The
osapi-orchestrator needs host-level facts to enable Ansible-style conditional
execution — "only run on Ubuntu hosts", "skip hosts with < 4GB RAM", "group
hosts by OS distribution".

Today the orchestrator can only target by hostname or label. It can't make
decisions based on what a host *is* (architecture, kernel, network interfaces,
cloud region).

## Design

### Fact Categories

**Phase 1 — Built-in facts (cheap, always collected):**

| Category | Facts | Source |
|----------|-------|--------|
| System | architecture, kernel_version, fqdn, service_mgr, pkg_mgr | `host.Provider` extensions |
| Hardware | cpu_count | `host.Provider` extension |
| Network | interfaces (name, ipv4, mac) | New `netinfo.Provider` |

**Phase 2 — Additional providers (opt-in):**

| Provider | Facts | Source |
|----------|-------|--------|
| Cloud | instance_id, region, instance_type, public_ip | Cloud metadata endpoints (AWS/GCP/Azure) |
| Local | arbitrary key-value data | JSON/YAML files in `/etc/osapi/facts.d/` |

All Phase 1 facts are sub-millisecond calls. Phase 2 providers may involve
network I/O (cloud metadata) or file I/O (local facts).

### Storage: Same API, Separate KV

The heartbeat serves two purposes today: liveness ("I'm alive") and state
("what I look like"). Splitting these lets each optimize independently.

**Registry KV (existing)** — lean heartbeat, frequent refresh:
- Hostname, labels, timestamps
- 10s refresh, 30s TTL
- ~200 bytes per agent

**Facts KV (new `agent-facts` bucket)** — richer data, less frequent:
- OS, architecture, kernel, CPU, memory, interfaces, load, uptime
- Extended facts from future providers
- 60s refresh, 5min TTL
- 1-10KB per agent (grows with future providers)

The API merges both KVs into a single `AgentInfo` response. Consumers never
know about the split.

### Provider Pattern (Not a Plugin System)

Facts are gathered through the existing provider layer — the same pattern
used for `hostProvider.GetOSInfo()`, `loadProvider.GetAverageStats()`, etc.
There is no plugin system and no `Collector` interface.

**Extend `host.Provider`** with new methods:
- `GetArchitecture() (string, error)`
- `GetKernelVersion() (string, error)`
- `GetFQDN() (string, error)`
- `GetCPUCount() (int, error)`
- `GetServiceManager() (string, error)`
- `GetPackageManager() (string, error)`

**New `netinfo.Provider`** for network interface facts:
- `GetInterfaces() ([]NetworkInterface, error)`

The facts writer calls these providers exactly like the heartbeat calls its
providers — errors are non-fatal, the agent writes whatever data it gathered.

Future cloud metadata and local facts would be additional providers added
to the agent when needed, following the same pattern.

### Data Structure

```go
type FactsRegistration struct {
    Architecture  string               `json:"architecture,omitempty"`
    KernelVersion string               `json:"kernel_version,omitempty"`
    CPUCount      int                  `json:"cpu_count,omitempty"`
    FQDN          string               `json:"fqdn,omitempty"`
    ServiceMgr    string               `json:"service_mgr,omitempty"`
    PackageMgr    string               `json:"package_mgr,omitempty"`
    Interfaces    []NetworkInterface   `json:"interfaces,omitempty"`
    Facts         map[string]any       `json:"facts,omitempty"`
}
```

The `Facts map[string]any` field is reserved for future providers that
produce unstructured data (cloud metadata, local facts).

### API Exposure

No new endpoints. Existing `GET /agent` and `GET /agent/{hostname}` return
`AgentInfo` which includes all facts. The API server reads both the registry
and facts KV buckets and merges them.

The orchestrator calls `Agent.List()` once and gets everything needed for
host filtering — no second API call.

### Orchestrator Integration

Facts enable four key patterns in the orchestrator DSL:

**1. Pre-routing host discovery (filter by facts):**
```go
hosts, _ := o.Discover(ctx, "_all",
    orchestrator.OS("Ubuntu"),
    orchestrator.Arch("amd64"),
    orchestrator.MinMemory(8 * GB),
)
```

**2. Fact-aware When guards:**
```go
o.CommandShell("_all", "apt upgrade -y").
    WhenFact(func(f orchestrator.Facts) bool {
        return f.OS.Distribution == "Ubuntu"
    })
```

**3. Group-by-fact (multi-distro playbooks):**
```go
groups, _ := o.GroupByFact(ctx, "os.distribution")
for distro, hosts := range groups {
    o.CommandShell(hosts[0], installCmd(distro))
}
```

**4. Facts in TaskFunc (custom logic):**
```go
o.TaskFunc("decide", func(ctx context.Context, r orchestrator.Results) (*sdk.Result, error) {
    agents, _ := r.ListAgents(ctx)
    // Use agent facts for decisions
})
```

### Configuration

```yaml
nats:
  facts:
    bucket: 'agent-facts'
    ttl: '5m'
    storage: 'file'
    replicas: 1

agent:
  facts:
    interval: '60s'
```

## What Changes Where

### OSAPI (this repo)

1. `internal/job/types.go` — add `NetworkInterface`, `FactsRegistration`,
   and new typed fields on `AgentInfo`
2. `internal/provider/node/host/types.go` — extend `Provider` interface
   with `GetArchitecture`, `GetKernelVersion`, `GetFQDN`, `GetCPUCount`,
   `GetServiceManager`, `GetPackageManager`
3. `internal/provider/node/host/ubuntu.go` (+ other platforms) — implement
   new methods
4. `internal/provider/network/netinfo/` — new provider for `GetInterfaces()`
5. `internal/agent/types.go` — add `factsKV` and `netinfoProvider` fields
6. `internal/agent/agent.go` — accept new provider, start facts loop
7. `internal/agent/facts.go` — facts writer (calls providers, writes KV)
8. `internal/agent/factory.go` — create netinfo provider
9. `internal/config/types.go` — add `NATSFacts` and `AgentFacts` config
10. `cmd/nats_helpers.go` — create facts KV bucket
11. `cmd/api_helpers.go` — wire factsKV into natsBundle and job client
12. `internal/job/client/client.go` — add `FactsKV` option
13. `internal/job/client/query.go` — merge facts into ListAgents/GetAgent
14. `internal/api/agent/gen/api.yaml` — extend AgentInfo schema
15. `internal/api/agent/agent_list.go` — update buildAgentInfo mapping
16. `osapi.yaml` — default config values
17. Documentation (see below)

### Documentation Updates

18. `docs/docs/sidebar/features/node-management.md` — update "Agent vs.
    Node" section to explain facts, add facts to "What It Manages" table
19. `docs/docs/sidebar/architecture/system-architecture.md` — add
    `agent-facts` KV bucket to component map, update NATS layers
20. `docs/docs/sidebar/architecture/job-architecture.md` — add section on
    facts collection, describe 60s interval and KV storage
21. `docs/docs/sidebar/usage/configuration.md` — add `nats.facts` and
    `agent.facts` config sections, env var table, section reference
22. `docs/docs/sidebar/usage/cli/client/agent/list.md` — update example
    output and column table with facts data
23. `docs/docs/sidebar/usage/cli/client/agent/get.md` — add facts fields
    to output example and field table
24. `docs/docs/sidebar/usage/cli/client/health/status.md` — add
    agent-facts bucket to KV buckets section

### SDK (osapi-sdk)

25. Sync api.yaml, regenerate — `AgentInfo` gets new fields automatically

### Orchestrator (osapi-orchestrator)

26. `Discover()` method — query `Agent.List()`, apply fact predicates
27. Fact predicates — `OS()`, `Arch()`, `MinMemory()`, `FactEquals()`, etc.
28. `WhenFact()` step method
29. `GroupByFact()` method

## What This Does NOT Change

- NATS routing unchanged — `_all`, `_any`, labels work as before
- No agent-side filtering — facts filter at publisher (orchestrator) side
- No new API endpoints — facts are richer `AgentInfo` data
- Labels remain the primary routing mechanism; facts are for conditional
  logic and discovery
- Existing heartbeat liveness behavior unchanged
- No plugin system — facts are gathered through the provider layer

## Phases

- **Phase 1**: Typed facts via providers, separate KV, API exposure, docs
- **Phase 2**: Cloud metadata provider, local facts provider
- **Phase 3**: Orchestrator DSL extensions (`Discover`, `WhenFact`,
  `GroupByFact`)
