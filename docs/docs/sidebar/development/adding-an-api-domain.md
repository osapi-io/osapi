---
sidebar_position: 5
---

# Adding an API Domain

When adding a new domain, follow existing domains as reference. Node-targeted
operations live under `internal/controller/api/node/{domain}/`. Controller-only
operations live under `internal/controller/api/{domain}/`. Read existing domains
before creating new ones — the codebase IS the reference.

## Cross-Layer Consistency (MANDATORY)

Every domain MUST be consistent across all layers: provider, agent processor,
API handler, SDK service, CLI commands, docs, and tests. When adding a new
domain, look at a recently completed domain (like `ntp` or `sysctl`) and
replicate the same set of artifacts across every layer. If something exists for
`sysctl` but not for your new domain, it's missing.

The principle: **pick any existing domain and `find`/`grep` for it across the
codebase. Your new domain should appear in all the same places.** This includes
code, tests, examples, SDK docs, CLI docs, feature docs, docusaurus config, and
permissions tables.

## Step 0: Provider Implementation

Providers are the operations layer — they execute the actual work on agent
hosts. Every operation under `/node/{hostname}/...` is backed by a provider. The
request flows:

```
CLI → SDK → REST API → Job Client → NATS → Agent → Provider
```

The provider runs on the agent, not the controller. It receives parameters from
the job payload and returns a result.

### Provider Types

Three provider patterns exist. Check existing providers for examples of each:

**Direct providers** interact with the system directly via commands or system
calls. No file management.

**Meta providers** delegate file writes to `file.Deployer` for SHA tracking,
idempotency, and template rendering.

**Direct-write providers** manage their own config files via `avfs.VFS` without
`file.Deployer`. They use the `osapi-` filename prefix to identify managed
files.

Meta providers depend on `file.Deployer` (the narrow interface):

```go
type Deployer interface {
    Deploy(ctx context.Context, req DeployRequest) (*DeployResult, error)
    Undeploy(ctx context.Context, req UndeployRequest) (*UndeployResult, error)
}
```

Meta providers store domain-specific metadata in the `FileState.Metadata` map
(e.g., schedule, interval, user for cron). The file provider persists this in
the file-state KV alongside SHA, path, and mode — one KV bucket for all
providers.

Reference: look at existing meta providers in the codebase.

### File Structure

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

For top-level providers: `internal/provider/{domain}/`. For categorized
providers: `internal/provider/{category}/{domain}/`. Look at existing providers
to see both patterns.

### Provider Interface

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

Every method takes `context.Context` as the first parameter. Result types
include `Changed bool` for mutations and `Error string` for per-operation error
reporting.

### Idempotency (MANDATORY)

All provider mutations follow Ansible-style desired-state semantics. Operations
MUST be idempotent:

| Operation  | Resource exists       | Resource absent       |
| ---------- | --------------------- | --------------------- |
| **Create** | `Changed: false`, nil | Creates it            |
| **Update** | Updates it            | Error (not found)     |
| **Delete** | Removes it            | `Changed: false`, nil |

- **Create** when the resource already exists returns success with
  `Changed: false` — the desired state (present) is already met.
- **Delete** when the resource doesn't exist returns success with
  `Changed: false` — the desired state (absent) is already met.
- **Update** when the resource doesn't exist returns an error — there is nothing
  to update.
- `ErrUnsupported` (wrong OS family) maps to `StatusSkipped` at the agent layer,
  which is distinct from `Changed: false`.

Every concrete provider struct MUST embed `provider.FactsAware` and include a
compile-time check:

```go
// Compile-time check: Debian must satisfy FactsSetter.
var _ provider.FactsSetter = (*Debian)(nil)

type Debian struct {
    provider.FactsAware
    logger *slog.Logger
    // ...
}
```

The provider must also be passed to `provider.WireProviderFacts()` in
`internal/agent/agent.go` so facts are injected at startup.

### Platform-Specific Implementations

OSAPI follows Ansible's OS family naming. Implementations are selected at
runtime via `platform.Detect()`:

- `debian.go` — Debian family (Ubuntu, Debian, Raspbian)
- `darwin.go` — macOS (for development)
- `linux.go` — generic Linux fallback

Unsupported platforms return `provider.ErrUnsupported`. The agent marks the job
as `skipped` (not `failed`) so the caller knows the operation isn't available on
that host rather than broken.

