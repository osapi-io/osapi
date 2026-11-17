---
name: build-an-endpoint
description: Build or change an endpoint group, the typed definition under internal/controller/endpoints/<group>/ of operations that run on the osapi controller itself, such as the agents, the jobs, the audit log, the files, health, facts and telemetry, and not on a host. Use when asked to add or change a controller endpoint, to move a hand-written controller group under internal/controller/api to the endpoint DSL, when osapi controller endpoint validate, TestEndpointParity or TestJoinedSpec fails, or to decide whether an operation is an endpoint or a behavior. Not for work on a host, which the build-a-behavior skill covers.
---

# Build an endpoint

An endpoint is one operation of the controller itself, such as `GET /api/agent`
or `POST /api/job/{id}/retry`. It runs on the controller, in the request. It has
no target and no result for each host. A group is one directory,
`internal/controller/endpoints/<group>/`, plus one import in
`internal/controller/endpoints/all`. The DSL is in `internal/controller/endpoint`.

This skill gives a procedure and decisions. It is not a catalog. The binary tells
about the DSL:

- `controller endpoint spec` tells each part of the DSL, what it does and when
  to use it, and what the DSL cannot do yet.
- `controller endpoint validate --help` tells each rule, with its fix. Each
  failure also ends with its fix.
- `controller endpoint list` and `controller endpoint get` tell what each group
  has.

Ask these tools. Do not use your memory. If a tool and this skill do not agree,
the tool is correct.

The DSL is new. Today it makes the spec of a group, and two tests hold that spec
to the hand-written spec. The controller still serves the hand-written spec,
handler and command of each group. Some parts are not built yet, such as the
handlers. `controller endpoint spec` lists them under "Not supported yet". If the
work needs one of them, stop and say so. That is a platform change, and a person
decides it. Do not add code around it.

The design, with what is built and what comes next, is
`components/osapi/endpoints.md` in osapi-io/specs. Read it first.

## Endpoint or behavior

| The operation | It is |
| --- | --- |
| Runs on the controller: the registry, the job queue, the audit log, the object store, the trace backend | An endpoint |
| Changes or reads a host, through a job to the agent | A behavior: use the build-a-behavior skill |
| Spans hosts, or orders several steps | Neither: an osapi-orchestrator workflow |

A user of osapi does not add endpoints. They are part of the platform. Thus this
skill is for a person who changes osapi itself.

## Find what you need

Ask the code and the binary before you read prose:

| Question | Ask |
| --- | --- |
| Each part of the DSL, what it does and when to use it, and what is not built | `go run . controller endpoint spec` |
| Each rule a definition must pass, with its fix | `go run . controller endpoint validate --help` |
| Which groups exist, and their methods | `go run . controller endpoint list` |
| The routes, permissions and operationIds of a group | `go run . controller endpoint get --name <group>` |
| The spec that the definition makes | `go run . controller endpoint get --name <group> --openapi` |
| Whether a definition breaks a rule | `go run . controller endpoint validate` |
| A complete group | `internal/controller/endpoints/agent` |
| The hand-written spec of a group | `internal/controller/api/<group>/gen/api.yaml` |
| How two specs are compared | `go doc ./internal/apicontract` |

## Move a hand-written group

Today a move makes the spec only. Do the groups in the order of the design page.
`controller endpoint spec` has each option and tag that the steps name.

1. Read the hand-written spec of the group. Each schema becomes a Go type, and
   each operation becomes a method.
2. Write `types.go`. Give each type the name of its schema: the spec takes the
   component name from the type. Use the tags that `spec` lists under "Tags on
   input and output fields". A schema that the spec has in place, with no name,
   is a type alias of a struct with no name, such as `Message` in the agent
   group.
3. Write `<group>.go`. Copy `Title` and `Tag` from the hand-written spec. The
   UI's client files get their names from them. Give each method the
   operationId of the hand-written spec with `.ID`. The Go SDK and the UI client
   get their names from it. Use `.Sub` for a fixed segment and `.Action` for a
   segment after the key. Map each 404, 409 and 503 with `.Errors`, with the
   description of the hand-written spec.
4. Add the import to `internal/controller/endpoints/all/all.go`.
5. Add a row to `moving` in
   `internal/controller/endpoints/all/parity_public_test.go`.
6. Run the checks until they pass.

When a difference is a bug in the hand-written spec or its handler, do not copy
the bug into the definition. Tell the user, and file an issue. If the user
agrees, the row accepts the line, with the issue in its reason. Drain and undrain
were the first case (osapi#788).

## Decide first

Ask the user when the request does not tell you. Give your recommendation with
each question.

| Ask | Answer becomes |
| --- | --- |
| Is it work on the controller, or on a host? | This skill, or build-a-behavior |
| Is it a new operation, or one that moves? | A new one has no hand-written spec and no parity row; a move keeps every name |
| Which permission does it need? | The default `<group>:read` or `<group>:write`, or `.Permission`, such as `agent:read` for facts |
| Does a probe call it without a token? | `.Public()`, for the health probes only |
| Which errors can it return, and what does each mean to the caller? | `.Errors`: 404 when it is not there, 409 for a state that does not allow it, 503 when a backend is not set. A failure of the controller is a 500, never a 409 |
| Does a success answer 201 or 204? | `.Status` |

## Check it

```bash
go run . controller endpoint validate
go test ./internal/controller/endpoint/ ./internal/controller/endpoints/... ./internal/apicontract/
mise exec -- just generate   # when a hand-written spec changed
mise exec -- just test
```

| Failure names | It means |
| --- | --- |
| `controller endpoint validate`, or `TestAllPublicTestSuite/TestValidate` | A rule of the definition. The problem names the rule and ends with its fix. `validate --help` lists each rule |
| `TestEndpointParity` | A line of the contract differs. `-` is only in the hand-written spec, `+` only in the generated spec. Change the definition until they agree, or see the bug rule above |
| `TestJoinedSpec` | The joined spec changed. If you changed the API on purpose, write the contract again: `go test ./internal/apicontract -run TestContractPublicTestSuite/TestJoinedSpec -update`, and keep the diff. If you did not, find what changed it |
| `the spec has no term for` | A new option of the DSL has no entry in `internal/controller/endpoint/spec.go`. Add one |

## Not yet

`controller endpoint spec` lists what the DSL cannot do yet, under "Not
supported yet". If a request needs one, say so and stop. Do not write a handler
for a group, and do not wire a group into the controller by hand.
