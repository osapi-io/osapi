---
sidebar_position: 1
sidebar_label: Configuration
---

# Configuration

A YAML file configures osapi. As an option, environment variables can replace
values of the file.

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

An environment variable can set each key: `OSAPI_`, then the key in capitals,
with underscores for its dots. `controller.api.port` is
`OSAPI_CONTROLLER_API_PORT`. An environment variable has priority over the file.
It applies also when the file does not have the key.

## Validation

osapi checks the configuration before a component starts. A bad value stops
osapi, with the field that failed. To check a file and start nothing:

```bash
osapi config validate -f ./osapi.yaml
```

Two keys have no default. You must set them:

| Key                                       | Purpose                       |
| ----------------------------------------- | ----------------------------- |
| `controller.api.security.signing_key`     | HS256 key for signing JWTs    |
| `controller.client.security.bearer_token` | JWT sent with client requests |

Make a signing key with `openssl rand -hex 32`. It must have 32 characters or
more. Make a bearer token with `osapi token generate`. Durations are Go
durations, for example `30s`, `5m` or `168h`. osapi does not accept a `d`
suffix.

Each bucket name and each namespace must be a name that NATS accepts for a
bucket: letters, digits, `-` and `_` only. Two settings must not name the same
bucket, because each role keeps its own keys in it. For example, when
`nats.kv.bucket` and `nats.registry.bucket` are both `job-queue`, osapi stops
with the two keys and the name. An empty bucket turns its role off, and osapi
does not compare it.

## Authentication

Each NATS connection supports pluggable authentication. Set the `auth.type`
field in the relevant section (`nats.server`, `controller.nats`, or
`agent.nats`):

| Type        | Description                          | Extra Fields                  |
| ----------- | ------------------------------------ | ----------------------------- |
| `none`      | No authentication (default)          | —                             |
| `user_pass` | Username and password                | `username`, `password`        |
| `nkey`      | NKey-based auth (server and clients) | See examples below            |
| `creds`     | A NATS credentials file (clients)    | `creds_file`                  |
| `operator`  | NATS operator mode (the server)      | `resolver_file`, `creds_file` |

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

With `type: nkey`, set `nkey_file` to a seed whose public key is in `nkeys`. The
setup and the heartbeat of the server connect with it, because they must
authenticate like all other clients.

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

With a NATS server in operator (decentralized JWT) mode, each process
authenticates with its own credentials file. The file has a user JWT, which the
account signs, and the seed of the user. The JWT has the permissions and the
expiry of the user. Thus, to add a host, you give it a credential. You do not
change the server.

`osapi nats setup` connects with the credentials of the controller, and makes
the streams and buckets. Thus the account of those credentials must have
JetStream on, and the user must have permission to make streams and key-value
buckets. The agents need only their own credentials.

```yaml
agent:
  nats:
    auth:
      type: creds
      creds_file: /etc/osapi/agent.creds
```

The embedded server runs in operator mode with `type: operator`. Its
`resolver_file` has the JWT of the operator, the system account and the JWT of
each account, as `nsc generate config --mem-resolver` writes them. Its
`creds_file` is the user that its setup and heartbeat connect as:

```yaml
nats:
  server:
    auth:
      type: operator
      resolver_file: /etc/osapi/resolver.conf
      creds_file: /etc/osapi/nats.creds
```

Against a NATS server osapi does not run,
[`osapi nats setup`](cli/nats/setup.md) creates osapi's streams and buckets with
the controller's connection.

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

A token can have a `permissions` claim. It replaces the permissions of the
roles. When the token has the claim, it gets only the listed permissions,
whatever its roles are. Make a token with direct permissions:

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

Set the same namespace in `nats.server.namespace`, `controller.nats.namespace`
and `agent.nats.namespace`, so that all components use the same names. An empty
string disables the prefix.
