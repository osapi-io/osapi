# List

List the kernel's parameters and which osapi manages.

Lists every parameter the kernel shows under /proc/sys, with its current value,
as sysctl -a does, and marks the ones osapi manages. A prefix, such as net.ipv4,
limits the list to that part of the tree. The kernel can lack a parameter that
osapi manages, for example when its module is not loaded. Then the list shows
the value that osapi wrote.

```bash
osapi client node sysctl list [flags]
```

## Examples

```bash
# Every parameter, and which osapi manages
osapi client node sysctl list --target web-01

# One part of the tree
osapi client node sysctl list --target web-01 --prefix net.ipv4

# A managed parameter the kernel does not have right now
osapi client node sysctl list --target web-01 --prefix net.bridge
```

## Result

Every parameter, and which osapi manages, on one host:

```ansi
$ osapi client node sysctl list --target web-01
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mKEY                        [0m[38;2;10;143;176mVALUE         [0m[38;2;10;143;176mMANAGED[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      kernel.panic               0             no
  [38;2;138;138;138mweb-01[0m    [38;2;103;234;147mok[0m      net.ipv4.ip_forward        1             yes
  [38;2;138;138;138mweb-01[0m    [38;2;103;234;147mok[0m      net.ipv4.ping_group_range  0 2147483647  no

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node sysctl list --target _all
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mKEY                        [0m[38;2;10;143;176mVALUE         [0m[38;2;10;143;176mMANAGED[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      kernel.panic               0             no
  [38;2;138;138;138mweb-01[0m    [38;2;103;234;147mok[0m      net.ipv4.ip_forward        1             yes
  [38;2;138;138;138mweb-01[0m    [38;2;103;234;147mok[0m      net.ipv4.ping_group_range  0 2147483647  no
  [1mweb-02[0m    [38;2;103;234;147mok[0m      kernel.panic               0             no
  [38;2;138;138;138mweb-02[0m    [38;2;103;234;147mok[0m      net.ipv4.ip_forward        1             yes
  [38;2;138;138;138mweb-02[0m    [38;2;103;234;147mok[0m      net.ipv4.ping_group_range  0 2147483647  no

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
      "hostname": "web-01",
      "key": "kernel.panic",
      "managed": false,
      "status": "ok",
      "value": "0"
    },
    {
      "hostname": "web-01",
      "key": "net.ipv4.ip_forward",
      "managed": true,
      "status": "ok",
      "value": "1"
    },
    {
      "hostname": "web-01",
      "key": "net.ipv4.ping_group_range",
      "managed": false,
      "status": "ok",
      "value": "0 2147483647"
    }
  ]
}
```

One part of the tree, on one host:

```ansi
$ osapi client node sysctl list --target web-01 --prefix net.ipv4
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mKEY                        [0m[38;2;10;143;176mVALUE         [0m[38;2;10;143;176mMANAGED[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      net.ipv4.ip_forward        1             yes
  [38;2;138;138;138mweb-01[0m    [38;2;103;234;147mok[0m      net.ipv4.ping_group_range  0 2147483647  no

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node sysctl list --target _all --prefix net.ipv4
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mKEY                        [0m[38;2;10;143;176mVALUE         [0m[38;2;10;143;176mMANAGED[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      net.ipv4.ip_forward        1             yes
  [38;2;138;138;138mweb-01[0m    [38;2;103;234;147mok[0m      net.ipv4.ping_group_range  0 2147483647  no
  [1mweb-02[0m    [38;2;103;234;147mok[0m      net.ipv4.ip_forward        1             yes
  [38;2;138;138;138mweb-02[0m    [38;2;103;234;147mok[0m      net.ipv4.ping_group_range  0 2147483647  no

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
      "hostname": "web-01",
      "key": "net.ipv4.ip_forward",
      "managed": true,
      "status": "ok",
      "value": "1"
    },
    {
      "hostname": "web-01",
      "key": "net.ipv4.ping_group_range",
      "managed": false,
      "status": "ok",
      "value": "0 2147483647"
    }
  ]
}
```

A managed parameter the kernel does not have right now, on one host:

```ansi
$ osapi client node sysctl list --target web-01 --prefix net.bridge
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mKEY                                 [0m[38;2;10;143;176mVALUE  [0m[38;2;10;143;176mMANAGED[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      net.bridge.bridge-nf-call-iptables  1      yes

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node sysctl list --target _all --prefix net.bridge
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mKEY                                 [0m[38;2;10;143;176mVALUE  [0m[38;2;10;143;176mMANAGED[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      net.bridge.bridge-nf-call-iptables  1      yes
  [1mweb-02[0m    [38;2;103;234;147mok[0m      net.bridge.bridge-nf-call-iptables  1      yes

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
      "hostname": "web-01",
      "key": "net.bridge.bridge-nf-call-iptables",
      "managed": true,
      "status": "ok",
      "value": "1"
    }
  ]
}
```

The command fails in these cases:

| Call                                  | Flags             | Fails with         |
| ------------------------------------- | ----------------- | ------------------ |
| A prefix that is not a parameter name | `--prefix ../etc` | invalid sysctl key |

## Permission

It needs the `sysctl:read` permission. `osapi token roles get --name <role>`
shows what a role grants.

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

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
