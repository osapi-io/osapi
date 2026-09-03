# AGENTS.md

Test: `just test` | Before committing: `just ready`

Read @CONTRIBUTING.md first. It covers prerequisites, setup, project structure,
code standards, testing, input validation, and how to add an API domain. All of
which apply to agents exactly as they apply to people. This file carries only
what is specific to agents.

## Running tools

Invoke tools through `mise`, not from your path:

```bash
mise exec -- just test
```

`mise` is active in a person's shell and supplies the versions `.mise.toml`
declares. An agent's shell has no activation, so a bare `just` resolves to
whatever is installed globally, usually an older version.

The symptom is a check that fails here and passes in continuous integration, on
a file nobody edited. When that happens, establish which version ran before
treating the failure as real.

## Never build with the Go toolchain directly

`go build ./...` and `go test ./...` fail in this repository. The UI is embedded
via `//go:embed dist/*`, which requires `ui/dist/` to hold files at compile
time. Use `just build`, `just test`, or `just ready`. Each builds the UI first.
See @CONTRIBUTING.md under "Building and running".

## Where the rules come from

@CONTRIBUTING.md names the specification under "Before you start". When a
convention here and the specification disagree, the specification wins. Say so
rather than following the code.

## Commit trailer

When committing via Claude Code, end the message with:

```
🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

## Task tracking

**Do not use superpowers.** Spec Kit governs specification, planning, and
implementation, and the design record for a change lives in
[osapi-io/specs](https://github.com/osapi-io/specs). A second workflow over that
ground gives two answers to which artifact is authoritative, and the answer that
loses is the one nobody reads. Nothing superpowers produces is committed.
