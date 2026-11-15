# Shell

Run a command through `/bin/sh -c`. You can use shell features, for example
pipes, redirects and variable expansion.

```bash
$ osapi client node command shell --command "ls -la /tmp | grep log"

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  EXIT  STDOUT
  web-01    ok      0     -rw-r--r-- 1 root root 4096 ...

  1 host: 1 ok
```

The table view shortens long output. Use `--json` to get the full response.

Use shell syntax, for example pipes and redirects:

```bash
$ osapi client node command shell \
    --command "df -h / | tail -1 | awk '{print \$5}'" \
    --timeout 10
```

Run in a specific working directory:

```bash
$ osapi client node command shell \
    --command "cat *.conf | wc -l" \
    --cwd /etc
```

Target a label to run on a group of servers:

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

Use `--stdout` to print only the stdout of the remote command:

```bash
$ osapi client node command shell --command "df -h / | tail -1" --stdout
/dev/sda1        50G   12G   35G  26% /
```

Use `--stderr` to print only stderr:

```bash
$ osapi client node command shell --command "cat /nonexistent" --stderr
cat: /nonexistent: No such file or directory
```

You can use the two flags together. Each line starts with the hostname:

```bash
$ osapi client node command shell --command "uname -r" --target _all --stdout
[web-01] 5.15.0-91-generic
[web-02] 5.15.0-91-generic
```

The exit code of the CLI is the exit code of the remote command.

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
