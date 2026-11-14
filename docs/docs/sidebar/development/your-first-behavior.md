---
sidebar_position: 5
---

# Your first behavior

In this tutorial, we will build a behavior: one that reads and sets a host's
message of the day, `/etc/motd`. By the end it will have an API route, CLI
commands, a docs page, telemetry, a setting and tests at every layer, and we
will have written the types, the provider and some examples. osapi builds the
rest from the definition, the tests included.

[Building a behavior](behaviors.md) is the reference for everything used here.
This page is the walk through it.

## What we will build

A `motd` behavior with two methods:

| Method   | Does                         | Served at                            |
| -------- | ---------------------------- | ------------------------------------ |
| `get`    | reads the message of the day | `GET /api/next/node/{hostname}/motd` |
| `update` | replaces it                  | `PUT /api/next/node/{hostname}/motd` |

and one setting, `behaviors.motd.path`, the file it manages.

## Before we start

We need a checkout of osapi with its tools, as [Contributing](contributing.md)
describes, and it should build and pass:

```bash
just deps
just test
```

With an agent, the whole tutorial is one request. In Claude Code, the
`build-a-behavior` skill in this repository carries these steps:

```text
Build a motd behavior that reads and sets /etc/motd.
```

The agent runs the commands below and checks its own work with the same tests.
Reading on is how you review what it wrote.

## Check what exists

Before writing anything, we ask the binary what it already has:

```bash
go run . behavior list
```

A behavior is the binary's own registry, so this list is always current. If
something close to `motd` were there, we would extend it rather than add a
second.

## Create it

```bash
go run . behavior new --name motd --description "The message shown at login"
```

That writes `internal/behaviors/motd/` and registers it with one import in
`internal/behaviors/all`. Nothing else outside the directory changes, now or
later.

```text
internal/behaviors/motd/
  behavior.go    behavior_public_test.go   the definition; validates it and runs the examples
  methods.go     methods_public_test.go    get and update, on a mocked provider
  provider.go                              the Provider interface
  debian.go      debian_public_test.go     the provider on Debian: keeps a value in a file
  settings.go                              behaviors.motd.path and its default
  types.go                                 the input and output types
  mocks/                                   the provider's double, generated
```

It is already a working behavior: a value kept in a file. Generate the mock and
the `gen/` directory, and notice that it passes before we change anything:

```bash
go generate ./internal/behaviors/motd/mocks ./internal/behaviors/all
go run . behavior validate
just test
```

`gen/` now holds a handler for each method, `api_get.go` and `api_update.go`,
the route table in `api.go`, the commands in `cli.go`, the spec in `api.yaml`,
and a test beside each Go file.

The rest of the tutorial turns that value into the message of the day.

## Point the setting at /etc/motd

Settings are the agent's standing configuration, and each gets a default, an
environment variable and an `agent start` flag:

```go title="internal/behaviors/motd/settings.go"
const defaultPath = "/etc/motd"

type Settings struct {
    Path string `mapstructure:"path" doc:"The file motd manages." validate:"required,startswith=/"`
}

var defaults = Settings{Path: defaultPath}
```

A value that changes with each call never goes here; that is the method's input.

## The types

The input and output types carry the whole contract in their tags: each input
field becomes a flag and a request field, and each output field a column and a
JSON field. The scaffold's `value` already fits, so we only describe it:

```go title="internal/behaviors/motd/types.go"
type UpdateInput struct {
    Value string `json:"value" validate:"required" doc:"The message, shown at login"`
}

type Entry struct {
    Value string `json:"value" doc:"The message"`
}

type Result struct {
    Changed bool `json:"changed" doc:"Whether the host changed"`
}
```

Notice `changed` on the write's output. Every write reports whether it changed
the host, which is what makes running it twice safe to reason about.

## The provider

The provider does the work. The scaffold's Debian provider already reads and
writes the file in the setting, through the agent's filesystem:

```go title="internal/behaviors/motd/debian.go"
func newDebian(h host.Host) (Provider, error) {
    return &debian{fs: h.FS, path: host.Settings[Settings](h).Path}, nil
}

func (d *debian) Set(_ context.Context, value string) (bool, error) {
    if current, err := d.fs.ReadFile(d.path); err == nil && string(current) == value {
        return false, nil
    }
    _ = d.fs.MkdirAll(filepath.Dir(d.path), 0o755)

    return true, d.fs.WriteFile(d.path, []byte(value), 0o644)
}
```

Notice that `Set` compares before it writes: setting the message it already has
changes nothing and says so. `host.Host` is what the agent hands every provider:
the platform, the filesystem, a command runner, the file state store and the
behavior's settings.

The definition says which platforms have a provider. On any other, the methods
report `skipped` rather than fail:

```go
behavior.Platform("debian", newDebian),
```

A provider declares every program it runs: `behavior.Sudo("program")` for one it
runs with sudo, which the agent's preflight checks, and
`behavior.Runs("program")` for one it runs without, which the agent warns about
when it is missing. The tests fail on a program run without its declaration.

## The examples

