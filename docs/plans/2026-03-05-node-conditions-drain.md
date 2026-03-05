# Node Conditions & Agent Drain Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to
> implement this plan task-by-task.

**Goal:** Add Kubernetes-inspired node conditions (MemoryPressure,
HighLoad, DiskPressure) and agent drain/cordon lifecycle to OSAPI.

**Architecture:** Conditions are evaluated agent-side on each heartbeat
tick using existing provider data, stored in AgentRegistration. Drain
uses append-only timeline events in the registry KV bucket (reusing the
existing `TimelineEvent` type from job lifecycle), with a separate drain
intent key the API writes and the agent reads on heartbeat. State
transitions trigger NATS consumer subscribe/unsubscribe.

**Tech Stack:** Go 1.25, NATS JetStream KV, Echo REST API, OpenAPI
codegen, testify/suite

**Design Doc:** `docs/plans/2026-03-05-node-conditions-drain-design.md`

---

## Task 1: Add Condition type and evaluation functions

**Files:**
- Create: `internal/agent/condition.go`
- Create: `internal/agent/condition_test.go`

**Step 1: Write the failing tests**

```go
// internal/agent/condition_test.go
package agent

import (
    "testing"
    "time"

    "github.com/stretchr/testify/suite"

    "github.com/retr0h/osapi/internal/job"
    "github.com/retr0h/osapi/internal/provider/node/disk"
    "github.com/retr0h/osapi/internal/provider/node/load"
    "github.com/retr0h/osapi/internal/provider/node/mem"
)

type ConditionTestSuite struct {
    suite.Suite
}

func TestConditionTestSuite(t *testing.T) {
    suite.Run(t, new(ConditionTestSuite))
}

func (s *ConditionTestSuite) TestEvaluateMemoryPressure() {
    tests := []struct {
        name       string
        stats      *mem.Stats
        threshold  int
        wantStatus bool
        wantReason string
    }{
        {
            name:       "above threshold",
            stats:      &mem.Stats{Total: 16000000000, Used: 15000000000, Free: 1000000000},
            threshold:  90,
            wantStatus: true,
        },
        {
            name:       "below threshold",
            stats:      &mem.Stats{Total: 16000000000, Used: 8000000000, Free: 8000000000},
            threshold:  90,
            wantStatus: false,
        },
        {
            name:       "nil stats",
            stats:      nil,
            threshold:  90,
            wantStatus: false,
        },
    }

    for _, tt := range tests {
        s.Run(tt.name, func() {
            c := evaluateMemoryPressure(tt.stats, tt.threshold, nil)
            s.Equal(tt.wantStatus, c.Status)
            s.Equal(job.ConditionMemoryPressure, c.Type)
        })
    }
}

func (s *ConditionTestSuite) TestEvaluateHighLoad() {
    tests := []struct {
        name       string
        loadAvg    *load.AverageStats
        cpuCount   int
        multiplier float64
        wantStatus bool
    }{
        {
            name:       "above threshold",
            loadAvg:    &load.AverageStats{OneMin: 5.0},
            cpuCount:   2,
            multiplier: 2.0,
            wantStatus: true,
        },
        {
            name:       "below threshold",
            loadAvg:    &load.AverageStats{OneMin: 1.0},
            cpuCount:   2,
            multiplier: 2.0,
            wantStatus: false,
        },
        {
            name:       "nil load",
            loadAvg:    nil,
            cpuCount:   2,
            multiplier: 2.0,
            wantStatus: false,
        },
        {
            name:       "zero cpus",
            loadAvg:    &load.AverageStats{OneMin: 5.0},
            cpuCount:   0,
            multiplier: 2.0,
            wantStatus: false,
        },
    }

    for _, tt := range tests {
        s.Run(tt.name, func() {
            c := evaluateHighLoad(tt.loadAvg, tt.cpuCount, tt.multiplier, nil)
            s.Equal(tt.wantStatus, c.Status)
            s.Equal(job.ConditionHighLoad, c.Type)
        })
    }
}

func (s *ConditionTestSuite) TestEvaluateDiskPressure() {
    tests := []struct {
        name       string
        disks      []disk.UsageStats
        threshold  int
        wantStatus bool
    }{
        {
            name: "one disk above threshold",
            disks: []disk.UsageStats{
                {Name: "/dev/sda1", Total: 100000, Used: 95000, Free: 5000},
            },
            threshold:  90,
            wantStatus: true,
        },
        {
            name: "all disks below threshold",
            disks: []disk.UsageStats{
                {Name: "/dev/sda1", Total: 100000, Used: 50000, Free: 50000},
            },
            threshold:  90,
            wantStatus: false,
        },
        {
            name:       "nil disks",
            disks:      nil,
            threshold:  90,
            wantStatus: false,
        },
    }

    for _, tt := range tests {
        s.Run(tt.name, func() {
            c := evaluateDiskPressure(tt.disks, tt.threshold, nil)
            s.Equal(tt.wantStatus, c.Status)
            s.Equal(job.ConditionDiskPressure, c.Type)
        })
    }
}

func (s *ConditionTestSuite) TestLastTransitionTimeTracking() {
    prev := []job.Condition{{
        Type: job.ConditionMemoryPressure, Status: false,
        LastTransitionTime: time.Now().Add(-5 * time.Minute),
    }}
    // Flip from false -> true: should update LastTransitionTime
    c := evaluateMemoryPressure(
        &mem.Stats{Total: 100, Used: 95, Free: 5}, 90, prev,
    )
    s.True(c.Status)
    s.True(c.LastTransitionTime.After(time.Now().Add(-1 * time.Second)))

    // Same status (true -> true): should keep old LastTransitionTime
    prev2 := []job.Condition{c}
    c2 := evaluateMemoryPressure(
        &mem.Stats{Total: 100, Used: 95, Free: 5}, 90, prev2,
    )
    s.True(c2.Status)
    s.Equal(c.LastTransitionTime, c2.LastTransitionTime)
}
```

