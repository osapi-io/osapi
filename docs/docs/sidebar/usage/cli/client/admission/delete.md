# Delete

Delete a machine.

The controller removes the key of the machine and each record of it. Unlike
revoke, it keeps no record, so a machine that asks again is new. Use it before a
machine is rebuilt or removed.

```bash
osapi client admission delete [flags]
```

## Examples

```bash
osapi client admission delete --machine-id d530664edc1c431eb5ef1e9dfd6617a5
```

## Permission

It needs the `admission:write` permission. `osapi token roles list` lists the
roles that grant it.

## Flags

| Flag           | Description                           | Default |
| -------------- | ------------------------------------- | ------- |
| `--machine-id` | The machine ID of the node (required) |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
