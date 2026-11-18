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

The DSL makes the spec, the routes, the decoders and the CLI commands of a
group. A group's handlers are typed functions. Each controller group is served,
and its commands are made, from its definition. If the work needs a part that
`controller endpoint spec` does not list, stop and say so. That is a platform
change, and a person decides it. Do not add code around it.

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

## Build a new group

1. Write it: `go run . controller endpoint new --name <group> --description "<one line>"`,
   then `go generate ./internal/controller/endpoints/all` and
   `mise exec -- just generate`. The group has a list and a get, an empty
   `Deps`, its types and the tests of its methods. It builds, validates and
   passes its tests before you change it.
2. Change it in this order: `types.go`, `deps.go`, `methods.go`, then
   `<group>.go`. Decide each method with the user first; "Decide first" below
   has the questions.
3. When `Deps` gets its first field, give it in
   `internal/controller/handlers.go`, as `agentDeps` does. While it is empty,
   the controller needs no wiring.
4. In `cmd/client_<group>.go`, the file of the group's command, add a `command.Render` for each method
   a person reads at a shell. A method without one prints its answer as JSON.
5. A new operation changes the API: run `TestJoinedSpec` with `-update`, and
   keep the diff.

## Move a hand-written group

Do the groups in the order of the design page. A move is two pull requests: the
spec, held to the hand-written spec, and then the switch. `controller endpoint
spec` has each option and tag that the steps name. `internal/controller/endpoints/agent`
is the example of each step.

The spec:

1. Read the hand-written spec of the group. Each schema becomes a Go type, and
   each operation becomes a method.
2. Write `types.go`. Give each type the name of its schema: the spec takes the
   component name from the type. Use the tags that `spec` lists under "Tags on
   input and output fields", with `example` and `doc` from the hand-written
   spec. A schema that the spec has in place, with no name, is a type alias of a
   struct with no name, such as `Message`. Copy the type's description with
   `Doc`: the join fails when two specs define one schema in different words.
3. Write `<group>.go`. Copy `Title` and `Tag` from the hand-written spec. The
   UI's client files get their names from them. Give each method the
   operationId of the hand-written spec with `.ID`. Use `.Sub` for a fixed
   segment and `.Action` for a segment after the key. Map each 400 that the
   method decides, 404, 409 and 503 with `.Errors`. A status whose body is the
   method's output, such as a probe's 503, is `.Answers`, not an error.
4. Add the import to `internal/controller/endpoints/all/all.go`, and a row to
   `moving` in `internal/controller/endpoints/all/parity_public_test.go`.
5. Run the checks until they pass.

The switch:

1. Write `deps.go`: the `Deps` struct of interfaces that the handlers use, with
   a `mocks/` package. Add `endpoint.Uses[Deps]()`.
2. Write `methods.go`: one function for each method,
   `func(ctx, Deps, In) (Out, error)`, with `.Handle(fn)`. Move the logic of
   the hand-written handler. Answer 404 or 409 with `endpoint.Errorf`, so the
   message names what the method was called with. A failure of the controller
   is a plain error, which is a 500.
3. Move each case of the hand-written handler tests into
   `methods_public_test.go`, on mocks, or into `http_public_test.go`, through
   the generated routes. Auth, 400 and each status go there.
4. Run `go generate ./internal/controller/endpoints/all`. It writes `gen/`.
5. Give the group's `Deps` in `internal/controller/handlers.go`, as `agentDeps`
   does. A nil pointer must be a nil interface.
6. Delete the hand-written package under `internal/controller/api/<group>` and
   its parity row. Run `mise exec -- just generate`. `TestJoinedSpec` must pass
   without `-update`: the contract of the API does not change. Read the diff of
   the Go SDK and the UI client, which must not change their names.

Each input field has a `validate` tag, such as `validate:"required,min=1,max=255"`.
A field with no limit is `validate:"omitempty"`. The server answers 400 to a
value that breaks a rule, and the spec carries the rule and the whole tag.
`validate` refuses an input field with no tag. A custom rule, such as
`sysctl_key`, needs its schema from `rule.Register`, beside its validator in
`internal/validation`. Without one the spec does not build. The endpoint DSL
and the behavior DSL use the same rules; `spec` prints them under
`validate:"..."`.

Give each rule of each input field a `.Fails` input that breaks it: the
method's input, with the other fields valid and one value wrong, such as a name
of 256 characters for `max=255`. The API refuses each one with 400, and a test
sends each one through the routes. `validate` names a rule that has none.
`required` needs none, because the generated tests leave the field out. You
write the bad value, because a generator cannot know one for a custom rule.

The CLI:

