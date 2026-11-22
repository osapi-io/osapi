# List

List the services of the host.

Lists each service that systemd has loaded, with its state, whether it starts at
boot and its description, through gohai. A service that osapi deployed shows by
the name it was deployed with, not its unit's osapi- name, and is managed. A Mac
has no systemd, and reports skipped.

```bash
osapi client node service list [flags]
```

## Examples

```bash
# A service that osapi deployed reads by its name
osapi client node service list --target web-01

# A host with two services
osapi client node service list --target web-01
```

## Permission

It needs the `service:read` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Flags

| Flag           | Description                                            | Default |
| -------------- | ------------------------------------------------------ | ------- |
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
