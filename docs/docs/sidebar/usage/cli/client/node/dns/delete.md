# Delete

Drop the DNS resolvers osapi set on an interface.

Removes the interface's netplan file and applies what is left. An interface
osapi set nothing on does not change.

```bash
osapi client node dns delete [flags]
```

## Examples

```bash
# An interface osapi set
osapi client node dns delete --target web-01 --interface-name osapi0

# An interface osapi set nothing on
osapi client node dns delete --target web-01 --interface-name osapi0
```

## Result

An interface osapi set, on one host:

```ansi
$ osapi client node dns delete --target web-01 --interface-name osapi0
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mINTERFACE NAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  osapi0

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node dns delete --target _all --interface-name osapi0
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mINTERFACE NAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  osapi0
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m  osapi0

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
      "interface_name": "osapi0",
      "status": "ok"
    }
  ]
}
```

An interface osapi set nothing on, on one host:

```ansi
$ osapi client node dns delete --target web-01 --interface-name osapi0
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mINTERFACE NAME[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      osapi0

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node dns delete --target _all --interface-name osapi0
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mINTERFACE NAME[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      osapi0
  [1mweb-02[0m    [38;2;103;234;147mok[0m      osapi0

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
      "interface_name": "osapi0",
      "status": "ok"
    }
  ]
}
```

With the default settings, each example does this on the host:

| Example                | On the host                                                        |
| ---------------------- | ------------------------------------------------------------------ |
| An interface osapi set | removes `/etc/netplan/osapi-dns-osapi0.yaml`; runs `netplan apply` |

The command fails in these cases:

| Call                                       | Flags                              | Fails with    |
| ------------------------------------------ | ---------------------------------- | ------------- |
| An interface name that is not alphanumeric | `--interface-name "osapi0;reboot"` | InterfaceName |

## Permission

It needs the `dns:write` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Flags

| Flag               | Description                                                                       | Default |
| ------------------ | --------------------------------------------------------------------------------- | ------- |
| `--interface-name` | The interface, such as eth0, or a fact such as @fact.interface.primary (required) |         |
| `-T, --target`     | Target: _all, _any, hostname, or label (group:web.dev)                            | `_all`  |

## Global flags

| Flag               | Description                                                 | Default                 |
| ------------------ | ----------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                |                         |
| `-j, --json`       | Enable JSON output                                          |                         |
| `-f, --osapi-file` | Path to config file                                         | `/etc/osapi/osapi.yaml` |
| `--url`            | Base URL the CLI client connects to (controller.client.url) | `http://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
