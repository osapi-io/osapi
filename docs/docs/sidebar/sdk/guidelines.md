---
sidebar_position: 1
---

# SDK Development Guidelines

Rules for developing the OSAPI Go SDK (`pkg/sdk/`). These apply to the client
library and any new SDK packages.

## Package Structure

```
pkg/sdk/
  client/          # HTTP client wrapping generated OpenAPI code
    gen/           # Generated code (DO NOT edit manually)
    osapi.go       # Client constructor, service wiring
    response.go    # Response[T], Collection[T], error helpers
    errors.go      # Typed error hierarchy
    node.go        # Request types (ExecRequest, ShellRequest, etc.)
    node_types.go  # SDK result types + gen→SDK conversions
    hostname.go    # HostnameService methods
    disk.go        # DiskService methods
    dns.go         # DNSService methods
    command.go     # CommandService methods
    cron.go        # CronService methods
    ...            # One file per domain service
  platform/        # Platform detection utilities
```

The orchestrator engine previously lived here but has been moved to
[osapi-orchestrator][]'s `internal/engine/` package.

[osapi-orchestrator]: https://github.com/osapi-io/osapi-orchestrator

## Never Expose Generated Types

The `gen/` package contains auto-generated OpenAPI client code. **No generated
type should appear in any public SDK method signature.** The SDK exists
specifically to hide `gen/` behind clean, stable types.

For every `gen.*` request or response type used internally, define an SDK-level
equivalent:

```go
// BAD — leaks gen type into public API
func (s *DockerService) Create(
    ctx context.Context,
    hostname string,
    body gen.DockerCreateRequest,    // consumer must import gen
) (*Response[Collection[DockerResult]], error)

// GOOD — SDK-defined type wraps gen internally
func (s *DockerService) Create(
    ctx context.Context,
    hostname string,
    opts DockerCreateOpts,           // SDK type, no gen import needed
) (*Response[Collection[DockerResult]], error)
```

Inside the method, build the `gen.*` request from the SDK type. Map zero values
to nil pointers where the gen type uses `*string`, `*bool`, etc.

## Result Types

### JSON Tags Required

Every exported struct field on every result/model type **must** have a
`json:"..."` tag with a snake_case key:

```go
// GOOD
type HostnameResult struct {
    Hostname string            `json:"hostname"`
    Error    string            `json:"error,omitempty"`
    Changed  bool              `json:"changed"`
    Labels   map[string]string `json:"labels,omitempty"`
}
```

Tags are required because:

- `StructToMap` (the bridge helper) uses JSON round-tripping to convert structs
  to `map[string]any`. Without tags, Go uses PascalCase field names which don't
  match the API's snake_case keys.
- Consumers may serialize SDK types to JSON for logging, storage, or forwarding.
  Consistent keys matter.

### omitempty Rules

- **Use `omitempty`** on: pointer fields, optional slices/maps, error strings,
  optional string fields
- **Do not use `omitempty`** on: `Changed bool` (must always be present),
  required fields like `Hostname`

### Collection Pattern

Multi-target operations return `Collection[T]`:

```go
type Collection[T any] struct {
    Results []T    `json:"results"`
    JobID   string `json:"job_id"`
}
```

Use `Collection.First()` for safe access to single-result responses instead of
indexing `Results[0]` directly.

### Changed Field

Every mutation result type must include `Changed bool`. The provider sets it,
the agent extracts it via `extractChanged()`, the API passes it through, and the
SDK exposes it. The full chain must be consistent.

## Response Pattern

All service methods return `*Response[T]`:

```go
type Response[T any] struct {
    Data    T
    rawJSON []byte
}
```

- `Data` — the typed SDK result
- `RawJSON()` — the raw HTTP response body for CLI `--json` mode

## Error Handling

### checkError

All service methods use `checkError()` to convert HTTP status codes into typed
errors:

```go
if err := checkError(
    resp.StatusCode(),
    resp.JSON400,
    resp.JSON401,
    resp.JSON403,
    resp.JSON500,
); err != nil {
    return nil, err
}
```

### Error Wrapping

Wrap errors with context at the SDK boundary:

```go
// GOOD
return nil, fmt.Errorf("docker create: %w", err)
return nil, fmt.Errorf("invalid audit ID: %w", err)

// BAD — no context
return nil, err
```

### Nil Response Guard

After `checkError`, always guard against nil response bodies:

```go
if resp.JSON200 == nil {
    return nil, &UnexpectedStatusError{APIError{
        StatusCode: resp.StatusCode(),
        Message:    "nil response body",
    }}
}
```

## Adding a New Service

When adding a new domain service to the SDK client:

1. **Create `{domain}.go`** — service struct + methods, each calling gen client
   and converting to SDK types
2. **Create `{domain}_types.go`** — SDK result types with JSON tags, SDK request
   types (wrapping gen types), and gen→SDK conversion functions
3. **Create `{domain}_public_test.go`** — tests using `httptest.Server` mocks,
   100% coverage
4. **Wire in `osapi.go`** — add service field to `Client`, initialize in `New()`
5. **Never import `gen` in examples or consumer code** — if a consumer needs to
   import `gen`, the SDK wrapper is incomplete

## Testing

- Use `httptest.Server` to mock API responses
- Test all HTTP status code paths (200, 400, 401, 403, 404, 500)
- Test nil response body path
- Test transport errors (unreachable server)
- Test all optional field branches in request type mapping
- Target 100% coverage on all SDK packages (excluding `gen/`)

## Consumer Guidance

SDK consumers (like `osapi-orchestrator`) should:

- Use SDK `client.*` types directly — do not redefine them locally
- Use `Collection.First()` instead of `Results[0]`
- Never import `gen` — if you need to, the SDK is missing a wrapper
- Never panic on SDK responses — always propagate errors
