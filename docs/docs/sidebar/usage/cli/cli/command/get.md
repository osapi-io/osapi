# Get

Describe one group: its commands, and the flags and refusals of a local group.

`--name` is the command path of the group, such as `token` or `agent`: a local
group, or the group of an endpoint or a behavior. Each command prints from the
output type of its method. A command of a local group also has its flags, each
with the validate tag of its field, and the inputs that its `.Fails` declares,
which `TestRefusals` runs.

```bash
osapi cli command get [flags]
```

## Examples

```bash
osapi cli command get --name token
osapi cli command get --name agent --json
```

## Flags

| Flag     | Description                                                      | Default |
| -------- | ---------------------------------------------------------------- | ------- |
| `--name` | The command path of the group, such as token or agent (required) |         |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
