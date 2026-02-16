# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

OSAPI is a Linux system management REST API and CLI written in Go 1.25. It uses NATS JetStream for distributed async job processing with a KV-first, stream-notification architecture.

## Development Reference

For setup, building, testing, and contributing, see the Docusaurus docs:

- @docs/docs/sidebar/development.md - Prerequisites, setup, code style, testing, commit conventions
- @docs/docs/sidebar/contributing.md - PR workflow and contribution guidelines
- @docs/docs/sidebar/testing.md - How to run tests and list just recipes
- @docs/docs/sidebar/architecture.md - Job system architecture (KV-first, subject routing, worker pipeline)

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

- Context-based lifecycle: `Start(ctx)` blocks until `ctx.Done()`, no explicit `Stop()` methods
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
