# Agent Facts, Routes, Fact References, and Timeline Fix

## Context

Agents collect system facts (OS, memory, load, interfaces, etc.) but lack two
useful capabilities: (1) knowing the primary network interface and full routing
table, and (2) allowing CLI/API parameters to reference agent facts dynamically.
For example, a user should be able to run:

```
osapi client network dns get --interface-name @fact.interface.primary --target _all
```

...and have each agent resolve `@fact.interface.primary` to its own primary
interface before executing the operation.

Additionally, the `agent get` CLI output is missing timeline events
(cordon/uncordon history) — the data path exists but timeline should always be
displayed.

This is a multi-phase effort. All phases stay on a single branch before pushing
upstream.

## Repo

All changes in `osapi` at `/Users/john/git/osapi-io/osapi/`.

---

## Phase 1: Fix Timeline Display and Configs

### Step 1.1: Sync local/nerd configs with osapi.yaml

`configs/osapi.local.yaml` and `configs/osapi.nerd.yaml` are missing sections
that exist in `osapi.yaml`:

- **`nats.state`** — missing in both. This is why timeline isn't showing:
  `stateKV` is nil so `GetAgentTimeline()` returns early.
- **`nats.facts`** — missing in `osapi.nerd.yaml`
- **`telemetry.metrics`** — missing in both
- **`agent.facts`** — missing in `osapi.nerd.yaml`
- **`agent.conditions`** — missing in both

Add these sections to both configs to match `osapi.yaml`.

### Step 1.2: Always show Timeline section in agent get CLI

File: `cmd/client_agent_get.go`

Line 169: change `if len(data.Timeline) > 0` to always display the Timeline
section. Show empty table or "No events" when empty.

### Step 1.3: Always show Timeline section in job get CLI

File: `internal/cli/ui.go`

Line 601: same fix — change `if len(resp.Timeline) > 0` to always display the
Timeline section for job details.

---

## Phase 2: Route Collection and Primary Interface

### Step 2.1: Add Route type to job types

File: `internal/job/types.go`

Add a `Route` struct:

```go
type Route struct {
    Destination string `json:"destination"`
    Gateway     string `json:"gateway"`
    Interface   string `json:"interface"`
    Mask        string `json:"mask,omitempty"`
    Metric      int    `json:"metric,omitempty"`
    Flags       string `json:"flags,omitempty"`
}
```

Add fields to `FactsRegistration`:

```go
PrimaryInterface string  `json:"primary_interface,omitempty"`
Routes           []Route `json:"routes,omitempty"`
```

Add same fields to `AgentInfo`.

### Step 2.2: Add route provider to netinfo

File: `internal/provider/network/netinfo/types.go`

Extend `Provider` interface:

```go
type Provider interface {
    GetInterfaces() ([]job.NetworkInterface, error)
    GetRoutes() ([]job.Route, error)
    GetPrimaryInterface() (string, error)
}
```

### Step 2.3: Linux route implementation

File: `internal/provider/network/netinfo/linux_get_routes.go` (build tag:
`//go:build linux`)

Parse `/proc/net/route` using Go (no exec). Use injectable `RouteReaderFn`
(defaults to `os.Open("/proc/net/route")`) for testing. The default route
(destination `00000000`) determines the primary interface.

Return all routes as `[]job.Route` and identify the primary interface from the
default route entry.

### Step 2.4: Darwin route stub

File: `internal/provider/network/netinfo/darwin_get_routes.go` (build tag:
`//go:build darwin`)

Stub that returns empty routes and empty primary interface (or uses a heuristic
like first interface with a default gateway). Darwin route detection can be
improved later.

### Step 2.5: Collect routes in agent facts

File: `internal/agent/facts.go`

In `writeFacts()`, call `a.netinfoProvider.GetRoutes()` and
`a.netinfoProvider.GetPrimaryInterface()`. Add results to `FactsRegistration`.

Cache `FactsRegistration` on the Agent struct as `cachedFacts` for use by fact
reference resolution (Phase 3).

### Step 2.6: Expose via API and CLI

- `internal/job/client/query.go` `mergeFacts()`: map new fields
- `internal/api/agent/gen/api.yaml`: add `primary_interface` and `routes` to
  AgentInfo schema
- `internal/api/agent/agent_list.go` `buildAgentInfo()`: map fields
- `cmd/client_agent_get.go`: display primary interface and routes
- SDK: update `Agent` type and agent spec

### Step 2.7: Tests

- `internal/provider/network/netinfo/linux_get_routes_public_test.go`:
  table-driven tests for `/proc/net/route` parsing (mock file content via
  `RouteReaderFn`)
- Update existing facts test to verify new fields

---

## Phase 3: `@fact.X` Resolution

### Step 3.1: Fact reference resolver

New file: `internal/agent/factref.go`

```go
func ResolveFacts(
    params map[string]any,
    facts *job.FactsRegistration,
) (map[string]any, error)
```

Walk all string values in the params map. For each string containing `@fact.X`,
resolve against the facts struct:

- `@fact.interface.primary` → `facts.PrimaryInterface`
- `@fact.hostname` → agent hostname
- `@fact.arch` → `facts.Architecture`
- `@fact.os` → `facts.OSInfo` distribution
- `@fact.kernel` → `facts.KernelVersion`
- Extensible: `@fact.custom.X` → `facts.Facts["X"]`

If a reference cannot be resolved, return an error (fail the job). Multiple
references in one string are supported:
`"@fact.interface.primary on @fact.hostname"` → `"eth0 on web-01"`.

### Step 3.2: Inject resolution in handler

File: `internal/agent/handler.go`

In `handleJobMessage()`, after unmarshalling the `jobRequest` (line ~163) and
before `processJobOperation()` (line ~225), call `ResolveFacts()` on the job
request parameters using the agent's cached facts. Replace the request params
with resolved values.

If resolution fails (unresolvable reference), fail the job with an error message
indicating which fact reference could not be resolved.

### Step 3.3: Tests

File: `internal/agent/factref_public_test.go`

Table-driven tests:

- Simple substitution (`@fact.interface.primary` → `eth0`)
- Multiple references in one string
- Nested map values
- Unknown fact reference → error
- No `@fact.` references → params unchanged
- Nil facts → error for any reference
- Custom facts via `@fact.custom.X`

---

## Phase 4 (Future): File Upload and Templates

Deferred — will be planned separately after Phases 1-3 are complete. Will use
NATS Object Store for blob storage and Go `text/template` for file content
rendering with fact data.

---

## Verification

After each phase:

```bash
go build ./...
just go::unit
just go::vet
```

Integration test after Phase 2:

```bash
# Start osapi, then:
go run main.go client agent get --hostname <host> --json | jq .primary_interface
go run main.go client agent get --hostname <host> --json | jq .routes
```

Integration test after Phase 3:

```bash
go run main.go client network dns get \
  --interface-name @fact.interface.primary --target _all
```
