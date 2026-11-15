# Shell

Execute a shell command.

You can use shell features, for example pipes, redirects and variable expansion.
The table view shortens long output. Use `--json` to get the full response. Use
`--stdout` or `--stderr` to print only that output of the remote command. The
exit code of the CLI is the exit code of the remote command. Thus you can use it
in scripts.

```bash
osapi client node command shell
```

## Examples

```bash
osapi client node command shell --command "ls -la /tmp | grep log"
osapi client node command shell \
  --command "df -h / | tail -1 | awk '{print \$5}'" \
  --timeout 10
osapi client node command shell \
  --command "cat *.conf | wc -l" \
  --cwd /etc
```

## Flags

| Flag        | Description                              | Default |
| ----------- | ---------------------------------------- | ------- |
| `--command` | The shell command to execute (required)  |         |
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
