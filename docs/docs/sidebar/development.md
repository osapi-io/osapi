---
sidebar_position: 2
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
- **[Node.js][]** - Required as a runtime for tools like `@redocly/cli`.
- **[Bun][]** - JavaScript package manager used for Docusaurus docs and
  installing tooling.
- **[just][]** - Task runner used for building, testing, formatting, and other
  development workflows. Install with `brew install just`.
- **[NATS CLI][]** - Command-line tools for interacting with NATS. Useful for debugging
  and monitoring during development. Install with `brew install nats-io/nats-tools/nats`.

### Claude Code

If you use [Claude Code][] for development, install the **commit-commands** plugin
from the default marketplace:

```
/plugin install commit-commands@claude-plugins-official
```

This provides `/commit` and `/commit-push-pr` slash commands that follow the
project's commit conventions automatically.

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
go run main.go overlay
```

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

See the [Testing](testing.md) page for details on running tests.

```bash
just test           # Run all tests (lint + unit + coverage + bats)
just go::unit       # Run unit tests only
just bats::test     # Run integration tests only
```

Unit tests should follow the Go convention of being located in a file named
`*_test.go` in the same package as the code being tested. Integration tests are
located in the `test` directory and executed by [Bats][].

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
[Bats]: https://github.com/bats-core/bats-core
[NATS CLI]: https://github.com/nats-io/natscli
<!-- prettier-ignore-end -->
