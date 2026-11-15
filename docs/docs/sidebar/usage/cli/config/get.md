# Get

Show one setting in full.

```bash
osapi config get [flags]
```

## Examples

```bash
osapi config get --key controller.api.port
osapi config get --key agent.labels -f ./osapi.yaml --json
```

## Flags

| Flag    | Description                                                 | Default |
| ------- | ----------------------------------------------------------- | ------- |
| `--key` | The setting to show, such as controller.api.port (required) |         |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../output.md) tells how the result is printed, and what `--json` gives.
