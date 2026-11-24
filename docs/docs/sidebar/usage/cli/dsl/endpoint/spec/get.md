# Get

Print the endpoint DSL: each option, method, refinement, error and field tag.

Print each part of the endpoint DSL, what it makes osapi do and when to use it,
and what it cannot do yet. The list comes from this binary. Thus it agrees with
the build that you run. Read it before you write a group, then check the result
with `osapi dsl endpoint validate`.

```bash
osapi dsl endpoint spec get
```

## Examples

```bash
osapi dsl endpoint spec get
osapi dsl endpoint spec get --json
```

## Result

The command prints:

```ansi
$ osapi dsl endpoint spec get
[1;38;2;0;213;255mDefinition:[0m
  [1mRegister(Define(name, doc, options...))[0m
    Declares the group and adds it to the registry. The commands and the
    tests read the registry. Each path of the group starts /api/<name>.
    [38;2;138;138;138mUse: One time, in <group>.go under[0m
    [38;2;138;138;138minternal/controller/endpoints/<group>: var Group =[0m
    [38;2;138;138;138mendpoint.Register(endpoint.Define(...)). Add its import to[0m
    [38;2;138;138;138minternal/controller/endpoints/all.[0m
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
          "does": "Declares the group and adds it to the registry. The commands and the tests read the registry. Each path of the group starts /api/\u003cname\u003e.",
          "when": "One time, in \u003cgroup\u003e.go under internal/controller/endpoints/\u003cgroup\u003e: var Group = endpoint.Register(endpoint.Define(...)). Add its import to internal/controller/endpoints/all."
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

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
