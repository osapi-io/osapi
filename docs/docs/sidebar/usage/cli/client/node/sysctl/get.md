# Get

Get a kernel parameter's current value, and whether osapi manages it.

```bash
osapi client node sysctl get [flags]
```

## Examples

```bash
# A parameter osapi manages
osapi client node sysctl get --target web-01 --key vm.swappiness

# A parameter osapi does not manage
osapi client node sysctl get --target web-01 --key vm.swappiness

# A value of several fields
osapi client node sysctl get --target web-01 --key net.ipv4.ping_group_range

# A managed parameter the kernel does not have right now
osapi client node sysctl get --target web-01 --key net.bridge.bridge-nf-call-iptables
```

## Result

Each example prints this:

A parameter osapi manages:

```ansi
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mKEY            [0m[38;2;10;143;176mVALUE  [0m[38;2;10;143;176mMANAGED[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      vm.swappiness  10     yes

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

A parameter osapi does not manage:

```ansi
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mKEY            [0m[38;2;10;143;176mVALUE  [0m[38;2;10;143;176mMANAGED[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      vm.swappiness  60     no

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

A value of several fields:

```ansi
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mKEY                        [0m[38;2;10;143;176mVALUE         [0m[38;2;10;143;176mMANAGED[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      net.ipv4.ping_group_range  0 2147483647  no

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

A managed parameter the kernel does not have right now:

```ansi
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mKEY                                 [0m[38;2;10;143;176mVALUE  [0m[38;2;10;143;176mMANAGED[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      net.bridge.bridge-nf-call-iptables  1      yes

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

The command fails in these cases:

| Call                                 | Flags                          | Fails with            |
| ------------------------------------ | ------------------------------ | --------------------- |
| A parameter the kernel does not have | `--key vm.nosuch`              | not found             |
| No key                               |                                | key must not be empty |
| A key that is not a parameter name   | `--key "vm.swappiness;reboot"` | invalid sysctl key    |
| A key holding a value                | `--key vm.swappiness=1`        | invalid sysctl key    |

## Permission

It needs the `sysctl:read` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Flags

| Flag           | Description                                            | Default |
| -------------- | ------------------------------------------------------ | ------- |
| `--key`        | Parameter name, such as net.ipv4.ip_forward (required) |         |
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
