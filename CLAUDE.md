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
just test          # Run all tests (lint + unit + coverage)
just go::unit      # Run unit tests only
just go::unit-int  # Run integration tests (requires running osapi)
just go::vet       # Run golangci-lint
just go::fmt       # Auto-format (gofumpt + golines)
go test -run TestName -v ./internal/job/...  # Run a single test
```

## Architecture (Quick Reference)

- **`cmd/`** - Cobra CLI commands (`client`, `node agent`, `controller.api`, `nats server`)
- **`internal/controller/api/`** - Echo REST API by domain. Node-targeted handlers are nested under `node/` (`node/hostname/`, `node/sysctl/`, `node/schedule/`, `node/docker/`, `node/command/`, `node/file/`, `node/network/`). Controller-only handlers: `job/`, `health/`, `audit/`, `agent/`, `file/`, `facts/`, `common/`. Types are OpenAPI-generated (`*.gen.go`). Combined OpenAPI spec: `internal/controller/api/gen/api.yaml`
- **`internal/job/`** - Job domain types, subject routing. `client/` for high-level ops
- **`internal/agent/`** - Node agent: consumer/handler/processor pipeline for job execution
- **`internal/telemetry/tracing/`** - OpenTelemetry tracer initialization, slog trace handler, context propagation\
- **`internal/telemetry/metrics/`** - Per-component Prometheus metrics server with isolated registries\
- **`internal/provider/`** - Operation implementations: `node/{host,disk,mem,load,sysctl}`, `network/{dns,ping}`, `scheduled/cron`, `container/docker`, `command`, `file`
- **`internal/telemetry/process/`** - Agent self-metrics (CPU%, RSS, goroutines) and process condition evaluation for heartbeat
- **`internal/controller/notify/`** - Pluggable condition notification system: watches registry KV for condition transitions, dispatches via `Notifier` interface (`log` backend)
- **`internal/config/`** - Viper-based config from `osapi.yaml`
- **`pkg/sdk/`** - Go SDK for programmatic REST API access (`client/` client library). See @docs/docs/sidebar/sdk/guidelines.md for SDK development rules
- Shared `nats-client` and `nats-server` are sibling repos linked via `replace` in `go.mod`
- **`github/`** - Temporary GitHub org config tooling (`repos.json` for declarative repo settings, `sync.sh` for drift detection via `gh` CLI). Untracked and intended to move to its own repo.

## Adding a New API Domain

When adding a new domain (e.g., `service`, `power`), follow existing
domains as reference. Node-targeted operations live under
`internal/controller/api/node/{domain}/` — follow `node/docker/` or
`node/schedule/`. Controller-only operations live under
`internal/controller/api/{domain}/` — follow `health/` or `audit/`.
Read the existing files before creating new ones.

### Step 0: Provider Implementation

Providers are the operations layer — they execute the actual work on
agent hosts. Every operation under `/node/{hostname}/...` is backed
by a provider. The request flows:

```
CLI → SDK → REST API → Job Client → NATS → Agent → Provider
```

The provider runs on the agent, not the controller. It receives
parameters from the job payload and returns a result.

#### Provider Types

**Direct providers** interact with the system directly:
- `node/host` — reads hostname, uptime, OS info
- `node/disk`, `node/mem`, `node/load` — reads system stats
- `network/dns` — reads/writes resolv.conf via `resolvectl`
- `network/ping` — executes ICMP ping
- `command` — executes arbitrary commands
- `docker` — manages Docker containers via the Docker SDK

Reference: `internal/provider/container/docker/` or `internal/provider/node/host/`

**Meta providers** don't write files directly — they delegate to
the file provider. This gives them SHA tracking, idempotency, drift
detection, and template rendering for free:
- `scheduled/cron` — deploys cron drop-in files and periodic scripts
- `node/sysctl` — deploys sysctl conf files to `/etc/sysctl.d/`
- Future: `systemd`, `apt sources`

Meta providers depend on `file.Deployer` (the narrow interface):
```go
type Deployer interface {
    Deploy(ctx context.Context, req DeployRequest) (*DeployResult, error)
    Undeploy(ctx context.Context, req UndeployRequest) (*UndeployResult, error)
}
```

Meta providers store domain-specific metadata in the
`FileState.Metadata` map (e.g., schedule, interval, user for cron).
The file provider persists this in the file-state KV alongside SHA,
path, and mode — one KV bucket for all providers.

Reference: `internal/provider/scheduled/cron/`

#### File Structure

Platform-specific providers:
```
internal/provider/{category}/{domain}/
  types.go              — Provider interface + domain types
  debian.go             — Debian-family implementation
  debian_{operation}.go — Per-operation file (large methods)
  debian_docker.go      — Container-aware variant (if needed)
  darwin.go             — macOS stub (returns ErrUnsupported)
  linux.go              — Generic Linux stub (returns ErrUnsupported)
  mocks/                — Generated gomock mocks
    generate.go         — //go:generate mockgen directive
