# List

List admission records.

Each node that asked to join, each node osapi admitted, and each decision
against a node. An admitted node that is not running is Offline, with when it
was last seen.

```bash
osapi client admission list [flags]
```

## Examples

```bash
osapi client admission list
osapi client admission list --state pending
```

## Permission

It needs the `admission:read` permission. `osapi token roles list` lists the
roles that grant it.

## Flags

| Flag      | Description                                                                                                   | Default |
| --------- | ------------------------------------------------------------------------------------------------------------- | ------- |
| `--state` | Only the records in this state: pending, admitted, rejected or revoked (pending, admitted, rejected, revoked) |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
