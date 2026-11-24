# Get

Show a role and each permission it grants.

```bash
osapi controller role get [flags]
```

## Examples

```bash
osapi controller role get --name admin
```

## Result

The command prints:

```ansi
$ osapi controller role get --name admin
[1;38;2;0;213;255mRole:[0m
  [38;2;10;143;176mName:[0m         admin
  [38;2;10;143;176mSource:[0m       built-in
  [38;2;10;143;176mPermissions:[0m  status:read, file:read, file:write, health:read
```

With `--json`, the command prints:

```json
{
  "name": "admin",
  "source": "built-in",
  "permissions": ["status:read", "file:read", "file:write", "health:read"]
}
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
