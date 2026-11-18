# Create

Manage a kernel parameter and apply it.

Writes osapi-&lt;key&gt;.conf to behaviors.sysctl.dir, /etc/sysctl.d by default,
and applies it with sysctl -p. It fails when sysctl -p fails, because the value
is not set. A second run with the same value changes nothing and reports
changed: false. If the kernel has a different value, for example after a failed
apply, it applies the value again.

```bash
osapi client next node sysctl create [flags]
```

## Examples

```bash
# Turn on IP forwarding
osapi client next node sysctl create --target web-01 --key net.ipv4.ip_forward --value 1
```

## Result

With the default settings, each example does this on the host:

| Example               | On the host                                                                                                                                     |
| --------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| Turn on IP forwarding | writes `/etc/sysctl.d/osapi-net.ipv4.ip_forward.conf`: `net.ipv4.ip_forward = 1`; runs `sysctl -p /etc/sysctl.d/osapi-net.ipv4.ip_forward.conf` |

The command fails in these cases:

| Call                               | Flags                                                | Fails with              |
| ---------------------------------- | ---------------------------------------------------- | ----------------------- |
| No key                             | `--value 1`                                          | key must not be empty   |
| A key that is not a parameter name | `--key ../etc/passwd --value 1`                      | invalid sysctl key      |
| A key that would read as a flag    | `--key -w --value 1`                                 | invalid sysctl key      |
| A key that would be a hidden file  | `--key .hidden --value 1`                            | invalid sysctl key      |
| No value                           | `--key vm.swappiness`                                | value must not be empty |
| A value on two lines               | `--key vm.swappiness --value "10\nkernel.panic = 0"` | line breaks             |

## Permission

It needs the `sysctl:write` permission. `osapi token roles` lists the roles that
grant it.

## Flags

| Flag           | Description                                            | Default |
| -------------- | ------------------------------------------------------ | ------- |
| `--key`        | Parameter name, such as net.ipv4.ip_forward (required) |         |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev) | `_all`  |
| `--value`      | Value to set (required)                                |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../../output.md) tells how the result is printed, and what
`--json` gives.
