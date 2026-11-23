---
sidebar_position: 23
---

# Log management

osapi reads the systemd journal of a host: its entries, and the sources that
wrote them. Two behaviors under the `log` group do it, `entry` and `source`.
Each runs `journalctl` on the host. The journal is what journald holds now, not
a fact of the host, so gohai does not read it. Both are read-only.

## Entries

`log entry list` runs `journalctl --output=json`, and returns the newest entries
last. Each entry has its time in RFC 3339, the syslog identifier that wrote it,
its priority by name, its message and its process ID. These filters select the
entries, and each one is optional:

| Filter     | Selects                                                                |
| ---------- | ---------------------------------------------------------------------- |
| `unit`     | the entries of one systemd unit, such as `nginx.service`               |
| `lines`    | how many of the newest entries, from 1 to 10000; 100 by default        |
| `since`    | the entries since a time, in each form that `journalctl --since` takes |
| `priority` | the entries at a priority or above: `emerg` to `debug`                 |

osapi gives each filter to `journalctl` in the same argument as its option, such
as `--unit=nginx.service`. Thus a value that starts with `-` is not read as an
option.

## Sources

`log source list` runs `journalctl --field=SYSLOG_IDENTIFIER`, and returns each
syslog identifier that has an entry, one time each, sorted. Use it to find a
name before you query the entries of a unit.

## CLI usage

```bash
# The 100 newest entries of a host
osapi client node log entry list --target web-01

# The 50 newest errors of the last hour
osapi client node log entry list --target web-01 \
  --lines 50 --since "1 hour ago" --priority err

# The entries of one unit
osapi client node log entry list --target web-01 --unit sshd.service

# The sources of the journal
osapi client node log source list --target web-01

# The 20 newest entries of each host
osapi client node log entry list --target _all --lines 20
```

Each command takes `--json` for the response body.

## Broadcast

Each command takes a broadcast target: `--target _all` for each agent, or a
label, such as `--target group:web`, for part of the fleet. The response has one
row for each entry, or source, of each host. A host that skipped or failed has
one row with its error.

## Supported platforms

| Platform        | Support                                         |
| --------------- | ----------------------------------------------- |
| Each Linux host | Full, when the host has journald                |
| macOS           | Skipped: macOS has no journal                   |
| A container     | Skipped: a container has no journald of its own |

The hand-written log domain read the Debian family only, and skipped the other
Linux families. The behaviors read each Linux host.

## Permissions

The Permission column of [Methods](#methods) gives the permission of each
method.

## Related

- [CLI reference](../usage/cli/client/node/log/log.mdx): the log commands
- [SDK reference](../sdk/client/node/log.md): the log service
- [Platform detection](../sdk/platform/detection.md): how osapi finds the
  platform of a host
