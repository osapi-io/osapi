# Exec

Execute a command directly without a shell interpreter. Arguments are passed to
the executable as-is, without shell expansion or interpretation.

```bash
$ osapi client node command exec --command ls --args "-la,/tmp"

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  CHANGED  STDOUT                          STDERR  EXIT CODE  DURATION
  server1   false    total 8 drwxrwxrwt 10 root r…           0          12ms
```

Long output is truncated in the table view. Use `--json` for the full response
data.

Execute a command in a specific working directory with a custom timeout:

```bash
$ osapi client node command exec \
    --command cat \
    --args "config.yaml" \
    --cwd /etc/osapi \
    --timeout 10
```

When targeting all hosts, the CLI prompts for confirmation:

```bash
$ osapi client node command exec --command uptime --target _all

  This will execute command on ALL hosts. Continue? [y/N] y

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  CHANGED  STDOUT                          STDERR  EXIT CODE  DURATION
  server1   false    13:21:06 up 42 days, 3:15, …            0          8ms
  server2   false    13:21:06 up 15 days, 1:02, …            0          11ms
```

Target by label to execute on a group of servers:

```bash
$ osapi client node command exec --command whoami --target group:web
```

## JSON Output

Use `--json` to get the full untruncated API response:

```bash
$ osapi client node command exec --command hostname --json
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `--command`    | The command to execute (**required**)                    |         |
| `--args`       | Command arguments (comma-separated)                      | `[]`    |
| `--cwd`        | Working directory for the command                        |         |
| `--timeout`    | Timeout in seconds (max 300)                             | `30`    |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
| `-j, --json`   | Output raw JSON response                                 |         |