```

SDK-based providers (no platform variants):
```
internal/provider/{category}/{domain}/
  types.go        — Provider interface + domain types
  {domain}.go     — Single implementation (e.g., docker.go)
  client.go       — API client interface for testing
  mocks/          — Generated gomock mocks
    generate.go   — //go:generate mockgen directive
```

For top-level providers: `internal/provider/{domain}/` (e.g.,
`internal/provider/command/`, `internal/provider/file/`).
For categorized providers: `internal/provider/{category}/{domain}/`
(e.g., `internal/provider/container/docker/`,
`internal/provider/scheduled/cron/`).

#### Provider Interface

```go
// types.go — package {domain}
type Provider interface {
    List(ctx context.Context) ([]Entry, error)
    Get(ctx context.Context, name string) (*Entry, error)
    Create(ctx context.Context, entry Entry) (*CreateResult, error)
    Update(ctx context.Context, entry Entry) (*UpdateResult, error)
    Delete(ctx context.Context, name string) (*DeleteResult, error)
}
```

Every method takes `context.Context` as the first parameter.
Result types include `Changed bool` for mutations and `Error string`
for per-operation error reporting.

Every concrete provider struct MUST embed `provider.FactsAware` and
include a compile-time check:

```go
// Compile-time check: Debian must satisfy FactsSetter.
var _ provider.FactsSetter = (*Debian)(nil)

type Debian struct {
    provider.FactsAware
    logger *slog.Logger
    // ...
}
```

The provider must also be passed to `provider.WireProviderFacts()`
in `internal/agent/agent.go` so facts are injected at startup.

#### Platform-Specific Implementations

OSAPI follows Ansible's OS family naming. Implementations are
selected at runtime via `platform.Detect()`:

- `debian.go` — Debian family (Ubuntu, Debian, Raspbian)
- `darwin.go` — macOS (for development)
- `linux.go` — generic Linux fallback

Unsupported platforms return `provider.ErrUnsupported`. The agent
marks the job as `skipped` (not `failed`) so the caller knows the
operation isn't available on that host rather than broken.

```go
// darwin.go
func (d *Darwin) List(
    _ context.Context,
) ([]Entry, error) {
    return nil, fmt.Errorf("cron: %w", provider.ErrUnsupported)
}
```

#### Provider Naming Conventions

There are three provider implementation patterns. The naming
convention determines the struct name, constructor, and file layout.

**1. Platform-specific providers** (most common)

One struct per OS family, each in its own file. Constructor names
follow `New{Platform}Provider()`. Methods that are large or testable
go in separate files named `{platform}_{operation}.go`.

| Struct   | Constructor              | File(s)                          |
| -------- | ------------------------ | -------------------------------- |
| `Debian` | `NewDebianProvider(...)` | `debian.go`, `debian_get_*.go`   |
| `Darwin` | `NewDarwinProvider(...)` | `darwin.go`, `darwin_get_*.go`   |
| `Linux`  | `NewLinuxProvider()`     | `linux.go`, `linux_get_*.go`     |

Examples: `node/host`, `node/disk`, `node/mem`, `node/load`,
`network/dns`, `network/ping`, `network/netinfo`.

**2. Container-aware platform providers**

When a provider's behavior differs inside a Docker container (e.g.,
hostname is read-only, DNS uses `/etc/resolv.conf` instead of
`resolvectl`), add a `DebianDocker` variant alongside the regular
`Debian` struct. The agent selects it via `platform.IsContainer()`.

| Struct          | Constructor                     | File(s)                              |
| --------------- | ------------------------------- | ------------------------------------ |
| `DebianDocker`  | `NewDebianDockerProvider(...)`  | `debian_docker.go`, `debian_docker_*.go` |

`DebianDocker` either embeds `Debian` (delegating reads, overriding
writes) or stands alone. It satisfies the same `Provider` interface.

```go
// agent_setup.go wiring
case "debian":
    if platform.IsContainer() {
        hostProvider = nodeHost.NewDebianDockerProvider()
    } else {
        hostProvider = nodeHost.NewDebianProvider(execManager)
    }
