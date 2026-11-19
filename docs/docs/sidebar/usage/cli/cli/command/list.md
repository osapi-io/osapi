# List

List each command of this binary, and the layer that made it.

The layer is `command` for a command of a local group, `endpoint` for a command
of an endpoint group, `behavior` for a command of a behavior, and `hand-written`
for a command in `cmd/` that no definition makes yet.

```bash
osapi cli command list
```

## Examples

```bash
osapi cli command list
osapi cli command list --json
```

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