**Step 2: Run tests to verify they fail**

Run: `go test -run TestConditionTestSuite -v ./internal/agent/`
Expected: FAIL — `evaluateMemoryPressure` not defined

**Step 3: Write minimal implementation**

```go
// internal/agent/condition.go
package agent

import (
    "fmt"
    "time"

    "github.com/retr0h/osapi/internal/job"
    "github.com/retr0h/osapi/internal/provider/node/disk"
    "github.com/retr0h/osapi/internal/provider/node/load"
    "github.com/retr0h/osapi/internal/provider/node/mem"
)

// findPrevCondition returns the previous condition of the given type,
// or nil if not found.
func findPrevCondition(
    condType string,
    prev []job.Condition,
) *job.Condition {
    for i := range prev {
        if prev[i].Type == condType {
            return &prev[i]
        }
    }
    return nil
}

// transitionTime returns the previous LastTransitionTime if status
// hasn't changed, otherwise returns now.
func transitionTime(
    condType string,
    newStatus bool,
    prev []job.Condition,
) time.Time {
    if p := findPrevCondition(condType, prev); p != nil {
        if p.Status == newStatus {
            return p.LastTransitionTime
        }
    }
    return time.Now()
}

func evaluateMemoryPressure(
    stats *mem.Stats,
    threshold int,
    prev []job.Condition,
) job.Condition {
    c := job.Condition{Type: job.ConditionMemoryPressure}
    if stats == nil || stats.Total == 0 {
        c.LastTransitionTime = transitionTime(c.Type, false, prev)
        return c
    }
    pct := float64(stats.Used) / float64(stats.Total) * 100
    c.Status = pct > float64(threshold)
    if c.Status {
        c.Reason = fmt.Sprintf(
            "memory %.0f%% used (%.1f/%.1f GB)",
            pct,
            float64(stats.Used)/1024/1024/1024,
            float64(stats.Total)/1024/1024/1024,
        )
    }
    c.LastTransitionTime = transitionTime(c.Type, c.Status, prev)
    return c
}

func evaluateHighLoad(
    loadAvg *load.AverageStats,
    cpuCount int,
    multiplier float64,
    prev []job.Condition,
) job.Condition {
    c := job.Condition{Type: job.ConditionHighLoad}
    if loadAvg == nil || cpuCount == 0 {
        c.LastTransitionTime = transitionTime(c.Type, false, prev)
        return c
    }
    threshold := float64(cpuCount) * multiplier
    c.Status = loadAvg.OneMin > threshold
    if c.Status {
        c.Reason = fmt.Sprintf(
            "load %.2f, threshold %.2f for %d CPUs",
            loadAvg.OneMin, threshold, cpuCount,
        )
    }
    c.LastTransitionTime = transitionTime(c.Type, c.Status, prev)
    return c
}

func evaluateDiskPressure(
    disks []disk.UsageStats,
    threshold int,
    prev []job.Condition,
) job.Condition {
    c := job.Condition{Type: job.ConditionDiskPressure}
    if len(disks) == 0 {
        c.LastTransitionTime = transitionTime(c.Type, false, prev)
        return c
    }
    for _, d := range disks {
        if d.Total == 0 {
            continue
        }
        pct := float64(d.Used) / float64(d.Total) * 100
        if pct > float64(threshold) {
            c.Status = true
            c.Reason = fmt.Sprintf(
                "%s %.0f%% used (%.1f/%.1f GB)",
                d.Name, pct,
                float64(d.Used)/1024/1024/1024,
                float64(d.Total)/1024/1024/1024,
            )
            break
        }
    }
    c.LastTransitionTime = transitionTime(c.Type, c.Status, prev)
    return c
}
```

