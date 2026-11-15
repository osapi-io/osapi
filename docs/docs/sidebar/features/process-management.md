---
sidebar_position: 20
---

# Process management

osapi shows the running processes on managed hosts, and can send signals to
specific processes. The agent collects the process data from `/proc` on Linux,
and from the system process table on other platforms.

## How it works

The process provider reads the process data from the operating system when the
request arrives. Facts are collected at intervals, but process data is always
live. Each request returns the processes that run at that time.

### List

Returns all running processes with their PID, name, user, state, CPU and memory
usage, and command line. The agent reads `/proc/{pid}/stat`,
`/proc/{pid}/status`, and `/proc/{pid}/cmdline` on Debian systems.

### Get

Returns detailed information about a single process identified by PID. If the
PID does not exist, the agent returns a 404 error.

### Signal

Sends a POSIX signal to a process by PID. The signals are `TERM`, `KILL`, `HUP`,
`INT`, `USR1`, `USR2`, `QUIT`, `STOP` and `CONT`. The agent calls
`syscall.Kill()`. If the process got the signal, it returns `changed: true`.

## Operations

| Operation | Description                            |
| --------- | -------------------------------------- |
| List      | List all running processes             |
| Get       | Get information about a process by PID |
| Signal    | Send a signal to a process by PID      |

## CLI usage

```bash
# List all processes on a host
osapi client node process list --target web-01

# Get details for a specific PID
osapi client node process get --target web-01 --pid 1234

# Send TERM signal to a process
osapi client node process signal --target web-01 \
  --pid 1234 --signal TERM

# Broadcast process list to all hosts
osapi client node process list --target _all
```

All commands accept `--json` to give the raw JSON output.

## Broadcast support

All process operations accept broadcast targets. Use `--target _all` to list
processes on all registered agents. Use a label selector, for example
`--target group:web`, to target a part of the fleet.

Responses always include per-host results:

```
  Job ID: 550e8400-e29b-41d4-a716-446655440000

  web-01
  PID   NAME      USER  STATE     CPU%  MEM%  COMMAND
  1     systemd   root  sleeping  0.0%  0.1%  /sbin/init
  1234  nginx     www   sleeping  2.3%  1.5%  nginx: worker process
```

Skipped and failed hosts appear with their error in the output.

## Supported platforms

| OS Family | Support |
| --------- | ------- |
| Debian    | Full    |
| Darwin    | Skipped |
| Linux     | Skipped |

On a platform that is not supported, process operations return
`status: skipped`. They do not fail. Refer to
[Platform Detection](../sdk/platform/detection.md) for how osapi finds the OS
family.

## Permissions

| Operation | Permission        |
| --------- | ----------------- |
| List, Get | `process:read`    |
| Signal    | `process:execute` |

Process listing and inspection require `process:read`, included in all built-in
roles. Sending signals requires `process:execute`, included only in the `admin`
role by default.

## Related

- [CLI Reference](../usage/cli/client/node/process/process.md): process commands
- [SDK Reference](../sdk/client/node/process.md): Process service
- [Platform Detection](../sdk/platform/detection.md): OS family detection
- [Configuration](../usage/configuration.md): full configuration reference
