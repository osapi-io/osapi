# Get

Get the disk usage of the host.

Reads the size, the used space and the free space of each local filesystem, in
bytes: the mounts of /proc/mounts on Linux and of mount on macOS, with their
sizes from df. Local is ext4, xfs and btrfs on Linux, and apfs and hfs on macOS.
Network and special filesystems are left out.

```bash
osapi client node disk get [flags]
```

## Examples

```bash
# A host with one local filesystem
osapi client node disk get --target web-01
```

## Permission

It needs the `disk:read` permission. `osapi token roles get --name <role>` shows
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
