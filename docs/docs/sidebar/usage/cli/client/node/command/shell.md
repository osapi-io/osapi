# Shell

Execute a command through `/bin/sh -c`. Supports shell features like pipes,
redirects, and variable expansion.

```bash
$ osapi client node command shell --command "ls -la /tmp | grep log"

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  CHANGED  STDOUT                          STDERR  EXIT CODE  DURATION
  false    -rw-r--r-- 1 root root 4096 …           0          15ms
```

Long output is truncated in the table view. Use `--json` for the full response
data.

Use shell syntax like pipes and redirects:

```bash
$ osapi client node command shell \
    --command "df -h / | tail -1 | awk '{print \$5}'" \
    --timeout 10
```

Execute in a specific working directory:

```bash
$ osapi client node command shell \
    --command "cat *.conf | wc -l" \
    --cwd /etc
```

Target by label to execute on a group of servers:

```bash
$ osapi client node command shell \
    --command "systemctl is-active nginx" \
    --target group:web
```

## JSON Output

Use `--json` to get the full untruncated API response:

```bash
$ osapi client node command shell --command "uname -r" --json
```

## Raw Output

Use `--stdout` to print only the remote command's stdout:

```bash
$ osapi client node command shell --command "df -h / | tail -1" --stdout
/dev/sda1        50G   12G   35G  26% /
```

Use `--stderr` to print only stderr:

```bash
$ osapi client node command shell --command "cat /nonexistent" --stderr
cat: /nonexistent: No such file or directory
```

Both flags can be combined. Each line is prefixed with the hostname:

```bash
$ osapi client node command shell --command "uname -r" --target _all --stdout
[web-01] 5.15.0-91-generic
[web-02] 5.15.0-91-generic
```

The CLI exit code matches the remote command's exit code.

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `--command`    | The shell command to execute (**required**)              |         |
| `--cwd`        | Working directory for the command                        |         |
| `--timeout`    | Timeout in seconds (max 300)                             | `30`    |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_all`  |
| `--stdout`     | Print only remote stdout                                 |         |
| `--stderr`     | Print only remote stderr                                 |         |
| `-j, --json`   | Output raw JSON response                                 |         |
