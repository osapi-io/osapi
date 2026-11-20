# List

List the behaviors this binary has.

An internal behavior has no API or CLI of its own. Only other behaviors use it.

```bash
osapi behavior list
```

## Examples

```bash
osapi behavior list
osapi behavior list --json
```

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../output.md) tells how the result is printed, and what `--json` gives.