```

Examples: `node/host` (embeds `Debian`, blocks `UpdateHostname`),
`network/dns` (standalone, reads `/etc/resolv.conf` directly).

**3. SDK-based providers** (no platform variants)

Providers that talk to an external API (not the OS) use a single
`Client` struct with `New()` / `NewWithClient()` constructors.
No `debian.go` / `darwin.go` / `linux.go` files — the provider
works the same on all platforms. Availability is checked at startup
(e.g., Docker daemon ping).

| Struct   | Constructor         | File(s)              |
| -------- | ------------------- | -------------------- |
| `Client` | `New()`             | `docker.go`          |
|          | `NewWithClient(c)`  | (same file, testing) |

```go
// agent_setup.go wiring — no platform switch
dockerClient, err := dockerNewFn()
if err == nil {
    if pingErr := dockerClient.Ping(ctx); pingErr == nil {
        dockerProvider = dockerClient
    }
}
```

Examples: `container/docker`.

#### FactsAware

Embed `provider.FactsAware` in the provider struct to access agent
facts (OS family, architecture, hostname, network interfaces) at
runtime. The agent wires facts via `provider.WireProviderFacts()`.

```go
type Debian struct {
    provider.FactsAware
    logger *slog.Logger
    fs     avfs.VFS
}
```

Facts are available in template rendering via `{{ .Facts.os_family }}`
when using the file provider's template support.

#### Agent Wiring

Two files connect a provider to the agent:

1. **`internal/agent/processor_{domain}.go`** — create helper
   functions that dispatch sub-operations to the provider. If the
   domain gets its own category (like `schedule`, `docker`), create
   a `NewXxxProcessor` factory. If the domain belongs under an
   existing category (like `node`), add a `case` to that category's
   processor and delegate to helpers in a new file:
   ```go
   // processor_{domain}.go
   func process{Domain}Operation(
       provider {domain}.Provider,
       logger *slog.Logger,
       req job.Request,
   ) (json.RawMessage, error) {
       // switch on sub-operation, call provider, marshal result
   }
   ```

2. **`cmd/agent_setup.go`** — create the provider and register it
   with the `ProviderRegistry`. For new categories, use a separate
   `Register` call. For existing categories (e.g., `node`), pass
   the provider to the existing processor factory:
   ```go
   // New category example (like schedule, docker):
   registry.Register("mydomain",
       agent.NewMyDomainProcessor(myProv, log),
       myProv)

   // Existing category example (like sysctl under node):
   registry.Register("node",
       agent.NewNodeProcessor(
           hostProv, diskProv, memProv, loadProv,
           sysctlProv,  // added to existing processor
           appConfig, log,
       ),
       hostProv, diskProv, memProv, loadProv, sysctlProv,
   )
   ```

That's it. No changes to `agent/types.go`, `agent/agent.go`, or
the `JobClient` interface. The registry handles dispatch and
FactsAware wiring automatically.

#### Provider Testing

- **Filesystem:** Use `avfs` — `memfs.New()` for in-memory,
  `failfs.New()` for targeted error injection. Never use `afero`.
- **Mocks:** Use gomock for all interfaces (`FileDeployer`,
  `KeyValue`, `ObjectStore`). Generated mocks live in
  `{package}/mocks/`.
- **Platform stubs:** Test that Darwin and Linux stubs return
  `ErrUnsupported` for every method.
- **export_test.go:** Use for testing unexported variable swaps
  (e.g., `marshalJSON`). Public tests import via the bridge.
- **Table-driven:** One suite method per provider method, all
  scenarios as rows.

### Step 1: OpenAPI Spec + Code Generation

For node-targeted domains, create
`internal/controller/api/node/{domain}/gen/` with three hand-written
files. For controller-only domains, create
`internal/controller/api/{domain}/gen/` instead:

- `api.yaml` — OpenAPI spec with paths, schemas, and `BearerAuth` security
- `cfg.yaml` — oapi-codegen config (`strict-server: true`, import-mapping
  for `common/gen`)
- `generate.go` — `//go:generate` directive

