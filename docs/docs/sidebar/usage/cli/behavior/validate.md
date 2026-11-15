# Validate

Check every behavior against the rules the platform relies on.

Exits 1 when anything fails, naming the behavior, the method and the rule. Each
problem in `--json` has `behavior`, `method` (when the problem is in one),
`rule` and `message`. [The checks](../../../development/behaviors.md#the-checks)
lists the rules.

```bash
osapi behavior validate
```

## Examples

```bash
osapi behavior validate
osapi behavior validate --json
```

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../output.md) tells how the result is printed, and what `--json` gives.
