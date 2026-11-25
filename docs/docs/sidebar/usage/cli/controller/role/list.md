# List

List the roles and the permissions each grants.

The list has the permissions of each behavior, so it is correct for the binary
that you run. A custom role from the config's `controller.api.security.roles`
gets a column of its own. Without a config file it shows the built-in roles
only. A behavior's permissions are `<name>:<verb>`, such as `dns:read` and
`dns:write`. A rule gives each built-in role its permissions. Use this list, not
a table in the docs.

```bash
osapi controller role list
```

## Examples

```bash
osapi controller role list
osapi controller role list --json
```

## Result

The command prints:

```ansi
$ osapi controller role list
[1;38;2;0;213;255mRoles (7):[0m
  [38;2;10;143;176mNAME        [0m[38;2;10;143;176mSOURCE[0m
  [1madmin[0m       built-in
  [1moperator[0m    built-in
  [1mautomation[0m  built-in
  [1mauditor[0m     built-in
  [1mread[0m        built-in
  [1mobserver[0m    built-in
  [1mdns-ops[0m     config
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
      "name": "operator",
      "source": "built-in",
      "permissions": null
    },
    {
      "name": "automation",
      "source": "built-in",
      "permissions": null
    },
    {
      "name": "auditor",
      "source": "built-in",
      "permissions": null
    },
    {
      "name": "read",
      "source": "built-in",
      "permissions": null
    },
    {
      "name": "observer",
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

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
