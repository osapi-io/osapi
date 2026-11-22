# Create

Add a static route to an interface.

Adds the route to the interface's netplan file, such as
/etc/netplan/osapi-osapi0-routes.yaml for osapi0, under a netplan definition
that is the interface's own, and applies it. A route osapi already has, the
same, does not change; one to the same destination by another gateway or metric
is refused: update changes it. A default route is refused: the interface
behavior sets the gateway. An interface the host does not have is not found. An
interface that a match defines, such as cloud-init's all for en*, is refused,
because the route would go on each interface the match takes. It runs on the
Debian family, and elsewhere, and in a container, reports skipped.

```bash
osapi client node network interface route create [flags]
```

## Examples

```bash
# A route
osapi client node network interface route create --target web-01 --name osapi0 --destination 198.51.100.0/24 --via 192.0.2.1

# A second route, with a metric
osapi client node network interface route create --target web-01 --name osapi0 --destination 2001:db8:1::/48 --via 2001:db8::1 --metric 200

# A route osapi already has
osapi client node network interface route create --target web-01 --name osapi0 --destination 198.51.100.0/24 --via 192.0.2.1
```

## Result

Each example prints this:

A route:

```ansi
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mDESTINATION[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  198.51.100.0/24

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

A second route, with a metric:

```ansi
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mDESTINATION[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  2001:db8:1::/48

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

A route osapi already has:

```ansi
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mDESTINATION[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      198.51.100.0/24

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

With the default settings, each example does this on the host:

| Example                       | On the host                                                                                                                                                                                                                                                                                                  |
| ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| A route                       | runs `netplan generate`; runs `netplan apply`; writes `/etc/netplan/osapi-osapi0-routes.yaml`: `network:\n  version: 2\n  ethernets:\n    osapi0:\n      routes:\n        - to: 198.51.100.0/24\n          via: 192.0.2.1`                                                                                   |
| A second route, with a metric | runs `netplan generate`; runs `netplan apply`; writes `/etc/netplan/osapi-osapi0-routes.yaml`: `network:\n  version: 2\n  ethernets:\n    osapi0:\n      routes:\n        - to: 198.51.100.0/24\n          via: 192.0.2.1\n        - to: 2001:db8:1::/48\n          via: 2001:db8::1\n          metric: 200` |

The command fails in these cases:

| Call                                                 | Flags                                                                             | Fails with                     |
| ---------------------------------------------------- | --------------------------------------------------------------------------------- | ------------------------------ |
| A default route                                      | `--name osapi0 --destination 0.0.0.0/0 --via 192.0.2.1`                           | default route                  |
| A gateway of the other family                        | `--name osapi0 --destination 198.51.100.0/24 --via 2001:db8::1`                   | not an address of the family   |
| Another route to a destination osapi has             | `--name osapi0 --destination 198.51.100.0/24 --via 192.0.2.2`                     | update changes it              |
| An interface in a match that takes other interfaces  | `--name enp0s1 --destination 198.51.100.0/24 --via 192.0.2.1`                     | also takes enp0s2              |
| An interface in a match that takes it alone, for now | `--name enp0s1 --destination 198.51.100.0/24 --via 192.0.2.1`                     | any interface it matches later |
| An interface the host does not have                  | `--name eth9 --destination 198.51.100.0/24 --via 192.0.2.1`                       | not found                      |
| A destination that is not CIDR                       | `--name osapi0 --destination 10.0.0.0 --via 192.0.2.1`                            | Destination                    |
| A gateway that is not an address                     | `--name osapi0 --destination 198.51.100.0/24 --via gateway`                       | Via                            |
| A metric below 0                                     | `--name osapi0 --destination 198.51.100.0/24 --via 192.0.2.1 --metric -1`         | Metric                         |
| A metric over 4294967295                             | `--name osapi0 --destination 198.51.100.0/24 --via 192.0.2.1 --metric 4294967296` | Metric                         |
| A name that is not an interface name                 | `--name "eth0;reboot" --destination 198.51.100.0/24 --via 192.0.2.1`              | Name                           |

## Permission

It needs the `network:write` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Flags

| Flag            | Description                                                                                | Default |
| --------------- | ------------------------------------------------------------------------------------------ | ------- |
| `--destination` | Where the route goes, in CIDR notation, such as 10.0.0.0/8; not a default route (required) |         |
| `--metric`      | The metric, from 0 to 4294967295; none when it is 0                                        |         |
| `--name`        | The interface, such as eth0, or a fact such as @fact.interface.primary (required)          |         |
| `-T, --target`  | Target: _all, _any, hostname, or label (group:web.dev)                                     | `_all`  |
| `--via`         | The gateway, an address of the destination's family, such as 192.168.1.1 (required)        |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../../../output.md) tells how the result is printed, and what
`--json` gives.
