---
sidebar_position: 1
sidebar_label: Overview
---

# Development

This guide covers the tools, setup, and conventions needed to work on OSAPI.

## Prerequisites

Install tools using [mise][]:

```bash
mise install
```

- **[Go][]** - OSAPI is written in Go. We always support the latest two major Go
  versions, so make sure your version is recent enough.
- **[Node.js][]** - Required as a runtime for tools like `@redocly/cli`, for
  building the Docusaurus docs site, and for building the embedded React UI in
  `ui/`.
- **[Bun][]** - JavaScript package manager and script runner used for the
  Docusaurus docs and the React UI.
- **[just][]** - Task runner used for building, testing, formatting, and other
  development workflows. Install with `brew install just`.
- **[NATS CLI][]** - Command-line tools for interacting with NATS. Useful for debugging
  and monitoring during development. Install with `brew install nats-io/nats-tools/nats`.

### Claude Code

If you use [Claude Code][] for development, install these plugins from the default
marketplace:

```
/plugin install commit-commands@claude-plugins-official
/plugin install superpowers@claude-plugins-official
```

- **commit-commands** — provides `/commit` and `/commit-push-pr` slash commands
  that follow the project's commit conventions automatically.
- **superpowers** — provides structured workflows for planning, TDD, debugging,
  code review, and git worktree isolation.

## Setup

Fetch shared justfiles and install all dependencies:

```bash
just fetch
just deps
```

## Code style

Go code should be formatted by [`gofumpt`][gofumpt] and linted using
[`golangci-lint`][golangci-lint]. Markdown and TypeScript files should be
formatted and linted by [Prettier][]. This style is enforced by CI.

```bash
just go::fmt-check   # Check formatting
just go::fmt         # Auto-fix formatting
just go::vet         # Run linter
```

## Running your changes

To run OSAPI with working changes:

```bash
just build     # Builds React UI + Go binary
./osapi controller start -f configs/osapi.yaml
```

:::important

Use `just build` (not `go build` directly). The `//go:embed` directive for the
UI assets requires `ui/dist/` to be populated at compile time — `just build`
runs `just react::build` first to satisfy this. The same applies to tests: use
`just test`, not `go test ./...`.

:::

## UI Development

The embedded React management dashboard has its own development workflow. See
the [UI Development](ui-development.md) guide for prerequisites, the development
server (`just react::dev`), code style, and component conventions.

## Documentation

OSAPI uses [Docusaurus][] to host a documentation server. Content is written in
Markdown and located in the `docs/docs` directory. All Markdown documents should
have an 80 character line wrap limit (enforced by Prettier).

```bash
just docs::start     # Start local docs server (requires bun)
just docs::build     # Build docs for production
just docs::fmt-check # Check docs formatting
```

## Testing

See the [Testing](testing.md) page for details on running tests and listing just
recipes.

```bash
just test           # Run all tests (lint + unit + coverage)
just go::unit       # Run unit tests only
just go::unit-int   # Run integration tests (requires running osapi)
```

Unit tests should follow the Go convention of being located in a file named
`*_test.go` in the same package as the code being tested. Integration tests are
located in `test/integration/` and use a `//go:build integration` tag. They
build and start a real `osapi` binary, so they require no external setup.

Use `testify/suite` with table-driven patterns and `validateFunc` callbacks.
**One suite method per function under test.** All scenarios for a function
(success, error codes, transport failures, nil responses) belong as rows in a
single table — never split into separate `TestFoo`, `TestFooError`,
`TestFooNilResponse` methods.

### File naming

Avoid generic file names like `helpers.go` or `utils.go`. Name files after what
they contain.

## Input Validation

All user input is validated through the `internal/validation` package, which
wraps `go-playground/validator`. Validation rules are declared in OpenAPI specs
via `x-oapi-codegen-extra-tags` and enforced at runtime by handler calls to
`validation.Struct()` or `validation.Var()`.

### Config validation

Config struct fields in `internal/config/types.go` use the same `validate` tags.
Validation runs at startup after `viper.Unmarshal()` — invalid values cause an
immediate exit with a clear error. Defaults are set via `viper.SetDefault()` in
`cmd/root.go` so most fields can be omitted. Use `go_duration` for Go duration
strings. Add `required` to fields with no sensible default.

### Validation rules

- **Required fields** use `validate: "required,..."` — the field must be present
  and non-zero.
- **Optional fields** use `validate: "omitempty,..."` — validation is skipped
  when the field is absent or zero-valued.
- **Enum constraints** use `validate: "oneof=a b c"` to restrict values.
- **Cross-field validation** uses `required_without` / `excluded_with` for
  mutually exclusive fields (e.g., cron `schedule` vs `interval`).

### Update endpoints with all-optional fields

When a PUT endpoint has all optional fields (e.g., user update, group update,
cron update), use `validation.AtLeastOneField(request.Body)` to reject empty
bodies with a 400. This prevents clients from sending meaningless no-op updates
or, worse, triggering destructive defaults. Place this call after
`validation.Struct()`:

```go
if errMsg, ok := validation.Struct(request.Body); !ok {
    return gen.PutXxx400JSONResponse{Error: &errMsg}, nil
}

if errMsg, ok := validation.AtLeastOneField(request.Body); !ok {
    return gen.PutXxx400JSONResponse{Error: &errMsg}, nil
}
```

### Defense-in-depth pattern

When `validation.Struct()` cannot currently fail (all fields use `omitempty`),
keep the call with a comment explaining why. This guards against future field
additions breaking validation silently:

```go
// Defense in depth: current fields use omitempty so validation
// always passes, but guards against future field additions.
if errMsg, ok := validation.Struct(request.Body); !ok {
    return gen.PostXxx400JSONResponse{Error: &errMsg}, nil
}
```

This pattern applies to action endpoints (power, docker stop) where an empty
body is valid — unlike update endpoints which must use `AtLeastOneField`.

## Before committing

Run `just ready` before committing to ensure generated code, package docs,
formatting, and lint are all up to date:

```bash
just ready
```

## Branching

All changes should be developed on feature branches. Create a branch from `main`
using the naming convention `type/short-description`, where `type` matches the
[Conventional Commits][] type:

- `feat/add-retry-logic`
- `fix/null-pointer-crash`
- `docs/update-api-reference`
- `refactor/simplify-handler`
- `chore/update-dependencies`

When using Claude Code's `/commit` command, a branch will be created
automatically if you are on `main`.

## Commit messages

Follow [Conventional Commits][] with the 50/72 rule:

- **Subject line**: max 50 characters, imperative mood, capitalized, no period
- **Body**: wrap at 72 characters, separated from subject by a blank line
- **Format**: `type(scope): description`
- **Types**: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`
- Summarize the "what" and "why", not the "how"

Try to write meaningful commit messages and avoid having too many commits on a
PR. Most PRs should likely have a single commit (although for bigger PRs it may
be reasonable to split it in a few). Git squash and rebase is your friend!

<!-- prettier-ignore-start -->
[mise]: https://mise.jdx.dev
[Go]: https://go.dev
[Node.js]: https://nodejs.org/en/
[Bun]: https://bun.sh
[just]: https://just.systems
[Claude Code]: https://claude.ai/code
[Anthropic Marketplace]: https://marketplace.anthropic.com
[gofumpt]: https://github.com/mvdan/gofumpt
[golangci-lint]: https://golangci-lint.run
[Prettier]: https://prettier.io/
[Docusaurus]: https://docusaurus.io
[Conventional Commits]: https://www.conventionalcommits.org
[NATS CLI]: https://github.com/nats-io/natscli
<!-- prettier-ignore-end -->
