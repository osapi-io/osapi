# Get

Print the group's part of the OpenAPI spec, as YAML.

It prints the spec as the definition makes it. Read it beside the hand-written
spec of the group to compare the prose. `TestEndpointParity` compares the rest.

```bash
osapi dsl endpoint openapi get [flags]
```

## Examples

```bash
osapi dsl endpoint openapi get --name agent
```

## Result

The command prints:

```ansi
$ osapi dsl endpoint openapi get --name agent
paths:
  /api/agent/{hostname}/drain:
    post:
      operationId: DrainAgent
      summary: Drain an agent
      tags:
        - agent_operations
```

With `--json`, the command prints:

```json
{
  "name": "agent",
  "yaml": "paths:\n  /api/agent/{hostname}/drain:\n    post:\n      operationId: DrainAgent\n      summary: Drain an agent\n      tags:\n        - agent_operations\n"
}
```

## Flags

| Flag     | Description                                  | Default |
| -------- | -------------------------------------------- | ------- |
| `--name` | The endpoint group, such as agent (required) |         |

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