**Step 4: Run tests to verify they pass**

Run: `go test -run TestConditionTestSuite -v ./internal/agent/`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/agent/condition.go internal/agent/condition_test.go
git commit -m "feat(agent): add condition evaluation functions"
```

---

## Task 2: Add Condition and State types to job domain

**Files:**
- Modify: `internal/job/types.go:273-331` (AgentRegistration,
  AgentInfo)

**Step 1: Write the types**

Add to `internal/job/types.go` after existing types:

```go
// Condition type constants.
const (
    ConditionMemoryPressure = "MemoryPressure"
    ConditionHighLoad       = "HighLoad"
    ConditionDiskPressure   = "DiskPressure"
)

// Agent state constants.
const (
    AgentStateReady    = "Ready"
    AgentStateDraining = "Draining"
    AgentStateCordoned = "Cordoned"
)

// Condition represents a node condition evaluated agent-side.
type Condition struct {
    Type               string    `json:"type"`
    Status             bool      `json:"status"`
    Reason             string    `json:"reason,omitempty"`
    LastTransitionTime time.Time `json:"last_transition_time"`
}

```

The existing `TimelineEvent` type (line 177) is already generic and
will be reused for agent state transitions — no new event type needed.

Add fields to `AgentRegistration`:

```go
Conditions []Condition `json:"conditions,omitempty"`
State      string      `json:"state,omitempty"`
```

Add fields to `AgentInfo`:

```go
Conditions []Condition      `json:"conditions,omitempty"`
State      string           `json:"state,omitempty"`
Timeline   []TimelineEvent  `json:"timeline,omitempty"`
```

**Step 2: Run existing tests**

Run: `go test ./internal/job/... -count=1`
Expected: PASS (additive change)

**Step 3: Commit**

```bash
git add internal/job/types.go
git commit -m "feat(job): add Condition type and agent state constants"
```

---

## Task 3: Add conditions config to AgentConfig

**Files:**
- Modify: `internal/config/types.go:262-277`
- Modify: `configs/osapi.yaml`
- Modify: `configs/osapi.local.yaml`

**Step 1: Add config struct**

Add to `internal/config/types.go`:

```go
// AgentConditions holds threshold configuration for node conditions.
type AgentConditions struct {
    MemoryPressureThreshold int     `mapstructure:"memory_pressure_threshold"`
    HighLoadMultiplier      float64 `mapstructure:"high_load_multiplier"`
    DiskPressureThreshold   int     `mapstructure:"disk_pressure_threshold"`
}
```

Add field to `AgentConfig`:

```go
Conditions AgentConditions `mapstructure:"conditions,omitempty"`
```

**Step 2: Set defaults in osapi.yaml and osapi.local.yaml**

```yaml
agent:
  conditions:
    memory_pressure_threshold: 90
    high_load_multiplier: 2.0
    disk_pressure_threshold: 90
```

**Step 3: Verify compilation**

Run: `go build ./...`
Expected: compiles

**Step 4: Commit**

```bash
git add internal/config/types.go configs/osapi.yaml configs/osapi.local.yaml
git commit -m "feat(config): add agent conditions threshold configuration"
```

---

## Task 4: Add disk stats to heartbeat and evaluate conditions

**Files:**
- Modify: `internal/agent/heartbeat.go:88-134` (writeRegistration)
- Modify: `internal/agent/types.go:45-81` (add prevConditions, cpuCount)

**Step 1: Add fields to Agent struct**

In `internal/agent/types.go`, add to Agent struct:

```go
// prevConditions tracks condition state between heartbeats.
prevConditions []job.Condition

// cpuCount cached from facts for HighLoad evaluation.
cpuCount int
```

**Step 2: Extend writeRegistration**

In `internal/agent/heartbeat.go`, after memory stats collection
(~line 111), add:

```go
// Collect disk stats (non-fatal).
var diskStats []disk.UsageStats
if stats, err := a.diskProvider.GetLocalUsageStats(); err == nil {
    diskStats = stats
}

