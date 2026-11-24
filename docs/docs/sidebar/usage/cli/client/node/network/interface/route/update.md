# Update

Change the gateway or the metric of a route.

Writes the route over the one osapi has to the same destination, and applies it.
A destination osapi has no route to is not found: create adds it.

```bash
osapi client node network interface route update [flags]
```

## Examples

```bash
# A new gateway and metric
osapi client node network interface route update --target web-01 --name osapi0 --destination 198.51.100.0/24 --via 192.0.2.2 --metric 10
```

## Result

A new gateway and metric, on one host:

```ansi
$ osapi client node network interface route update --target web-01 --name osapi0 --destination 198.51.100.0/24 --via 192.0.2.2 --metric 10
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mDESTINATION[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  198.51.100.0/24

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node network interface route update --target _all --name osapi0 --destination 198.51.100.0/24 --via 192.0.2.2 --metric 10
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mDESTINATION[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  198.51.100.0/24
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m  198.51.100.0/24

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
      "destination": "198.51.100.0/24",
      "hostname": "web-01",
      "status": "ok"
    }
  ]
}
```

With the default settings, each example does this on the host:

| Example                  | On the host                                                                                                                                                                                                                                      |
| ------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| A new gateway and metric | runs `netplan generate`; runs `netplan apply`; writes `/etc/netplan/osapi-osapi0-routes.yaml`: `network:\n  version: 2\n  ethernets:\n    osapi0:\n      routes:\n        - to: 198.51.100.0/24\n          via: 192.0.2.2\n          metric: 10` |

The command fails in these cases:

| Call                                 | Flags                                                                             | Fails with  |
| ------------------------------------ | --------------------------------------------------------------------------------- | ----------- |
| A destination osapi has no route to  | `--name osapi0 --destination 198.51.100.0/24 --via 192.0.2.1`                     | not found   |
| A destination that is not CIDR       | `--name osapi0 --destination 10.0.0.0 --via 192.0.2.1`                            | Destination |
| A gateway that is not an address     | `--name osapi0 --destination 198.51.100.0/24 --via gateway`                       | Via         |
| A metric below 0                     | `--name osapi0 --destination 198.51.100.0/24 --via 192.0.2.1 --metric -1`         | Metric      |
| A metric over 4294967295             | `--name osapi0 --destination 198.51.100.0/24 --via 192.0.2.1 --metric 4294967296` | Metric      |
| A name that is not an interface name | `--name "eth0;reboot" --destination 198.51.100.0/24 --via 192.0.2.1`              | Name        |

## Permission

It needs the `interface:write` permission.
`osapi controller role get --name <role>` shows what a role grants.

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
