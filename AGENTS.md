# AGENTS.md

Test: `just test` | Before you commit: `just ready`

Read [CONTRIBUTING.md](CONTRIBUTING.md) first. It gives:

- the prerequisites and the setup
- the code style and the test conventions
- the pull request workflow

These rules apply to agents and to people. This file has only the rules that are
specific to agents. They apply to each coding agent.

## Run tools

Run tools through `mise`, not from your path:

```bash
mise exec -- just test
```

`mise` is active in the shell of a person. It supplies the versions that
`.mise.toml` declares. The shell of an agent does not activate `mise`. Thus a
bare `just` runs the version that is installed globally. That version is usually
older.

The symptom is a check that fails here but passes in CI, on a file that nobody
changed. If this occurs, find the version that ran. Do this before you accept
the failure as real.

## Never build with the Go toolchain directly

`go build ./...` and `go test ./...` fail in this repository. The UI is embedded
via `//go:embed dist/*`, which requires `ui/dist/` to hold files at compile
time. Use `just build`, `just test`, or `just ready`. Each builds the UI first.
See [CONTRIBUTING.md](CONTRIBUTING.md#building-and-running).

## Where the rules come from

[CONTRIBUTING.md](CONTRIBUTING.md#before-you-start) names the design pages. If a
rule here and a design page do not agree, obey the design page. Say that they do
not agree. Do not use the code to decide.

## Commit trailer

When an agent writes a commit, end the commit message with a trailer that names
the agent. Use the trailer that the documentation of the agent gives.

## Design pages

**The design is in the design pages.** Write or change the page in
[osapi-io/specs](https://github.com/osapi-io/specs) before you build. If the
build shows that the page is wrong, correct the page. A design page in this
repository becomes incorrect when the code changes. No check finds the
difference.