#### HTTP Verb Conventions

Mutable domains MUST use separate verbs for create and update:

- `POST` — create a new resource (key/name in request body)
- `PUT` — update an existing resource (key/name from path parameter)
- `GET` — read/list resources
- `DELETE` — remove a resource

Do NOT combine create and update into a single "set" or "upsert"
endpoint. The cron domain is the reference: `POST` creates,
`PUT /{name}` updates. This separation gives clear 404 semantics
(update fails if not found, create fails if already exists) and
matches REST conventions.

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
  NOT generate tags on `RequestObject` structs in **strict-server
  mode** (upstream limitation — see oapi-codegen issue). Keep the
  `x-oapi-codegen-extra-tags` in the spec for documentation and add
  a YAML comment noting validation is handled manually. Path params
  that need validation beyond `format: uuid` (e.g., `valid_target`)
  use a shared helper like `node.validateHostname()` which calls
  `validation.Var()`.

**IMPORTANT — every endpoint with user input MUST have:**
1. `x-oapi-codegen-extra-tags` with `validate:` tags on all request
   body properties and query params in the OpenAPI spec
2. `validation.Struct(request.Params)` in the handler for query params,
   `validation.Struct(request.Body)` for request bodies
3. A `400` response defined in the OpenAPI spec for endpoints that
   accept user input
4. HTTP wiring tests (`TestXxxHTTP` / `TestXxxRBACHTTP` methods in the
   `*_public_test.go` suite) that send raw HTTP through the full Echo
   middleware stack and verify:
   - Validation errors return correct status codes and error messages
   - RBAC: 401 (no token), 403 (wrong permissions), 200 (valid token)

### Step 2: Handler Implementation

For node-targeted domains, create
`internal/controller/api/node/{domain}/`. For controller-only
domains, create `internal/controller/api/{domain}/`:

- `types.go` — domain struct, dependency interfaces (e.g., `Checker`)
- `{domain}.go` — `New()` factory, compile-time interface check:
  `var _ gen.StrictServerInterface = (*Domain)(nil)`
- One file per endpoint (e.g., `{operation}_get.go`). Every handler
  that accepts user input MUST call `validation.Struct()` and return
  a 400 on failure.
- Tests: `{operation}_get_public_test.go` (testify/suite, table-driven).
  Must cover validation failures (400), success, and error paths.
  Each public test suite also includes HTTP wiring methods:
  - `TestXxxHTTP` — sends raw HTTP through the full Echo middleware
    stack to verify validation (valid input, invalid input → 400).
  - `TestXxxRBACHTTP` — verifies auth middleware: no token (401),
    wrong permissions (403), valid token (200). Uses `api.New()` +
    `server.GetXxxHandler()` + `server.RegisterHandlers()` to wire
    through `scopeMiddleware`.
  See existing examples in `internal/controller/api/node/docker/`,
  `internal/controller/api/job/`, and
  `internal/controller/api/audit/`.

#### Broadcast Support (MANDATORY for node-targeted operations)

Every operation under `/node/{hostname}/...` MUST support broadcast
targeting (`_all`, `_any`, hostname, label selectors). The handler
checks `job.IsBroadcastTarget(hostname)` and routes to a broadcast
function. Both single-target and broadcast paths return the same
collection response shape.

**Response pattern** — all node-targeted operations return:
```json
{
  "job_id": "...",
  "results": [
    {"hostname": "web-01", "error": "", ...domain fields...},
    {"hostname": "web-02", "error": "unsupported", ...}
  ]
}
```

Every result item MUST have `hostname` and `error` fields.
Single-target returns 1 result; broadcast returns N results.
Failed/skipped agents appear as entries with `error` set.

**Handler pattern:**
```go
func (s *Handler) PostOperation(ctx, request) {
    validate(request)
    hostname := request.Hostname
    if job.IsBroadcastTarget(hostname) {
        return s.postOperationBroadcast(ctx, hostname, ...)
    }
    // Single-target: wrap in collection with 1 result.
}
```

**Job client** — the `JobClient` interface has 4 generic methods:
`Query`, `QueryBroadcast`, `Modify`, `ModifyBroadcast`. Handlers
call these with a category string and operation constant. No new
methods are needed when adding operations. Example:
```go
jobID, resp, err := s.JobClient.Modify(
    ctx, hostname, "node", job.OperationSysctlCreate, data)
```

