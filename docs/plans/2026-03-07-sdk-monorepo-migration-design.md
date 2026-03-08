# SDK Monorepo Migration

**Date:** 2026-03-07 **Status:** Design **Author:** @retr0h

## Problem

The SDK living in a separate repo (`osapi-io/osapi-sdk`) creates friction:

- OpenAPI specs must be synced via gilt overlay from osapi's `main` branch
- Every API change requires a two-repo dance: merge osapi, run `just generate`
  in the SDK, merge SDK, update `go.mod` in osapi
- Per-example directories each with their own `go.mod` are a maintenance burden
- The SDK has no external consumers — it's only used by the osapi CLI

## Solution

Move the SDK into the osapi repo as `pkg/sdk/`. Two incremental PRs.

## Package Layout

```
pkg/sdk/
├── osapi/                    ← PR 1
│   ├── gen/
│   │   ├── cfg.yaml          ← points to ../../../internal/api/gen/api.yaml
│   │   ├── generate.go       ← just oapi-codegen, no gilt
│   │   └── client.gen.go
│   ├── osapi.go
│   ├── transport.go
│   ├── errors.go
│   ├── response.go
│   ├── types.go
│   ├── agent.go
│   ├── agent_types.go
│   ├── audit.go
│   ├── audit_types.go
│   ├── file.go
│   ├── file_types.go
│   ├── health.go
│   ├── health_types.go
│   ├── job.go
│   ├── job_types.go
│   ├── metrics.go
│   ├── node.go
│   ├── node_types.go
│   └── *_test.go
└── orchestrator/             ← PR 2
    ├── plan.go
    ├── task.go
    ├── options.go
    ├── result.go
    ├── runner.go
    └── *_test.go
```

Import paths change to:

- `github.com/retr0h/osapi/pkg/sdk/osapi`
- `github.com/retr0h/osapi/pkg/sdk/orchestrator`

## Spec Generation

No more gilt. The `cfg.yaml` in `pkg/sdk/osapi/gen/` references the server's
combined spec directly:

```yaml
# cfg.yaml
input: ../../../internal/api/gen/api.yaml
```

Single source of truth. Specs can never drift. Regenerate with
`go generate ./pkg/sdk/...`.

## Examples

Flatten from per-directory modules to individual files in two directories:

```
examples/sdk/
├── osapi/
│   ├── go.mod              ← replace ../../../pkg/sdk
│   ├── go.sum
│   ├── health.go           ← go run health.go
│   ├── node.go
│   ├── agent.go
│   ├── audit.go
│   ├── command.go
│   ├── file.go
│   ├── job.go
│   ├── metrics.go
│   └── network.go
└── orchestrator/
    ├── go.mod              ← replace ../../../pkg/sdk
    ├── go.sum
    ├── basic.go
    ├── parallel.go
    ├── guards.go
    ├── hooks.go
    ├── retry.go
    ├── broadcast.go
    ├── error_strategy.go
    ├── file_deploy.go
    ├── only_if_changed.go
    ├── only_if_failed.go
    ├── result_decode.go
    ├── task_func.go
    └── task_func_results.go
```

All files are `package main`. Run with `go run health.go`.

## Documentation

### Docusaurus SDK Sidebar

New top-level sidebar section:

```
docs/docs/sidebar/sdk/
├── sdk.md                  ← Overview with DocCardList
├── client/
│   ├── client.md           ← Client overview, New(), options, transport
│   ├── agent.md
│   ├── audit.md
│   ├── file.md
│   ├── health.md
│   ├── job.md
│   ├── metrics.md
│   └── node.md
└── orchestrator/
    ├── orchestrator.md     ← Overview, Plan/Task/Run
    ├── operations.md       ← Built-in operations reference
    ├── hooks.md            ← Hooks and error strategies
    └── examples.md         ← Example walkthroughs
```

Content migrated from the osapi-sdk `docs/osapi/` and `docs/orchestration/`
directories. Landing page uses `<DocCardList />` cards.

### README and CLAUDE.md Updates

- **README.md**: Add SDK link in the docs/features section. Remove sibling repo
  references.
- **CLAUDE.md**: Update SDK references to reflect `pkg/sdk/` location. Simplify
  "Adding a New API Domain" Step 5 — no gilt, just `go generate ./pkg/sdk/...`.
  Remove sibling repo references but keep SDK documentation (now pointing to
  in-repo paths).
- **docusaurus.config.ts**: Add "SDK" to the navbar Features dropdown.

## Cleanup

### PR 1 (SDK client)

- Copy `osapi-sdk/pkg/osapi/` → `pkg/sdk/osapi/`
- Update `pkg/sdk/osapi/gen/cfg.yaml` to reference `internal/api/gen/api.yaml`
- Remove `generate.go` gilt step (oapi-codegen only)
- Flatten `osapi-sdk/examples/osapi/` → `examples/sdk/osapi/`
- Update all `cmd/*.go` imports: `github.com/osapi-io/osapi-sdk/pkg/osapi` →
  `github.com/retr0h/osapi/pkg/sdk/osapi`
- Remove `github.com/osapi-io/osapi-sdk` from `go.mod`
- Create Docusaurus client pages
- Update README.md, CLAUDE.md

### PR 2 (Orchestrator)

- Copy `osapi-sdk/pkg/orchestrator/` → `pkg/sdk/orchestrator/`
- Update orchestrator imports to use new SDK client path
- Flatten `osapi-sdk/examples/orchestration/` → `examples/sdk/orchestrator/`
- Create Docusaurus orchestrator pages

### Post-merge

- User archives `osapi-io/osapi-sdk` repo on GitHub

## Scalability Note: `kv.Keys()`

Not related to this migration but documented here for context — the SDK's
`QueueStats()` and `List()` methods rely on the server's `kv.Keys()` call. See
the [Job Architecture](../docs/sidebar/architecture/job-architecture.md)
performance section for the known scalability constraint and mitigation
approaches.