// Evaluate conditions.
conditions := []job.Condition{
    evaluateMemoryPressure(
        memStats,
        a.appConfig.Agent.Conditions.MemoryPressureThreshold,
        a.prevConditions,
    ),
    evaluateHighLoad(
        loadAvg,
        a.cpuCount,
        a.appConfig.Agent.Conditions.HighLoadMultiplier,
        a.prevConditions,
    ),
    evaluateDiskPressure(
        diskStats,
        a.appConfig.Agent.Conditions.DiskPressureThreshold,
        a.prevConditions,
    ),
}
a.prevConditions = conditions
```

Add `Conditions: conditions` to the `AgentRegistration` literal.

**Step 3: Set cpuCount from facts**

In `internal/agent/facts.go` (the `writeFacts` function), after
collecting `CPUCount`, add:

```go
a.cpuCount = cpuCount
```

**Step 4: Run tests**

Run: `go test ./internal/agent/... -count=1`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/agent/heartbeat.go internal/agent/types.go internal/agent/facts.go
git commit -m "feat(agent): evaluate node conditions on heartbeat tick"
```

---

## Task 5: Add drain timeline event storage functions

**Files:**
- Modify: `internal/job/client/agent.go:39-85`
- Create: `internal/job/client/agent_timeline_test.go`

**Step 1: Write failing tests**

```go
// internal/job/client/agent_timeline_test.go
package client_test

// Test WriteAgentTimelineEvent writes append-only key to registryKV.
// Test ComputeAgentState returns latest state from timeline events.
// Test GetAgentTimeline returns sorted timeline events.
```

Table-driven tests:
- `WriteAgentTimelineEvent` writes key like
  `timeline.{hostname}.{event}.{unix_nano}`
- `ComputeAgentState` with no events returns "Ready"
- `ComputeAgentState` with drain event returns "Draining"
- `ComputeAgentState` with cordoned event returns "Cordoned"
- `ComputeAgentState` with undrain event returns "Ready"

**Step 2: Run tests to verify they fail**

Run: `go test -run TestAgentTimeline -v ./internal/job/client/`
Expected: FAIL

**Step 3: Implement**

Add to `internal/job/client/agent.go`:

```go
// WriteAgentTimelineEvent writes an append-only timeline event
// for an agent state transition. Reuses the same TimelineEvent
// type used by job lifecycle events.
func (c *Client) WriteAgentTimelineEvent(
    _ context.Context,
    hostname, event, message string,
) error {
    now := time.Now()
    key := fmt.Sprintf(
        "timeline.%s.%s.%d",
        job.SanitizeHostname(hostname),
        event,
        now.UnixNano(),
    )
    data, _ := json.Marshal(job.TimelineEvent{
        Timestamp: now,
        Event:     event,
        Hostname:  hostname,
        Message:   message,
    })
    _, err := c.registryKV.Put(key, data)
    return err
}

// GetAgentTimeline returns sorted timeline events for a hostname.
func (c *Client) GetAgentTimeline(
    ctx context.Context,
    hostname string,
) ([]job.TimelineEvent, error) {
    prefix := "timeline." + job.SanitizeHostname(hostname) + "."
    // List keys with prefix, unmarshal, sort by Timestamp
    // Return sorted events
}

// ComputeAgentState returns the current state from timeline events.
func ComputeAgentState(
    events []job.TimelineEvent,
) string {
    if len(events) == 0 {
        return job.AgentStateReady
    }
    latest := events[len(events)-1]
    switch latest.Event {
    case "drain":
        return job.AgentStateDraining
    case "cordoned":
        return job.AgentStateCordoned
    case "undrain", "ready":
        return job.AgentStateReady
    default:
        return job.AgentStateReady
    }
}
```

Add `WriteAgentTimelineEvent`, `GetAgentTimeline` to the `JobClient`
interface in `internal/job/client/types.go`. Regenerate mocks.

**Step 4: Run tests**

Run: `go test -run TestAgentTimeline -v ./internal/job/client/`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/job/client/agent.go internal/job/client/agent_timeline_test.go \
    internal/job/client/types.go internal/job/client/mock_*.go
