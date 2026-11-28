# Update

Set the DNS resolvers of an interface.

Writes a netplan file of the interface's own, such as
/etc/netplan/osapi-dns-eth0.yaml for eth0, under the netplan definition that
matches the interface, and applies it. A definition that matches several
interfaces, such as cloud-init's all for en\*, gives each of them the resolvers.
An empty list writes nothing for that list, so the interface keeps what DHCP
gives it. override_dhcp drops the servers and domains DHCP gives, and then an
empty list keeps the interface's current ones in the file; an interface with no
resolvers yet starts from none. It runs on the Debian family, and elsewhere, and
in a container, reports skipped.

```bash
osapi client node dns update [flags]
```

## Examples

```bash
# Set the servers of an interface
osapi client node dns update --target web-01 --interface-name osapi0 --servers 9.9.9.9

# Set the search domains, and keep the servers
osapi client node dns update --target web-01 --interface-name osapi0 --search-domains corp.example --override-dhcp

# An interface with no resolvers yet, with DHCP's dropped
osapi client node dns update --target web-01 --interface-name osapi0 --servers 9.9.9.9 --override-dhcp
```

## Result

Set the servers of an interface, on one host:

```ansi
$ osapi client node dns update --target web-01 --interface-name osapi0 --servers 9.9.9.9
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mINTERFACE NAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  osapi0

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node dns update --target _all --interface-name osapi0 --servers 9.9.9.9
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

Set the search domains, and keep the servers, on one host:

```ansi
$ osapi client node dns update --target web-01 --interface-name osapi0 --search-domains corp.example --override-dhcp
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mINTERFACE NAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  osapi0

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node dns update --target _all --interface-name osapi0 --search-domains corp.example --override-dhcp
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

An interface with no resolvers yet, with DHCP's dropped, on one host:

```ansi
$ osapi client node dns update --target web-01 --interface-name osapi0 --servers 9.9.9.9 --override-dhcp
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mINTERFACE NAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  osapi0

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node dns update --target _all --interface-name osapi0 --servers 9.9.9.9 --override-dhcp
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

With the default settings, each example does this on the host:

| Example                                                 | On the host                                                                                                                                                                                                                                                                                                                                                                  |
| ------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Set the servers of an interface                         | writes `/etc/netplan/osapi-dns-osapi0.yaml`: `network:\n  version: 2\n  ethernets:\n    osapi0:\n      nameservers:\n        addresses:\n          - 9.9.9.9`; runs `netplan generate`; runs `netplan apply`                                                                                                                                                                 |
| Set the search domains, and keep the servers            | writes `/etc/netplan/osapi-dns-osapi0.yaml`: `network:\n  version: 2\n  ethernets:\n    osapi0:\n      nameservers:\n        addresses:\n          - 1.1.1.1\n          - 8.8.8.8\n        search:\n          - corp.example\n      dhcp4-overrides:\n        use-dns: false\n      dhcp6-overrides:\n        use-dns: false`; runs `netplan generate`; runs `netplan apply` |
| An interface with no resolvers yet, with DHCP's dropped | writes `/etc/netplan/osapi-dns-osapi0.yaml`: `network:\n  version: 2\n  ethernets:\n    osapi0:\n      nameservers:\n        addresses:\n          - 9.9.9.9\n      dhcp4-overrides:\n        use-dns: false\n      dhcp6-overrides:\n        use-dns: false`; runs `netplan generate`; runs `netplan apply`                                                                 |

The command fails in these cases:

| Call                                                        | Flags                                                                                                                                                                                                                                                                                                                                                                                          | Fails with                     |
| ----------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------ |
| No servers and no search domains                            | `--interface-name osapi0`                                                                                                                                                                                                                                                                                                                                                                      | give servers or search domains |
| An interface the host does not have, to keep the servers of | `--interface-name eth9 --search-domains corp.example`                                                                                                                                                                                                                                                                                                                                          | not found                      |
| A server that is not an IP address                          | `--interface-name osapi0 --servers one.one`                                                                                                                                                                                                                                                                                                                                                    | ip                             |
| A search domain that is not a host name                     | `--interface-name osapi0 --search-domains 'bad domain'`                                                                                                                                                                                                                                                                                                                                        | hostname                       |
| More than 16 servers                                        | `--interface-name osapi0 --servers 1.1.1.1 --servers 1.1.1.1 --servers 1.1.1.1 --servers 1.1.1.1 --servers 1.1.1.1 --servers 1.1.1.1 --servers 1.1.1.1 --servers 1.1.1.1 --servers 1.1.1.1 --servers 1.1.1.1 --servers 1.1.1.1 --servers 1.1.1.1 --servers 1.1.1.1 --servers 1.1.1.1 --servers 1.1.1.1 --servers 1.1.1.1 --servers 1.1.1.1`                                                    | Servers                        |
| More than 16 search domains                                 | `--interface-name osapi0 --search-domains a.b --search-domains a.b --search-domains a.b --search-domains a.b --search-domains a.b --search-domains a.b --search-domains a.b --search-domains a.b --search-domains a.b --search-domains a.b --search-domains a.b --search-domains a.b --search-domains a.b --search-domains a.b --search-domains a.b --search-domains a.b --search-domains a.b` | SearchDomains                  |
| An interface name that is not alphanumeric                  | `--interface-name 'osapi0;reboot' --servers 9.9.9.9`                                                                                                                                                                                                                                                                                                                                           | InterfaceName                  |

## Permission

It needs the `dns:write` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Flags

| Flag               | Description                                                                       | Default |
| ------------------ | --------------------------------------------------------------------------------- | ------- |
| `--interface-name` | The interface, such as eth0, or a fact such as @fact.interface.primary (required) |         |
| `--override-dhcp`  | Use only these servers: drop the ones DHCP gives                                  |         |
| `--search-domains` | The search domains; the interface's own when empty                                |         |
| `--servers`        | The DNS servers, IPv4 or IPv6; the interface's own when empty                     |         |
| `-T, --target`     | Target: _all, _any, hostname, or label (group:web.dev)                            | `_all`  |

## Global flags

| Flag               | Description                                                                      | Default                  |
| ------------------ | -------------------------------------------------------------------------------- | ------------------------ |
| `-d, --debug`      | Enable or disable debug mode                                                     |                          |
| `-j, --json`       | Enable JSON output                                                               |                          |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml`  |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `https://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
