# Get

Get the timezone of the host.

Reads the IANA name of the timezone and its offset from UTC now: the zone that
/etc/localtime links to, or /etc/timezone on Linux.

```bash
osapi client node timezone get [flags]
```

## Examples

```bash
# A Linux host on UTC
osapi client node timezone get --target web-01

# A Mac on UTC
osapi client node timezone get --target web-01
```

## Permission

It needs the `timezone:read` permission. `osapi token roles get --name <role>`
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
