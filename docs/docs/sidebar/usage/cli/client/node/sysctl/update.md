# Update

Set a kernel parameter, manage it, and apply it.

The same as create, addressed by the parameter: osapi manages the parameter from
then on whether it did before or not. Setting the value it already has reports
changed: false.

```bash
osapi client node sysctl update [flags]
```

## Examples

```bash
# Lower swappiness
osapi client node sysctl update --target web-01 --key vm.swappiness --value 10

# A parameter osapi did not manage
osapi client node sysctl update --target web-01 --key vm.swappiness --value 10
```

## Result

Lower swappiness, on one host:

```ansi
$ osapi client node sysctl update --target web-01 --key vm.swappiness --value 10
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mKEY[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  vm.swappiness

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node sysctl update --target _all --key vm.swappiness --value 10
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mKEY[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  vm.swappiness
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m  vm.swappiness

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
      "key": "vm.swappiness",
      "status": "ok"
    }
  ]
}
```

A parameter osapi did not manage, on one host:

```ansi
$ osapi client node sysctl update --target web-01 --key vm.swappiness --value 10
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mKEY[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  vm.swappiness

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node sysctl update --target _all --key vm.swappiness --value 10
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mKEY[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  vm.swappiness
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m  vm.swappiness

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
      "key": "vm.swappiness",
      "status": "ok"
    }
  ]
}
```

With the default settings, each example does this on the host:

| Example                          | On the host                                                                                                                    |
| -------------------------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| Lower swappiness                 | writes `/etc/sysctl.d/osapi-vm.swappiness.conf`: `vm.swappiness = 10`; runs `sysctl -p /etc/sysctl.d/osapi-vm.swappiness.conf` |
| A parameter osapi did not manage | writes `/etc/sysctl.d/osapi-vm.swappiness.conf`: `vm.swappiness = 10`; runs `sysctl -p /etc/sysctl.d/osapi-vm.swappiness.conf` |

The command fails in these cases:

| Call                               | Flags                                                | Fails with              |
| ---------------------------------- | ---------------------------------------------------- | ----------------------- |
| No key                             | `--value 10`                                         | key must not be empty   |
| A key that is not a parameter name | `--key "vm.swappiness;reboot" --value 1`             | invalid sysctl key      |
| A key holding a value              | `--key vm.swappiness=1 --value 1`                    | invalid sysctl key      |
| No value                           | `--key vm.swappiness`                                | value must not be empty |
| A value on two lines               | `--key vm.swappiness --value "10\nkernel.panic = 0"` | line breaks             |

## Permission

It needs the `sysctl:write` permission. `osapi token roles get --name <role>`
shows what a role grants.

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

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
