# Create

Manage a kernel parameter and apply it.

Writes osapi-net.ipv4.ip_forward.conf for net.ipv4.ip_forward to
behaviors.sysctl.dir, /etc/sysctl.d by default, and applies it with sysctl -p.
It fails when sysctl -p fails, because the value is not set. A second run with
the same value changes nothing and reports changed: false. If the kernel has a
different value, for example after a failed apply, it applies the value again.

```bash
osapi client node sysctl create [flags]
```

## Examples

```bash
# Turn on IP forwarding
osapi client node sysctl create --target web-01 --key net.ipv4.ip_forward --value 1
```

## Result

Turn on IP forwarding, on one host:

```ansi
$ osapi client node sysctl create --target web-01 --key net.ipv4.ip_forward --value 1
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mKEY[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  net.ipv4.ip_forward

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node sysctl create --target _all --key net.ipv4.ip_forward --value 1
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mKEY[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  net.ipv4.ip_forward
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m  net.ipv4.ip_forward

  3 hosts: [38;2;103;234;147m2 changed[0m, [38;5;222m1 skipped[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m

[1;38;2;255;0;255mDetails:[0m
  [1mmac-01[0m  [38;5;222mskipped[0m  operation not supported on this OS family
```

With `--json`, the API's response:

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "results": [
    {
      "changed": true,
      "hostname": "web-01",
      "key": "net.ipv4.ip_forward",
      "status": "ok"
    }
  ]
}
```

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

It needs the `sysctl:write` permission.
`osapi controller role get --name <role>` shows what a role grants.

## Flags

| Flag           | Description                                            | Default |
| -------------- | ------------------------------------------------------ | ------- |
| `--key`        | Parameter name, such as net.ipv4.ip_forward (required) |         |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev) | `_all`  |
| `--value`      | Value to set (required)                                |         |

## Global flags

| Flag               | Description                                                                      | Default                  |
| ------------------ | -------------------------------------------------------------------------------- | ------------------------ |
| `-d, --debug`      | Enable or disable debug mode                                                     |                          |
| `-j, --json`       | Enable JSON output                                                               |                          |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml`  |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `https://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