git commit -m "feat(job): add append-only timeline events for agent drain"
```

---

## Task 6: Add drain/undrain API endpoints

**Files:**
- Modify: `internal/api/agent/gen/api.yaml`
- Create: `internal/api/agent/agent_drain.go`
- Create: `internal/api/agent/agent_drain_public_test.go`

**Step 1: Add to OpenAPI spec**

Add to `internal/api/agent/gen/api.yaml`:

```yaml
/agent/{hostname}/drain:
  post:
    operationId: drainAgent
    summary: Drain an agent
    description: Stop the agent from accepting new jobs.
    security:
      - BearerAuth:
          - "agent:write"
    parameters:
      - name: hostname
        in: path
        required: true
        schema:
          type: string
    responses:
      "200":
        description: Agent drain initiated.
        content:
          application/json:
            schema:
              type: object
              properties:
                message:
                  type: string
      "404":
        description: Agent not found.
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/ErrorResponse"
      "409":
        description: Agent already in requested state.
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/ErrorResponse"

/agent/{hostname}/undrain:
  post:
    operationId: undrainAgent
    summary: Undrain an agent
    description: Resume accepting jobs on a drained agent.
    security:
      - BearerAuth:
          - "agent:write"
    parameters:
      - name: hostname
        in: path
        required: true
        schema:
          type: string
    responses:
      "200":
        description: Agent undrain initiated.
        content:
          application/json:
            schema:
              type: object
              properties:
                message:
                  type: string
      "404":
        description: Agent not found.
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/ErrorResponse"
      "409":
        description: Agent not in draining/cordoned state.
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/ErrorResponse"
```

Add `agent:write` to BearerAuth scopes. Add `state` and `conditions`
fields to `AgentInfo` schema. Add `NodeCondition` schema.

Run: `just generate` to regenerate `*.gen.go`.

**Step 2: Write failing tests**

```go
// internal/api/agent/agent_drain_public_test.go
// Table-driven tests for DrainAgent and UndrainAgent:
// - 200: agent found and drain initiated
// - 404: agent not found
// - 409: already draining/cordoned
// - HTTP wiring: RBAC (401, 403 without agent:write, 200 with agent:write)
```

**Step 3: Implement handlers**

```go
// internal/api/agent/agent_drain.go
package agent

func (a *Agent) DrainAgent(
    ctx context.Context,
    request gen.DrainAgentRequestObject,
) (gen.DrainAgentResponseObject, error) {
    hostname := request.Hostname

    // 1. Verify agent exists
    agentInfo, err := a.JobClient.GetAgent(ctx, hostname)
    if err != nil {
        return gen.DrainAgent404JSONResponse{...}, nil
    }

    // 2. Check not already draining
    if agentInfo.State == job.AgentStateDraining ||
        agentInfo.State == job.AgentStateCordoned {
        return gen.DrainAgent409JSONResponse{...}, nil
    }

    // 3. Write drain intent key
    // 4. Write state event
    return gen.DrainAgent200JSONResponse{...}, nil
}

func (a *Agent) UndrainAgent(
    ctx context.Context,
    request gen.UndrainAgentRequestObject,
) (gen.UndrainAgentResponseObject, error) {
    // Similar: verify exists, check state, delete drain key, write event
}
```

**Step 4: Run tests**

Run: `go test ./internal/api/agent/... -count=1`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/api/agent/gen/api.yaml internal/api/agent/gen/*.gen.go \
    internal/api/agent/agent_drain.go internal/api/agent/agent_drain_public_test.go
git commit -m "feat(api): add drain/undrain endpoints with RBAC"
```

---

## Task 7: Add agent:write permission

**Files:**
- Modify: `internal/authtoken/permissions.go:27-37` (add constant)
- Modify: `internal/authtoken/permissions.go:53-81` (add to admin role)

**Step 1: Add permission constant**

```go
PermAgentWrite Permission = "agent:write"
```

**Step 2: Add to admin role**

In `DefaultRolePermissions`, add `PermAgentWrite` to the `admin` slice.

**Step 3: Run tests**

Run: `go test ./internal/authtoken/... -count=1`
Expected: PASS

**Step 4: Commit**

```bash
git add internal/authtoken/permissions.go
git commit -m "feat(auth): add agent:write permission for drain operations"
```

---

## Task 8: Wire drain endpoints into server

**Files:**
- Modify: `internal/api/handler_agent.go:34-61`
- Modify: `internal/api/handler_agent_public_test.go`

**Step 1: Update handler registration**

The `GetAgentHandler` already wires all agent gen handlers through
`scopeMiddleware`. After regenerating the OpenAPI code (Task 6), the
new `DrainAgent` and `UndrainAgent` methods on the strict server
interface will be picked up automatically by `RegisterHandlers`.

No code change needed in `handler_agent.go` unless
`unauthenticatedOperations` needs updating (it doesn't — drain
requires auth).

**Step 2: Verify compilation**

Run: `go build ./...`
Expected: compiles