The examples are what the docs pages show, and they are also the provider's
tests. Each one says what the host looks like first and what the call changes:

```go title="internal/behaviors/motd/behavior.go"
behavior.Read("get", Get).
    Describe("Get the message of the day").
    Example("Before anything is set", GetInput{}, Entry{}).
    Example("The message", GetInput{}, Entry{Value: "Authorized use only."},
        behavior.After(UpdateInput{Value: "Authorized use only."}),
    ),
behavior.Write("update", Update).
    Describe("Replace the message of the day").
    Example("A new message", UpdateInput{Value: "Back at 5pm."}, Result{Changed: true},
        behavior.Wrote(defaultPath, "Back at 5pm."),
    ),
```

`After` runs another method first, here `update`, so `get` has something to
read. `Wrote` says the file holds exactly that text afterwards. A provider that
runs commands names them with `Ran`, and one that reads a command's output gives
it with `Prints`. What a method refuses is declared the same way, with
`.Fails("No message", UpdateInput{}, "value must not be empty")`, when the
provider checks for it.

## Generate and confirm

```bash
go generate ./internal/behaviors/all
go run . behavior validate
go run . behavior get --name motd
go run . config get --key behaviors.motd.path
```

`validate` checks the definition against every rule the platform relies on.
`get` shows the routes, permissions, platforms, types, the setting and the
telemetry, exactly as the binary registered them. `config get` shows the
setting's default, its environment variable `OSAPI_BEHAVIORS_MOTD_PATH` and its
`agent start` flag `--motd-path`.

## The tests we did not write

```bash
just test
```

We wrote no test. The scaffold's tests cover its own code:
`methods_public_test.go` calls each method on a mocked provider, and
`debian_public_test.go` runs the provider on an in-memory filesystem.
`behavior_public_test.go` validates the definition and has one line,
`behaviortest.Run(s.T(), motd.Behavior)`, that runs the examples and refusals
through the methods and the provider. The rest is generated from the definition:

| What                    | Checks                                                                                                                                                                                            |
| ----------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| The provider, on Debian | Each example on a fake host gives its output and its effects; `update` run twice reports `changed: false`; a failing call never panics                                                            |
| The API                 | `gen/api_get_public_test.go` and `gen/api_update_public_test.go`: each example through the real handler, the response matching the spec, broadcast, 401, 403, 500, and a missing `value` is a 400 |
| The CLI                 | `gen/cli_public_test.go`: each example's flags, given to the generated command, send the example's input                                                                                          |
| The live server         | `go test -tags integration` runs `get` against a real agent, and the writes too with `OSAPI_INTEGRATION_WRITE_BEHAVIORS=1`; `just behavior-int-linux` runs both on Debian in a container          |

To see them for `motd` alone:

```bash
go test ./internal/behaviors/motd/... -v
```

A provider test of our own is only for what an example cannot show, such as a
parser's edge cases.

## Run it

Start osapi with a token, as [Getting Started](../intro.md) shows, then:

```bash
osapi client next node motd update --value "Back at 5pm."
osapi client next node motd get
osapi client next node motd update --value "Back at 5pm."
```

Notice the second `update` reports no change: the host already has that message.

## What we built

One directory, one import line, and a behavior with:

- `GET` and `PUT` routes, permission checks granted to the built-in roles,
  broadcast targeting and the `{job_id, results[]}` envelope;
- `osapi client next node motd get` and `update`, with typed flags;
- an OpenAPI spec, under API (next);
- a docs page under Behaviors, with each example rendered as the CLI prints it
  and what it changes on the host;
- a provider the agent builds on Debian, skipped elsewhere;
- a span and three metrics on every call;
- a setting with a default, an environment variable, a flag and validation;
- provider, API, CLI and integration tests, from the examples.

Next: [Building a behavior](behaviors.md) has every option and tag, and
[Behaviors](../behaviors/behaviors.mdx) shows the ones osapi ships.

## Changing a hand-written domain

The domains that are not behaviors yet still have to be maintained. Their rules
are in the specifications repository, stated once, and the `add-a-domain` skill
there carries the build order and cites them:

| Subject                                                         | Document                                                                                     |
| --------------------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| What a domain consists of, and how to tell one is half-finished | [Building a domain](https://github.com/osapi-io/specs/blob/main/components/osapi/domains.md) |
| What a provider returns, and the three idempotency outcomes     | [Providers](https://github.com/osapi-io/specs/blob/main/components/osapi/providers.md)       |
| Broadcast targeting, and the shape every result returns         | [The job system](https://github.com/osapi-io/specs/blob/main/components/osapi/job-system.md) |
| What a new SDK service owes                                     | [The Go SDK](https://github.com/osapi-io/specs/blob/main/components/osapi/sdk.md)            |
| Where a new permission goes                                     | [Permissions](https://github.com/osapi-io/specs/blob/main/components/osapi/permissions.md)   |
| Running a command safely                                        | [Running commands](https://github.com/osapi-io/specs/blob/main/components/osapi/exec.md)     |

A new area of a host is a behavior, not a new hand-written domain.
