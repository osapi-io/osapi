# Get

Describe one endpoint group: its routes, permissions and operation IDs.

`osapi controller endpoint openapi get` prints the group's part of the OpenAPI
spec.

```bash
osapi controller endpoint get [flags]
```

## Examples

```bash
osapi controller endpoint get --name agent
osapi controller endpoint get --name agent --json
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

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