See `internal/controller/api/node/node_status_get.go` for the
reference implementation.

### Step 3: Server Wiring (4 files in `internal/controller/api/`)

For node-targeted domains, the handler shim is named
`handler_node_{domain}.go` with a `GetNode{Domain}Handler()` method.
For controller-only domains, use `handler_{domain}.go` with
`Get{Domain}Handler()`.

- `handler_node_{domain}.go` (or `handler_{domain}.go`) — wraps the
  handler with `NewStrictHandler` + `scopeMiddleware`. Define
  `unauthenticatedOperations` map if any endpoints skip auth.
- `types.go` — add `{domain}Handler` field to `Server` struct +
  `With{Domain}Handler()` option func
- `handler.go` — call `Get{Domain}Handler()` in `CreateHandlers()` and
  append results
- `handler_public_test.go` — add `TestGet{Domain}Handler` with test cases
  for both unauthenticated and authenticated paths

### Step 4: Startup Wiring

- `cmd/controller_start.go` — initialize the handler with real
  dependencies and pass `api.With{Domain}Handler(h)` to `api.New()`

### Step 5: Update SDK

The SDK client library lives in `pkg/sdk/client/`. Its generated HTTP client
uses the same combined OpenAPI spec as the server
(`internal/controller/api/gen/api.yaml`). Follow the rules in
@docs/docs/sidebar/sdk/guidelines.md — especially: never expose `gen`
types in public method signatures, add JSON tags to all result types,
and wrap errors with context.

**When modifying existing API specs:**

1. Make changes to the domain's `gen/api.yaml` (under `api/node/{domain}/`
   for node-targeted domains or `api/{domain}/` for controller-only domains)
2. Run `just generate` to regenerate server code (this also regenerates the
   combined spec via `redocly join`)
3. Run `go generate ./pkg/sdk/client/gen/...` to regenerate the SDK client
4. Update the SDK service wrappers in `pkg/sdk/client/{domain}.go` if new
   response codes were added
5. Update CLI switch blocks in `cmd/` if new response codes were added

**When adding a new API domain:**

1. Add a service wrapper in `pkg/sdk/client/{domain}.go`
2. Run `go generate ./pkg/sdk/client/gen/...` to pick up the new domain's
   spec from the combined `api.yaml`
3. Add an SDK example in `examples/sdk/client/{domain}.go`
4. Add an SDK doc page in `docs/docs/sidebar/sdk/client/{domain}.md`
   with methods table, request types, usage examples, and permissions

#### SDK example conventions

SDK examples live in `examples/sdk/client/`, one file per domain.
Follow the same principles as the orchestrator examples:

- **One domain per file**: demonstrate the domain's SDK operations.
  Don't mix in other domains.
- **Self-contained**: for read-only operations, just call and print.
  For mutating operations, cleanup at the start so the example is
  repeatable.
- **Print results**: decode and print at least one result so the
  example isn't silent.
- **Keep it short**: under ~100 lines of code (excluding license).
- **Handle errors inline**: use `log.Fatalf` for unexpected errors.
  For operations that may fail on some platforms, check the error
  and print a message instead of crashing.

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
- Update `docs/docs/sidebar/usage/configuration.md` — add any new
  permissions to the roles table and permissions comments in the
  YAML reference
- Update `docs/docs/sidebar/features/authentication.md` — add new
  permissions to the roles/permissions tables
- Update `docs/docs/sidebar/architecture/architecture.md` — add link
  to the new feature page in the features list
- Update `docs/docs/sidebar/architecture/api-guidelines.md` — add
  new endpoints to the path pattern table
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

Three test layers:
- **Unit tests** (`*_test.go`, `*_public_test.go`) — fast, mocked
  dependencies, run with `just go::unit`. Includes `TestXxxHTTP` /
  `TestXxxRBACHTTP` methods that send raw HTTP through real Echo
  middleware with mocked backends.
- **Integration tests** (`test/integration/`) — build and start a real
  `osapi` binary, exercise CLI commands end-to-end. Guarded by
  `//go:build integration` tag, run with `just go::unit-int`. New API
  domains should include a `{domain}_test.go` smoke suite. Write tests
  (mutations) must be guarded by `skipWrite(s.T())` so CI can run
  read-only tests by default (`OSAPI_INTEGRATION_WRITES=1` enables
  writes).

