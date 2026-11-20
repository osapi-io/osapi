---
name: add-a-domain
description: Add or extend an osapi API domain, a provider plus every layer it must appear in. Covers the provider implementation and its platform variants, agent processor and registry wiring, the OpenAPI spec and validation tags, the Echo handler with broadcast targeting, handler registration and startup wiring, the SDK service, CLI commands, and the docs and permission tables a new domain must be added to. Use when asked to add a domain, add a provider, add an operation or endpoint to an existing domain, wire a provider into the agent, add an SDK service, add CLI commands for a domain, or when asked what a new domain has to touch, why a domain feels half-finished, or which layer is missing. Also use when reviewing a domain for cross-layer consistency against an existing one.
compatibility: Requires an osapi checkout with mise and just available. Commands run through `mise exec -- just`.
license: MIT
metadata:
  author: osapi-io
  source: https://github.com/osapi-io/osapi
---

# Add a domain

Domains move to [behaviors](https://github.com/osapi-io/specs/blob/main/components/osapi/behaviors.md). A behavior
is one typed definition under `internal/behaviors/<name>/`. osapi generates the
routes, agent dispatch, CLI, spec and docs from it, under `/api` and
`osapi client`, next to the hand-written domains that have not moved yet ([the
plan](https://github.com/osapi-io/specs/blob/main/components/osapi/behavior-migration.md)).

Before you start, ask which kind of work it is. A new behavior uses the
`build-a-behavior` skill (`go run . skill get --name build-a-behavior`), which
starts from `osapi behavior create`. It
does not use the steps below. This skill covers the hand-written domains: each
domain that did not move yet.

The operations of the controller itself (agent, audit, facts, file, health, job
and telemetry) are not domains of a host. They move to the endpoint DSL
([the design](https://github.com/osapi-io/specs/blob/main/components/osapi/endpoints.md)). The `build-an-endpoint`
skill (`go run . skill get --name build-an-endpoint`) covers them.
A new operation of the controller does not use the steps below.

## 1. Establish the reference domain

Do this on each run, before you write anything. Select a domain that somebody
completed recently, with the same shape as the new domain, and read it:

- `sysctl` for a direct-write domain
- `cron` for a meta provider that deploys files
- `docker` for a domain that connects to an external API
- `ntp` for the simplest CRUD shape

```bash
grep -rl 'sysctl\|Sysctl' --include='*.go' --include='*.yaml' --include='*.md' . \
  | grep -vE '/gen/|/node_modules/|docs/docs/gen|\.worktrees' | sort
```

That list is the specification for the new domain. It ran to 81 files in
September 2026. The codebase is the reference, not this skill: where they
disagree, read the code and fix the skill
([why](https://github.com/osapi-io/specs/blob/main/CONSTITUTION.md)).

Never copy a domain's files wholesale. Read one, then write the new one.

## 2. Route

| The task | Read |
| --- | --- |
| The operations themselves, platform variants, idempotency | [provider.md](references/provider.md) |
| Getting the provider called by a job | [agent.md](references/agent.md) |
| Endpoints, validation, handler, broadcast, registration | [api.md](references/api.md) |
| The Go SDK service consumers call | [sdk.md](references/sdk.md) |
| `osapi client ...` commands | [cli.md](references/cli.md) |
| Feature page, generated CLI pages, permission tables, navbars | [docs.md](references/docs.md) |
| A whole new domain, end to end | All six, in that order |

## 3. The order that avoids rework

The layers depend on each other in one direction. If you go out of order, you
must generate or write again.

1. **Provider**: the operations, with its own tests that pass.
2. **Agent**, processor and registry, so a job reaches the provider.
3. **OpenAPI spec**, then `just generate`, which produces the server, the
   combined spec, and the SDK's generated client together.
4. **Handler**, with validation and broadcast, then registration and startup.
5. **SDK service**, which wraps the generated client.
6. **CLI**, which wraps the SDK.
7. **Docs and tables**, feature page, `just cli-docs-generate`, permissions, navbars.

A new permission is decided at step 3 and lands in step 7. Write it down when
you choose it; it is the thing most often missed.

## 4. Cross-layer checklist

Run this before you say that a domain is complete. Each line is a location where
a domain occurs. A domain that is missing from one of them is only half
complete.

```
provider    internal/provider/{category}/{domain}/    types.go, platform files, mocks/
agent       internal/agent/processor_{domain}.go      dispatch to the provider
            internal/agent/providers.go               a hostEnv method
            internal/agent/setup.go                   Register in buildRegistry
api         internal/controller/api/node/{domain}/gen/  api.yaml, cfg.yaml, generate.go
            internal/controller/api/node/{domain}/    types.go, {domain}.go, handler.go,
                                                      one file per endpoint, tests
            internal/controller/handlers.go           one Handler(...) line
sdk         pkg/sdk/client/{domain}.go                service methods
            pkg/sdk/client/{domain}_types.go          result types, gen -> SDK conversion
            pkg/sdk/client/osapi.go                   Client field, wired in New()
            examples/sdk/client/{domain}.go           one runnable example
cli         cmd/client_node_{domain}.go               parent command
            cmd/client_node_{domain}_{op}.go          one per endpoint
docs        docs/docs/sidebar/features/{domain}-management.md
            (CLI pages)                               just cli-docs-generate writes them
            docs/docusaurus.config.ts                 features and SDK dropdowns
            features/features.md, features/authentication.md,
            usage/configuration.md                   permission tables
```

Prove it. Do not only read it again. Run the `grep -rl` from step 1 against the
new domain. It must return a list with the same shape as the list of the
reference domain.

## 5. Verify

```bash
mise exec -- just react-build            # populate ui/dist first, see below
mise exec -- just generate               # specs, server, SDK client, docs
mise exec -- just ready                  # generate, format, lint, build
mise exec -- just test                   # format checks, lint, unit, coverage gate
```

The coverage gate is 100%. Thus a branch without tests fails the build, and
does not merge. `just test` also runs `go-mod-check`, which fails when a nested
module under `examples/` is untidy: a new SDK example adds one. And it runs
`examples-check`, which vets each SDK example against this SDK: a change to an
SDK type or method fails there until each example uses the new shape.

Never run `go build ./...` or `go test ./...` directly. The build embeds the UI
with `//go:embed dist/*`. Thus both fail unless `ui/dist/` has files. The `just`
recipes put them there first.

In a new clone or worktree, run `just react-build` before `just ready`. `ready`
lints before it builds the UI. Thus the linter cannot import the embed package,
and reports a `typecheck` error on a file that nobody touched.

**`just test` checks the format of the site and of the UI. `just ready` does
not.** `md-fmt` formats only markdown, and excludes `docs/`. Prettier owns
`docs/` and `ui/`. `just test` runs `docs-fmt-check` and `react-fmt-check` for
them. Thus run `just test` before you push a change that touches `docs/` or
`ui/`. A change to a domain always touches `docs/`, as
[docs.md](references/docs.md) says.

In September 2026, three merged security fixes left five pages unformatted,
because the gate did not check the site then. main was red until
osapi-io/osapi#512. In November 2026, three UI files merged unformatted in the
same way, until osapi-io/osapi#721 added `react-fmt-check` to `just test`.

## Rules

1. **The reference domain answers style questions.** For names, file splits,
   table output and error words, read the sibling domain. Do not invent. A
   different word means a different rule.
2. **Idempotency is necessary.** Create on an existing resource is
   `Changed: false`. Delete on a missing resource is `Changed: false`. Update on
   a missing resource is an error. The table in
   [provider.md](references/provider.md) is the contract. It makes these
   operations safe to run many times.
3. **Each node-targeted operation supports broadcast.** It accepts `_all`,
   `_any`, a hostname and label selectors. It returns the same collection shape
   when one host answers and when forty answer.
4. **The OpenAPI spec declares validation.** `validation.Struct()` in the
   handler enforces it. The provider checks it again where a value becomes a
   path or a command argument.
5. **Never edit generated code by hand.** Change the spec and generate again. An
   edit by hand stays until the next `just generate`, and then it is gone.
6. **Stop at the first difference from the code.** This skill comes from the
   codebase of September 2026. When a step does not match now, correct the skill
   in its own change before you continue. Then the next run does not hit the
   same problem.
