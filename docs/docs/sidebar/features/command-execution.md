---
sidebar_position: 3
---

# Command Execution

OSAPI can execute arbitrary commands on managed hosts. Command execution runs
through the [job system](job-system.md), so the API server never runs commands
directly -- agents handle all execution.

## What It Does

| Operation | Description                                            |
| --------- | ------------------------------------------------------ |
| Exec      | Run a command directly without a shell interpreter     |
| Shell     | Run a command through `/bin/sh -c` with shell features |

**Exec** invokes the command binary directly with an explicit argument list. No
shell interpretation occurs, so metacharacters like `|`, `>`, and `*` are passed
literally. This is the safer option for structured automation.

**Shell** passes the command string to `/bin/sh -c`, enabling pipes, redirects,
variable expansion, and other shell features. Use this for ad-hoc debugging or
when you need shell syntax.

Both operations support a configurable working directory and a timeout (default
30 seconds, maximum 300 seconds). Results include stdout, stderr, the exit code,
and execution duration.

## How It Works

Command execution follows the same request flow as all OSAPI operations:

1. The CLI (or API client) posts a request to the API server.
2. The API server creates a job and publishes it to NATS.
3. An agent picks up the job, executes the command, and writes the result back
   to NATS KV.
4. The API server collects the result and returns it to the client.

You can target a specific host, broadcast to all hosts with `_all`, or route by
label. When targeting `_all`, the CLI prompts for confirmation before
proceeding. See [CLI Reference](../usage/cli/client/node/command/command.mdx)
for usage and examples, or the
[API Reference](/gen/api/command-execution-api-command-operations) for the REST
endpoints.

## Use Cases

- **Ad-hoc debugging** -- quickly check a process table, inspect a log file, or
  verify a configuration on a remote host without SSH.
- **Automation fallback** -- run a one-off command that does not yet have a
  dedicated OSAPI endpoint.
- **Fleet-wide checks** -- broadcast a command to all hosts to verify a package
  version, check disk space, or confirm a service is running.

## Security Model

Command execution is a privileged operation. The `command:execute` permission is
required for both `exec` and `shell` endpoints. Only the built-in `admin` role
includes this permission by default. The `write` and `read` roles do not.

To grant command execution to a custom role:

```yaml
api:
  server:
    security:
      roles:
        ops:
          permissions:
            - command:execute
            - node:read
            - health:read
```

Or grant it directly on a token:

```bash
osapi token generate -r read -u user@example.com \
  -p command:execute
```

## Configuration

Command execution uses the general job infrastructure. No domain-specific
configuration is required. See [Configuration](../usage/configuration.md) for
NATS, agent, and authentication settings.

## Permissions

| Operation | Permission        |
| --------- | ----------------- |
| Exec      | `command:execute` |
| Shell     | `command:execute` |

Only the `admin` role includes `command:execute` by default. Grant it to other
roles or tokens explicitly when needed.

## Related

- [CLI Reference](../usage/cli/client/node/command/command.mdx) -- command
  execution commands (exec, shell)
- [API Reference](/gen/api/command-execution-api-command-operations) -- REST API
  documentation
- [Job System](job-system.md) -- how async job processing works
- [Authentication & RBAC](authentication.md) -- permissions and roles
- [Architecture](../architecture/architecture.md) -- system design overview
