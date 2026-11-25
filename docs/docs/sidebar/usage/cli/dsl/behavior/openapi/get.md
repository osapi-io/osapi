# Get

Print a served behavior's part of the OpenAPI spec, as YAML.

It prints the spec as the definition makes it: the part that go generate writes
to the behavior's gen/api.yaml, and puts in the full spec of the API. An
internal behavior serves no route, so it has none.

```bash
osapi dsl behavior openapi get [flags]
```

## Examples

```bash
osapi dsl behavior openapi get --name sysctl
```

## Result

The command prints:

```ansi
$ osapi dsl behavior openapi get --name sysctl
paths:
  /api/node/{hostname}/sysctl/{key}:
    put:
      operationId: PutNodeSysctl
      summary: Set a kernel parameter, manage it, and apply it
      tags:
        - sysctl_operations
```

With `--json`, the command prints:

```json
{
  "name": "sysctl",
  "yaml": "paths:\n  /api/node/{hostname}/sysctl/{key}:\n    put:\n      operationId: PutNodeSysctl\n      summary: Set a kernel parameter, manage it, and apply it\n      tags:\n        - sysctl_operations\n"
}
```

## Flags

| Flag     | Description                                           | Default |
| -------- | ----------------------------------------------------- | ------- |
| `--name` | The behavior, from osapi dsl behavior list (required) |         |

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
