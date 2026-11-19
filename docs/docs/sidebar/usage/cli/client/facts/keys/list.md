# List

List the fact keys.

Lists each fact key that an input can refer to as @fact.&lt;key&gt;, with its
description.

```bash
osapi client facts keys list
```

## Examples

```bash
osapi client facts keys list
osapi client facts keys list --json
```

## Permission

It needs the `agent:read` permission. `osapi token roles list` lists the roles
that grant it.

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
