---
sidebar_position: 1
sidebar_label: Configuration
---

# Configuration

OSAPI is configured through a YAML file and optional environment variable
overrides.

## Config File

By default OSAPI looks for `/etc/osapi/osapi.yaml`. Override the path with the
`-f` / `--osapi-file` flag:

```bash
osapi -f /path/to/osapi.yaml controller start
```

## Every setting

The binary lists its own settings, so this page does not repeat them:

```bash
osapi config list                           # every key, its value, source and environment variable
osapi config list --json                    # the same, with each key's type, default, rules and description
osapi config get --key controller.api.port  # one key in full
```

Both read the file and the environment the way `osapi start` does, and work
without a config file. The source of each value is `env`, `file`, `default` or
`unset`. Passwords, signing keys and tokens print as `********`.

## Environment variables

Every key can be set with an environment variable instead: `OSAPI_`, then the
key in capitals with its dots as underscores. `controller.api.port` is
`OSAPI_CONTROLLER_API_PORT`. An environment variable wins over the file, and
applies even when the file does not mention the key.

## Validation

Configuration is checked before any component starts, and a bad value stops
osapi with the field that failed. Check a file without starting anything:

```bash
osapi config validate -f ./osapi.yaml
```

Two keys have no default and must be set:

| Key                                       | Purpose                       |
| ----------------------------------------- | ----------------------------- |
| `controller.api.security.signing_key`     | HS256 key for signing JWTs    |
| `controller.client.security.bearer_token` | JWT sent with client requests |

Generate a signing key with `openssl rand -hex 32`; it must be at least 32
characters. Generate a bearer token with `osapi token generate`. Durations are
Go durations, such as `30s`, `5m` or `168h`; a `d` suffix is not accepted.

## Authentication

Each NATS connection supports pluggable authentication. Set the `auth.type`
field in the relevant section (`nats.server`, `controller.nats`, or
`agent.nats`):

| Type        | Description                          | Extra Fields           |
| ----------- | ------------------------------------ | ---------------------- |
| `none`      | No authentication (default)          | —                      |
| `user_pass` | Username and password                | `username`, `password` |
| `nkey`      | NKey-based auth (server and clients) | See examples below     |

### Server-Side Auth

The embedded NATS server (`nats.server.auth`) accepts a list of users or nkeys:

```yaml
nats:
  server:
    auth:
      type: user_pass
      users:
        - username: osapi
          password: '<secret>'
```

### Client-Side Auth

Controller and agent connections (`controller.nats.auth`, `agent.nats.auth`)
authenticate as a single identity:

```yaml
controller:
  nats:
    auth:
      type: user_pass
      username: osapi
      password: '<secret>'
```

## Permissions

OSAPI uses fine-grained `resource:verb` permissions for access control. Each API
endpoint requires a specific permission. Built-in roles expand to a default set
of permissions:

| Role    | Permissions                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| ------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `admin` | `agent:read`, `agent:write`, `node:read`, `node:write`, `network:read`, `network:write`, `job:read`, `job:write`, `health:read`, `audit:read`, `command:execute`, `command:shell`, `file:read`, `file:write`, `docker:read`, `docker:write`, `docker:execute`, `schedule:read`, `schedule:write`, `sysctl:read`, `sysctl:write`, `ntp:read`, `ntp:write`, `timezone:read`, `timezone:write`, `power:execute`, `process:read`, `process:execute`, `user:read`, `user:write`, `package:read`, `package:write`, `log:read`, `certificate:read`, `certificate:write`, `service:read`, `service:write` |
| `write` | `agent:read`, `node:read`, `node:write`, `network:read`, `network:write`, `job:read`, `job:write`, `health:read`, `file:read`, `file:write`, `docker:read`, `docker:write`, `schedule:read`, `schedule:write`, `sysctl:read`, `sysctl:write`, `ntp:read`, `ntp:write`, `timezone:read`, `timezone:write`, `process:read`, `user:read`, `user:write`, `package:read`, `package:write`, `log:read`, `certificate:read`, `certificate:write`, `service:read`, `service:write`                                                                                                                        |
| `read`  | `agent:read`, `node:read`, `network:read`, `job:read`, `health:read`, `file:read`, `docker:read`, `schedule:read`, `sysctl:read`, `ntp:read`, `timezone:read`, `process:read`, `user:read`, `package:read`, `log:read`, `certificate:read`, `service:read`                                                                                                                                                                                                                                                                                                                                        |

### Custom Roles

You can define custom roles in the `controller.api.security.roles` section.
Custom roles override the default permission mapping for the same name, or
define entirely new role names:

```yaml
controller:
  api:
    security:
      roles:
        ops:
          permissions:
            - node:read
            - health:read
        netadmin:
          permissions:
            - network:read
            - network:write
            - health:read
```

### Direct Permissions

Tokens can carry a `permissions` claim that overrides role-based expansion. When
the claim is present, only the listed permissions are granted regardless of the
token's roles. Generate a token with direct permissions:

```bash
osapi token generate -r admin -u user@example.com \
  -p node:read -p health:read
```

## Namespace

The `namespace` field on NATS connections prefixes all subject names and
infrastructure names. This allows multiple OSAPI deployments to share a single
NATS cluster without collisions.

| Without namespace       | With `namespace: osapi` |
| ----------------------- | ----------------------- |
| `jobs.query._any`       | `osapi.jobs.query._any` |
| `JOBS` (stream)         | `osapi-JOBS`            |
| `job-queue` (KV bucket) | `osapi-job-queue`       |

Set the same namespace value in `nats.server.namespace`,
`controller.nats.namespace`, and `agent.nats.namespace` so all components agree
on naming. An empty string disables prefixing.
