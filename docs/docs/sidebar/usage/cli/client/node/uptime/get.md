# Get

Get the uptime of the host.

Reads how long the host has run since it booted: /proc/uptime on Linux, and the
boot time that sysctl gives on macOS. The uptime is in days, hours and minutes,
and in seconds.

```bash
osapi client node uptime get [flags]
```

## Examples

```bash
# A host up for a day, two hours and five minutes
osapi client node uptime get --target web-01
```

## Permission

It needs the `uptime:read` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Flags

| Flag           | Description                                            | Default |
| -------------- | ------------------------------------------------------ | ------- |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev) | `_all`  |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
