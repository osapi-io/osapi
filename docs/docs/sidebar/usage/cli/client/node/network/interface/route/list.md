# List

List the routes of an interface.

Reads the routes of the main table that go out of the interface: ip route on
Linux, netstat on macOS. proto says where a route came from, such as dhcp,
kernel or static. An interface with no routes has none; one the host does not
have is not found.

```bash
osapi client node network interface route list [flags]
```

## Examples

```bash
# An interface with routes
osapi client node network interface route list --target web-01 --name eth0

# An interface with no routes
osapi client node network interface route list --target web-01 --name eth1

# A Mac
osapi client node network interface route list --target web-01 --name en0
```

## Result

Each example prints this:

An interface with routes:

```ansi
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mDESTINATION   [0m[38;2;10;143;176mGATEWAY     [0m[38;2;10;143;176mINTERFACE  [0m[38;2;10;143;176mFAMILY  [0m[38;2;10;143;176mMETRIC  [0m[38;2;10;143;176mPROTO   [0m[38;2;10;143;176mSCOPE[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      default       10.0.0.1    eth0       inet    100     dhcp
  [38;2;138;138;138mweb-01[0m    [38;2;103;234;147mok[0m      10.0.0.0/24               eth0       inet    0       kernel  link
  [38;2;138;138;138mweb-01[0m    [38;2;103;234;147mok[0m      10.20.0.0/16  10.0.0.254  eth0       inet    50      static

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

An interface with no routes:

```ansi
[1;38;2;0;213;255mResults:[0m
  [38;2;138;138;138mNone.[0m
```

A Mac:

```ansi
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mDESTINATION  [0m[38;2;10;143;176mGATEWAY      [0m[38;2;10;143;176mINTERFACE  [0m[38;2;10;143;176mFAMILY  [0m[38;2;10;143;176mMETRIC  [0m[38;2;10;143;176mPROTO  [0m[38;2;10;143;176mSCOPE[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      default      192.168.0.1  en0        inet    0

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

The command fails in these cases:

| Call                                 | Flags                  | Fails with |
| ------------------------------------ | ---------------------- | ---------- |
| An interface the host does not have  | `--name eth9`          | not found  |
| A name that is not an interface name | `--name "eth0;reboot"` | Name       |

## Permission

It needs the `network:read` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Flags

| Flag           | Description                                                                       | Default |
| -------------- | --------------------------------------------------------------------------------- | ------- |
| `--name`       | The interface, such as eth0, or a fact such as @fact.interface.primary (required) |         |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev)                            | `_all`  |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../../../output.md) tells how the result is printed, and what
`--json` gives.
