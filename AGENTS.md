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

## Do not build with the Go toolchain directly

`go build ./...` and `go test ./...` fail in this repository. The build embeds
the UI through `//go:embed dist/*`. Thus `ui/dist/` must hold files at compile
time. Use `just build`, `just test` or `just ready`. Each one builds the UI
first. Refer to [CONTRIBUTING.md](CONTRIBUTING.md#building-and-running).

## Where the rules come from

[CONTRIBUTING.md](CONTRIBUTING.md#before-you-start) names the design pages. If a
rule here and a design page do not agree, obey the design page. Say that they do
not agree. Do not use the code to decide.

## Track work in issues

Work that you agree to do later is an issue in the repository where the change
goes. A problem that you find during a different change is also an issue. Open
the issue when you agree to the work or find the problem. Before you open it,
look for an issue that has the work.

Your task list, your memory and the text of the session are not records. They
stop when the session stops, and nobody else can read them. Use a task list only
for the steps of the change that you do now. Each item that is not done when the
change merges becomes an issue, or you delete it.
[Tracking](https://github.com/osapi-io/specs/blob/main/CONSTITUTION.md#tracking)
has the rule.

## Commit trailer

When an agent writes a commit, end the commit message with a trailer that names
the agent. Use the trailer that the documentation of the agent gives.

## Design pages

**The design is in the design pages.** Write or change the page in
[osapi-io/specs](https://github.com/osapi-io/specs) before you build. If the
build shows that the page is wrong, correct the page. A design page in this
repository becomes incorrect when the code changes. No check finds the
difference.
