# The parts that each DSL shares

| Part              | Where it is                                     | What a new DSL does                                                  |
| ----------------- | ----------------------------------------------- | -------------------------------------------------------------------- |
| Spec              | `internal/guide`: `guide.Spec`, `guide.Term`    | Declares a `Spec` with a term for each exported part of the DSL      |
| Checks            | `internal/guide`: `guide.Checks`, `guide.Check` | Declares a `Checks` with each rule, what it requires and its fix     |
| Validate          | `<dsl> validate`                                | Returns problems, each with its rule and the fix from `Checks`       |
| Learning commands | `<dsl> spec get`, `<dsl> validate --help`       | Prints the `Spec` and the `Checks`, with `--json`                    |
| Typed values      | named types and constants in the DSL package    | Gives each value from a fixed set a type, and a validate rule for it |
| Struct tags       | `command.KnownTags`, `command.TagKeys`          | Uses the same list of tag keys as the other DSLs                     |
| Permissions       | `internal/authtoken/permissions.go`             | Declares the permission of each method that the API serves           |
| Generation        | `internal/cli/command/gen/genfile`              | Writes its generated files through `genfile.Writer`                  |
| Scaffold          | `<dsl> create`, and a `*-scaffold-check` recipe | Writes a new definition that builds and passes its tests             |
| Docs page         | `cmd/commands.go` (`dslPages`), `cli docs dsl`  | Adds its `Spec` as a page under `development/dsl/`                   |
| Skill             | `internal/skills/resources/dev-skills/`         | Adds a skill that sends an agent to the learning commands            |

## Spec and checks

The terms and the rules are Go tables of the types in `internal/guide`. The
`spec get` command, the validate help and each problem read the same table. Thus
what a command tells cannot be different from what the code does.

- **`guide.Spec`** has kinds and terms. Each `guide.Term` has the name as an
  author writes it, what it makes osapi do, and when to use it. `behavior.Spec`,
  `endpoint.Spec` and `command.Spec` are examples.
- **`guide.Checks`** has each rule of validate: the rule name that a problem
  reports, what it requires, and the fix. `Checks.Fix(rule)` gives the fix that
  a problem carries, and `Checks.Help()` gives the validate help.

Two tests keep the tables complete. Copy them:

- **A spec test** parses the exported functions and methods of the DSL package,
  and fails on one that has no term. The behavior DSL has
  `TestSpecHasEveryPartOfTheDSL`. The endpoint DSL and the command DSL have
  `TestSpecCoversTheDSL`, with a list of exported names that are not parts of
  the DSL.
- **A checks test** fails on a rule that has no check. The behavior DSL has
  `TestChecksHaveEveryRule`, which reads each `Rule` constant. The validate
  tests of the endpoint DSL and the command DSL also fail on a check that no
  test case triggers.

## Validate and typed values

Validate runs on each registered definition. A test of each registry runs it,
such as the tests of `internal/controller/endpoints/all` and
`internal/cli/commands/all`. Thus `just test` fails on a definition that breaks
a rule.

A value that selects from a fixed set is a named type with constants, such as
`behavior.Verb`, `behavior.PlatformName`, `behavior.GroupName`, `command.Parent`
and `endpoint.Reason`. Go converts an untyped string literal to a named type.
Thus `behavior.Platform("debain", ...)` compiles, and a validate rule must
refuse a value outside the set, such as `behavior.KnownPlatforms()`. Names,
documentation text, program names and path patterns stay strings.

A struct tag of an input or an output field uses a key of `command.KnownTags`.
The command DSL refuses an unknown key of an output field, and the behavior DSL
refuses one of an input field, with the same list. Go reads a tag only by its
exact key, so a misspelled key such as `secrte` does nothing and no compiler
reports it. Add a new tag key to `command.KnownTags`.

`internal/dslcheck` fails on a new bare `string` in the exported API of a DSL
that its allowlist does not give a reason for, and on a string literal where a
typed value goes (osapi-io/osapi#1078).

## Permissions

The Permission column of [Methods](#methods) gives the permission of each method.
## Generation and the scaffold

A generator writes its files through `genfile.Writer`. `Clean`, `Write`, `Go`,
`Remove` and `Registry` only record a change, and `Flush` makes the changes at
the end of a run that succeeds. Thus a generator that fails or panics part way
changes nothing on disk. `Go` turns a panic of jennifer into an error that names
the file. jennifer's `Lit` panics on a named type, so write
`jen.Lit(string(v))`.

`go-generate-check` fails when the generated files differ from the committed
files. Give the new DSL a `<dsl> create` command that writes a new definition,
and a `*-scaffold-check` recipe in the `justfile` that runs it in a copy of the
checkout and runs the tests of what it wrote. `behavior-scaffold-check`,
`endpoint-scaffold-check` and `command-scaffold-check` are examples.

## Docs and skills

Each page of a DSL under `development/dsl/` comes from its `Spec`.
`cli docs dsl generate` writes them from the list in `dslPages` in
`cmd/commands.go`, and `cli-docs-gen-check` fails when a page differs from the
binary. A generated page starts with the `guide.Generated` comment, and the generator removes a page with the comment that it did not write. A hand-written page has no comment, and the generator keeps it.

A skill keeps only judgment: what to ask the person, and which way to select. It
sends the agent to `spec get` and `validate` for each fact. The skills are in
`internal/skills/resources/dev-skills/`, and `osapi skill get --name <skill>`
prints one.

## Records at run time

`behavior.Record` gives a context, and each `behavior.Call` under it adds a
`behavior.Step`. Today the examples and `osapi behavior chain get` use it, on
the fake host of the tests. A real job that records its steps is planned in
[osapi-io/osapi#1076](https://github.com/osapi-io/osapi/issues/1076). A new DSL
whose definitions run on the agent records its calls the same way.

## Extensions

A DSL that extensions use must let a definition register without a recompile of
osapi. The registry, the spec and the checks must not depend on generated code
in osapi, and the definition must be able to come from outside the module.
[osapi-io/osapi#1064](https://github.com/osapi-io/osapi/issues/1064) has the
design questions.
