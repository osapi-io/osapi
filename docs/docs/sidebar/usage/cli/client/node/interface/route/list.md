# List

List the routes of an interface.

Reads the routes of the main table that go out of the interface: ip route on
Linux, netstat on macOS. proto says where a route came from, such as dhcp,
kernel or static. An interface with no routes has none; one the host does not
have is not found.

```bash
osapi client node interface route list [flags]
```

## Examples

```bash
# An interface with routes
osapi client node interface route list --target web-01 --name eth0

# An interface with no routes
osapi client node interface route list --target web-01 --name eth1

# A Mac
osapi client node interface route list --target web-01 --name en0
```

## Result

An interface with routes, on one host:

```ansi
$ osapi client node interface route list --target web-01 --name eth0
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mDESTINATION   [0m[38;2;10;143;176mGATEWAY     [0m[38;2;10;143;176mINTERFACE  [0m[38;2;10;143;176mFAMILY  [0m[38;2;10;143;176mMETRIC  [0m[38;2;10;143;176mPROTO   [0m[38;2;10;143;176mSCOPE[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      default       10.0.0.1    eth0       inet    100     dhcp
  [38;2;138;138;138mweb-01[0m    [38;2;103;234;147mok[0m      10.0.0.0/24               eth0       inet    0       kernel  link
  [38;2;138;138;138mweb-01[0m    [38;2;103;234;147mok[0m      10.20.0.0/16  10.0.0.254  eth0       inet    50      static

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node interface route list --target _all --name eth0
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mDESTINATION   [0m[38;2;10;143;176mGATEWAY     [0m[38;2;10;143;176mINTERFACE  [0m[38;2;10;143;176mFAMILY  [0m[38;2;10;143;176mMETRIC  [0m[38;2;10;143;176mPROTO   [0m[38;2;10;143;176mSCOPE[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      default       10.0.0.1    eth0       inet    100     dhcp
  [38;2;138;138;138mweb-01[0m    [38;2;103;234;147mok[0m      10.0.0.0/24               eth0       inet    0       kernel  link
  [38;2;138;138;138mweb-01[0m    [38;2;103;234;147mok[0m      10.20.0.0/16  10.0.0.254  eth0       inet    50      static
  [1mweb-02[0m    [38;2;103;234;147mok[0m      default       10.0.0.1    eth0       inet    100     dhcp
  [38;2;138;138;138mweb-02[0m    [38;2;103;234;147mok[0m      10.0.0.0/24               eth0       inet    0       kernel  link
  [38;2;138;138;138mweb-02[0m    [38;2;103;234;147mok[0m      10.20.0.0/16  10.0.0.254  eth0       inet    50      static

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
      "destination": "default",
      "family": "inet",
      "gateway": "10.0.0.1",
      "hostname": "web-01",
      "interface": "eth0",
      "metric": 100,
      "proto": "dhcp",
      "scope": "",
      "status": "ok"
    },
    {
      "destination": "10.0.0.0/24",
      "family": "inet",
      "gateway": "",
      "hostname": "web-01",
      "interface": "eth0",
      "metric": 0,
      "proto": "kernel",
      "scope": "link",
      "status": "ok"
    },
    {
      "destination": "10.20.0.0/16",
      "family": "inet",
      "gateway": "10.0.0.254",
      "hostname": "web-01",
      "interface": "eth0",
      "metric": 50,
      "proto": "static",
      "scope": "",
      "status": "ok"
    }
  ]
}
```

An interface with no routes, on one host:

```ansi
$ osapi client node interface route list --target web-01 --name eth1
[1;38;2;0;213;255mResults:[0m
  [38;2;138;138;138mNone.[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node interface route list --target _all --name eth1
[1;38;2;0;213;255mResults:[0m

  1 host: [38;5;222m1 skipped[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m

[1;38;2;255;0;255mDetails:[0m
  [1mmac-01[0m  [38;5;222mskipped[0m  operation not supported on this OS family
```

With `--json`, the API's response:

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "results": []
}
```

A Mac, on one host:

```ansi
$ osapi client node interface route list --target web-01 --name en0
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mDESTINATION  [0m[38;2;10;143;176mGATEWAY      [0m[38;2;10;143;176mINTERFACE  [0m[38;2;10;143;176mFAMILY  [0m[38;2;10;143;176mMETRIC  [0m[38;2;10;143;176mPROTO  [0m[38;2;10;143;176mSCOPE[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      default      192.168.0.1  en0        inet    0

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node interface route list --target _all --name en0
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mDESTINATION  [0m[38;2;10;143;176mGATEWAY      [0m[38;2;10;143;176mINTERFACE  [0m[38;2;10;143;176mFAMILY  [0m[38;2;10;143;176mMETRIC  [0m[38;2;10;143;176mPROTO  [0m[38;2;10;143;176mSCOPE[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      default      192.168.0.1  en0        inet    0
  [1mweb-02[0m    [38;2;103;234;147mok[0m      default      192.168.0.1  en0        inet    0

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
      "destination": "default",
      "family": "inet",
      "gateway": "192.168.0.1",
      "hostname": "web-01",
      "interface": "en0",
      "metric": 0,
      "proto": "",
      "scope": "",
      "status": "ok"
    }
  ]
}
```

The command fails in these cases:

| Call                                 | Flags                  | Fails with |
| ------------------------------------ | ---------------------- | ---------- |
| An interface the host does not have  | `--name eth9`          | not found  |
| A name that is not an interface name | `--name "eth0;reboot"` | Name       |

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

[Output](../../../../output.md) tells how the result is printed, and what
`--json` gives.
