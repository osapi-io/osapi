# Get

Print the group's part of the OpenAPI spec, as YAML.

It prints the spec as the definition makes it. Read it beside the hand-written
spec of the group to compare the prose. `TestEndpointParity` compares the rest.

```bash
osapi controller endpoint openapi get [flags]
```

## Examples

```bash
osapi controller endpoint openapi get --name agent
```

## Flags

| Flag     | Description                                  | Default |
| -------- | -------------------------------------------- | ------- |
| `--name` | The endpoint group, such as agent (required) |         |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
