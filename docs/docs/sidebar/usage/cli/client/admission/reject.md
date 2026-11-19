# Reject

Reject a pending machine.

Rejects a pending machine, and keeps a record of the decision, so a machine that
asks again is flagged.

```bash
osapi client admission reject [flags]
```

## Examples

```bash
osapi client admission reject --machine-id d530664edc1c431eb5ef1e9dfd6617a5 --reason "unknown host"
```

## Permission

It needs the `admission:write` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Flags

| Flag           | Description                                                     | Default |
| -------------- | --------------------------------------------------------------- | ------- |
| `--machine-id` | The machine ID of the node (required)                           |         |
| `--reason`     | Why the machine is rejected; the machine and the record keep it |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
