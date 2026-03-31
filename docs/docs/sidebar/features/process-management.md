---
sidebar_position: 20
---

# Process Management

OSAPI provides read-only visibility into running processes and the ability to
send signals to specific processes on managed hosts. Process information is
collected via `/proc` on Linux and the system process table on other platforms.

## How It Works

The process provider reads process information from the operating system at
request time. Unlike facts (which are collected periodically), process data is
always live -- each request returns the current snapshot of running processes.

### List

Returns all running processes with their PID, name, user, state, CPU and memory
usage, and command line. The agent reads `/proc/{pid}/stat`,
`/proc/{pid}/status`, and `/proc/{pid}/cmdline` on Debian systems.

### Get

Returns detailed information about a single process identified by PID. If the
PID does not exist, the agent returns a 404 error.

### Signal

Sends a POSIX signal to a process by PID. Supported signals: `TERM`, `KILL`,
`HUP`, `INT`, `USR1`, `USR2`, `QUIT`, `STOP`, `CONT`. The agent calls
`syscall.Kill()` and returns `changed: true` if the signal was delivered.

## Operations

| Operation | Description                            |
| --------- | -------------------------------------- |
| List      | List all running processes             |
| Get       | Get information about a process by PID |
| Signal    | Send a signal to a process by PID      |

## CLI Usage

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

All commands support `--json` for raw JSON output.

## Broadcast Support

All process operations support broadcast targeting. Use `--target _all` to list
processes on every registered agent, or use a label selector like
`--target group:web` to target a subset.

Responses always include per-host results:

```
  Job ID: 550e8400-e29b-41d4-a716-446655440000

  web-01
  PID   NAME      USER  STATE     CPU%  MEM%  COMMAND
  1     systemd   root  sleeping  0.0%  0.1%  /sbin/init
  1234  nginx     www   sleeping  2.3%  1.5%  nginx: worker process
```

Skipped and failed hosts appear with their error in the output.

## Supported Platforms

| OS Family | Support |
| --------- | ------- |
| Debian    | Full    |
| Darwin    | Skipped |
| Linux     | Skipped |

On unsupported platforms, process operations return `status: skipped` instead of
failing. See [Platform Detection](../sdk/platform/detection.md) for details on
OS family detection.

## Permissions

| Operation | Permission        |
| --------- | ----------------- |
| List, Get | `process:read`    |
| Signal    | `process:execute` |

Process listing and inspection require `process:read`, included in all built-in
roles. Sending signals requires `process:execute`, included only in the `admin`
role by default.

## Related

- [CLI Reference](../usage/cli/client/node/process/process.md) -- process
  commands
- [SDK Reference](../sdk/client/operations/process.md) -- Process service
- [Platform Detection](../sdk/platform/detection.md) -- OS family detection
- [Configuration](../usage/configuration.md) -- full configuration reference
