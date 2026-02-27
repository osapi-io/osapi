# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

OSAPI is a Linux system management REST API and CLI written in Go 1.25. It uses NATS JetStream for distributed async job processing with a KV-first, stream-notification architecture.

## Development Reference

For setup, building, testing, and contributing, see the Docusaurus docs:

- @docs/docs/sidebar/development/development.md - Prerequisites, setup, code style, testing, commit conventions
- @docs/docs/sidebar/development/contributing.md - PR workflow and contribution guidelines
- @docs/docs/sidebar/development/testing.md - How to run tests and list just recipes
- @docs/docs/sidebar/architecture/principles.md - Guiding principles (simplicity, minimalism, design philosophy)
- @docs/docs/sidebar/architecture/api-guidelines.md - API design guidelines (REST conventions, endpoint structure)
- @docs/docs/sidebar/usage/configuration.md - Configuration reference (osapi.yaml, env overrides)
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

- **`cmd/`** - Cobra CLI commands (`client`, `node agent`, `api server`, `nats server`)
- **`internal/api/`** - Echo REST API by domain (`node/`, `network/`, `job/`, `command/`, `common/`). Types are OpenAPI-generated (`*.gen.go`)
- **`internal/job/`** - Job domain types, subject routing. `client/` for high-level ops, `worker/` for consumer/handler/processor pipeline
- **`internal/provider/`** - Operation implementations: `node/{host,disk,mem,load}`, `network/{dns,ping}`
- **`internal/config/`** - Viper-based config from `osapi.yaml`
- **`osapi-sdk`** - External SDK for programmatic REST API access (sibling repo, linked via `replace` in `go.mod`)
- Shared `nats-client` and `nats-server` are sibling repos linked via `replace` in `go.mod`
- **`github/`** - Temporary GitHub org config tooling (`repos.json` for declarative repo settings, `sync.sh` for drift detection via `gh` CLI). Untracked and intended to move to its own repo.

## Adding a New API Domain

When adding a new domain (e.g., `service`, `power`), follow the `health`
domain as a reference. Read the existing files before creating new ones.

### Step 1: OpenAPI Spec + Code Generation

Create `internal/api/{domain}/gen/` with three hand-written files:

- `api.yaml` — OpenAPI spec with paths, schemas, and `BearerAuth` security
- `cfg.yaml` — oapi-codegen config (`strict-server: true`, import-mapping
  for `common/gen`)
- `generate.go` — `//go:generate` directive

#### Validation in OpenAPI Specs

The OpenAPI spec is the **source of truth** for input validation. All user
input must be validated, and the spec must declare how:

- **Request body properties**: Add `x-oapi-codegen-extra-tags` with
  `validate:` tags. These generate Go struct tags that
  `validation.Struct()` enforces at runtime.
  ```yaml
  properties:
    address:
      type: string
      x-oapi-codegen-extra-tags:
        validate: required,ip
  ```
- **Path parameters (UUID)**: Use `format: uuid` on the schema. This
  causes oapi-codegen to generate `openapi_types.UUID` type, and the
  router validates the format before the handler runs. No manual
  validation needed in the handler.
  ```yaml
  parameters:
    - name: id
      in: path
      required: true
      schema:
        type: string
        format: uuid
  ```
- **Query parameters**: Place `x-oapi-codegen-extra-tags` at the
  **parameter level** (sibling of `name`/`in`/`schema`), NOT inside
  `schema:`. At parameter level, oapi-codegen generates `validate:`
  tags on the `*Params` struct fields. Use `enum` for constrained
  string values (generates `oneof` validation).
  ```yaml
  parameters:
    - name: limit
      in: query
      required: false
      x-oapi-codegen-extra-tags:
        validate: omitempty,min=1,max=100
      schema:
        type: integer
        default: 20
        minimum: 1
        maximum: 100
  ```
  Then in the handler, validate with a single call:
  ```go
  if errMsg, ok := validation.Struct(request.Params); !ok {
      return gen.GetFoo400JSONResponse{Error: &errMsg}, nil
  }
  ```
  **NOTE:** `x-oapi-codegen-extra-tags` on **path parameters** does
  NOT generate tags on request object structs. Path params that need
  validation beyond `format: uuid` (e.g., alphanum checks) still
  require manual validation with a temporary struct.

**IMPORTANT — every endpoint with user input MUST have:**
1. `x-oapi-codegen-extra-tags` with `validate:` tags on all request
   body properties and query params in the OpenAPI spec
2. `validation.Struct(request.Params)` in the handler for query params,
   `validation.Struct(request.Body)` for request bodies
3. A `400` response defined in the OpenAPI spec for endpoints that
   accept user input
4. An integration test (`*_integration_test.go`) that sends raw HTTP
   through the full Echo middleware stack and verifies:
   - Validation errors return correct status codes and error messages
   - RBAC: 401 (no token), 403 (wrong permissions), 200 (valid token)

### Step 2: Handler Implementation

Create `internal/api/{domain}/`:

- `types.go` — domain struct, dependency interfaces (e.g., `Checker`)
- `{domain}.go` — `New()` factory, compile-time interface check:
  `var _ gen.StrictServerInterface = (*Domain)(nil)`