**Step 3: Add handler test cases**

Add test cases to `handler_agent_public_test.go` for drain/undrain
handler registration.

**Step 4: Commit**

```bash
git add internal/api/handler_agent.go internal/api/handler_agent_public_test.go
git commit -m "feat(api): wire drain/undrain handlers into server"
```

---

## Task 9: Add drain detection to agent heartbeat

**Files:**
- Modify: `internal/agent/heartbeat.go:88-134`
- Modify: `internal/agent/server.go:32-61`
- Create: `internal/agent/drain.go`
- Create: `internal/agent/drain_test.go`

**Step 1: Write failing tests**

```go
// internal/agent/drain_test.go
// Test checkDrainFlag: returns true when drain key exists
// Test checkDrainFlag: returns false when drain key absent
// Test handleDrainTransition: unsubscribes consumers when draining
// Test handleUndrainTransition: resubscribes consumers when undrained
```

**Step 2: Implement drain detection**

```go
// internal/agent/drain.go
package agent

// checkDrainFlag reads drain.{hostname} from registryKV.
func (a *Agent) checkDrainFlag(
    ctx context.Context,
    hostname string,
) bool {
    key := "drain." + job.SanitizeHostname(hostname)
    _, err := a.registryKV.Get(ctx, key)
    return err == nil
}

// handleDrainDetection checks drain flag on each heartbeat.
func (a *Agent) handleDrainDetection(
    ctx context.Context,
    hostname string,
) {
    drainRequested := a.checkDrainFlag(ctx, hostname)

    switch {
    case drainRequested && a.state == job.AgentStateReady:
        a.state = job.AgentStateDraining
        a.unsubscribeConsumers()
        // Write timeline event: "drain", "Drain initiated"
        // When WaitGroup drains, transition to Cordoned

    case !drainRequested && a.state == job.AgentStateCordoned:
        a.state = job.AgentStateReady
        a.resubscribeConsumers(ctx, hostname)
        // Write timeline event: "undrain", "Resumed accepting jobs"
    }
}
```

**Step 3: Add state field to Agent struct**

In `internal/agent/types.go`:

```go
state string // Ready, Draining, Cordoned
```

Initialize to `job.AgentStateReady` in `Start()`.

**Step 4: Call from heartbeat**

In `writeRegistration()`, add `a.handleDrainDetection(ctx, hostname)`
and include `State: a.state` in the registration.

**Step 5: Run tests**

Run: `go test ./internal/agent/... -count=1`
Expected: PASS

**Step 6: Commit**

```bash
git add internal/agent/drain.go internal/agent/drain_test.go \
    internal/agent/heartbeat.go internal/agent/types.go internal/agent/server.go
git commit -m "feat(agent): detect drain flag and manage consumer lifecycle"
```

---

## Task 10: Extend buildAgentInfo with conditions and state

**Files:**
- Modify: `internal/api/agent/agent_list.go:59-171` (buildAgentInfo)
- Modify: `internal/api/agent/agent_list_public_test.go`
- Modify: `internal/job/client/query.go:479-493`
  (agentInfoFromRegistration)

**Step 1: Update agentInfoFromRegistration**

Add to the returned `AgentInfo`:

```go
Conditions: reg.Conditions,
State:      reg.State,
```

**Step 2: Update buildAgentInfo**

Map conditions and state from `job.AgentInfo` to `gen.AgentInfo`:

```go
if len(a.Conditions) > 0 {
    conditions := make([]gen.NodeCondition, 0, len(a.Conditions))
    for _, c := range a.Conditions {
        nc := gen.NodeCondition{
            Type:               gen.NodeConditionType(c.Type),
            Status:             c.Status,
            LastTransitionTime: c.LastTransitionTime,
        }
        if c.Reason != "" {
            nc.Reason = &c.Reason
        }
        conditions = append(conditions, nc)
    }
    info.Conditions = &conditions
}

if a.State != "" {
    state := gen.AgentInfoState(a.State)
    info.State = &state
}
```

**Step 3: Update status derivation**

Change status logic: if `a.State` is set, use it; otherwise default
to `Ready` (existing behavior).

**Step 4: Add test cases**

Add table-driven test case for agent with conditions and
Draining/Cordoned states.

**Step 5: Run tests**

Run: `go test ./internal/api/agent/... -count=1`
Expected: PASS

**Step 6: Commit**

```bash
git add internal/api/agent/agent_list.go internal/api/agent/agent_list_public_test.go \
    internal/job/client/query.go
git commit -m "feat(api): expose conditions and state in agent responses"
```

---

