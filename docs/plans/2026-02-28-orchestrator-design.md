# osapi-orchestrator Design

## Goal

A Go library that gives operators orchestration primitives on top of the
osapi-sdk. Operators write Go programs that define tasks with dependencies,
and the library handles DAG resolution, parallel execution, idempotency
reporting, and error handling.

## Motivation

OSAPI provides primitives: submit a job, target a host, get a result. But
there's no way to express "run A then B" or "only run C if A changed
something" or "run these three things in parallel, then converge." Today the
only sequencing option is polling `job get` in a loop (what `job run` does
for a single job).

Operators managing fleets need orchestration: install a package, then
configure DNS, then start a service — across multiple hosts, with
dependencies between steps, and accurate reporting of what changed.

## Approach

**Client-side orchestration library.** The orchestrator runs on the
operator's machine (or a control node), calls the SDK, and tracks execution.
No new server components. The API server stays stateless. Agents stay dumb.

This follows the Ansible model (push from client) rather than the
Chef/Puppet model (pull from agent) or the Kubernetes model (server-side
controllers). The escape hatch is clean: if server-side durability is needed
later, a `POST /orchestrate` endpoint can accept the same task definitions
and run them internally.

**Provider-level idempotency.** The orchestration library doesn't implement
idempotency — it trusts the platform. Each write provider checks current
state before mutating and returns an accurate `changed` field. The DNS
provider already does this. All future write providers follow the same
pattern. The orchestrator consumes `changed` for conditionals and reporting.

## Core Concepts

Four primitives:

| Concept    | What it is                                                        |
| ---------- | ----------------------------------------------------------------- |
| **Task**   | A unit of work — wraps an SDK call or custom function             |
| **Plan**   | A DAG of tasks with dependency edges                              |
| **Runner** | Resolves the DAG, executes in topological order, parallelizes     |
| **Report** | Per-task status and aggregate convergence summary                 |

## Task Definition

Two styles, both returning the same `*Result`:

### Declarative — standard SDK operations

```go
installPkg := plan.Task("install-nginx", &orchestrator.Op{
    Operation: "command.exec",
    Target:    "_all",
    Params: map[string]any{
        "command": "apt",
        "args":    []string{"install", "-y", "nginx"},
    },
})
```

### Functional — custom logic

```go
verify := plan.TaskFunc("verify-nginx", func(
    ctx context.Context,
    client *osapi.Client,
) (*orchestrator.Result, error) {
    resp, err := client.Command.Exec(ctx, "nginx", []string{"-t"}, "_all")
    if err != nil {
        return nil, err
    }
    return &orchestrator.Result{Changed: false}, nil
})
```

## Dependencies and Execution Order

```go
createUser := plan.Task("create-user", &orchestrator.Op{...})
installNginx := plan.Task("install-nginx", &orchestrator.Op{...})
configureDNS := plan.Task("configure-dns", &orchestrator.Op{...})
startNginx := plan.Task("start-nginx", &orchestrator.Op{...})

installNginx.DependsOn(createUser)
startNginx.DependsOn(installNginx, configureDNS)
```

DAG:

```
createUser ──→ installNginx ──→ startNginx
configureDNS ─────────────────↗
```

The runner executes `createUser` and `configureDNS` in parallel, then
`installNginx` after `createUser` completes, then `startNginx` after both
`installNginx` and `configureDNS` complete.

## Conditional Execution

```go
// Only run if dependency actually changed something
startNginx.DependsOn(installNginx).OnlyIfChanged()

// Custom guard
startNginx.When(func(results orchestrator.Results) bool {
    return results.Get("install-nginx").Changed
})
```

## Error Handling

```go
plan := orchestrator.NewPlan(
    orchestrator.OnError(orchestrator.StopAll), // default
)

// Per-task override:
installNginx.OnError(orchestrator.Continue)
```

Three strategies:

- `StopAll` — fail fast, cancel everything (default)
- `Continue` — skip dependents, keep running independent tasks
- `Retry(n)` — retry n times before failing

## Running and Reporting

```go
report, err := plan.Run(ctx)

fmt.Println(report.Summary())
// 4 tasks: 2 changed, 1 unchanged, 1 skipped
// Total duration: 12.3s

for _, r := range report.Tasks {
    fmt.Printf("%s: %s (changed=%v, duration=%s)\n",
        r.Name, r.Status, r.Changed, r.Duration)
}
```

## How Idempotency Fits

The orchestration library delegates to the SDK, which delegates to the API,
which delegates to agents, which run providers. Providers own idempotency:

```
Plan.Run()
  → Task calls SDK
    → SDK calls API
      → API creates job → Agent runs provider
        → Provider checks state, mutates only if needed
        → Returns changed: true/false
      → Result flows back
  → Task gets Result{Changed: bool}
  → Runner uses Changed for conditionals and reporting
```

The DNS provider is the reference implementation: read current state, compare
to desired, skip if equal, mutate if different, report accurately. All
future write providers follow this pattern.

For command exec/shell: always `changed: true`. These are inherently
non-idempotent. The orchestration layer handles command idempotency via
guards (`When`, `OnlyIfChanged`) — not the provider.

## Provider Idempotency Standard

Every write provider MUST:

1. Read current state before mutating
2. Compare desired vs current — return `{Changed: false}` if equal
3. Mutate only if different — return `{Changed: true}`
4. Preserve unspecified fields (partial updates keep existing values)

Integration tests for write providers MUST verify:

- First call returns `changed: true`
- Second identical call returns `changed: false`

## Project Structure

`osapi-orchestrator` is a separate repo under the `osapi-io` org. It mirrors
the `osapi-sdk` project scaffolding:

```
osapi-orchestrator/
├── .github/              # Same workflows as osapi-sdk
│   ├── workflows/
│   │   ├── commitlint.yml
│   │   ├── depreview.yml
│   │   ├── go.yml
│   │   ├── greetings.yml
│   │   ├── labeler.yml
│   │   ├── release.yml
│   │   ├── reportcard.yml
│   │   └── stale.yml
│   ├── codecov.yml
│   ├── dependabot.yml
│   └── labeler.yml
├── .golangci.yml
├── .goreleaser.yaml
├── .mise.toml
├── .coverignore
├── .gitignore
├── justfile              # Shared justfiles via osapi-io-justfiles
├── AI_POLICY.md
├── CLAUDE.md
├── LICENSE
├── README.md
├── docs/
│   ├── contributing.md
│   └── development.md
├── examples/
│   ├── simple/           # Basic sequential tasks
│   └── webserver/        # Multi-host nginx deployment
├── orchestrator/
│   ├── plan.go           # Plan, NewPlan
│   ├── task.go           # Task, TaskFunc, DependsOn, When
│   ├── runner.go         # DAG resolution, parallel execution
│   ├── result.go         # Result, Report, Summary
│   └── options.go        # OnError, Retry, OnlyIfChanged
├── go.mod                # depends on osapi-sdk
└── go.sum
```

## What Comes Later (Not This Design)

- **YAML DSL** — parse YAML into Plan/Task structs at runtime
- **Dry-run mode** — `plan.DryRun(ctx)` shows what would execute
- **Checkpointing** — save progress to disk, resume after crash
- **Event triggers** — run a plan when agent comes online
- **Server-side execution** — `POST /orchestrate` for durable runs