```go
// darwin.go
func (d *Darwin) List(
    _ context.Context,
) ([]Entry, error) {
    return nil, fmt.Errorf("cron: %w", provider.ErrUnsupported)
}
```

### Provider Naming Conventions

There are three provider implementation patterns. The naming convention
determines the struct name, constructor, and file layout.

**1. Platform-specific providers** (most common)

One struct per OS family, each in its own file. Constructor names follow
`New{Platform}Provider()`. Methods that are large or testable go in separate
files named `{platform}_{operation}.go`.

| Struct   | Constructor              | File(s)                        |
| -------- | ------------------------ | ------------------------------ |
| `Debian` | `NewDebianProvider(...)` | `debian.go`, `debian_get_*.go` |
| `Darwin` | `NewDarwinProvider(...)` | `darwin.go`, `darwin_get_*.go` |
| `Linux`  | `NewLinuxProvider()`     | `linux.go`, `linux_get_*.go`   |

Most providers under `node/` and `network/` follow this pattern.

**2. Container-aware platform providers**

When a provider's behavior differs inside a Docker container (e.g., hostname is
read-only, DNS uses `/etc/resolv.conf` instead of `resolvectl`), add a
`DebianDocker` variant alongside the regular `Debian` struct. The agent selects
it via `platform.IsContainer()`.

| Struct         | Constructor                    | File(s)                                  |
| -------------- | ------------------------------ | ---------------------------------------- |
| `DebianDocker` | `NewDebianDockerProvider(...)` | `debian_docker.go`, `debian_docker_*.go` |

`DebianDocker` either embeds `Debian` (delegating reads, overriding writes) or
stands alone. It satisfies the same `Provider` interface.

```go
// agent_setup.go wiring
case "debian":
    if platform.IsContainer() {
        hostProvider = nodeHost.NewDebianDockerProvider()
    } else {
        hostProvider = nodeHost.NewDebianProvider(execManager)
    }
```

Examples: `node/host` (embeds `Debian`, blocks `UpdateHostname`), `network/dns`
(standalone, reads `/etc/resolv.conf` directly).

**3. SDK-based providers** (no platform variants)

Providers that talk to an external API (not the OS) use a single `Client` struct
with `New()` / `NewWithClient()` constructors. No `debian.go` / `darwin.go` /
`linux.go` files — the provider works the same on all platforms. Availability is
checked at startup (e.g., Docker daemon ping).

| Struct   | Constructor        | File(s)              |
| -------- | ------------------ | -------------------- |
| `Client` | `New()`            | `docker.go`          |
|          | `NewWithClient(c)` | (same file, testing) |

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

### FactsAware

Embed `provider.FactsAware` in the provider struct to access agent facts (OS
family, architecture, hostname, network interfaces) at runtime. The agent wires
facts via `provider.WireProviderFacts()`.

```go
type Debian struct {
    provider.FactsAware
    logger *slog.Logger
    fs     avfs.VFS
}
```

Facts are available in template rendering via `{{ .Facts.os_family }}` when
using the file provider's template support.

### Agent Wiring

Two files connect a provider to the agent:

1. **`internal/agent/processor_{domain}.go`** — create helper functions that
   dispatch sub-operations to the provider. If the domain gets its own category
   (like `schedule`, `docker`), create a `NewXxxProcessor` factory. If the
   domain belongs under an existing category (like `node`), add a `case` to that
   category's processor and delegate to helpers in a new file:

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

2. **`cmd/agent_setup.go`** — create the provider and register it with the
   `ProviderRegistry`. For new categories, use a separate `Register` call. For
   existing categories (e.g., `node`), pass the provider to the existing
   processor factory:

   ```go
   // New category example (like schedule, docker):
   registry.Register("mydomain",
       agent.NewMyDomainProcessor(myProv, log),
       myProv)

   // Existing category example (node):
   // Add your provider as a parameter to NewNodeProcessor
   // and include it in the providers list for FactsAware wiring.
   // Read cmd/agent_setup.go to see the current parameter list.
   ```

That's it. No changes to `agent/types.go`, `agent/agent.go`, or the `JobClient`
interface. The registry handles dispatch and FactsAware wiring automatically.

### Provider Testing

Testing conventions are in the root `CONTRIBUTING.md` under "Testing". The
interfaces a provider typically mocks are `FileDeployer`, `KeyValue`, and
`ObjectStore`; their generated mocks live in `{package}/mocks/`.

