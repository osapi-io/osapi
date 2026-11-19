# Get

Describe one local group: its commands, their flags and the inputs they refuse.

`--name` is the command path of the group, such as `token`. Each flag has the
validate tag of its field, and each command has the inputs that its `.Fails`
declares, which `TestRefusals` runs.

```bash
osapi cli command get [flags]
```

## Examples

```bash
osapi cli command get --name token
osapi cli command get --name token --json
```

## Flags

| Flag     | Description                                                   | Default |
| -------- | ------------------------------------------------------------- | ------- |
| `--name` | The command path of the local group, such as token (required) |         |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
