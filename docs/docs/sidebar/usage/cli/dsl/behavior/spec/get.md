# Get

Print the behavior DSL: each option, method, example part and field tag.

Print each part of the behavior DSL, what it makes osapi do and when to use it.
The list comes from this binary. Thus it agrees with the build that you run.
Read it before you write a behavior, then check the result with
`osapi dsl behavior validate`.

```bash
osapi dsl behavior spec get
```

## Examples

```bash
osapi dsl behavior spec get
osapi dsl behavior spec get --json
```

## Result

The command prints:

```ansi
$ osapi dsl behavior spec get
[1;38;2;0;213;255mDefinition:[0m
  [1mRegister(Define(name, doc, options...))[0m
    Declares the behavior and adds it to the registry. The agent, the
    server, the CLI and the generator read the registry.
    [38;2;138;138;138mUse: One time, in behavior.go: var Behavior =[0m
    [38;2;138;138;138mbehavior.Register(behavior.Define(...)).[0m
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
          "does": "Declares the behavior and adds it to the registry. The agent, the server, the CLI and the generator read the registry.",
          "when": "One time, in behavior.go: var Behavior = behavior.Register(behavior.Define(...))."
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
