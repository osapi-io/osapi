# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

OSAPI is a Linux system management REST API and CLI written in Go 1.25. It uses NATS JetStream for distributed async job processing with a KV-first, stream-notification architecture.

## Development Reference

For setup, building, testing, and contributing, see the Docusaurus docs:

- @docs/docs/sidebar/development.md - Prerequisites, setup, code style, testing, commit conventions
- @docs/docs/sidebar/contributing.md - PR workflow and contribution guidelines
- @docs/docs/sidebar/testing.md - How to run tests and list just recipes
- @docs/docs/sidebar/principles.md - Guiding principles (simplicity, minimalism, design philosophy)
- @docs/docs/sidebar/api-guidelines.md - API design guidelines (REST conventions, endpoint structure)
- @docs/docs/sidebar/configuration.md - Configuration reference (osapi.yaml, env overrides)
- @docs/docs/sidebar/architecture/architecture.md - Architecture overview (links to system and job architecture)

Quick reference for common commands:

```bash
just deps          # Install all dependencies
just test          # Run all tests (lint + unit + coverage + bats)
just go::unit      # Run unit tests only
just go::vet       # Run golangci-lint
just go::fmt       # Auto-format (gofumpt + golines)
go test -run TestName -v ./internal/job/...  # Run a single test
```

## Architecture (Quick Reference)

- **`cmd/`** - Cobra CLI commands (`client`, `job`, `api server`, `job worker`)
- **`internal/api/`** - Echo REST API by domain (`system/`, `network/`, `job/`, `common/`). Types are OpenAPI-generated (`*.gen.go`)
- **`internal/job/`** - Job domain types, subject routing. `client/` for high-level ops, `worker/` for consumer/handler/processor pipeline
- **`internal/provider/`** - Operation implementations: `system/{host,disk,mem,load}`, `network/{dns,ping}`
- **`internal/config/`** - Viper-based config from `osapi.yaml`
- **`internal/client/`** - Generated REST API client
- Shared `nats-client` and `nats-server` are sibling repos linked via `replace` in `go.mod`

## Adding a New API Domain

When adding a new domain (e.g., `service`, `power`), follow the `health`
domain as a reference. Read the existing files before creating new ones.

### Step 1: OpenAPI Spec + Code Generation

Create `internal/api/{domain}/gen/` with three hand-written files:

- `api.yaml` — OpenAPI spec with paths, schemas, and `BearerAuth` security
- `cfg.yaml` — oapi-codegen config (`strict-server: true`, import-mapping
  for `common/gen`)
- `generate.go` — `//go:generate` directive

### Step 2: Handler Implementation

Create `internal/api/{domain}/`:

- `types.go` — domain struct, dependency interfaces (e.g., `Checker`)
- `{domain}.go` — `New()` factory, compile-time interface check:
  `var _ gen.StrictServerInterface = (*Domain)(nil)`
- One file per endpoint (e.g., `{operation}_get.go`)
- Tests: `{operation}_get_public_test.go` (testify/suite, table-driven)

### Step 3: Server Wiring (4 files in `internal/api/`)

- `handler_{domain}.go` — `Get{Domain}Handler()` method that wraps the
  handler with `NewStrictHandler` + `scopeMiddleware`. Define
  `unauthenticatedOperations` map if any endpoints skip auth.
- `types.go` — add `{domain}Handler` field to `Server` struct +
  `With{Domain}Handler()` option func
- `handler.go` — call `Get{Domain}Handler()` in `CreateHandlers()` and
  append results
- `handler_public_test.go` — add `TestGet{Domain}Handler` with test cases
  for both unauthenticated and authenticated paths

### Step 4: Startup Wiring

- `cmd/api_server_start.go` — initialize the handler with real
  dependencies and pass `api.With{Domain}Handler(h)` to `api.New()`

### Step 5: Generate Client Code

Run `just generate` which:
1. `redocly join` merges all `internal/api/*/gen/api.yaml` into
   `internal/client/gen/api.yaml`
2. `go generate` creates `client.gen.go` with typed
   `Get{Op}WithResponse()` methods

### Step 6: Client Wrappers

- `internal/client/handler.go` — add `{Domain}Handler` interface to
  `CombinedHandler` composition
- `internal/client/{domain}_{operation}.go` — thin wrapper per endpoint
  calling `c.Client.Get{Op}WithResponse(ctx)`
- `internal/client/client_public_test.go` — add test methods to existing
  suite

### Step 7: CLI Commands

- `cmd/client_{domain}.go` — parent command registered under `clientCmd`
- `cmd/client_{domain}_{operation}.go` — one subcommand per endpoint
- All commands support `--json` for raw output
- Use `printStyledMap` and `printStyledTable` from `cmd/ui.go` for
  formatted output

### Step 8: Verify

```bash
just generate        # regenerate specs + code
go build ./...       # compiles
just go::unit        # tests pass
just go::vet         # lint passes
```

## Code Standards (MANDATORY)

### Function Signatures

ALL function signatures MUST use multi-line format:
```go
func FunctionName(
    param1 type1,
    param2 type2,
) (returnType, error) {
}
```

### Testing

- ALL tests in `internal/job/` MUST use `testify/suite` with table-driven patterns
- Internal tests: `*_test.go` in same package (e.g., `package job`) for private functions
- Public tests: `*_public_test.go` in test package (e.g., `package job_test`) for exported functions
- Table-driven structure with `validateFunc` callbacks

### Go Patterns

- Non-blocking lifecycle: `Start()` returns immediately, `Stop(ctx)` shuts down with deadline
- Error wrapping: `fmt.Errorf("context: %w", err)`
- Early returns over nested if-else
- Unused parameters: rename to `_`
- Import order: stdlib, third-party, local (blank-line separated)

### Linting

golangci-lint with: errcheck, errname, goimports, govet, prealloc, predeclared, revive, staticcheck. Generated files (`*.gen.go`, `*.pb.go`) are excluded from formatting.

### Branching

See @docs/docs/sidebar/development.md#branching for full conventions.

When committing changes via `/commit`, create a feature branch first if
currently on `main`. Branch names use the pattern `type/short-description`
(e.g., `feat/add-dns-retry`, `fix/memory-leak`, `docs/update-readme`).

### Commit Messages

See @docs/docs/sidebar/development.md#commit-messages for full conventions.

Follow [Conventional Commits](https://www.conventionalcommits.org/) with the
50/72 rule. Format: `type(scope): description`.

When committing via Claude Code, end with:
- `🤖 Generated with [Claude Code](https://claude.ai/code)`
- `Co-Authored-By: Claude <noreply@anthropic.com>`

## Task Tracking

Work is tracked as markdown files in `.tasks/`. See @.tasks/README.md for format details.

```
.tasks/
├── backlog/          # Tasks not yet started
├── in-progress/      # Tasks actively being worked on
├── done/             # Completed tasks
└── sessions/         # Session work logs (per Claude Code session)
```

When starting a session:
1. Check `.tasks/in-progress/` for ongoing work
2. Check `.tasks/backlog/` for next tasks
3. Move task files between directories as status changes
4. Log session work in `.tasks/sessions/YYYY-MM-DD.md`
