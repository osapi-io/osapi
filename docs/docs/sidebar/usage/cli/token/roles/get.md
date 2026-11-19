# Get

Show a role and each permission it grants.

```bash
osapi token roles get [flags]
```

## Examples

```bash
osapi token roles get --name admin
```

## Flags

| Flag     | Description                                                        | Default |
| -------- | ------------------------------------------------------------------ | ------- |
| `--name` | The role, such as admin, or a custom role of the config (required) |         |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