1. Give each method `.Command(short, long)` with the help of the hand-written
   command, `.Example("--flag value")` with its examples, and `.Notes` and
   `.Result` with the `reference.Notes` and `reference.Result` annotations of
   the hand-written command, so its page keeps those sections. Each input field
   is a flag, named for its json name with `_` as `-`; a field of an
   `in:"body"` struct is a flag too. Keep the flag of the hand-written command
   with the `flag` and `short` tags, such as `-T/--target`.
   A `[]byte` field is a file. In the API it is a part of an `in:"form"`
   multipart body. In the CLI its flag is the path of a local file, and the
   command sends the content of that file, such as `file upload --file`.
   Each command is `<group> <resource> <verb>`, such as `telemetry trace get`:
   the verb is `list`, `get`, `create`, `update`, `delete`, or an action such
   as `retry`, never a noun. A verb that the group has as a method already
   gets `.Use`, such as `agent pending list`. A command that `.Sub` makes has
   a hand-written parent, so its help says what it holds.
2. In `cmd/client_<group>.go`, the file of the group's command, add a
   `command.Render` for each method, with the output code of the hand-written
   command. A command that acts on the answer, such as an export that writes a
   file, uses `command.Acts`, which runs with `--json` too. `--json` prints the
   API's answer otherwise, not a shape of its own.
3. Run `go generate ./internal/controller/endpoints/all`. It writes
   `gen/cli.go`.
4. Before you delete the hand-written commands, save their help:
   `go run . help client <group> <command> --json`. Compare it with the help of
   the generated commands: the same flags, types and required flags. Tell the
   user each difference.
   Run `mise exec -- just cli-docs-generate` and read the diff of each
   command's page under `docs/docs/sidebar/usage/cli/client/<group>/`: no
   section may go.
5. Delete `cmd/client_<group>_<method>.go` for each method. A command that is
   not a method of the group, such as `agent key`, stays.
6. Run `mise exec -- just cli-docs-generate`, and run each command against a
   controller.

When the hand-written spec leaves out what its route already does, fix the
hand-written spec in the spec pull request, so the two agree with no accepted
line. Audit's get answered 400 for an ID that is not a UUID, and its spec did
not list the 400. Run `mise exec -- just generate`, map the new status in the
SDK service, and run `TestJoinedSpec` with `-update`. The diff of the contract
is that status only.

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
| Which errors can it return, and what does each mean to the caller? | `.Errors`: 400 (`Invalid`) for a request it cannot do, 404 when it is not there, 409 for a state that does not allow it, 503 when a backend is not set. A failure of the controller is a 500, never a 409. A status with the output as its body, such as a probe's 503, is `.Answers` |
| Does a success answer 201 or 204? | `.Status`: 201 when it creates a resource, with `.Location` naming the output field of its ID; 204 when it has nothing to return, such as a delete or an action. Never 202: each method answers when its work is done |
| Is it a list? | Its output is `items` and `total_items` |

The conventions of the API are in `components/osapi/api.md` in osapi-io/specs.
`TestConventions` in `internal/apicontract` fails on an operation that breaks one.

## Check it

```bash
go run . controller endpoint validate
go generate ./internal/controller/endpoints/all
mise exec -- just generate   # joins gen/api.yaml into the main spec: SDK, UI and API docs
go test ./internal/controller/endpoint/... ./internal/controller/endpoints/... ./internal/apicontract/
mise exec -- just test
```

| Failure names | It means |
| --- | --- |
| `controller endpoint validate`, or `TestAllPublicTestSuite/TestValidate` | A rule of the definition. The problem names the rule and ends with its fix. `validate --help` lists each rule |
| `go-generate-check` | A `gen/` file is stale: run `go generate ./internal/controller/endpoints/all` |
| `Conflict on components => schemas` from `just generate` | Two specs define one schema in different words: copy the description with `Doc` and each `example` from the other spec |
| `endpoint: no ... given to the server` at startup | The controller does not give the group's `Deps`: add it in `internal/controller/handlers.go` |
| `is both a hand-written command and the <group> group's` at startup | Delete the hand-written command of that method |
| `endpoint-scaffold-check` | A group that `controller endpoint new` writes no longer passes its tests: change the scaffold in `internal/controller/endpoint/scaffold_files.go`, and run its golden test with `-update` |
| `TestEndpointParity` | A line of the contract differs. `-` is only in the hand-written spec, `+` only in the generated spec. Change the definition until they agree, or see the bug rule above |
| `TestJoinedSpec` | The joined spec changed. If you changed the API on purpose, write the contract again: `go test ./internal/apicontract -run TestContractPublicTestSuite/TestJoinedSpec -update`, and keep the diff. If you did not, find what changed it |
| `the spec has no term for` | A new option of the DSL has no entry in `internal/controller/endpoint/spec.go`. Add one |

Never edit a path that `controller endpoint spec` lists under "Generated, never
edited". Change the definition and generate. Do not write CLI docs for an
endpoint.

## What the DSL does not do

`controller endpoint spec` lists each part of the DSL. If a request needs a part
that is not in the list, say so and stop. Do not write a handler for a group,
and do not wire a group into the controller by hand.