## Step 1: OpenAPI Spec + Code Generation

For node-targeted domains, create `internal/controller/api/node/{domain}/gen/`
with three hand-written files. For controller-only domains, create
`internal/controller/api/{domain}/gen/` instead:

- `api.yaml` — OpenAPI spec with paths, schemas, and `BearerAuth` security
- `cfg.yaml` — oapi-codegen config (`strict-server: true`, import-mapping for
  `common/gen`)
- `generate.go` — `//go:generate` directive

### HTTP Verb Conventions

Mutable domains MUST use separate verbs for create and update:

- `POST` — create a new resource (key/name in request body)
- `PUT` — update an existing resource (key/name from path parameter)
- `GET` — read/list resources
- `DELETE` — remove a resource

Do NOT combine create and update into a single "set" or "upsert" endpoint. The
cron domain is the reference: `POST` creates, `PUT /{name}` updates. This
separation gives clear 404 semantics (update fails if not found, create fails if
already exists) and matches REST conventions.

### Validation in OpenAPI Specs

The OpenAPI spec is the **source of truth** for input validation. All user input
must be validated, and the spec must declare how:

- **Request body properties**: Add `x-oapi-codegen-extra-tags` with `validate:`
  tags. These generate Go struct tags that `validation.Struct()` enforces at
  runtime.
  ```yaml
  properties:
    address:
      type: string
      x-oapi-codegen-extra-tags:
        validate: required,ip
  ```
- **Path parameters (UUID)**: Use `format: uuid` on the schema. This causes
  oapi-codegen to generate `openapi_types.UUID` type, and the router validates
  the format before the handler runs. No manual validation needed in the
  handler.
  ```yaml
  parameters:
    - name: id
      in: path
      required: true
      schema:
        type: string
        format: uuid
  ```
- **Query parameters**: Place `x-oapi-codegen-extra-tags` at the **parameter
  level** (sibling of `name`/`in`/`schema`), NOT inside `schema:`. At parameter
  level, oapi-codegen generates `validate:` tags on the `*Params` struct fields.
  Use `enum` for constrained string values (generates `oneof` validation).
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
  **NOTE:** `x-oapi-codegen-extra-tags` on **path parameters** does NOT generate
  tags on `RequestObject` structs in **strict-server mode** (upstream limitation
  — see oapi-codegen issue). Keep the `x-oapi-codegen-extra-tags` in the spec
  for documentation and add a YAML comment noting validation is handled
  manually. Path params that need validation beyond `format: uuid` (e.g.,
  `valid_target`) use a shared helper like `node.validateHostname()` which calls
  `validation.Var()`.

**IMPORTANT — every endpoint with user input MUST have:**

1. `x-oapi-codegen-extra-tags` with `validate:` tags on all request body
   properties and query params in the OpenAPI spec
2. `validation.Struct(request.Params)` in the handler for query params,
   `validation.Struct(request.Body)` for request bodies
3. A `400` response defined in the OpenAPI spec for endpoints that accept user
   input
4. HTTP wiring tests (`TestXxxHTTP` / `TestXxxRBACHTTP` methods in the
   `*_public_test.go` suite) that send raw HTTP through the full Echo middleware
   stack and verify:
   - Validation errors return correct status codes and error messages
   - RBAC: 401 (no token), 403 (wrong permissions), 200 (valid token)

**Defense-in-depth validation**: When validation calls cannot currently fail
(e.g., all fields use `omitempty`), keep the call but add a comment explaining
why. This guards against future field additions breaking validation silently:

```go
// Defense in depth: current fields use omitempty so validation
// always passes, but guards against future field additions.
if errMsg, ok := validation.Struct(request.Body); !ok {
    return gen.PostFoo400JSONResponse{Error: &errMsg}, nil
}
```

## Step 2: Handler Implementation

For node-targeted domains, create `internal/controller/api/node/{domain}/`. For
controller-only domains, create `internal/controller/api/{domain}/`:

- `types.go` — domain struct, dependency interfaces (e.g., `Checker`)
- `{domain}.go` — `New()` factory, compile-time interface check:
  `var _ gen.StrictServerInterface = (*Domain)(nil)`
