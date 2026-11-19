# API layer

Four things: the OpenAPI spec, the handlers, the registration, and one line of
startup wiring.

Node-targeted domains live under `internal/controller/api/node/{domain}/`,
controller-only domains directly under `internal/controller/api/{domain}/`.

## The rules are in the corpus, not here

[Building a domain](https://github.com/osapi-io/specs/blob/main/components/osapi/domains.md) states each rule
that this layer obeys. Read it before you write an endpoint. This file has only
what that page does not have: where the files go, their names, and the scaffold
to start from.

The split is deliberate. A second statement of a rule here becomes different
from the rule in the corpus. Then the copy that an agent loads first applies.
[The constitution](https://github.com/osapi-io/specs/blob/main/CONSTITUTION.md) exists to stop that failure.

| What you need to know | Where |
| --- | --- |
| Node-targeted or controller-only, and which shape applies | [node-targeted or controller-only](https://github.com/osapi-io/specs/blob/main/components/osapi/domains.md#node-targeted-or-controller-only) |
| The spec is the source of truth for validation, and the three places a tag goes | [validation](https://github.com/osapi-io/specs/blob/main/components/osapi/domains.md#validation) |
| Path parameters are the trap, and what actually validates one | [validation](https://github.com/osapi-io/specs/blob/main/components/osapi/domains.md#validation) |
| Separate verbs for create and update, and why a combined upsert is forbidden | [verbs](https://github.com/osapi-io/specs/blob/main/components/osapi/domains.md#verbs) |
| The six API design guidelines, including path versus query parameters | [design guidelines](https://github.com/osapi-io/specs/blob/main/components/osapi/domains.md#design-guidelines) |
| Broadcast is mandatory, and both paths return the same collection shape | [broadcast is not optional](https://github.com/osapi-io/specs/blob/main/components/osapi/domains.md#broadcast-is-not-optional) |
| What `{hostname}` accepts, a literal, `_any`, `_all`, a label selector | [routing and targeting](https://github.com/osapi-io/specs/blob/main/components/osapi/job-system.md#routing-and-targeting) |
| The job client has four methods, and an operation adds none | [the job client](https://github.com/osapi-io/specs/blob/main/components/osapi/domains.md#the-job-client-needs-nothing-added) |
| `Handler()` returns route closures, and the `Server` struct does not change | [wiring](https://github.com/osapi-io/specs/blob/main/components/osapi/domains.md#wiring) |
| The permission a new endpoint needs, and where it is declared | [adding a permission](https://github.com/osapi-io/specs/blob/main/components/osapi/permissions.md#adding-a-permission) |
| What the tests owe: validation proved to fire, the RBAC trio, every declared status | [what a domain's tests owe](https://github.com/osapi-io/specs/blob/main/components/osapi/domains.md#what-a-domains-tests-owe) |
| A domain appears everywhere an existing domain appears | [every layer, or not done](https://github.com/osapi-io/specs/blob/main/components/osapi/domains.md#a-domain-is-in-every-layer-or-it-is-not-done) |

The job system decides what a caller sees when an agent does not answer. This
layer does not. Refer to [what a caller gets
back](https://github.com/osapi-io/specs/blob/main/components/osapi/job-system.md#what-a-caller-gets-back) for
the four statuses of each host, and to [three
limits](https://github.com/osapi-io/specs/blob/main/components/osapi/job-system.md#three-limits-bounding-different-things)
for the clocks. Domain code does not handle the timeout. CLI and SDK output must
not say that the operation ran.

Read the package of the reference domain together with the specification. The
specification says what must be true. `sysctl` and `cron` show it in use.

## File layout

```
internal/controller/api/node/{domain}/
  types.go                    domain struct, dependency interfaces
  {domain}.go                 New(), var _ gen.StrictServerInterface = (*Domain)(nil)
  validate.go                 path-parameter validators
  handler.go                  Handler(): construct, wrap in auth, return routes
  {operation}_{verb}.go       one file per endpoint
  {operation}_{verb}_public_test.go
  gen/
    api.yaml                  paths, schemas, BearerAuth, the permission each endpoint needs
    cfg.yaml                  oapi-codegen config, strict-server: true
    generate.go               the //go:generate directive
```

Everything else in `gen/` is generated. `mise exec -- just generate` regenerates
the server, joins the combined spec at `internal/controller/api/gen/api.yaml`
with redocly, and regenerates the SDK client and the API doc pages.

## Scaffolding

A handler validates, then delegates to the job client. It never touches the
operating system.

```go
jobID, resp, err := s.JobClient.Modify(
    ctx, hostname, "node", job.OperationSysctlCreate, data)
```

```go
if job.IsBroadcastTarget(hostname) {
    return s.postOperationBroadcast(ctx, hostname, entry)
}
// single target: one result in the same collection envelope
```

Registration and startup: copy the reference domain's `handler.go`, change the
package and type names, and add one line to `handlers` in
`internal/controller/handlers.go` with its import:

```go
handlers = append(handlers,
    {domain}API.Handler(d.log, d.jobs, signingKey, customRoles)...)
```

## Three rules the corpus does not yet hold

Stated here because they are real and nothing else states them. Both belong in
the corpus and neither is there yet.

**A custom validation rule is a registered validator.** It goes in
`internal/validation`, with a hint in `customHints`. Thus the 400 says what
shape was expected, and does not only name the tag. `sysctl_key` and
`cron_schedule` are the
pattern.

**`IsBroadcastTarget` has one implementation and never a second.** It lives in
`internal/job/subjects.go`. What the corpus does not say is that a domain must
not write its own target parser.

**Select a permission by the possible damage, not by the endpoint group.** If
two operations can do different amounts of damage, they need two permissions,
also if their shape is the same. Add a new permission to the permission
constants, `builtinPermissions`, `builtinRolePermissions` and the SDK.
`osapi token roles list` then prints it. Also add it to the roles table in
`usage/configuration.md`. Refer to [docs.md](docs.md). A permission that is in
the spec but in no role gets to nobody.

## Tests

Testing conventions are osapi's `CONTRIBUTING.md`, under "Testing". What this
layer adds to a public suite:

- `TestXxxHTTP`, raw HTTP through the full Echo middleware stack: valid input
  succeeds, invalid input returns 400 with the message.
- `TestXxxRBACHTTP`, no token is 401, a token without the permission is 403, a
  token with it succeeds.

Cover validation failure, success, a provider error surfaced from the job, and
the broadcast path. Where a status code is declared in the spec, the handler is
expected to be able to return it.
