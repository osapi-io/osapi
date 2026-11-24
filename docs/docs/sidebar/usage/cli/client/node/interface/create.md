# Create

Configure a network interface.

Writes a netplan file of the interface's own, such as
/etc/netplan/osapi-osapi0.yaml for osapi0, under a netplan definition that is
the interface's own, and applies it. An interface the host does not have is not
found. dhcp4 and dhcp6 are false unless given, so the interface has only the
addresses given. An interface that a match defines, such as cloud-init's all for
en\*, is refused, because the config would change each interface the match
takes, now and later. An interface osapi already configures does not change:
update replaces its config. It runs on the Debian family, and elsewhere, and in
a container, reports skipped.

```bash
osapi client node interface create [flags]
```

## Examples

```bash
# A static address
osapi client node interface create --target web-01 --name osapi0 --addresses 10.0.1.5/24

# Each setting
osapi client node interface create --target web-01 --name osapi0 --dhcp6 --addresses 2001:db8::5/64 --mtu 9000 --mac-address 52:54:00:aa:bb:cc --wakeonlan

# An interface osapi already configures
osapi client node interface create --target web-01 --name osapi0 --addresses 10.0.1.5/24
```

## Result

A static address, on one host:

```ansi
$ osapi client node interface create --target web-01 --name osapi0 --addresses 10.0.1.5/24
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mosapi0[0m

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node interface create --target _all --name osapi0 --addresses 10.0.1.5/24
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mosapi0[0m
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m  [1mosapi0[0m

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
      "name": "osapi0",
      "status": "ok"
    }
  ]
}
```

Each setting, on one host:

```ansi
$ osapi client node interface create --target web-01 --name osapi0 --dhcp6 --addresses 2001:db8::5/64 --mtu 9000 --mac-address 52:54:00:aa:bb:cc --wakeonlan
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mosapi0[0m

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node interface create --target _all --name osapi0 --dhcp6 --addresses 2001:db8::5/64 --mtu 9000 --mac-address 52:54:00:aa:bb:cc --wakeonlan
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mosapi0[0m
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m  [1mosapi0[0m

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
      "name": "osapi0",
      "status": "ok"
    }
  ]
}
```

An interface osapi already configures, on one host:

```ansi
$ osapi client node interface create --target web-01 --name osapi0 --addresses 10.0.1.5/24
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mosapi0[0m

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node interface create --target _all --name osapi0 --addresses 10.0.1.5/24
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mosapi0[0m
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [1mosapi0[0m

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
      "name": "osapi0",
      "status": "ok"
    }
  ]
}
```

With the default settings, each example does this on the host:

| Example          | On the host                                                                                                                                                                                                                                                                                                 |
| ---------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| A static address | writes `/etc/netplan/osapi-osapi0.yaml`: `network:\n  version: 2\n  ethernets:\n    osapi0:\n      dhcp4: false\n      dhcp6: false\n      addresses:\n        - 10.0.1.5/24`; runs `netplan generate`; runs `netplan apply`                                                                                |
| Each setting     | writes `/etc/netplan/osapi-osapi0.yaml`: `network:\n  version: 2\n  ethernets:\n    osapi0:\n      dhcp4: false\n      dhcp6: true\n      addresses:\n        - 2001:db8::5/64\n      mtu: 9000\n      macaddress: 52:54:00:aa:bb:cc\n      wakeonlan: true`; runs `netplan generate`; runs `netplan apply` |

The command fails in these cases:

| Call                                                 | Flags                                                                                                                                                                                                                                                                                                                                                                                                                                   | Fails with                     |
| ---------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------ |
| An interface in a match that takes other interfaces  | `--name enp0s1 --dhcp4`                                                                                                                                                                                                                                                                                                                                                                                                                 | also takes enp0s2              |
| An interface in a match that takes it alone, for now | `--name enp0s1 --dhcp4`                                                                                                                                                                                                                                                                                                                                                                                                                 | any interface it matches later |
| An interface the host does not have                  | `--name eth9 --dhcp4`                                                                                                                                                                                                                                                                                                                                                                                                                   | not found                      |
| An address without a prefix                          | `--name osapi0 --addresses 10.0.1.5`                                                                                                                                                                                                                                                                                                                                                                                                    | Addresses                      |
| More than 16 addresses                               | `--name osapi0 --addresses 10.0.1.5/24 --addresses 10.0.1.5/24 --addresses 10.0.1.5/24 --addresses 10.0.1.5/24 --addresses 10.0.1.5/24 --addresses 10.0.1.5/24 --addresses 10.0.1.5/24 --addresses 10.0.1.5/24 --addresses 10.0.1.5/24 --addresses 10.0.1.5/24 --addresses 10.0.1.5/24 --addresses 10.0.1.5/24 --addresses 10.0.1.5/24 --addresses 10.0.1.5/24 --addresses 10.0.1.5/24 --addresses 10.0.1.5/24 --addresses 10.0.1.5/24` | Addresses                      |
| An IPv4 gateway that is not IPv4                     | `--name osapi0 --gateway4 fe80::1`                                                                                                                                                                                                                                                                                                                                                                                                      | Gateway4                       |
| An IPv6 gateway that is not IPv6                     | `--name osapi0 --gateway6 10.0.1.1`                                                                                                                                                                                                                                                                                                                                                                                                     | Gateway6                       |
| An MTU below 68                                      | `--name osapi0 --mtu 10`                                                                                                                                                                                                                                                                                                                                                                                                                | MTU                            |
| An MTU above 9000                                    | `--name osapi0 --mtu 9001`                                                                                                                                                                                                                                                                                                                                                                                                              | MTU                            |
| A MAC address netplan does not take                  | `--name osapi0 --mac-address 0000.5e00.5301`                                                                                                                                                                                                                                                                                                                                                                                            | MACAddress                     |
| A name that is not an interface name                 | `--name "eth0;reboot"`                                                                                                                                                                                                                                                                                                                                                                                                                  | Name                           |

## Permission

It needs the `interface:write` permission.
`osapi controller role get --name <role>` shows what a role grants.

## Flags

| Flag            | Description                                                                       | Default |
| --------------- | --------------------------------------------------------------------------------- | ------- |
| `--addresses`   | The static addresses, in CIDR notation, such as 10.0.0.5/24                       |         |
| `--dhcp4`       | Take an IPv4 address from DHCP                                                    |         |
| `--dhcp6`       | Take an IPv6 address from DHCP                                                    |         |
| `--gateway4`    | The IPv4 default gateway                                                          |         |
| `--gateway6`    | The IPv6 default gateway                                                          |         |
| `--mac-address` | The MAC address to set                                                            |         |
| `--mtu`         | The MTU, from 68 to 9000                                                          |         |
| `--name`        | The interface, such as eth0, or a fact such as @fact.interface.primary (required) |         |
| `-T, --target`  | Target: _all, _any, hostname, or label (group:web.dev)                            | `_all`  |
| `--wakeonlan`   | Turn on wake-on-LAN                                                               |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