Conventions:
- ALL tests MUST use `testify/suite` with table-driven patterns
- Public tests: `*_public_test.go` in test package (e.g.,
  `package job_test`) for exported functions. This is the default —
  all new tests should be public tests.
- Suite naming: `*_public_test.go` → `{Name}PublicTestSuite`
- Table-driven structure with `validateFunc` callbacks
- One suite method per function under test — all scenarios (success,
  errors, edge cases) as rows in one table
- Avoid generic file names like `helpers.go` or `utils.go` — name
  files after what they contain
- **`types.go` is for types only**: `types.go` files MUST contain
  only type definitions (structs, interfaces, constants, type
  aliases). Never put functions or methods in `types.go` — put
  them in a file named after what they do (e.g., `nats.go` for
  NATS config methods, `options.go` for option functions).
- **Test file naming**: every test file MUST have a corresponding
  production file with a matching name. `foo_public_test.go` tests
  `foo.go`. Never create test files with names that don't match a
  production file (e.g., don't create `check_error_public_test.go`
  if the code lives in `response.go` — name it
  `response_public_test.go`). If a production file is too large and
  you want to split tests by concern, split the production file
  first (e.g., `agent.go` → `agent.go` + `agent_drain.go` +
  `agent_timeline.go`), then create matching test files.

#### Mocking

- **Always use gomock** (`go:generate mockgen`) for interface mocks.
  Generated mocks live in `{package}/mocks/` directories alongside
  their source interfaces. Never hand-roll mock structs.
- **export_test.go pattern** for testing unexported internals: create
  an `export_test.go` file in the production package that exposes
  unexported variables or functions to the `_test` package:
  ```go
  // export_test.go — package file
  package file

  func SetMarshalJSON(fn func(interface{}) ([]byte, error)) {
      marshalJSON = fn
  }
  func ResetMarshalJSON() { marshalJSON = json.Marshal }
  ```
  Public tests then call `file.SetMarshalJSON(...)` and
  `defer file.ResetMarshalJSON()`. This avoids internal tests,
  import cycles, and hand-rolled stubs.
- **TearDownSubTest** — use `suite.TearDownSubTest()` to reset
  swapped variables between table-driven sub-tests, not `defer`
  inside the loop.
- **Filesystem testing** — use `avfs` (`memfs.New()` for in-memory,
  `failfs.New()` for targeted error injection). Never use
  `afero`. The only exception for hand-rolled types is stdlib
  interfaces like `fs.FS` or `net.Conn` where gomock is impractical.

### Go Patterns

- Non-blocking lifecycle: `Start()` returns immediately, `Stop(ctx)` shuts down with deadline
- Error wrapping: `fmt.Errorf("context: %w", err)`
- Early returns over nested if-else
- Unused parameters: rename to `_`
- Import order: stdlib, third-party, local (blank-line separated)

### Logging

All logging uses Go's `log/slog` structured logger. Follow these rules:

- **Subsystem labels**: Every component that holds a logger MUST wrap it
  with `logger.With(slog.String("subsystem", "..."))` at construction
  time. This auto-tags every log line from that component. Examples:
  `"agent"`, `"agent.seed"`, `"api.schedule"`, `"provider.file"`,
  `"job.client"`, `"metrics"`, `"controller.heartbeat"`.
- **Always use typed attributes**: Use `slog.String("key", val)`,
  `slog.Int("key", val)`, `slog.Bool("key", val)`, `slog.Any("key", val)`.
  Never use positional pairs like `"key", val` — they compile but
  bypass type safety and are inconsistent with the codebase.
- **Standard field names**: `error` for errors, `hostname` for hosts,
  `path` for file paths, `job_id` for job IDs, `name` for entry names,
  `addr` for addresses.
- **Error fields**: Use `slog.String("error", err.Error())` for string
  context or `slog.Any("error", err)` to preserve the error type.
- **Log levels**: `Debug` for operation dispatch and idempotency skips,
  `Info` for lifecycle events and state changes, `Warn` for degraded
  but functional states, `Error` for failures that need attention.

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

Implementation planning and execution uses the superpowers plugin workflows
(`writing-plans` and `executing-plans`). Plans live in `docs/plans/`.
