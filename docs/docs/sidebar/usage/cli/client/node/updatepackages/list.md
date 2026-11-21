# List

List the packages the index has newer versions of.

Runs apt list --upgradable: each package with its installed version and the
version the index has. Refresh the index first for what the repositories have
now.

```bash
osapi client node updatepackages list [flags]
```

## Examples

```bash
# Two updates
osapi client node updatepackages list --target web-01
```

## Permission

It needs the `updatepackages:read` permission.
`osapi token roles get --name <role>` shows what a role grants.

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
