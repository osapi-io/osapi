# Get

Get the operating system of the host.

Reads the distribution and its version: the ID and VERSION_ID of /etc/os-release
on Linux, and the system type and product version that sysctl gives on macOS,
such as darwin 15.5.

```bash
osapi client node os get [flags]
```

## Examples

```bash
# An Ubuntu host
osapi client node os get --target web-01
```

## Permission

It needs the `os:read` permission. `osapi token roles get --name <role>` shows
what a role grants.

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