- One file per endpoint (e.g., `{operation}_get.go`). Every handler
  that accepts user input MUST call `validation.Struct()` and return
  a 400 on failure.
- Tests: `{operation}_get_public_test.go` (testify/suite, table-driven).
  Must cover validation failures (400), success, and error paths.
- Integration tests: `{operation}_get_integration_test.go` — sends raw
  HTTP through the full Echo middleware stack. Every integration test
  MUST include:
  - **Validation tests**: valid input, invalid input (400 responses)
  - **RBAC tests**: no token (401), wrong permissions (403), valid
    token (200). Uses `api.New()` + `server.GetXxxHandler()` +
    `server.RegisterHandlers()` to wire through `scopeMiddleware`.
  See existing examples in `internal/api/job/` and
  `internal/api/audit/`.

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

### Step 5: Update SDK

The `osapi-sdk` (sibling repo) provides the generated HTTP client used by
the CLI. The SDK syncs its `api.yaml` files from this repo via `gilt`
overlay (configured in `osapi-sdk/.gilt.yml`). When `just generate` runs
in the SDK, gilt pulls the latest specs from osapi's `main` branch and
regenerates the client code.

**When adding a new API domain:**

1. Add the domain's `api.yaml` to `osapi-sdk/pkg/osapi/gen/{domain}/`
2. Run `just generate` in the SDK repo to regenerate the merged spec and
   client code
3. Add a service wrapper in `osapi-sdk/pkg/osapi/{domain}.go`

**When modifying existing API specs** (adding responses, parameters, or
schemas to existing endpoints):

1. Make changes to `internal/api/{domain}/gen/api.yaml` in this repo
2. Run `just generate` here to regenerate server code
3. After merging to `main`, run `just generate` in `osapi-sdk` — gilt
   will pull the updated specs and regenerate the client
4. Update the SDK service wrappers and CLI switch blocks if new response
   codes were added (e.g., adding a 404 response requires a
   `case http.StatusNotFound:` in the CLI)

### Step 6: CLI Commands

- `cmd/client_{domain}.go` — parent command registered under `clientCmd`
- `cmd/client_{domain}_{operation}.go` — one subcommand per endpoint
- All commands support `--json` for raw output
- Use `printKV` for inline key-value output and `printStyledTable` for
  multi-row tabular data (both in `cmd/ui.go`)
- Use flags (e.g., `--job-id`, `--audit-id`) instead of positional args
  for resource IDs
- Handle **all** API response codes in the `switch resp.StatusCode()`
  block: 200, 400 (`handleUnknownError`), 401/403 (`handleAuthError`),
  404 (`handleUnknownError`), 500 (`handleUnknownError`). Match the
  responses declared in the OpenAPI spec.

### Step 7: Documentation

- `docs/docs/sidebar/features/{domain}.md` — feature page describing
  what the domain manages, how it works, configuration, permissions,
  and links to CLI, API, and architecture docs. Follow the consistent
  template used by existing feature pages (see `features/` directory).
- `docs/docs/sidebar/usage/cli/client/{domain}/{domain}.md` — parent
  page with `<DocCardList />` for sidebar navigation
- `docs/docs/sidebar/usage/cli/client/{domain}/{operation}.md` — one page
  per CLI subcommand with usage examples and `--json` output
- Update `docs/docusaurus.config.ts` — add the new feature to the
  "Features" navbar dropdown
- Update `docs/docs/sidebar/usage/configuration.md` — add any new config
  sections (env vars, YAML reference, section reference table)
- Update `docs/docs/sidebar/architecture/system-architecture.md` — add
  endpoints to the health/endpoint tables if applicable

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

See @docs/docs/sidebar/development/development.md#branching for full conventions.

When committing changes via `/commit`, create a feature branch first if
currently on `main`. Branch names use the pattern `type/short-description`
(e.g., `feat/add-dns-retry`, `fix/memory-leak`, `docs/update-readme`).

### Commit Messages

See @docs/docs/sidebar/development/development.md#commit-messages for full conventions.

Follow [Conventional Commits](https://www.conventionalcommits.org/) with the
50/72 rule. Format: `type(scope): description`.

When committing via Claude Code, end with:
- `🤖 Generated with [Claude Code](https://claude.ai/code)`
- `Co-Authored-By: Claude <noreply@anthropic.com>`

## Task Tracking

Work is tracked as markdown files in `docs/docs/sidebar/development/tasks/`. These
render on the documentation site. See
@docs/docs/sidebar/development/tasks/README.md for format details.

```
docs/docs/sidebar/development/tasks/
├── backlog/          # Tasks not yet started
├── in-progress/      # Tasks actively being worked on
├── done/             # Completed tasks
└── sessions/         # Session work logs (per Claude Code session)
```

When starting a session:
1. Check `docs/docs/sidebar/development/tasks/in-progress/` for ongoing work
2. Check `docs/docs/sidebar/development/tasks/backlog/` for next tasks
3. Move task files between directories as status changes
4. Log session work in `docs/docs/sidebar/development/tasks/sessions/YYYY-MM-DD.md`