- One file per endpoint (e.g., `{operation}_get.go`). Every handler that accepts
  user input MUST call `validation.Struct()` and return a 400 on failure.
- Tests: `{operation}_get_public_test.go` (testify/suite, table-driven). Must
  cover validation failures (400), success, and error paths. Each public test
  suite also includes HTTP wiring methods:
  - `TestXxxHTTP` — sends raw HTTP through the full Echo middleware stack to
    verify validation (valid input, invalid input → 400).
  - `TestXxxRBACHTTP` — verifies auth middleware: no token (401), wrong
    permissions (403), valid token (200). Uses `api.New()` +
    `{domain}.Handler()` + `server.RegisterHandlers()` to wire through
    `ScopeMiddleware`. Follow existing handler test files in the codebase.

### Broadcast Support (MANDATORY for node-targeted operations)

Every operation under `/node/{hostname}/...` MUST support broadcast targeting
(`_all`, `_any`, hostname, label selectors). The handler checks
`job.IsBroadcastTarget(hostname)` and routes to a broadcast function. Both
single-target and broadcast paths return the same collection response shape.

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

Every result item MUST have `hostname` and `error` fields. Single-target returns
1 result; broadcast returns N results. Failed/skipped agents appear as entries
with `error` set.

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

**Job client** — the `JobClient` interface has 4 generic methods: `Query`,
`QueryBroadcast`, `Modify`, `ModifyBroadcast`. Handlers call these with a
category string and operation constant. No new methods are needed when adding
operations. Example:

```go
jobID, resp, err := s.JobClient.Modify(
    ctx, hostname, "node", job.OperationSysctlCreate, data)
```

Read existing handlers in the codebase for reference.

## Step 3: Handler Registration

Each domain package exports a `Handler()` function that creates the handler,
wraps it with auth middleware, and returns route registration closures. No
changes to the `Server` struct are needed.

Create `handler.go` in your domain package:

```go
// internal/controller/api/node/{domain}/handler.go
package {domain}

func Handler(
    logger *slog.Logger,
    jobClient client.JobClient,
    signingKey string,
    customRoles map[string][]string,
) []func(e *echo.Echo) {
    var tokenManager api.TokenValidator = authtoken.New(logger)
    h := New(logger, jobClient)
    strictHandler := gen.NewStrictHandler(h,
        []gen.StrictMiddlewareFunc{
            func(handler strictecho.StrictEchoHandlerFunc,
                _ string,
            ) strictecho.StrictEchoHandlerFunc {
                return api.ScopeMiddleware(
                    handler, tokenManager, signingKey,
                    gen.BearerAuthScopes, customRoles,
                )
            },
        },
    )
    return []func(e *echo.Echo){
        func(e *echo.Echo) {
            gen.RegisterHandlers(e, strictHandler)
        },
    }
}
```

Add a `handler_public_test.go` that tests route registration and middleware
execution. Follow existing domain handler tests.

## Step 4: Startup Wiring

- `cmd/controller_setup.go` — add one line to `registerControllerHandlers`:
  ```go
  handlers = append(handlers,
      {domain}API.Handler(log, jc, signingKey, customRoles)...)
  ```
  Add the import for your domain package.

## Step 5: Update SDK

The SDK client library lives in `pkg/sdk/client/`. Its generated HTTP client
uses the same combined OpenAPI spec as the server
(`internal/controller/api/gen/api.yaml`). Follow the rules in the
[SDK Development Guidelines](../sdk/guidelines.md) — especially: never expose
`gen` types in public method signatures, add JSON tags to all result types, and
wrap errors with context.

**When modifying existing API specs:**

1. Make changes to the domain's `gen/api.yaml` (under `api/node/{domain}/` for
   node-targeted domains or `api/{domain}/` for controller-only domains)
2. Run `just generate` to regenerate server code (this also regenerates the
   combined spec via `redocly join`)
3. Run `go generate ./pkg/sdk/client/gen/...` to regenerate the SDK client
4. Update the SDK service wrappers in `pkg/sdk/client/{domain}.go` if new
   response codes were added
5. Update CLI switch blocks in `cmd/` if new response codes were added

**When adding a new API domain:**

1. Add a service with four files in `pkg/sdk/client/`:
   - `{service}.go` — `{Service}Service` struct + methods
   - `{service}_types.go` — SDK result types + gen→SDK conversions
   - `{service}_public_test.go` — service method tests
   - `{service}_types_public_test.go` — conversion function tests Each service
     gets its own files — do NOT add methods or types to an existing service's
     files.
