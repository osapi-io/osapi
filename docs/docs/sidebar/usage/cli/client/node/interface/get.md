# Get

Get a network interface of the host.

Reads the interface as list does.

```bash
osapi client node interface get [flags]
```

## Examples

```bash
# An interface with a DHCP address
osapi client node interface get --target web-01 --name eth0

# A Mac
osapi client node interface get --target web-01 --name en0
```

## Result

An interface with a DHCP address, on one host:

```ansi
$ osapi client node interface get --target web-01 --name eth0
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME  [0m[38;2;10;143;176mSTATE  [0m[38;2;10;143;176mMAC                [0m[38;2;10;143;176mMTU   [0m[38;2;10;143;176mIPV4      [0m[38;2;10;143;176mIPV6         [0m[38;2;10;143;176mADDRESSES                                                [0m[38;2;10;143;176mDHCP4  [0m[38;2;10;143;176mPRIMARY[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1meth0[0m  up     52:54:00:12:34:56  1500  10.0.0.5  2001:db8::5  10.0.0.5/24, 2001:db8::5/64, fe80::5054:ff:fe12:3456/64  yes    yes

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node interface get --target _all --name eth0
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME  [0m[38;2;10;143;176mSTATE  [0m[38;2;10;143;176mMAC                [0m[38;2;10;143;176mMTU   [0m[38;2;10;143;176mIPV4      [0m[38;2;10;143;176mIPV6         [0m[38;2;10;143;176mADDRESSES                                                [0m[38;2;10;143;176mDHCP4  [0m[38;2;10;143;176mPRIMARY[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1meth0[0m  up     52:54:00:12:34:56  1500  10.0.0.5  2001:db8::5  10.0.0.5/24, 2001:db8::5/64, fe80::5054:ff:fe12:3456/64  yes    yes
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [1meth0[0m  up     52:54:00:12:34:56  1500  10.0.0.5  2001:db8::5  10.0.0.5/24, 2001:db8::5/64, fe80::5054:ff:fe12:3456/64  yes    yes

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
      "addresses": [
        "10.0.0.5/24",
        "2001:db8::5/64",
        "fe80::5054:ff:fe12:3456/64"
      ],
      "dhcp4": true,
      "hostname": "web-01",
      "ipv4": "10.0.0.5",
      "ipv6": "2001:db8::5",
      "mac": "52:54:00:12:34:56",
      "mtu": 1500,
      "name": "eth0",
      "primary": true,
      "state": "up",
      "status": "ok"
    }
  ]
}
```

A Mac, on one host:

```ansi
$ osapi client node interface get --target web-01 --name en0
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME  [0m[38;2;10;143;176mSTATE  [0m[38;2;10;143;176mMAC                [0m[38;2;10;143;176mMTU   [0m[38;2;10;143;176mIPV4          [0m[38;2;10;143;176mIPV6  [0m[38;2;10;143;176mADDRESSES        [0m[38;2;10;143;176mDHCP4  [0m[38;2;10;143;176mPRIMARY[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1men0[0m   up     3c:22:fb:00:00:01  1500  192.168.0.40        192.168.0.40/24  no     yes

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node interface get --target _all --name en0
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME  [0m[38;2;10;143;176mSTATE  [0m[38;2;10;143;176mMAC                [0m[38;2;10;143;176mMTU   [0m[38;2;10;143;176mIPV4          [0m[38;2;10;143;176mIPV6  [0m[38;2;10;143;176mADDRESSES        [0m[38;2;10;143;176mDHCP4  [0m[38;2;10;143;176mPRIMARY[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1men0[0m   up     3c:22:fb:00:00:01  1500  192.168.0.40        192.168.0.40/24  no     yes
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [1men0[0m   up     3c:22:fb:00:00:01  1500  192.168.0.40        192.168.0.40/24  no     yes

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
      "addresses": ["192.168.0.40/24"],
      "dhcp4": false,
      "hostname": "web-01",
      "ipv4": "192.168.0.40",
      "ipv6": "",
      "mac": "3c:22:fb:00:00:01",
      "mtu": 1500,
      "name": "en0",
      "primary": true,
      "state": "up",
      "status": "ok"
    }
  ]
}
```

The command fails in these cases:

| Call                                 | Flags                      | Fails with |
| ------------------------------------ | -------------------------- | ---------- |
| An interface the host does not have  | `--name eth9`              | not found  |
| A name that is not an interface name | `--name "eth0;reboot"`     | Name       |
| A fact the agent does not know       | `--name @fact.nonexistent` | Name       |

## Permission

It needs the `interface:read` permission.
`osapi controller role get --name <role>` shows what a role grants.

## Flags

| Flag           | Description                                                                       | Default |
| -------------- | --------------------------------------------------------------------------------- | ------- |
| `--name`       | The interface, such as eth0, or a fact such as @fact.interface.primary (required) |         |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev)                            | `_all`  |

## Global flags

| Flag               | Description                                                 | Default                 |
| ------------------ | ----------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                |                         |
| `-j, --json`       | Enable JSON output                                          |                         |
| `-f, --osapi-file` | Path to config file                                         | `/etc/osapi/osapi.yaml` |
| `--url`            | Base URL the CLI client connects to (controller.client.url) | `http://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
