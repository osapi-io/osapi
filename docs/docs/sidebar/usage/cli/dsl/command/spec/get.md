# Get

Print the command DSL: each option, method, refinement, field tag and check.

Print each part of the command DSL, what it makes osapi do and when to use it.
The list comes from this binary. Thus it agrees with the build that you run.
Read it before you write a local group, then check the result with
`osapi dsl command validate`.

```bash
osapi dsl command spec get
```

## Examples

```bash
osapi dsl command spec get
osapi dsl command spec get --json
```

## Result

The command prints:

```ansi
$ osapi dsl command spec get
[1;38;2;0;213;255mDefinition:[0m
  [1mRegister(Define(name, doc, options...))[0m
    Declares a group of local commands and adds it to the registry. A local
    command runs in the CLI process and calls no API, such as osapi
    controller token generate. Its command is osapi <name>, and doc is its
    help.
    [38;2;138;138;138mUse: One time, in <group>.go under internal/cli/commands/<group>: var[0m
    [38;2;138;138;138mGroup = command.Register(command.Define(...)). Add its import to[0m
    [38;2;138;138;138minternal/cli/commands/all. A command that calls the API is an endpoint[0m
    [38;2;138;138;138mor a behavior, not a local command.[0m
```

With `--json`, the command prints:

```json
{
  "kinds": [
    {
      "kind": "definition",
      "title": "Definition",
      "terms": [
        {
          "kind": "definition",
          "name": "Register(Define(name, doc, options...))",
          "does": "Declares a group of local commands and adds it to the registry. A local command runs in the CLI process and calls no API, such as osapi controller token generate. Its command is osapi \u003cname\u003e, and doc is its help.",
          "when": "One time, in \u003cgroup\u003e.go under internal/cli/commands/\u003cgroup\u003e: var Group = command.Register(command.Define(...)). Add its import to internal/cli/commands/all. A command that calls the API is an endpoint or a behavior, not a local command."
        }
      ]
    }
  ]
}
```

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
