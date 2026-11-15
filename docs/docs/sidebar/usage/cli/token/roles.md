# Roles

List the roles and the permissions each grants.

The list has the permissions of each behavior, so it is correct for the binary
that you run. A custom role from the config's `controller.api.security.roles`
gets a column of its own. Without a config file it shows the built-in roles
only. A behavior's permissions are `<name>:read` and `<name>:write`, or the
group's when it is nested under one. `admin` and `write` grant both, and `read`
grants the read permission. Use this list, not a table in the docs.

```bash
osapi token roles
```

## Examples

```bash
osapi token roles
osapi token roles --json
```

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../output.md) tells how the result is printed, and what `--json` gives.
