# Pending

List agents awaiting enrollment.

When PKI is enabled, a new agent sends an enrollment request, and waits here
until an administrator accepts or rejects it with `agent accept` or
`agent reject`.

```bash
osapi client agent pending
```

## Examples

```bash
osapi client agent pending
```

## Permission

It needs the `agent:read` permission. `osapi token roles` lists the roles that
grant it.

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
