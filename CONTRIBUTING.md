# Contributing

Thank you for your contribution to osapi.

## Before you start

- Read the [Code of Conduct](CODE_OF_CONDUCT.md). It applies to all
  communication in this repository.
- **Design pages.** The design pages of this repository are in
  [osapi-io/specs](https://github.com/osapi-io/specs), under
  `components/osapi/`. Read them before you change the behavior. When you change
  the behavior, change the design page in the same work. A design page in this
  repository becomes incorrect when the code changes. No check finds the
  difference.
- **Learn the project.** Read these pages in this order:
  1. [Guiding Principles][principles]. The design philosophy and the values of
     the project.
  2. [Architecture][system-architecture]. The processes, where the state is, and
     the behaviors.
  3. [API Design Guidelines][api-guidelines]. The REST conventions and the
     structure of an endpoint.
  4. [How a request runs][job-architecture]. Targets, the KV-first job, the
     dispatch in the agent, the statuses and broadcast.
- **Examine the work that exists.** Look for a pull request or an issue about
  the change. If you find one, use what it says in your work.
- **Backward compatibility.** Make sure that your change does not stop a program
  that uses osapi. A change that keeps compatibility is more likely to merge. If
  your change cannot keep it, open an issue first. Then the maintainers can
  discuss the API change before you do the work.

## Prerequisites

Install the tools with [mise]:

```bash
mise install
```

- **[Go].** osapi is written in Go. The project supports the two latest major
  versions of Go. Use one of them.
- **[Node.js].** The runtime for tools such as `@redocly/cli`. It also builds
  the Docusaurus site and the React UI in `ui/`.
- **[Bun].** The JavaScript package manager and script runner for the Docusaurus
  site and the React UI.
- **[uv].** A Python package runner. `just md-fmt` uses it to run [mdformat]. It
  installs nothing into the repository.
- **[just].** The task runner. Use it to build, test and format.
- **[NATS CLI].** Command-line tools for NATS. Use them to debug and monitor
  during development. Install them with `brew install nats-io/nats-tools/nats`.

### Coding agents

This repository works with any coding agent that reads `AGENTS.md`. `CLAUDE.md`
only points to `AGENTS.md`. Prepare your agent as its documentation tells you.
This repository does not need a plugin or an extension.

Agents read the skills of this repository from `.claude/skills`.
`.agents/skills` is a link to that directory, for agents that read that path.

## Setup

Download the shared justfiles and install the dependencies:

```bash
just fetch
just deps
```

## Project structure

- **`cmd/`.** The Cobra CLI commands (`client`, `node agent`, `controller.api`,
  `nats server`).
- **`internal/controller/`.** `controller.New` makes the controller process: the
  NATS connection, the REST API with each handler, the heartbeat and the
  watchers. A new API handler goes in `handlers.go`.
- **`internal/controller/api/`.** The Echo REST API. The handlers that target a
  node are under `node/{domain}/`. The handlers for the controller only are at
  the top level (`job/`, `health/`). Each domain has its own `gen/` with an
  OpenAPI spec. The combined spec is `internal/controller/api/gen/api.yaml`.
- **`internal/job/`.** The job domain types and the subject routes. `client/`
  holds the high-level operations.
- **`internal/agent/`.** The node agent: the consumer, handler and processor
  pipeline that runs jobs. `agent.Setup` makes the agent, with the provider of
  each area for the platform of the host. A new provider goes in `providers.go`,
  and its registration in `buildRegistry` in `setup.go`.
- **`internal/provider/`.** The implementations of the operations, in groups by
  category, then by domain.
- **`internal/config/`.** The configuration from `osapi.yaml`, read with Viper.
- **`internal/telemetry/`.** Traces, metrics, and the metrics that the agent
  records about itself.
- **`internal/controller/notify/`.** The system that sends notifications about
  conditions.
- **`internal/behaviors/`.** The behaviors. Each behavior is a directory. Its
  definition builds its routes, agent dispatch, commands, spec and docs.
  [Building a behavior][building-a-behavior] gives the layout that each behavior
  follows.
- **`internal/behavior/`.** The framework that defines behaviors, and its tools:
  the host that a provider is built from, the test harnesses, the spec builder
  and the generator. `command/` runs the generated commands, and `server/`
  serves the generated routes.
- **`internal/pki/`.** The keys and signatures of an agent and the controller.
  Both processes use it.
- **`internal/bus/`.** The NATS options, and the buckets and streams that osapi
  makes. Each process that connects to NATS uses it. `server/` runs the embedded
  NATS server and writes its heartbeat.
- **`internal/heartbeat/`.** It writes the registration of the controller or the
  NATS server to the registry KV bucket. The agent has its own heartbeat.
- **`internal/atomicfile/`.** It writes a file so that a reader sees the old
  file or the new file, never a part.
- **`internal/cli/`.** The output of the CLI: the theme, tables, help, errors
  and logs.
- **`internal/skills/`.** The agent skills, in `resources/`, and the code that
  installs them (`osapi init`) and serves them (`osapi skill`). Each skill is
  one directory, with its references and its evals.
  [Developing a skill][developing-a-skill] tells how to change, add and measure
  a skill.
- **`pkg/sdk/`.** The Go SDK. A program uses it to access the REST API.
- **`ui/`.** React 19, TypeScript, Vite and Tailwind CSS v4. The build embeds
  the UI in the Go binary.

### Where a package goes

- **`pkg/` is what osapi promises to code outside it:** the SDK. `pkg/` imports
  nothing from `internal/`. `internal/` can import `pkg/`, but `pkg/` cannot
  import `internal/`. `just test` fails if a `pkg/` package imports a package
  under `internal/`.
- **`internal/` is all the code that makes osapi work:** the agent, the
  controller, the job system, configuration, telemetry and behaviors. No code
  outside this module imports it, so it can change freely.
- **One process, one tree.** `internal/agent/` and `internal/controller/` never
  import each other. Code that both use goes in a top-level package, for example
  `internal/pki/`, `internal/bus/` or `internal/job/`.
- **The rules for each Go package are in [Go patterns](#go-patterns).** In
  osapi, `internal/behavior/command/` and `internal/behavior/server/` are under
  the framework that they belong to. `internal/telemetry/` groups `metrics`,
  `tracing`, `httplog` and `process`.
- **Some old packages do not obey these rules yet.**
  [#717](https://github.com/osapi-io/osapi/issues/717) lists them. Obey the
  rules in a new package.
- **A package that only one tree uses goes in that tree.** An example is
  `internal/behavior/behaviortest`, the fake host that only the tests of
  behaviors use. Do not put such a package at the top level, next to packages
  that are not related.
- **Declare an interface where you use it.** `internal/behavior/host` declares
  `Exec`, the part of the command runner that a provider calls. It does not use
  the type of the agent.
- **Code in `internal/` returns an error. Only `cmd/` ends the process.** A
  command calls `cli.LogFatal` with the error that it gets. Nothing in
  `internal/` calls `LogFatal` or `os.Exit`.
- **`cmd/` only parses flags and calls `internal/`.** A start command reads its
  flags and the config, calls `controller.New`, `agent.Setup` or `server.New`,
  and gives the result to `serve`. It holds no logic of its own that needs a
  test.
- **Embedded files go in `resources/` of the package that owns them.** The
  package that reads the files also embeds them, with
  `//go:embed resources/...`. Do not put embedded files in `cmd/`, or in a
  package that only passes them to a different package. The current resources:
  - `internal/skills/resources/`: the agent skills. `skills/` has the skills
    that `osapi init` installs. `dev-skills/` has the skills that `osapi skill`
    serves to a checkout of osapi.
  - `internal/behavior/resources/scaffold/`: the templates that
    `osapi behavior new` writes.
  - `internal/agent/resources/templates/`: the protected `osapi/` objects that
    the agent puts in the object store.
  - `ui/dist/`: the build of the UI, which `ui` embeds.

[Architecture][system-architecture] tells how the processes work together. The
[specifications](https://github.com/osapi-io/specs/tree/main/components/osapi)
describe the package layout, the handler structure and the provider pattern.

`nats-client` and `nats-server` are sibling repositories. `go.mod` pins their
module versions. `go.mod` has no `replace` directive.

## Code style

[gofumpt] formats the Go code. [golangci-lint] examines the Go code for
problems. CI makes sure that both pass.

```bash
just go-fmt-check   # Check formatting
just go-fmt         # Auto-fix formatting
just go-vet         # Run linter
```

`.golangci.yml` lists the linters that run. Read the list there. A copy of the
list in a page becomes incorrect when the configuration changes. Formatting does
not apply to generated files (`*.gen.go`, `*.pb.go`).

[Prettier] formats the TypeScript and the CSS in `ui/`. [ESLint] examines them
for problems. `just test` checks the format, so CI makes sure that it is
correct.

```bash
just react-fmt-check   # Check UI formatting
just react-fmt         # Auto-fix UI formatting
just react-lint        # Run ESLint
```

### Documentation

[mdformat] formats the markdown files, through `uvx`. CI makes sure that the
format is correct.

```bash
just md-fmt-check   # Check formatting
just md-fmt         # Auto-fix formatting
```

mdformat does not format the Docusaurus site under `docs/`. Prettier formats the
site, through the `docusaurus` module. CI makes sure that this format is correct
too. If you change a file under `docs/`, run `just docs-fmt-check`.

```bash
just docs-fmt-check   # Check formatting of the site
just docs-fmt         # Auto-fix formatting of the site
```

The CLI reference pages under `docs/docs/sidebar/usage/cli/` are generated from
the cobra commands. Do not edit them. Change the command in `cmd/`, then run
`just cli-docs-generate`:

- `Short` is one line, and `Long` tells what the command does in a few
  sentences. `--help` shows both.
- `Example` has one to three command lines, and no output.
- The annotations from `internal/cli/reference` are for the page only:
  `reference.Result` tells when the command returns `changed: true` or `false`
  and when it fails, `reference.Permission` names the permission that it needs,
  and `reference.Notes` has more detail. Put a fact in its annotation, not in
  `Long`. `just test` fails when a page does not agree with its command.
  `output.md` is the one page in that directory that a person writes.

```bash
just cli-docs-generate   # Write the CLI reference pages
```

Write each page in Simplified Technical English, as
[the osapi-io constitution](https://github.com/osapi-io/specs/blob/main/CONSTITUTION.md)
requires.

## Code standards

### Function signatures

If a function has parameters, write its signature on more than one line. Put
each parameter on its own line. Put the last `)` and the return types on their
own line:

```go
func FunctionName(
    param1 type1,
    param2 type2,
) (returnType, error) {
}
```

Write a function that has no parameters on one line:

```go
func Name() string {
}
```

Then a new parameter shows as one added line in a diff, not as a changed
signature.

### File names

Give a file the name of what it holds. Do not use `helpers.go`, `utils.go` or
names of that type. They tell where code is, not what it is. They also collect
code that has no other place.

`types.go` holds only type declarations: structs, interfaces, constants and
aliases. Put a function in a file that has the name of what the function does.

A test file has the name of the production file that it tests. If a test file
becomes too large to read, split the production file first. Then each test file
keeps its production file.

A test file that holds no tests has the name of what it holds. Do not give it a
name that tells only that other tests use it. `support`, `fixture` and `helper`
tell where the code is. `sha256_public_test.go` and `test_agent_public_test.go`
tell what the files hold. If only one test file uses the code, put the code in
that file.

### Go patterns

- Wrap errors with `fmt.Errorf("context: %w", err)`. Then the chain names each
  layer. `errors.Is` and `errors.As` can examine the chain.
- Return early. Do not put the successful path inside conditions.
- Give an unused parameter the name `_`.
- Put imports in three groups, with a blank line between them: the standard
  library, third-party packages, then local packages.
- A package holds one kind of thing, and its name tells what. Do not make a
  package named `util`, `common`, `helpers` or `shared`.
- The name of a package is the last part of its path. An exported name does not
  repeat it inside a longer name: `atomicfile.Write`, not
  `atomicfile.WriteFileAtomic`. The main type can have the name of the package,
  as `time.Time` does.
- A caller does not need an alias. If two packages that a file imports have the
  same name, rename one of them. Mock packages are the exception. Each is named
  `mocks`, so a test that uses two of them gives one an alias. A clash with a
  package outside the module, or between two names that are both correct, also
  takes an alias.
- A directory can group packages that belong together. Do not put many packages
  under one package.
  [A Go package is one thing](https://github.com/osapi-io/specs/blob/main/CONSTITUTION.md#a-go-package-is-one-thing)
  tells why.

### Test doubles

For an interface that this organization defines, generate the double with
`mockgen` and commit it. Do not write a struct by hand for it.

A generated mock is in a `mocks` package next to the code that it mocks. A
`generate.go` file in that package holds the directive:

```go
package mocks

//go:generate go tool go.uber.org/mock/mockgen -source=../types.go -destination=types.gen.go -package=mocks
```

The tool dependencies of the module supply the generator. Thus each checkout
runs the version that `go.mod` records. The names of the generated files end in
`.gen.go`. Commit these files. Do not put mocks in `gen/`. That name is for API
code generation.

If the interface is **unexported**, a sibling package cannot hold the mock. The
mock must import the package to name the types of the interface. The tests of
the package must import the mock. Thus generate the mock into the package. Give
the mock a destination that only tests use. Then the mock library is not a
dependency of code that imports the package:

```go
// generate.go, in the package that declares the interface
package thispackage

//go:generate go tool go.uber.org/mock/mockgen -source=thing.go -destination=thing.gen_test.go -package=thispackage
```

In both cases, the directives are in a `generate.go` that holds no code. The
name of the generated file has `.gen`, so a reader knows not to edit it.

If many tests set the same expectations, write a constructor that returns a
configured mock. Do not write a type by hand for it. The generated mock is still
the type that satisfies the interface.

Write these three doubles by hand, because a generated one gives no advantage:

- A double for a standard library interface: `net.Conn`, `fs.File`, `io.Writer`,
  `slog.Handler`. These interfaces do not change when our code changes.
- A double that has a real implementation of the behavior under test, for
  example signatures from a key pair that the test generates.
- A recorder for a dependency that a goroutine calls, if the test cannot wait
  for that goroutine. A generated mock counts the calls at a time that the test
  cannot control. Write that reason where you define the recorder.

### File headers

Start each `.go` file with the MIT license header. Copy the format from a Go
file in the repository. In a file with a build tag, put `//go:build` on line 1,
then a blank line, then the header.

### Logging

All logs use `log/slog`, the structured logger of Go.

- **Subsystem labels.** When you construct a component that holds a logger, wrap
  the logger with `logger.With(slog.String("subsystem", "..."))`. Then each line
  from that component has the label. Examples: `"agent"`, `"agent.seed"`,
  `"api.schedule"`, `"provider.file"`, `"job.client"`, `"metrics"`,
  `"controller.heartbeat"`.
- **Typed attributes.** Use `slog.String("key", val)`, `slog.Int`, `slog.Bool`
  and `slog.Any`. Do not use positional pairs such as `"key", val`. They
  compile, but they have no type safety.
- **Standard field names.** Use `error` for errors, `hostname` for hosts, `path`
  for file paths, `job_id` for job IDs, `name` for the names of entries and
  `addr` for addresses.
- **Error fields.** Use `slog.String("error", err.Error())` to log the error as
  a string. Use `slog.Any("error", err)` to keep the type of the error.
- **Log levels.** Use `Debug` for the dispatch of an operation, and for an
  operation that the provider skips because there is no change. Use `Info` for
  lifecycle events and state changes. Use `Warn` for a state that is degraded
  but works. Use `Error` for a failure that needs attention.

### Lifecycle

A component has a lifecycle that does not block. `Start()` returns immediately.
`Stop(ctx)` stops the component before a deadline. The `component` interface in
`cmd/shutdown.go` has these two methods. A start command gives its components to
`serve`. `serve` starts them, waits for the signal, then stops them with one
context that ends after `shutdownTimeout`.

### Filesystem access

Use [avfs]. Use `memfs.New()` for work in memory. Use `failfs.New()` to inject
an error at a specified operation. Do not use `afero`.

## Tests

```bash
just test           # Run all tests (lint + unit + coverage)
just go-unit       # Run unit tests only
just go-unit-cov   # Generate coverage report
just go-unit-int   # Run integration tests
go test -run TestName -v ./internal/job/...  # Run a single test
```

The coverage target is 99.9%. `just test` fails when the total coverage is less
than the target. Thus a change that adds code without tests fails on your
machine and in CI:

```bash
just go-unit-cov-check   # Report coverage and fail below the target
```

`.github/codecov.yml` and the shared `go` justfile module both declare the
target. Change the two together.

`.coverignore` limits what the coverage measures. It excludes `/cmd/`, `/gen/`,
`/mocks/`, `/ui/` and `main.go`. A percentage that includes these paths does not
agree with the gate.

### Test file conventions

- Public tests: `*_public_test.go`, in the `_test` package of the package. They
  test the exported API. Write this type of test first.
- Internal tests: `*_test.go`, in the same package. Use them only for what the
  exported API cannot reach.
- Suite names: `*_public_test.go` has `{Name}PublicTestSuite`, and `*_test.go`
  has `{Name}TestSuite`.
- Use `testify/suite` with table-driven cases.
- Write one suite method for each function under test. Success, errors and edge
  cases are rows in one table, not different methods.
- `export_test.go` gives external tests access to unexported symbols, through an
  alias or a setter. Do not use an alias to test again what the test of the
  caller already reaches. Use it for a helper that has its own contract.

In this repository, an external test is in a `_test` package next to the code
that it tests. The tables have `validateFunc` callbacks.

### Test layers

- **Unit tests** (`*_test.go`, `*_public_test.go`). They are fast and use mocked
  dependencies. Public suites also have HTTP test methods (`TestXxxHTTP`,
  `TestXxxRBACHTTP`). These methods send raw HTTP through the full Echo
  middleware stack, with mocked backends.
- **Integration tests** (`test/integration/`). They build and start a real
  `osapi` binary and run CLI commands from end to end. A
  `//go:build integration` tag guards them. The harness selects random ports,
  generates a JWT and starts the server. Thus the tests need no external setup.

For a new API domain, add a `{domain}_test.go` smoke suite under
`test/integration/`. Guard each test that makes a change with
`skipWrite(s.T())`. Then CI runs only the read-only tests by default. To enable
the tests that write, set `OSAPI_INTEGRATION_WRITES=1`.

### Test helpers

- Use `export_test.go` to give the `_test` package access to an unexported
  variable or function. Do not write an internal test or a stub by hand for it.
- The setter form is `SetXFn(fn)`, with a matching `ResetXFn()`. The setter
  replaces the dependency. The reset puts the original back.
- To reset replaced variables between table-driven sub-tests, use
  `suite.TearDownSubTest()`. Do not use `defer` in the loop.
- Platform stubs: test that the Darwin and Linux stubs return `ErrUnsupported`
  for each method.

## Building and running

```bash
just build     # Builds the React UI, then the Go binary
export OSAPI_CONTROLLER_API_SECURITY_SIGNING_KEY=$(openssl rand -hex 32)
./osapi controller start -f configs/osapi.yaml
```

`configs/osapi.yaml` has a placeholder `signing_key`. The placeholder fails the
startup validation on purpose. Thus you must set the environment variable above
to run the controller on your machine. The environment variable overrides the
value in the file. [Configuration] tells how `OSAPI_` environment variables map
to configuration keys.

**Always build through `just`.** The `//go:embed dist/*` directive in
`ui/embed.go` requires files in `ui/dist/` at compile time. `just build` runs
`just react-build` first to make these files. Thus `go build ./...` and
`go test ./...` fail if you run them directly. `just build`, `just test` and
`just ready` all build the UI first.

## Input validation

The `internal/validation` package validates all user input. It wraps
[go-playground/validator]. The OpenAPI specs declare the rules through
`x-oapi-codegen-extra-tags`. At runtime, the handlers call `validation.Struct()`
or `validation.Var()` to apply the rules.

The fields of the configuration structs in `internal/config/types.go` use the
same `validate` tags. Validation runs at startup, after `viper.Unmarshal()`.
Thus an invalid value stops the process immediately, with a clear error.
`viper.SetDefault()` in `cmd/root.go` sets the defaults, so you can omit most
fields. Use `go_duration` for Go duration strings. Add `required` to a field
that has no sensible default.

- **Required fields** use `validate:"required,..."`.
- **Optional fields** use `validate:"omitempty,..."`.
- **Enum constraints** use `validate:"oneof=a b c"`.
- **Cross-field validation** uses `required_without` or `excluded_with` for
  fields that exclude each other, such as the cron `schedule` and `interval`.

### Update endpoints with all-optional fields

If all fields of a PUT endpoint are optional (user update, group update, cron
update), call `validation.AtLeastOneField(request.Body)` after
`validation.Struct()`. This call rejects an empty body with a 400. Without it,
an update can do nothing. Worse, it can apply destructive defaults.

```go
if errMsg, ok := validation.Struct(request.Body); !ok {
    return gen.PutXxx400JSONResponse{Error: &errMsg}, nil
}

if errMsg, ok := validation.AtLeastOneField(request.Body); !ok {
    return gen.PutXxx400JSONResponse{Error: &errMsg}, nil
}
```

### Defense in depth

If `validation.Struct()` cannot fail now because all fields use `omitempty`,
keep the call. Add a comment that tells why. Then the call validates a field
that someone adds later:

```go
// Defense in depth: current fields use omitempty so validation
// always passes, but guards against future field additions.
if errMsg, ok := validation.Struct(request.Body); !ok {
    return gen.PostXxx400JSONResponse{Error: &errMsg}, nil
}
```

This applies to action endpoints (power, docker stop), where an empty body is
valid. Update endpoints are different. They must use `AtLeastOneField`.

## Adding a new API domain

A new area of a host is a behavior: one directory under `internal/behaviors/`.
The definition of the behavior builds the routes, commands, spec, docs and
telemetry. [Your first behavior][your-first-behavior] builds one from the start.

Each hand-written domain that is not a behavior yet must be consistent across
all of its layers. The layers are the provider, the agent processor, the API
handler, the SDK service, the CLI commands, the docs and the tests. **Select an
existing domain. Search for it across the codebase with `find` and `grep`. Your
new domain must be in all the same places.** If something exists for `sysctl`
but not for your domain, it is missing.

### Rules a domain has to follow

These rules apply to all domains. This page states each rule because you need it
to make the change. The page behind each link gives the reason.

- **One endpoint does not both create and update.** `POST` creates, with the
  name in the body. `PUT /{name}` updates, with the name from the path. If one
  endpoint does both, a 404 has no meaning.
  [Why](https://github.com/osapi-io/specs/blob/main/components/osapi/domains.md)
- **A create of a thing that exists is not an error. An update of a thing that
  is absent is an error.** A create is idempotent and an update is not. The
  caller that sends an update has stated that the thing exists.
  [Why](https://github.com/osapi-io/specs/blob/main/components/osapi/providers.md)
- **A permission that the built-in roles do not include reaches nobody.** The
  permission exists and the handler checks it, but no token can hold it. Nothing
  reports that the endpoint is unreachable. A behavior adds its permissions to
  the roles from its definition. For a hand-written domain, add them to
  `builtinRolePermissions`.
  [Why](https://github.com/osapi-io/specs/blob/main/components/osapi/permissions.md)
- **In the provider, validate again each value that becomes a path, a filename
  or a command argument.** Validation in the handler is not sufficient. The
  system stores the job and runs it later, so the provider is a second caller
  that comes after the handler. GHSA-7fjw-v3g9-326g was the result of this
  error.
  [Why](https://github.com/osapi-io/specs/blob/main/components/osapi/providers.md)
- **Send a secret to a command through stdin, never through an argument.** Use
  `RunPrivilegedCmdWithStdin`. The logs record arguments, and the process table
  shows them. GHSA-6gc6-px2x-q95j was the result of this error.
  [Why](https://github.com/osapi-io/specs/blob/main/components/osapi/exec.md)
- **A value that starts with a dash becomes a flag** if it reaches a command
  without a guard. The provider prevents that before it calls `internal/exec`.
  [Why](https://github.com/osapi-io/specs/blob/main/components/osapi/exec.md)
- **Run every command through `internal/exec`.** If you start a process
  directly, you lose the timeout, the argument logs, the stdin path for secrets
  and the testability. Nothing in the build tells you.
  [Why](https://github.com/osapi-io/specs/blob/main/components/osapi/exec.md)
- **Ten minutes is a maximum, not a default.** `internal/exec` always wraps the
  context of each command. A caller can ask for less time, but it cannot ask for
  more. A job that needs twenty minutes does not get them.
  [Why](https://github.com/osapi-io/specs/blob/main/components/osapi/exec.md)
- **Put each field that holds a secret in the audit denylist.** Redaction
  matches on the field name. Thus the audit log records a new field as sent,
  until you add the field to `sensitiveFields`. No test fails if you forget. A
  behavior tags the field `secret:"true"` instead. The tag redacts the field,
  and you do not edit the list.
  [Why](https://github.com/osapi-io/specs/blob/main/components/osapi/audit.md)

## UI contributions

The React management dashboard is in `ui/`. The build embeds it in the Go
binary. [UI Development][ui-development] gives the prerequisites, the
development server, the code style and the component conventions.
[UI Architecture][ui-architecture] tells how the binary embeds the UI. It also
gives the component layers and the flow that generates the SDK.

## Documentation

osapi uses [Docusaurus] for its documentation site. The pages are Markdown files
under `docs/docs/`, wrapped at 80 characters. [Documentation](#documentation)
under Code style tells how to format them.

```bash
just docs-start     # Start local docs server
just docs-build     # Build docs for production
```

## Before you commit

Run `just ready` before you commit. It makes sure that generated code, package
documentation, format and lint are current. In this repository, it also builds
the Go binary and the UI. Then your commit agrees with what CI checks:

```bash
just ready
```

## Branches

Make all changes on a branch. Make the branch from `main`. Give it the name
`type/short-description`, where `type` is the [Conventional Commits] type:

- `feat/add-retry-logic`
- `fix/null-pointer-crash`
- `docs/update-api-reference`
- `refactor/simplify-handler`
- `chore/update-dependencies`

## Commit messages

Use [Conventional Commits] with the 50/72 rule:

- **Subject line.** 50 characters or fewer, in the imperative mood, with no
  period at the end.
- **Body.** Wrap it at 72 characters. Put a blank line between the subject and
  the body.
- **Format.** `type(scope): description`.
- **Types.** `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`,
  `chore`.
- Tell what changed and why, not how.

Write commit messages that help the reader. Keep the number of commits in a pull
request small. Most pull requests have one commit. A large pull request can have
a few. Use squash and rebase to combine commits.

## Submit a pull request

- **Describe your changes.** Tell what changed and why. Then a reviewer does not
  have to read the diff to find the reason.
- **Links to issues and pull requests.** Link to related issues and pull
  requests. Tell how your change is different from that work, or how it adds to
  it.
- **Examples.** Add examples or screenshots that show the effect of your change.
- **Draft pull requests.** If your change is not complete but you want to
  discuss it, open the pull request as a draft. Start the discussion in a
  comment, not in the description. Then you can change the description later and
  keep the discussion.

## AI usage

We write this repository with the help of AI tools. The
[AI Usage Policy](AI_POLICY.md) applies to all contributions. Tell which tool
you used. Make sure that you can explain what your change does without the help
of an AI tool.

## FAQ

> I want to contribute. Where do I start?

All contributions are welcome, from a typo fix to a new feature. You can also
vote or comment on issues, or help answer questions.

> I need help. Where do I ask?

Open a [Discussion] on GitHub.

[api-guidelines]: docs/docs/sidebar/architecture/api-guidelines.md
[avfs]: https://github.com/avfs/avfs
[building-a-behavior]: docs/docs/sidebar/development/behaviors.md
[bun]: https://bun.sh
[configuration]: docs/docs/sidebar/usage/configuration.md
[conventional commits]: https://www.conventionalcommits.org
[developing-a-skill]: docs/docs/sidebar/development/skills.md
[discussion]: https://github.com/osapi-io/osapi/discussions
[docusaurus]: https://docusaurus.io
[eslint]: https://eslint.org
[go]: https://go.dev
[go-playground/validator]: https://github.com/go-playground/validator
[gofumpt]: https://github.com/mvdan/gofumpt
[golangci-lint]: https://golangci-lint.run
[job-architecture]: docs/docs/sidebar/architecture/request-flow.md
[just]: https://just.systems
[mdformat]: https://pypi.org/project/mdformat/
[mise]: https://mise.jdx.dev
[nats cli]: https://github.com/nats-io/natscli
[node.js]: https://nodejs.org
[prettier]: https://prettier.io
[principles]: docs/docs/sidebar/architecture/principles.md
[system-architecture]: docs/docs/sidebar/architecture/architecture.md
[ui-architecture]: docs/docs/sidebar/architecture/ui.md
[ui-development]: docs/docs/sidebar/development/ui-development.md
[uv]: https://docs.astral.sh/uv/
[your-first-behavior]: docs/docs/sidebar/development/your-first-behavior.md
