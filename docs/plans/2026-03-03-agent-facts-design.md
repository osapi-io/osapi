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
| System | architecture, kernel_version, fqdn, service_mgr, pkg_mgr | `runtime.GOARCH`, `host.KernelVersion()`, `os.Hostname()` |
| Hardware | cpu_count | `runtime.NumCPU()` or `cpu.Counts()` |
| Network | interfaces (name, ipv4, mac), default gateway | `net.Interfaces()` |

**Phase 2 — Pluggable collectors (opt-in):**

| Collector | Facts | Source |
|-----------|-------|--------|
| Cloud | instance_id, region, instance_type, public_ip, availability_zone | Cloud metadata endpoints (AWS/GCP/Azure `169.254.169.254`) |
| Local | arbitrary key-value data | JSON/YAML files in `/etc/osapi/facts.d/` |

All Phase 1 facts are sub-millisecond calls. Phase 2 collectors may involve
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
- Extended facts from pluggable collectors (cloud, local)
- 60s refresh, 5min TTL
- 1-10KB per agent (grows with extensible facts)

The API merges both KVs into a single `AgentInfo` response. Consumers never
know about the split.

### Fact Collector Interface

Extensible via a provider pattern in the agent:

```go
// internal/agent/facts/collector.go
type Collector interface {
    Name() string
    Collect(ctx context.Context) (map[string]any, error)
}
```

Built-in collectors (system, hardware, network) always run. Pluggable
collectors (cloud, local) are opt-in via config. Collector errors are
non-fatal — the agent writes whatever data it gathered.

### Data Structure

Common facts get typed fields for compile-time safety. Extended facts go
into a flexible map for forward compatibility:

```go
type AgentRegistration struct {
    // Existing fields (move to facts KV)
    OSInfo       *host.OSInfo          `json:"os_info,omitempty"`
    Uptime       string                `json:"uptime,omitempty"`
    LoadAverages *load.AverageStats    `json:"load_averages,omitempty"`
    MemoryStats  *mem.Stats            `json:"memory_stats,omitempty"`

    // New typed facts
    Architecture  string               `json:"architecture,omitempty"`
    KernelVersion string               `json:"kernel_version,omitempty"`
    CPUCount      int                  `json:"cpu_count,omitempty"`
    FQDN          string               `json:"fqdn,omitempty"`
    ServiceMgr    string               `json:"service_mgr,omitempty"`
    PackageMgr    string               `json:"package_mgr,omitempty"`
    Interfaces    []NetworkInterface   `json:"interfaces,omitempty"`

    // Extended facts from pluggable collectors
    Facts map[string]any `json:"facts,omitempty"`
}
```

### API Exposure

No new endpoints. Existing `GET /node` and `GET /node/{hostname}` return
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
    collectors:
      - system
      - hardware
      - network
      # - cloud    # auto-detect cloud platform
      # - local    # read /etc/osapi/facts.d/
    # local_dir: /etc/osapi/facts.d
```

## What Changes Where

### OSAPI (this repo)

1. `internal/job/types.go` — add new typed fields + `Facts map[string]any`
   to `AgentRegistration` and `AgentInfo`
2. `internal/agent/facts/` — new package with `Collector` interface and
   built-in collectors (system, hardware, network)
3. `internal/agent/agent.go` — initialize fact collectors, start fact
   refresh loop (separate from heartbeat)
4. `internal/agent/heartbeat.go` — slim down to just liveness fields
   (hostname, labels, timestamps)
5. `internal/config/` — add `nats.facts` and `agent.facts` config sections
6. `internal/api/agent/gen/api.yaml` — extend `AgentInfo` schema with new
   fact fields
7. `internal/job/client/query.go` — `ListAgents` and `GetAgent` merge
   registry + facts KVs
8. `internal/api/` — wire facts KV into API server startup

### SDK (osapi-sdk)

9. Sync api.yaml, regenerate — `AgentInfo` gets new fields automatically
10. No code changes needed

### Orchestrator (osapi-orchestrator)

11. `Discover()` method — query `Agent.List()`, apply fact predicates
12. Fact predicates — `OS()`, `Arch()`, `MinMemory()`, `FactEquals()`, etc.
13. `WhenFact()` step method
14. `GroupByFact()` method

## What This Does NOT Change

- NATS routing unchanged — `_all`, `_any`, labels work as before
- No agent-side filtering — facts filter at publisher (orchestrator) side
- No new API endpoints — facts are richer `AgentInfo` data
- Labels remain the primary routing mechanism; facts are for conditional
  logic and discovery
- Existing heartbeat liveness behavior unchanged

## Phases

- **Phase 1**: Typed facts (system, hardware, network), separate KV,
  `Collector` interface, API exposure, SDK sync
- **Phase 2**: Cloud metadata collector, local facts collector,
  `agent.facts` config section
- **Phase 3**: Orchestrator DSL extensions (`Discover`, `WhenFact`,
  `GroupByFact`)
