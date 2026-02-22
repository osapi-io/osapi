# Shell

Execute a command through `/bin/sh -c`. Supports shell features like
pipes, redirects, and variable expansion.

```bash
$ osapi client command shell --command "ls -la /tmp | grep log"

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━┳━━━━━━━━━━━┳━━━━━━━━━━┓
  ┃ STDOUT                        ┃ STDERR ┃ EXIT CODE ┃ DURATION ┃
  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╋━━━━━━━━╋━━━━━━━━━━━╋━━━━━━━━━━┫
  ┃ -rw-r--r-- 1 root root 4096  ┃        ┃ 0         ┃ 15ms     ┃
  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┻━━━━━━━━┻━━━━━━━━━━━┻━━━━━━━━━━┛
```

Use shell syntax like pipes and redirects:

```bash
$ osapi client command shell \
    --command "df -h / | tail -1 | awk '{print \$5}'" \
    --timeout 10
```

Execute in a specific working directory:

```bash
$ osapi client command shell \
    --command "cat *.conf | wc -l" \
    --cwd /etc
```

When targeting all hosts, the CLI prompts for confirmation:

```bash
$ osapi client command shell --command "hostname -f" --target _all

  This will execute shell command on ALL hosts. Continue? [y/N] y

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  ┏━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━┳━━━━━━━━━━━┳━━━━━━━━━━┓
  ┃ HOSTNAME ┃ STDOUT                ┃ STDERR ┃ EXIT CODE ┃ DURATION ┃
  ┣━━━━━━━━━━╋━━━━━━━━━━━━━━━━━━━━━━━╋━━━━━━━━╋━━━━━━━━━━━╋━━━━━━━━━━┫
  ┃ server1  ┃ server1.example.com   ┃        ┃ 0         ┃ 5ms      ┃
  ┃ server2  ┃ server2.example.com   ┃        ┃ 0         ┃ 7ms      ┃
  ┗━━━━━━━━━━┻━━━━━━━━━━━━━━━━━━━━━━━┻━━━━━━━━┻━━━━━━━━━━━┻━━━━━━━━━━┛
```

Target by label to execute on a group of servers:

```bash
$ osapi client command shell \
    --command "systemctl is-active nginx" \
    --target group:web
```

## JSON Output

Use `--json` to get the raw API response:

```bash
$ osapi client command shell --command "uname -r" --json
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `--command`    | The shell command to execute (**required**)               |         |
| `--cwd`        | Working directory for the command                        |         |
| `--timeout`    | Timeout in seconds (max 300)                             | `30`    |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
| `--json`       | Output raw JSON response                                 |         |