## Task 11: Add timeline to GetAgent response

**Files:**
- Modify: `internal/job/client/query.go:423-445` (GetAgent)
- Modify: `internal/job/client/query_public_test.go`

**Step 1: Extend GetAgent to fetch timeline events**

After building `AgentInfo`, fetch timeline events:

```go
timeline, err := c.GetAgentTimeline(ctx, hostname)
if err == nil {
    info.Timeline = timeline
}
```

**Step 2: Add test cases**

Test GetAgent returns timeline events when present.

**Step 3: Run tests**

Run: `go test ./internal/job/client/... -count=1`
Expected: PASS

**Step 4: Commit**

```bash
git add internal/job/client/query.go internal/job/client/query_public_test.go
git commit -m "feat(job): include timeline events in GetAgent response"
```

---

## Task 12: Update SDK with conditions, state, drain/undrain

**Files:**
- Modify: `osapi-sdk/pkg/osapi/gen/agent/api.yaml` (copy from osapi)
- Modify: `osapi-sdk/pkg/osapi/agent.go` (add Drain, Undrain methods)
- Modify: `osapi-sdk/pkg/osapi/agent_types.go` (add conditions, state,
  timeline to Agent type)
- Create: `osapi-sdk/pkg/osapi/types.go` (promote TimelineEvent to
  shared type)
- Modify: `osapi-sdk/pkg/osapi/job_types.go` (remove TimelineEvent,
  import from types.go)

**Step 1: Promote TimelineEvent to shared type**

Move `TimelineEvent` from `job_types.go` to a new `types.go`:

```go
// pkg/osapi/types.go

// TimelineEvent represents a lifecycle event. Used by both job
// timelines and agent state transition history.
type TimelineEvent struct {
    Timestamp string
    Event     string
    Hostname  string
    Message   string
    Error     string
}
```

Update `job_types.go` to remove the `TimelineEvent` definition —
`JobDetail.Timeline` now references the shared type.

**Step 2: Sync OpenAPI spec**

Copy `internal/api/agent/gen/api.yaml` to
`osapi-sdk/pkg/osapi/gen/agent/api.yaml`.

Run `redocly join` + `go generate` in the SDK.

**Step 3: Add domain types**

```go
// In agent_types.go
type Agent struct {
    // ... existing fields ...
    State      string
    Conditions []Condition
    Timeline   []TimelineEvent  // shared type from types.go
}

type Condition struct {
    Type               string
    Status             bool
    Reason             string
    LastTransitionTime time.Time
}
```

**Step 4: Add Drain/Undrain methods**

```go
func (s *AgentService) Drain(
    ctx context.Context,
    hostname string,
) (*Response[any], error) {
    // POST /agent/{hostname}/drain
}

func (s *AgentService) Undrain(
    ctx context.Context,
    hostname string,
) (*Response[any], error) {
    // POST /agent/{hostname}/undrain
}
```

**Step 4: Run SDK tests**

Run: `go test ./pkg/osapi/... -count=1`
Expected: PASS

**Step 5: Commit (in osapi-sdk repo)**

```bash
git add pkg/osapi/
git commit -m "feat(agent): add conditions, state, drain/undrain support"
```

---

## Task 13: Add CONDITIONS column to agent list CLI

**Files:**
- Modify: `cmd/client_agent_list.go`

**Step 1: Add CONDITIONS column**

In the table builder for `agent list`, add a column that joins active
condition type names:

```go
conditions := "-"
if len(agent.Conditions) > 0 {
    active := make([]string, 0)
    for _, c := range agent.Conditions {
        if c.Status {
            active = append(active, c.Type)
        }
    }
    if len(active) > 0 {
        conditions = strings.Join(active, ",")
    }
}
```

Headers: `HOSTNAME`, `STATUS`, `CONDITIONS`, `LABELS`, `AGE`, `LOAD`,
`OS`

**Step 2: Use State for STATUS column**

Replace hardcoded "Ready" with `agent.State` (defaulting to "Ready"
if empty).

**Step 3: Run `go build ./cmd/...`**

Expected: compiles

**Step 4: Commit**

```bash
git add cmd/client_agent_list.go
git commit -m "feat(cli): add CONDITIONS column and state to agent list"
```

---

## Task 14: Add conditions and timeline to agent get CLI

**Files:**
- Modify: `cmd/client_agent_get.go:58-141`

**Step 1: Add state to agent get output**

After the Status KV line, display the State:

```go
if data.State != "" && data.State != "Ready" {
    cli.PrintKV("State", data.State)
}
```

**Step 2: Add conditions section**

