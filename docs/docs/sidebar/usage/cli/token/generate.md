# Generate

Generate a new token.

```bash
osapi token generate
```

## Examples

```bash
osapi token generate --subject user123 --roles=read
osapi token generate --subject svc@example.com --roles=read \
  -p node:read -p health:read
```

## Flags

| Flag                | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     | Default |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- |
| `-p, --permissions` | Direct permissions (overrides role expansion; allowed: agent:read, agent:write, node:read, node:write, network:read, network:write, job:read, job:write, health:read, audit:read, command:execute, command:shell, file:read, file:write, docker:read, docker:write, docker:execute, schedule:read, schedule:write, sysctl:read, sysctl:write, ntp:read, ntp:write, timezone:read, timezone:write, power:execute, process:read, process:execute, user:read, user:write, package:read, package:write, log:read, certificate:read, certificate:write, service:read, service:write) |         |
| `-r, --roles`       | Roles for the token (allowed: admin, write, read) (required)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |         |
| `-u, --subject`     | Subject for the token (e.g., user ID or unique identifier) (required)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |         |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../output.md) tells how the result is printed, and what `--json` gives.