2. Add a field to the `Client` struct in `osapi.go` and wire it in `New()`
3. Run `go generate ./pkg/sdk/client/gen/...` to pick up the new domain's spec
   from the combined `api.yaml`
4. Add an SDK example in `examples/sdk/client/{service}.go` — one file per SDK
   service (e.g., `hostname.go`, `disk.go`, `ntp.go`). The example file name
   matches the Client field name in lowercase.
5. Add an SDK doc page under the appropriate category subdirectory in
   `docs/docs/sidebar/sdk/client/`. SDK docs are grouped by concern (e.g.,
   `node-info/`, `system-config/`, `operations/`). Place the new page in the
   matching group — look at the existing directory structure to find the right
   one. Use the Client field name as the page title (e.g., `# Power`), NOT the
   Go struct name. Update `client.md` to add the service to its category table.
6. Add the new service to the SDK navbar dropdown in `docs/docusaurus.config.ts`
   under the matching category header. The dropdown is grouped the same way as
   the sidebar.

### SDK method naming

Method naming, type exposure, result-field tags, and error handling are
specified in the `sdk-standards` capability in
[osapi-io/specs](https://github.com/osapi-io/specs). When a convention here and
the specification disagree, the specification wins.

### SDK example conventions

SDK examples live in `examples/sdk/client/`, one file per SDK service. Follow
the same principles as the orchestrator examples:

- **One service per file**: demonstrate the service's SDK operations. Don't mix
  in other services.
- **Self-contained**: for read-only operations, just call and print. For
  mutating operations, cleanup at the start so the example is repeatable.
- **Print results**: decode and print at least one result so the example isn't
  silent.
- **Keep it short**: under ~100 lines of code (excluding license).
- **Handle errors inline**: use `log.Fatalf` for unexpected errors. For
  operations that may fail on some platforms, check the error and print a
  message instead of crashing.

## Step 6: CLI Commands

- `cmd/client_node_{domain}.go` — parent command registered under
  `clientNodeCmd` (for node-targeted domains)
- `cmd/client_node_{domain}_{operation}.go` — one subcommand per endpoint (e.g.,
  `client_node_sysctl_get.go`)
- All commands support `--json` for raw output
- Use `cli.PrintKV` for inline key-value output and `cli.PrintCompactTable` for
  multi-row tabular data (both in `internal/cli/ui.go`)
- Use flags (e.g., `--job-id`, `--audit-id`) instead of positional args for
  resource IDs
- Handle **all** API response codes in the `switch resp.StatusCode()` block:
  200, 400 (`handleUnknownError`), 401/403 (`handleAuthError`), 404
  (`handleUnknownError`), 500 (`handleUnknownError`). Match the responses
  declared in the OpenAPI spec.

## Step 7: Documentation

- `docs/docs/sidebar/features/{domain}-management.md` — feature page. Follow
  existing feature pages for the template.
- `docs/docs/sidebar/usage/cli/client/node/{domain}/{domain}.md` — CLI landing
  page with `<DocCardList />`
- `docs/docs/sidebar/usage/cli/client/node/{domain}/{verb}.md` — one page per
  CLI subcommand (e.g., `get.md`, `create.md`, `update.md`)
- Update `docs/docusaurus.config.ts`:
  - Add the new feature to the "Features" navbar dropdown
  - Add the new SDK service to the "SDK" → "Client Library" dropdown
- Update `docs/docs/sidebar/features/features.md` — add the new domain to the
  features landing page table
- Update `docs/docs/sidebar/usage/configuration.md` — add any new permissions to
  the roles table and permissions comments in the YAML reference
- Update `docs/docs/sidebar/features/authentication.md` — add new permissions to
  the roles/permissions tables
- Update `docs/docs/sidebar/architecture/architecture.md` — add link to the new
  feature page in the features list
- Update `docs/docs/sidebar/architecture/api-guidelines.md` — add new endpoints
  to the path pattern table
- Update `docs/docs/sidebar/architecture/system-architecture.md` — add endpoints
  to the health/endpoint tables if applicable

## Step 8: Verify

```bash
just generate        # regenerate specs + code
go build ./...       # compiles
just go-unit        # tests pass
just go-vet         # lint passes
```
