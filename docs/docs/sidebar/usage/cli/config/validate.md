# Validate

Check a config file before osapi starts with it.

The command starts nothing. Exits 1 when the config is not valid.

```bash
osapi config validate
```

## Examples

```bash
osapi config validate
osapi config validate -f ./osapi.yaml --json
```

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../output.md) tells how the result is printed, and what `--json` gives.
