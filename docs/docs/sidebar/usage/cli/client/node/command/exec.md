# Exec

Execute a command directly.

The executable gets the arguments as they are, without shell expansion. The
table view shortens long output. Use `--json` to get the full response. Use
`@fact.*` references to put live system values in the command. Refer to
[System Facts](../../../../../features/system-facts.md) for all `@fact.*`
references. Use `--stdout` or `--stderr` to print only that output of the remote
command. The exit code of the CLI is the exit code of the remote command. Thus
you can use it in scripts. Each agent resolves its own facts, so `@fact.*` works
with a broadcast target.

```bash
osapi client node command exec
```

## Examples

```bash
osapi client node command exec --command ls --args "-la,/tmp"
osapi client node command exec \
  --command cat \
  --args "config.yaml" \
  --cwd /etc/osapi \
  --timeout 10
osapi client node command exec --command whoami --target group:web
```

## Flags

| Flag        | Description                              | Default |
| ----------- | ---------------------------------------- | ------- |
| `--args`    | Command arguments (comma-separated)      |         |
| `--command` | The command to execute (required)        |         |
| `--cwd`     | Working directory for the command        |         |
| `--stderr`  | Print only remote stderr                 |         |
| `--stdout`  | Print only remote stdout                 |         |
| `--timeout` | Timeout in seconds (default 30, max 300) | `30`    |

## Global flags

| Flag               | Description                                            | Default                 |
| ------------------ | ------------------------------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                           |                         |
| `-j, --json`       | Enable JSON output                                     |                         |
| `-f, --osapi-file` | Path to config file                                    | `/etc/osapi/osapi.yaml` |
| `-T, --target`     | Target: _all, _any, hostname, or label (group:web.dev) | `_all`                  |
| `--url`            | URL the client will connect to                         | `http://0.0.0.0:8080`   |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