```go
if len(data.Conditions) > 0 {
    condRows := make([][]string, 0, len(data.Conditions))
    for _, c := range data.Conditions {
        status := "false"
        if c.Status {
            status = "true"
        }
        reason := ""
        if c.Reason != "" {
            reason = c.Reason
        }
        since := cli.FormatAge(time.Since(c.LastTransitionTime)) + " ago"
        condRows = append(condRows, []string{c.Type, status, reason, since})
    }
    sections = append(sections, cli.Section{
        Title:   "Conditions",
        Headers: []string{"TYPE", "STATUS", "REASON", "SINCE"},
        Rows:    condRows,
    })
}
```

**Step 3: Add timeline section**

Same pattern as `DisplayJobDetail` in `internal/cli/ui.go:600-615`:

```go
if len(data.Timeline) > 0 {
    timelineRows := make([][]string, 0, len(data.Timeline))
    for _, te := range data.Timeline {
        timelineRows = append(timelineRows, []string{
            te.Timestamp, te.Event, te.Hostname, te.Message, te.Error,
        })
    }
    sections = append(sections, cli.Section{
        Title:   "Timeline",
        Headers: []string{"TIMESTAMP", "EVENT", "HOSTNAME", "MESSAGE", "ERROR"},
        Rows:    timelineRows,
    })
}
```

**Step 4: Run `go build ./cmd/...`**

Expected: compiles

**Step 5: Commit**

```bash
git add cmd/client_agent_get.go
git commit -m "feat(cli): display conditions and timeline in agent get"
```

---

## Task 15: Add agent drain/undrain CLI commands

**Files:**
- Create: `cmd/client_agent_drain.go`
- Create: `cmd/client_agent_undrain.go`

**Step 1: Create drain command**

```go
// cmd/client_agent_drain.go
var clientAgentDrainCmd = &cobra.Command{
    Use:   "drain",
    Short: "Drain an agent",
    Long:  `Stop an agent from accepting new jobs. In-flight jobs complete.`,
    Run: func(cmd *cobra.Command, _ []string) {
        ctx := cmd.Context()
        hostname, _ := cmd.Flags().GetString("hostname")

        resp, err := sdkClient.Agent.Drain(ctx, hostname)
        if err != nil {
            cli.HandleError(err, logger)
            return
        }

        if jsonOutput {
            fmt.Println(string(resp.RawJSON()))
            return
        }

        fmt.Printf("Agent %s drain initiated\n", hostname)
    },
}
```

**Step 2: Create undrain command**

Similar pattern for `undrain`.

**Step 3: Register commands**

```go
func init() {
    clientAgentCmd.AddCommand(clientAgentDrainCmd)
    clientAgentDrainCmd.Flags().String("hostname", "", "Hostname of the agent to drain")
    _ = clientAgentDrainCmd.MarkFlagRequired("hostname")
}
```

**Step 4: Run `go build ./cmd/...`**

Expected: compiles

**Step 5: Commit**

```bash
git add cmd/client_agent_drain.go cmd/client_agent_undrain.go
git commit -m "feat(cli): add agent drain and undrain commands"
```

---

## Task 16: Update documentation

**Files:**
- Modify: `docs/docs/sidebar/features/agent-management.md` (or create)
- Modify: `docs/docs/sidebar/usage/configuration.md`
- Modify: `docs/docs/sidebar/usage/cli/client/agent/`

**Step 1: Add conditions and drain docs**

Document:
- Condition types and thresholds
- Drain lifecycle (Ready → Draining → Cordoned)
- CLI commands (`agent drain`, `agent undrain`)
- Configuration section for `agent.conditions`

**Step 2: Update permission table**

Add `agent:write` to the permissions table in configuration.md.

**Step 3: Commit**

```bash
git add docs/
git commit -m "docs: add node conditions and agent drain documentation"
```

---

## Task 17: Final verification

**Step 1: Regenerate**

Run: `just generate`
Expected: no diff

**Step 2: Build**

Run: `go build ./...`
Expected: compiles

**Step 3: Unit tests**

Run: `just go::unit`
Expected: PASS

**Step 4: Lint**

Run: `just go::vet`
Expected: clean

**Step 5: Coverage check**

Run: `go test -coverprofile=coverage.out ./internal/agent/... ./internal/job/client/... ./internal/api/agent/...`
Expected: condition.go, drain.go, agent_drain.go at 100%

---

## Verification

```bash
just generate        # regenerate specs + code
go build ./...       # compiles
just go::unit        # tests pass
just go::vet         # lint passes
```
