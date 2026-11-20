# Status

Show the health of the controller.

Each controller, with each of its sub-components, from the health status of the
controller.

```bash
osapi client controller status
```

## Examples

```bash
osapi client controller status
osapi client controller status --json
```

## Permission

It needs the `health:read` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
