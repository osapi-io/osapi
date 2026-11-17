# Validate

Check every endpoint group against the rules of a definition.

It exits 1 when a rule fails, and names the group, the method, the rule and the
fix.

```bash
osapi endpoint validate
```

## Examples

```bash
osapi endpoint validate
osapi endpoint validate --json
```

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../output.md) tells how the result is printed, and what `--json` gives.
