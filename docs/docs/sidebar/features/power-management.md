---
sidebar_position: 19
---

# Power Management

OSAPI manages power state on target hosts via systemd. Reboot and shutdown
operations are scheduled through `shutdown(8)` and take effect after an optional
delay. Operations are dispatched asynchronously through the job system so the
controller returns immediately.

## How It Works

Power operations call `shutdown` on the agent host:

- **Reboot**: schedules `shutdown -r +<delay>` (or `shutdown -r now` when delay
  is 0)
- **Shutdown**: schedules `shutdown -h +<delay>` (or `shutdown -h now` when
  delay is 0)

The agent records the scheduled action and returns `changed: true`. If the
operation is unsupported on the platform (e.g., macOS agents), the job returns
`status: skipped` instead of failing.

### Delay Behaviour

The `delay` field specifies seconds to wait before the power event occurs. When
`delay` is 0, the operation executes immediately (`now`). The system broadcasts
a wall message to logged-in users when a non-zero delay or message is provided.

## Operations

| Operation | Description                                  |
| --------- | -------------------------------------------- |
| Reboot    | Schedule a host reboot with optional delay   |
| Shutdown  | Schedule a host shutdown with optional delay |

## CLI Usage

```bash
# Reboot a host immediately
osapi client node power reboot --target web-01

# Reboot with a 60-second delay and message
osapi client node power reboot --target web-01 \
  --delay 60 --message "Maintenance reboot"

# Shut down a host immediately
osapi client node power shutdown --target web-01

# Broadcast reboot to all hosts with a delay
osapi client node power reboot --target _all \
  --delay 30 --message "Scheduled reboot"
```

All commands support `--json` for raw JSON output.

## Broadcast Support

All power operations support broadcast targeting. Use `--target _all` to reboot
or shut down every registered agent, or use a label selector like
`--target group:web` to target a subset.

Responses always include per-host results:

```
  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  CHANGED  ERROR  ACTION  DELAY
  web-01    ok      true            reboot  30
  web-02    ok      true            reboot  30
```

Skipped and failed hosts appear with their error in the output.

## Supported Platforms

| OS Family | Support |
| --------- | ------- |
| Debian    | Full    |
| Darwin    | Skipped |
| Linux     | Skipped |

On unsupported platforms, power operations return `status: skipped` instead of
failing. See [Platform Detection](../sdk/platform/detection.md) for details on
OS family detection.

## Permissions

| Operation        | Permission      |
| ---------------- | --------------- |
| Reboot, Shutdown | `power:execute` |

Power management requires the `power:execute` permission. Only the `admin` role
includes this permission by default. Do not grant it to untrusted tokens.

## Related

- [CLI Reference](../usage/cli/client/node/power/power.md) — power commands
- [Platform Detection](../sdk/platform/detection.md) — OS family detection
- [Configuration](../usage/configuration.md) — full configuration reference
