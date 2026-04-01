# Exec

Execute a command directly without a shell interpreter. Arguments are passed to
the executable as-is, without shell expansion or interpretation.

```bash
$ osapi client node command exec --command ls --args "-la,/tmp"

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  CHANGED  STDOUT                          STDERR  EXIT CODE  DURATION
  false    total 8 drwxrwxrwt 10 root r…           0          12ms
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

Target by label to execute on a group of servers:

```bash
$ osapi client node command exec --command whoami --target group:web
```

Use `@fact.*` references to inject live system values. Each agent resolves its
own facts, so this works correctly with broadcast targeting:

```bash
$ osapi client node command exec \
    --command ip --args "addr,show,dev,@fact.interface.primary" \
    --target _all
```

See [System Facts](../../../../../features/system-facts.md) for all available
`@fact.*` references.

## JSON Output

Use `--json` to get the full untruncated API response:

```bash
$ osapi client node command exec --command hostname --json
```

## Raw Output

Use `--stdout` to print only the remote command's stdout, without the table
wrapper:

```bash
$ osapi client node command exec --command ls --args "-la" --stdout
total 48
drwxr-xr-x  12 john  staff  384 Mar  2 10:00 .
-rw-r--r--   1 john  staff 1234 Mar  2 09:30 main.go
```

Use `--stderr` to print only stderr:

```bash
$ osapi client node command exec --command ls --args "/nonexistent" --stderr
ls: cannot access '/nonexistent': No such file or directory
```

Both flags can be combined. Each line is prefixed with the hostname:

```bash
$ osapi client node command exec --command hostname --target _all --stdout
[web-01] web-01.example.com
[web-02] web-02.example.com
```

The CLI exit code matches the remote command's exit code, making it scriptable:

```bash
$ osapi client node command exec --command "test" --args "-f,/etc/hosts" --stdout && echo exists
exists
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `--command`    | The command to execute (**required**)                    |         |
| `--args`       | Command arguments (comma-separated)                      | `[]`    |
| `--cwd`        | Working directory for the command                        |         |
| `--timeout`    | Timeout in seconds (max 300)                             | `30`    |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_all`  |
| `--stdout`     | Print only remote stdout                                 |         |
| `--stderr`     | Print only remote stderr                                 |         |
| `-j, --json`   | Output raw JSON response                                 |         |
