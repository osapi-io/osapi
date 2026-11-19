# Validate

Validate a token for authenticity and claims.

```bash
osapi token validate [flags]
```

## Examples

```bash
osapi token validate --token eyJhbGciOiJIUzI1NiIs...
```

## Flags

| Flag          | Description                 | Default |
| ------------- | --------------------------- | ------- |
| `-t, --token` | The Token string (required) |         |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../output.md) tells how the result is printed, and what `--json` gives.
