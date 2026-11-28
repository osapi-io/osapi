# Delete

Stop the management of a kernel parameter.

Removes the file that osapi wrote for it and reloads with sysctl --system. Then
a setting for the key in a different file takes effect. If there is no such
setting, the parameter keeps its current value until the next reboot. A delete
of a parameter that osapi does not manage reports changed: false.

```bash
osapi client node sysctl delete [flags]
```

## Examples

```bash
# Remove a parameter
osapi client node sysctl delete --target web-01 --key vm.swappiness

# A parameter osapi does not manage
osapi client node sysctl delete --target web-01 --key vm.swappiness

# A parameter osapi already stopped managing
osapi client node sysctl delete --target web-01 --key vm.swappiness
```

## Result

Remove a parameter, on one host:

```ansi
$ osapi client node sysctl delete --target web-01 --key vm.swappiness
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mKEY[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  vm.swappiness

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node sysctl delete --target _all --key vm.swappiness
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

A parameter osapi does not manage, on one host:

```ansi
$ osapi client node sysctl delete --target web-01 --key vm.swappiness
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mKEY[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      vm.swappiness

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node sysctl delete --target _all --key vm.swappiness
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mKEY[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      vm.swappiness
  [1mweb-02[0m    [38;2;103;234;147mok[0m      vm.swappiness

  3 hosts: [38;2;103;234;147m2 ok[0m, [38;5;222m1 skipped[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m

[1;38;2;255;0;255mDetails:[0m
  [1mmac-01[0m  [38;5;222mskipped[0m  operation not supported on this OS family
```

With `--json`, the API's response:

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "results": [
    {
      "changed": false,
      "hostname": "web-01",
      "key": "vm.swappiness",
      "status": "ok"
    }
  ]
}
```

A parameter osapi already stopped managing, on one host:

```ansi
$ osapi client node sysctl delete --target web-01 --key vm.swappiness
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mKEY[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      vm.swappiness

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node sysctl delete --target _all --key vm.swappiness
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mKEY[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      vm.swappiness
  [1mweb-02[0m    [38;2;103;234;147mok[0m      vm.swappiness

  3 hosts: [38;2;103;234;147m2 ok[0m, [38;5;222m1 skipped[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m

[1;38;2;255;0;255mDetails:[0m
  [1mmac-01[0m  [38;5;222mskipped[0m  operation not supported on this OS family
```

With `--json`, the API's response:

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "results": [
    {
      "changed": false,
      "hostname": "web-01",
      "key": "vm.swappiness",
      "status": "ok"
    }
  ]
}
```

With the default settings, each example does this on the host:

| Example            | On the host                                                              |
| ------------------ | ------------------------------------------------------------------------ |
| Remove a parameter | removes `/etc/sysctl.d/osapi-vm.swappiness.conf`; runs `sysctl --system` |

The command fails in these cases:

| Call                               | Flags                          | Fails with            |
| ---------------------------------- | ------------------------------ | --------------------- |
| No key                             |                                | key must not be empty |
| A key that is not a parameter name | `--key 'vm.swappiness;reboot'` | invalid sysctl key    |
| A key holding a value              | `--key vm.swappiness=1`        | invalid sysctl key    |

## Permission

It needs the `sysctl:write` permission.
`osapi controller role get --name <role>` shows what a role grants.

## Flags

| Flag           | Description                                            | Default |
| -------------- | ------------------------------------------------------ | ------- |
| `--key`        | Parameter name, such as net.ipv4.ip_forward (required) |         |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev) | `_all`  |

## Global flags

| Flag               | Description                                                                      | Default                  |
| ------------------ | -------------------------------------------------------------------------------- | ------------------------ |
| `-d, --debug`      | Enable or disable debug mode                                                     |                          |
| `-j, --json`       | Enable JSON output                                                               |                          |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml`  |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `https://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
