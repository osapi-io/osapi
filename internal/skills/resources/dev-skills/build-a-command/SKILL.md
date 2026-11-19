---
name: build-a-command
description: Build or change a group of local commands, the typed definition under internal/cli/commands/<group>/ of osapi commands that run in the CLI process and call no API, such as osapi token generate. Use when asked to add or change a local command, to move a hand-written command under cmd/ to the command DSL, when osapi cli command validate, TestRefusals or TestCommandShape fails, or to decide whether a command is a local command, an endpoint or a behavior. Not for a command that calls the API: the build-an-endpoint and build-a-behavior skills cover those.
---

# Build a command

A local command runs in the CLI process. It calls no API. Examples are
`osapi token generate` and `osapi token roles list`. A group of local
commands is one directory, `internal/cli/commands/<group>/`, plus one import in
`internal/cli/commands/all`. The DSL is in `internal/cli/command`.

This skill gives a procedure and decisions. It is not a catalog. The binary tells
you about the DSL:

- `cli command spec` tells each part of the DSL, what it does and when to use
  it.
- `cli command validate --help` tells each rule, with its fix. Each failure
  also ends with its fix.
- `cli command list` tells each command of the binary, and the layer that made
  it. `cli command get --name <group>` tells the commands, the flags and the
  refused inputs of a group.

Ask these tools. Do not use your memory. If a tool and this skill do not agree,
the tool is correct.

The DSL makes the cobra command, the flags, the help, the page and the checks of
each method. The method is a typed function. If the work needs a part that
`cli command spec` does not list, stop and say so. That is a platform change,
and a person decides it. Do not add code around it.

The design is `components/osapi/commands.md` in osapi-io/specs. Read it first.

## Local command, endpoint or behavior

| The command | It is |
| --- | --- |
| Runs on the machine of the person, with no API: the config, a file, a token | A local command: this skill |
| Calls the API for work on the controller: the registry, the job queue, the audit log | An endpoint: use the build-an-endpoint skill |
| Calls the API for work on a host | A behavior: use the build-a-behavior skill |
| Runs for a long time and stops on a signal, such as `agent start` | Not yet: the DSL has no method for a process. Keep it in `cmd/` |

## Find what you need

| Question | Ask |
| --- | --- |
| Each part of the DSL, what it does and when to use it | `go run . cli command spec` |
| Each rule a definition must pass, with its fix | `go run . cli command validate --help` |
| Each command of the binary, and the layer that made it | `go run . cli command list` |
| The commands, flags and refused inputs of a group | `go run . cli command get --name <group>` |
| Whether a definition breaks a rule | `go run . cli command validate` |
| A complete group | `internal/cli/commands/token` |

## Build a new group

1. Write it: `go run . cli command new --name <group> --description "<one line>"`,
   then `go generate ./internal/cli/commands/all`. The group has a list and a
   get, an empty `Deps`, its types and the tests of its methods. It builds,
   validates and passes its tests before you change it.
2. Change it in this order: `types.go`, `deps.go`, `methods.go`, then
   `<group>.go`. Decide each method with the user first. "Decide first"
   below has the questions.
3. When `Deps` gets its first field, give it with `command.Provide` in
   `cmd/commands.go`, as the token group does. Add a `mocks/` package for the
   tests. While `Deps` is empty, the CLI needs no wiring.
4. Run `go generate ./internal/cli/commands/all` and
   `mise exec -- just cli-docs-generate`.

## Move a hand-written command

1. Before you change the command, save its help:
   `go run . <command> --help --json`.
2. Write the definition. Each flag of the old command is a field of the input,
   with its json name, its `short` tag, its `doc` and its `validate` tag. A
   check in the old command, such as a list of the roles that are allowed, is a
   validate rule now. Do not write the check in the handler; declare the rule.
3. Move the logic into `methods.go`, one function for each method. Each thing
   that the logic reads from the process, such as the config or a client, goes
   into `Deps` as an interface. Delete the output code: the output type prints
   the output. Tag the fields of the output type until it prints as the old
   command did. Tell the user each difference.
4. Delete the hand-written files from `cmd/`, and run
   `go generate ./internal/cli/commands/all`.
5. Compare the help with the saved help: the same flags, types, defaults and
   required flags. Tell the user each difference. Run
   `mise exec -- just cli-docs-generate`, and read the diff of each page.

## The rules to keep

- Each value is a flag. A command takes no positional argument.
- The last word of each command is a verb or an action. A noun goes before the
  verb with `.Sub`, such as `token roles list`. Add a word to
  `command.Actions` only for a real action, never for a noun.
- Give each rule of each input field a `.Fails` input that breaks it: the
  input with the other fields valid and one value wrong. A required field needs
  one too, with the field empty. `TestRefusals` runs each one, and expects a
  usage error. You write the bad value, because a generator cannot know one for
  a custom rule.
- A rule that the standard tags cannot say is a custom rule. Declare it with
  `rule.Declare`, with its check and its message.
- `--json` prints the output of the method. Do not make a second shape for it.
- The output type prints the output. Give each field the tags that it needs:
  `render`, `label`, `none`, `empty`, `prefix` and `table:"-"`, and a heading
  with `_ struct{} \`title:"..."\``. `osapi cli command spec` has each tag.
- If the output type cannot print the output, change the output type first.
  A method that does a different thing with each value, such as a matrix, is
  two methods: a list and a get. Write a renderer only when no output type can
  give the shape, such as a tree. Give it `.Because("...")`.
  `osapi cli command validate` fails a renderer without a reason.

## Decide first

Ask the user when the request does not tell you. Give your recommendation with
each question.

| Ask | Answer becomes |
| --- | --- |
| Does it call the API? | This skill, or build-an-endpoint or build-a-behavior |
| Is it a new command, or one that moves? | A move keeps the name, the flags and the output; tell the user each difference |
| Where does it go in the tree? | Under the root, or under a command with `Under`, such as `controller` |
| What does it read from the machine? | A field of `Deps` for each: the config, the file system, a clock |
| What does a person read at a shell? | The tags of the output type; a renderer only for a shape no type gives, with `.Because` |

## Check it

```bash
go run . cli command validate
go generate ./internal/cli/commands/all
mise exec -- just cli-docs-generate
go test ./internal/cli/... ./cmd/
mise exec -- just test
```

| Failure names | It means |
| --- | --- |
| `cli command validate`, or `TestAllPublicTestSuite/TestValidate` in `internal/cli/commands/all` | A rule of the definition. The problem names the rule and ends with its fix |
| `TestRefusals` | A `.Fails` input does not give a usage error: it breaks no rule, or it breaks a rule of another field |
| `TestCommandShape` | The last word of a command is not a verb or an action |
| `go-generate-check` | A `gen/` file is stale: run `go generate ./internal/cli/commands/all` |
| `cli-docs-gen-check` | A CLI page is stale: run `mise exec -- just cli-docs-generate` |
| `is not given: add command.Provide` when a command runs | The group's `Deps` has a field, and `cmd/commands.go` does not give it |
| `command-scaffold-check` | A group that `cli command new` writes no longer passes its tests: change `internal/cli/command/scaffold_files.go`, and run its golden test with `-update` |
| `the spec has no term for` | A new option of the DSL has no entry in `internal/cli/command/spec.go`. Add one |

Never edit a path that `cli command spec` lists under "Generated, never edited".
Change the definition and generate. Do not write CLI docs for a local command.
