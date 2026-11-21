# Update

Refresh the package index.

Runs apt-get update. It reports changed: true when the Release file of a
repository under /var/lib/apt/lists changed, and changed: false when each
repository was unchanged.

```bash
osapi client node updatepackages update [flags]
```

## Examples

```bash
# The repositories are unchanged since the last refresh
osapi client node updatepackages update --target web-01
```

## Permission

It needs the `updatepackages:write` permission.
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
