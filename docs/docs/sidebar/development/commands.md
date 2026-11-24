---
sidebar_position: 7.2
---

# Building a command

A local command runs in the CLI process and calls no API, such as
`osapi controller token generate`. A command that calls the API is an
[endpoint](endpoints.md) or a [behavior](behaviors.md). A command that runs for
a long time and stops on a signal, such as `agent start`, is not in the DSL yet:
it stays in `cmd/`.

A group of local commands is one directory, `internal/cli/commands/<group>/`,
and one import in `internal/cli/commands/all`. The DSL is in
`internal/cli/command`. From the definition, osapi makes the cobra command, the
flags, the help, the reference page and the checks of each method. The method is
a typed Go function. `internal/cli/commands/token` is a complete group.

## Steps

1. Write the layout:
   `go run . dsl command create --name <group> --description "..." --under <word>`.
   `--under` is the word of the root where the group runs: `controller`, `agent`
   or `nats`, or `dsl` for a command for authors.
2. Read the [Command DSL](dsl/command.md) reference, or
   `go run . dsl command spec get`. It lists each part of the DSL, what it does
   and when to use it.
3. Change the definition, the types and the methods.
4. Run `go generate ./internal/cli/commands/all` and `just cli-docs-generate`.
5. Run `go run . dsl command validate`. Each problem ends with its fix, and
   `--help` lists each rule.
6. Run the tests of the group.

`go run . dsl command list` shows each command of the binary and the layer that
made it. `go run . dsl command get --name <group>` shows the commands, flags and
refused inputs of a group.

The `build-a-command` [skill](skills.md) gives the same steps to an agent.
