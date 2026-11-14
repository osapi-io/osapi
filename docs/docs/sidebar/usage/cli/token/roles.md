# Roles

List the roles and every permission each grants, as this build of osapi has
them. The list includes each behavior's permissions, so it is current for the
binary you run; this is the reference rather than a table in the docs:

```bash
$ osapi token roles

Permissions (37):
  PERMISSION         ADMIN  WRITE  READ
  agent:read         yes    yes    yes
  agent:write        yes
  node:read          yes    yes    yes
  ...
  sysctl:read        yes    yes    yes
  sysctl:write       yes    yes
```

A custom role from the config's `controller.api.security.roles` gets a column of
its own. Without a config file it shows the built-in roles only.

A behavior's permissions are `<name>:read` and `<name>:write`, or the group's
when it is nested under one. `admin` and `write` grant both, and `read` grants
the read permission.

```bash
$ osapi token roles --json
{"roles":[{"name":"admin","source":"built-in","permissions":["agent:read",...]}]}
```

## Flags

| Flag     | Description             | Default |
| -------- | ----------------------- | ------- |
| `--json` | Print the roles as JSON |         |
