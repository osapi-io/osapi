# List

List the roles and the permissions each grants.

The list has the permissions of each behavior, so it is correct for the binary
that you run. A custom role from the config's `controller.api.security.roles`
gets a column of its own. Without a config file it shows the built-in roles
only. A behavior's permissions are `<name>:read` and `<name>:write`, or the
group's when it is nested under one. `admin` and `write` grant both, and `read`
grants the read permission. Use this list, not a table in the docs.

```bash
osapi token roles list
```

## Examples

```bash
osapi token roles list
osapi token roles list --json
```

## Result

The command prints:

```ansi
$ osapi token roles list
[1;38;2;0;213;255mRoles (4):[0m
  [38;2;10;143;176mNAME     [0m[38;2;10;143;176mSOURCE[0m
  [1madmin[0m    built-in
  [1mwrite[0m    built-in
  [1mread[0m     built-in
  [1mdns-ops[0m  config
```

With `--json`, the command prints:

```json
{
  "roles": [
    {
      "name": "admin",
      "source": "built-in",
      "permissions": null
    },
    {
      "name": "write",
      "source": "built-in",
      "permissions": null
    },
    {
      "name": "read",
      "source": "built-in",
      "permissions": null
    },
    {
      "name": "dns-ops",
      "source": "config",
      "permissions": null
    }
  ]
}
```

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
