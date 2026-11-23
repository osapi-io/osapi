# Get

Print the behavior DSL: each option, method, example part and field tag.

Print each part of the behavior DSL, what it makes osapi do and when to use it.
The list comes from this binary. Thus it agrees with the build that you run.
Read it before you write a behavior, then check the result with
`osapi behavior validate`.

```bash
osapi behavior spec get
```

## Examples

```bash
osapi behavior spec get
osapi behavior spec get --json
```

## Result

The command prints:

```ansi
$ osapi behavior spec get
[1;38;2;0;213;255mDefinition:[0m
  Register(Define(name, doc, options...))
    Declares the behavior and adds it to the registry. The agent, the
    server, the CLI and the generator read the registry.
    Use: One time, in behavior.go: var Behavior =
    behavior.Register(behavior.Define(...)).
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

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
