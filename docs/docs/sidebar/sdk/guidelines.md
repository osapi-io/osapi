---
sidebar_position: 1
---

# SDK Development Guidelines

How to develop the OSAPI Go SDK (`pkg/sdk/`), with worked examples. Applies to
the client library and any new SDK packages.

:::note

This page shows the rules in use. The
[SDK document](https://github.com/osapi-io/specs/blob/main/components/osapi/sdk.md)
states them. It gives the five naming rules that come from the 117 methods, and
what a service must do. It also gives the package layout, and why each method
returns `Response[T]`.

Method naming was stated nowhere for a long time, and earlier versions of this
page deferred to an `sdk-standards` capability that was never written. Both are
fixed: the rules exist now, and the document names the seven existing methods
that break them.

:::

## Never Expose Generated Types

The `gen/` package contains auto-generated OpenAPI client code. **No generated
type is in the signature of a public SDK method.** The purpose of the SDK is to
hide `gen/` behind clean, stable types.

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
to nil pointers where the gen type uses `*string`, `*bool`.

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

- `StructToMap` (the bridge helper) converts structs through JSON to
  `map[string]any`. Without tags, Go uses PascalCase field names, which are
  different from the snake_case keys of the API.
- Consumers can serialize SDK types to JSON, to log, store or send them.
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

Use `Collection.First()` to get a single result safely. Do not use `Results[0]`
directly.

### Changed Field

Every mutation result type must include `Changed bool`. The provider sets it,
the agent extracts it through `extractChanged()`, the API passes it through, and
the SDK exposes it. The full chain must be consistent.

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

To add a new domain service to the SDK client:

1. **Create `{domain}.go`**: the service struct and its methods. Each method
   calls the gen client and converts the result to SDK types.
1. **Create `{domain}_types.go`**: the SDK result types with JSON tags, the SDK
   request types (around the gen types), and the gen→SDK conversion functions.
1. **Create `{domain}_public_test.go`**: tests with `httptest.Server` mocks, at
   100% coverage.
1. **Connect it in `osapi.go`**: add the service field to `Client`, and
   initialize it in `New()`.
1. **Never import `gen` in examples or consumer code.** If a consumer must
   import `gen`, the SDK wrapper is not complete.

## Testing

- Use `httptest.Server` to mock API responses
- Test all HTTP status code paths (200, 400, 401, 403, 404, 500)
- Test nil response body path
- Test transport errors (unreachable server)
- Test all optional field branches in request type mapping
- Target 100% coverage on all SDK packages (excluding `gen/`)

## Consumer Guidance

SDK consumers, for example `osapi-orchestrator`, must do these things:

- Use SDK `client.*` types directly. do not redefine them locally
- Use `Collection.First()` instead of `Results[0]`
- Never import `gen`. if you need to, the SDK is missing a wrapper
- Never panic on SDK responses. always propagate errors
