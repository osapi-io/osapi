# List

List the kernel's parameters and which osapi manages.

Lists every parameter the kernel shows under /proc/sys, with its current value,
as sysctl -a does, and marks the ones osapi manages. A prefix, such as net.ipv4,
limits the list to that part of the tree. The kernel can lack a parameter that
osapi manages, for example when its module is not loaded. Then the list shows
the value that osapi wrote.

```bash
osapi client next node sysctl list [flags]
```

## Examples

```bash
# Every parameter, and which osapi manages
osapi client next node sysctl list --target web-01

# One part of the tree
osapi client next node sysctl list --target web-01 --prefix net.ipv4

# A managed parameter the kernel does not have right now
osapi client next node sysctl list --target web-01 --prefix net.bridge
```

## Result

The command fails in these cases:

| Call                                  | Flags             | Fails with         |
| ------------------------------------- | ----------------- | ------------------ |
| A prefix that is not a parameter name | `--prefix ../etc` | invalid sysctl key |

## Permission

It needs the `sysctl:read` permission. `osapi token roles` lists the roles that
grant it.

## Flags

| Flag           | Description                                                 | Default |
| -------------- | ----------------------------------------------------------- | ------- |
| `--prefix`     | Only this parameter and the ones under it, such as net.ipv4 |         |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev)      | `_all`  |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../../output.md) tells how the result is printed, and what
`--json` gives.
