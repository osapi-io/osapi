# Update

Set the hostname of the host.

Sets the static hostname with hostnamectl set-hostname, which keeps it after a
reboot. A host that already has the name does not change and reports changed:
false. It runs on the Debian family only, and elsewhere reports skipped. In a
container the runtime owns the hostname, so there it also reports skipped.

```bash
osapi client node hostname update [flags]
```

## Examples

```bash
# Rename a host
osapi client node hostname update --target web-01 --name web-02

# A host that already has the name
osapi client node hostname update --target web-01 --name web-01
```

## Result

With the default settings, each example does this on the host:

| Example       | On the host                            |
| ------------- | -------------------------------------- |
| Rename a host | runs `hostnamectl set-hostname web-02` |

The command fails in these cases:

| Call                              | Flags                                                                                                                                                                                                                                                                   | Fails with       |
| --------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------- |
| A name that is not a hostname     | `--name "web 01;reboot"`                                                                                                                                                                                                                                                | invalid hostname |
| A name longer than 253 characters | `--name aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa` | invalid hostname |
| No name                           |                                                                                                                                                                                                                                                                         | invalid hostname |

## Permission

It needs the `hostname:write` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Flags

| Flag           | Description                                            | Default |
| -------------- | ------------------------------------------------------ | ------- |
| `--name`       | The hostname to set, such as web-01 (required)         |         |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev) | `_all`  |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
