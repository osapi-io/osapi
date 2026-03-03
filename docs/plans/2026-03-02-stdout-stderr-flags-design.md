# Design: --stdout and --stderr flags for command exec/shell

## Problem

Running a remote command with OSAPI requires piping through `jq` to see
the actual output:

```bash
osapi client node command exec --command ls --json | jq -r '.results[0].stdout'
```

The default table view truncates stdout to 50 characters and flattens
multi-line output. There's no way to get raw command output directly.

## Solution

Add `--stdout` and `--stderr` flags to `node command exec` and
`node command shell` CLI commands. These print the remote command's
raw output directly to the terminal.

## Behavior

### Flags

- `--stdout` — print remote stdout to terminal stdout
- `--stderr` — print remote stderr to terminal stderr (fd 2)
- Both together — print stdout to fd 1, stderr to fd 2
- Mutually exclusive with `--json`
- Neither flag — current behavior (table display)

### Single-host output

Raw output, no decoration:

```
$ osapi client node command exec --command ls --args "-la" --stdout
total 48
drwxr-xr-x  12 john  staff  384 Mar  2 10:00 .
-rw-r--r--   1 john  staff 1234 Mar  2 09:30 main.go
```

### Multi-host output

Hostname-prefixed per line, hostname dimmed with lipgloss:

```
$ osapi client node command exec --target _all --command hostname --stdout
  web-01  web-01.example.com
  web-02  web-02.example.com
  db-01   db-01.example.com
```

Multi-line stdout from multiple hosts:

```
$ osapi client node command exec --target _all --command ls --stdout
  web-01  file1
  web-01  file2
  web-02  file1
  web-02  file3
```

### Exit code propagation

The CLI process exits with the remote command's exit code. For
multi-host, exits non-zero if any host returned non-zero.

## Scope

- CLI-only change — no API or protocol changes
- Applies to `node command exec` and `node command shell`
- No streaming — still synchronous request/response via NATS
- No architecture changes

## Files to change

- `cmd/client_node_command_exec.go` — add flags + output logic
- `cmd/client_node_command_shell.go` — add flags + output logic
- `docs/docs/sidebar/usage/cli/client/node/command-exec.md` — document
  flags with examples
- `docs/docs/sidebar/usage/cli/client/node/command-shell.md` — document
  flags with examples
- Tests for the new output paths

## Non-goals

- True streaming (would require WebSocket/SSE + architecture changes)
- Interactive commands (stdin passthrough)
- Changes to API response format
