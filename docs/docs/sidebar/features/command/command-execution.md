# Command execution

osapi can run any command on managed hosts. The
[command behavior](../../reference/behaviors/command/command/command.md)
supplies the command execution operations. Command execution goes through the
[job system](../platform/job-system.md). Thus the API server does not run
commands. The agents run them.

## What it does

| Operation | Description                                            |
| --------- | ------------------------------------------------------ |
| Exec      | Run a command directly without a shell interpreter     |
| Shell     | Run a command through `/bin/sh -c` with shell features |

**Exec** runs the command binary directly, with an argument list. No shell reads
the command. Thus characters like `|`, `>` and `*` go to the command as they
are. This is the safer option for automation.

**Shell** gives the command string to `/bin/sh -c`. Then you can use pipes,
redirects, variable expansion and other shell features. Use it to find problems
quickly, or when you need shell syntax.

The two operations take a working directory and a timeout (30 seconds by
default, 300 seconds at most). The results have stdout, stderr, the exit code
and the duration. A command that exits with a code that is not zero is a result,
not an error. The API refuses a command that starts with `-`.

## How it works

Command execution uses the same request flow as all osapi operations:

1. The CLI or an API client sends a request to the API server.
1. The API server makes a job and publishes it to NATS.
1. An agent receives the job, executes the command, and writes the result back
   to NATS KV.
1. The API server collects the result and returns it to the client.

You can target one host, broadcast to all hosts with `_all`, or target a label.
Refer to [CLI Reference](../../usage/cli/client/node/command/command.mdx) for
usage and examples, or the
[API Reference](/gen/api/command-api-command-operations) for the REST endpoints.

## Fact references

Command arguments accept `@fact.*` references. The agent that runs the command
changes them to live system values. This is useful with broadcast targets, where
each agent puts in its own values:

```bash
$ osapi client node command exec \
    --command ip --args "addr,show,dev,@fact.interface.primary" \
    --target _all
```

Refer to [System Facts](../node/system-facts.md) for the full list of available
references.

## Use cases

- **Find problems quickly.** Check a process table, examine a log file, or
  verify a configuration on a remote host without SSH.
- **Fill a gap in automation.** Run a command that has no osapi endpoint yet.
- **Checks on all of the fleet.** Broadcast a command to all hosts to verify a
  package version, check the disk space, or make sure that a service runs.

## Security model

**Command execution is off by default.** The agent refuses each `exec` and
`shell` job until you set `agent.command.enabled: true`. The agent makes this
check, not the API. Thus a job that gets to the agent without the API is refused
too.

A command runs as the user of the agent. If the agent runs as root, each command
runs as root. Turn it on only on hosts where that is acceptable.

> **Upgrade note.** Before this release, command execution was always on. After
> an upgrade, `exec`, `shell` and container `exec` jobs fail until you set
> `agent.command.enabled: true` on each agent that must run them.

Command execution has two permissions. Thus a role can run commands without
shell access:

- `command:execute` guards the `exec` endpoint (no shell interpretation).
- `command:shell` guards the `shell` endpoint (`/bin/sh -c`, with pipes,
  redirects, and expansion).

Only the built-in `admin` role includes either permission by default. No other
built-in role does. `exec` is an admin-only action. `shell` has the `shell`
verb, and only the admin role gets a permission with that verb.

> **Upgrade note.** A custom role or token that has only `command:execute`
> cannot use the `shell` endpoint now. It also needs `command:shell`. Add
> `command:shell` to each custom role or token that must keep shell access.

To grant command execution to a custom role:

```yaml
controller:
  api:
    security:
      roles:
        ops:
          permissions:
            - command:execute
            - command:shell
            - status:read
            - health:read
```

Or grant it directly on a token:

```bash
osapi controller token generate -r read -u user@example.com \
  -p command:execute -p command:shell
```

## Configuration

Turn command execution on, in the config of each agent that must run commands:

```yaml
agent:
  command:
    enabled: true
```

When it is off, each `exec` and `shell` job fails with an error that names
`agent.command.enabled`. Refer to
[Configuration](../../operate/plan.md#the-config-file) for NATS, agent and
authentication settings.

## Permissions

The [command](../../reference/behaviors/command/command/command.md) reference
page gives the permission of each method.

## Related

- [CLI Reference](../../usage/cli/client/node/command/command.mdx): command
  execution commands (exec, shell)
- [System Facts](../node/system-facts.md): available `@fact.*` references
- [API Reference](/gen/api/command-api-command-operations): REST API
  documentation
- [Job System](../platform/job-system.md): how async job processing works
- [Authentication & RBAC](../../operate/secure/authentication.md): permissions
  and roles
- [Architecture](../../architecture/architecture.md): system design overview
