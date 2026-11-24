# Power management

osapi manages the power state of target hosts through systemd. The agent
schedules reboot and shutdown with `shutdown(8)`. They take effect after an
optional delay. The operations go asynchronously through the job system, so the
controller returns immediately.

## How it works

Power operations call `shutdown` on the agent host:

- **Reboot**: schedules `shutdown -r -- +N`
- **Shutdown**: schedules `shutdown -h -- +N`

`--` ends the options of `shutdown`. Thus a message that starts with `-` is the
message, not an option. The agent records the scheduled action and returns
`changed: true` before the system stops. If the platform does not support the
operation, for example on macOS agents or in a container, the job returns
`status: skipped`. Power is a behavior: `osapi dsl behavior get --name power`
shows its definition.

### Minimum delay

The delay is always 1 minute or more, also when `delay` is 0 or not given. Thus
the agent has time to write the job result to NATS and to send the response to
the API. Then it stops cleanly, before the system powers off.

The `delay` field in the request body is in seconds. The provider changes it to
minutes for the `shutdown` command (1 minute or more, rounded up). The response
gives the real delay in seconds. Thus the caller knows when the action takes
effect.

## Operations

| Operation | Description                                  |
| --------- | -------------------------------------------- |
| Reboot    | Schedule a host reboot with optional delay   |
| Shutdown  | Schedule a host shutdown with optional delay |

## CLI usage

```bash
# Reboot a host (1-minute minimum delay)
osapi client node power reboot --target web-01

# Reboot with a 60-second delay and message
osapi client node power reboot --target web-01 \
  --delay 60 --message "Maintenance reboot"

# Shut down a host (1-minute minimum delay)
osapi client node power shutdown --target web-01

# Broadcast reboot to all hosts with a delay
osapi client node power reboot --target _all \
  --delay 30 --message "Scheduled reboot"
```

All commands accept `--json` to give the raw JSON output.

## Broadcast support

All power operations accept broadcast targets. Use `--target _all` to reboot or
stop all registered agents. Use a label selector, for example
`--target group:web`, to target a part of the fleet.

Responses always include per-host results:

```
  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  CHANGED  ERROR  ACTION  DELAY
  web-01    ok      true            reboot  30
  web-02    ok      true            reboot  30
```

Skipped and failed hosts appear with their error in the output.

## Supported platforms

| OS Family | Support |
| --------- | ------- |
| Debian    | Full    |
| Darwin    | Skipped |
| Linux     | Skipped |

On a platform that is not supported, power operations return `status: skipped`.
They do not fail. Refer to
[Platform Detection](../../sdk/library/platform-detection.md) for how osapi
finds the OS family.

## Permissions

The [power](../../reference/behaviors/node/power/power.md) reference page gives
the permission of each method.

## Related

- [CLI Reference](../../usage/cli/client/node/power/power.mdx): power commands
- [Platform Detection](../../sdk/library/platform-detection.md): OS family
  detection
- [Configuration](../../usage/configuration.md): full configuration reference
