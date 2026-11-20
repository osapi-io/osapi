# Get

Describe one endpoint group: its routes, permissions and operation IDs.

`--openapi` prints the group's part of the spec, as the definition makes it.
Read it beside the hand-written spec of the group to compare the prose.
`TestEndpointParity` compares the rest.

```bash
osapi controller endpoint get [flags]
```

## Examples

```bash
osapi controller endpoint get --name agent
osapi controller endpoint get --name agent --json
osapi controller endpoint get --name agent --openapi
```

## Flags

| Flag        | Description                                        | Default |
| ----------- | -------------------------------------------------- | ------- |
| `--name`    | The endpoint group, such as agent (required)       |         |
| `--openapi` | Print the group's part of the OpenAPI spec as YAML |         |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
