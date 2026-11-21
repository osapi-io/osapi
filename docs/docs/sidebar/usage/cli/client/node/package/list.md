# List

List the packages installed on the host.

Lists each installed package with its version: dpkg on the Debian family, rpm on
the other Linux families, and Homebrew on macOS.

```bash
osapi client node package list [flags]
```

## Examples

```bash
# A Debian host
osapi client node package list --target web-01

# A Fedora host
osapi client node package list --target web-01

# A Mac
osapi client node package list --target web-01
```

## Permission

It needs the `package:read` permission. `osapi token roles get --name <role>`
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
