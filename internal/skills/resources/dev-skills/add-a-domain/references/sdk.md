# SDK service

`pkg/sdk/client` is the public Go API. Its generated client comes from the same
combined spec as the server, so `just generate` covers it.

## The rules are in the corpus, not here

| What you need to know | Where |
| --- | --- |
| What a service owes, and the four files per service | [what a service owes](https://github.com/osapi-io/specs/blob/main/components/osapi/sdk.md#what-a-service-owes) |
| Method naming, and the seven methods that break it | [method naming](https://github.com/osapi-io/specs/blob/main/components/osapi/sdk.md#method-naming) |
| How the package is laid out | [package layout](https://github.com/osapi-io/specs/blob/main/components/osapi/sdk.md#how-the-package-is-laid-out) |
| No `gen` type in a public signature, and the envelope every method returns | [the envelope](https://github.com/osapi-io/specs/blob/main/components/osapi/sdk.md#every-method-returns-the-same-envelope) |
| JSON tags, wrapped errors, one service per file | [the rest of the conventions](https://github.com/osapi-io/specs/blob/main/components/osapi/sdk.md#the-rest-of-the-conventions) |
| A domain appears everywhere an existing domain appears | [every layer, or not done](https://github.com/osapi-io/specs/blob/main/components/osapi/domains.md#a-domain-is-in-every-layer-or-it-is-not-done) |

There is no SDK standard across repositories. The conventions in the corpus and
below apply, because you can check them against `pkg/sdk/client/`. If somebody
wants a shared standard with `osapi-orchestrator`, that is a separate feature.
If you invent one in a reference file, you make a rule that nobody agreed to.

## Files, one service per domain

```
pkg/sdk/client/
  {domain}.go                     {Domain}Service struct and methods
  {domain}_types.go               result types, request types, gen -> SDK conversion
  {domain}_public_test.go         method tests
  {domain}_types_public_test.go   conversion tests
  osapi.go                        Client field, wired in New()
examples/sdk/client/{domain}.go   one runnable example
```

Never add methods or types to another service's files.

## Generated types stay inside

No `gen.*` type appears in a public signature. That is the whole reason the SDK
exists.

```go
// Wrong: the consumer now imports gen
func (s *DockerService) Create(
    ctx context.Context,
    hostname string,
    body gen.DockerCreateRequest,
) (*Response[Collection[DockerResult]], error)

// Right: an SDK type, converted internally
func (s *DockerService) Create(
    ctx context.Context,
    hostname string,
    opts DockerCreateOpts,
) (*Response[Collection[DockerResult]], error)
```

Build the `gen` request inside the method, mapping zero values to nil where the
generated type uses a pointer.

If a consumer has to import `gen`, the wrapper is incomplete.

## Result types

Every exported field carries a snake_case `json` tag. `StructToMap` converts by
JSON round-trip, so an untagged field arrives as PascalCase and matches nothing.

```go
type HostnameResult struct {
    Hostname string            `json:"hostname"`
    Error    string            `json:"error,omitempty"`
    Changed  bool              `json:"changed"`
    Labels   map[string]string `json:"labels,omitempty"`
}
```

`omitempty` on pointers, optional slices and maps, error strings and optional
strings. Not on `Changed`, and not on a required field. A mutation result must
say `"changed": false`. It must not leave it out.

`Changed` travels provider to agent to API to SDK. Breaking the chain anywhere
makes every layer above it lie.

Operations with more than one target return `Collection[T]`, which has `Results`
and `JobID`. Use `Collection.First()`. Do not use an index.

## Responses and errors

Methods return `*Response[T]`: `Data` for the typed result, `RawJSON()` for the
CLI's `--json` mode.

```go
if err := checkError(
    resp.StatusCode(),
    resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON500,
); err != nil {
    return nil, fmt.Errorf("{domain} create: %w", err)
}

if resp.JSON200 == nil {
    return nil, &UnexpectedStatusError{APIError{
        StatusCode: resp.StatusCode(),
        Message:    "nil response body",
    }}
}
```

Each status that the spec declares gets a branch. Wrap with context at the
boundary, so that `errors.Is` and `errors.As` still work above it. Guard the nil
body after `checkError`, because a 200 with no body can occur.

## Examples

One file per service in `examples/sdk/client/`, named for the `Client` field in
lowercase.

- One service per file, nothing mixed in.
- Complete alone: read-only examples call and print. Examples that change things
  remove their old results first, so that they can run two times.
- Print at least one result, so the example is not silent.
- Approximately 100 lines or less, without the licence header.
- `log.Fatalf` for unexpected errors. For an operation that the host may not
  support, print the message. Do not crash.
- Never import `gen`.

`examples/` has its own module. Thus a new example needs `just go-mod`, and a
commit of the tidy result. `just test` checks it.

## Tests

- `httptest.Server` for the API.
- Every declared status path: 200, 400, 401, 403, 404, 500.
- The nil-body path and a transport error against an unreachable server.
- Every optional field branch in request conversion.
- 100% coverage on SDK packages, without `gen/`.
