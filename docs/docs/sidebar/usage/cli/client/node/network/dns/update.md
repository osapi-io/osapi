# Update

Set the DNS resolvers of an interface.

Writes a netplan file of the interface's own, such as
/etc/netplan/osapi-dns-eth0.yaml for eth0, under the netplan definition that
matches the interface, and applies it. A definition that matches several
interfaces, such as cloud-init's all for en*, gives each of them the resolvers.
An empty list keeps the interface's servers or domains, and an interface with no
resolvers yet starts from none. override_dhcp drops the servers DHCP gives. It
runs on the Debian family, and elsewhere, and in a container, reports skipped.

```bash
osapi client node network dns update [flags]
```

## Examples

```bash
# Set the servers of an interface
osapi client node network dns update --target web-01 --interface-name osapi0 --servers 9.9.9.9

# Set the search domains, and keep the servers
osapi client node network dns update --target web-01 --interface-name osapi0 --search-domains corp.example --override-dhcp

# An interface with no resolvers yet
osapi client node network dns update --target web-01 --interface-name osapi0 --servers 9.9.9.9
```

## Result

With the default settings, each example does this on the host:

| Example                                      | On the host                                                                                                                                                                                                                                                                                                                                                                  |
| -------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Set the servers of an interface              | writes `/etc/netplan/osapi-dns-osapi0.yaml`: `network:\n  version: 2\n  ethernets:\n    osapi0:\n      nameservers:\n        addresses:\n          - 9.9.9.9\n        search:\n          - example.com`; runs `netplan generate`; runs `netplan apply`                                                                                                                       |
| Set the search domains, and keep the servers | writes `/etc/netplan/osapi-dns-osapi0.yaml`: `network:\n  version: 2\n  ethernets:\n    osapi0:\n      nameservers:\n        addresses:\n          - 1.1.1.1\n          - 8.8.8.8\n        search:\n          - corp.example\n      dhcp4-overrides:\n        use-dns: false\n      dhcp6-overrides:\n        use-dns: false`; runs `netplan generate`; runs `netplan apply` |
| An interface with no resolvers yet           | writes `/etc/netplan/osapi-dns-osapi0.yaml`: `network:\n  version: 2\n  ethernets:\n    osapi0:\n      nameservers:\n        addresses:\n          - 9.9.9.9`; runs `netplan generate`; runs `netplan apply`                                                                                                                                                                 |

The command fails in these cases:

| Call                                                        | Flags                                                                                                                                                                                                                                                                                                                                                                                          | Fails with                     |
| ----------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------ |
| No servers and no search domains                            | `--interface-name osapi0`                                                                                                                                                                                                                                                                                                                                                                      | give servers or search domains |
| An interface the host does not have, to keep the servers of | `--interface-name eth9 --search-domains corp.example`                                                                                                                                                                                                                                                                                                                                          | not found                      |
| A server that is not an IP address                          | `--interface-name osapi0 --servers one.one`                                                                                                                                                                                                                                                                                                                                                    | ip                             |
| A search domain that is not a host name                     | `--interface-name osapi0 --search-domains "bad domain"`                                                                                                                                                                                                                                                                                                                                        | hostname                       |
| More than 16 servers                                        | `--interface-name osapi0 --servers 1.1.1.1 --servers 1.1.1.1 --servers 1.1.1.1 --servers 1.1.1.1 --servers 1.1.1.1 --servers 1.1.1.1 --servers 1.1.1.1 --servers 1.1.1.1 --servers 1.1.1.1 --servers 1.1.1.1 --servers 1.1.1.1 --servers 1.1.1.1 --servers 1.1.1.1 --servers 1.1.1.1 --servers 1.1.1.1 --servers 1.1.1.1 --servers 1.1.1.1`                                                    | Servers                        |
| More than 16 search domains                                 | `--interface-name osapi0 --search-domains a.b --search-domains a.b --search-domains a.b --search-domains a.b --search-domains a.b --search-domains a.b --search-domains a.b --search-domains a.b --search-domains a.b --search-domains a.b --search-domains a.b --search-domains a.b --search-domains a.b --search-domains a.b --search-domains a.b --search-domains a.b --search-domains a.b` | SearchDomains                  |
| An interface name that is not alphanumeric                  | `--interface-name "osapi0;reboot" --servers 9.9.9.9`                                                                                                                                                                                                                                                                                                                                           | InterfaceName                  |

## Permission

It needs the `network:write` permission. `osapi token roles get --name <role>`
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

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../../output.md) tells how the result is printed, and what
`--json` gives.
