---
sidebar_position: 20
---

# Process management

The [process behavior](../reference/behaviors/process/process.md) shows the
processes of managed hosts, and sends a signal to one process. The agent reads
the processes through gohai's process collector: from `/proc` on each Linux
family, and from `ps` on macOS.

## How it works

The agent reads the processes when the request arrives. Facts are collected at
intervals, but process data is always live.

- **List**: lists each process with its PID, name, user, state, CPU, memory,
  command line and start time. On Linux, gohai reads `/proc/{pid}/stat`,
  `/proc/{pid}/status` and `/proc/{pid}/cmdline`. On macOS, it runs `ps`. The
  CPU value on Linux is the average since the process started. On macOS, it is
  the `%cpu` of `ps`. Each row of the result is one process of one host.
- **Get**: reads one process by its PID, without the others. A PID that the host
  does not have is not found, and the API returns 404.
- **Signal**: sends `TERM`, `KILL`, `HUP`, `INT`, `USR1` or `USR2` to one
  process with `kill`, through sudo when the agent uses it. The agent reads the
  process first, so a PID that the host does not have is not found. A PID below
  1 is refused, because `kill` sends to a group of processes for 0 and for a
  negative PID. A signal always reports `changed: true`.

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

All process operations accept broadcast targets. Use `--target _all` to list the
processes of each registered agent. Use a label selector, for example
`--target group:web`, to target a part of the fleet. Each row of the result
names its host. A host that failed or that was skipped shows its error.

## Supported platforms

| OS Family         | List and get | Signal |
| ----------------- | ------------ | ------ |
| Each Linux family | Full         | Full   |
| Darwin            | Full         | Full   |

## Permissions

The Permission column of [Methods](#methods) gives the permission of each
method.

## Related

- [Process behavior](../reference/behaviors/process/process.md): each method,
  its examples and what it refuses
- [CLI Reference](../usage/cli/client/node/process/process.mdx): process
  commands
- [SDK Reference](../sdk/client/process.md): Process service
- [Platform Detection](../sdk/platform/detection.md): OS family detection
- [Configuration](../usage/configuration.md): full configuration reference
