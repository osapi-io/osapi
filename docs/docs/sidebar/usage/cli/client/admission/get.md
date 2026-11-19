# Get

Get a machine's admission record.

Returns one machine's admission record by machine ID.

```bash
osapi client admission get [flags]
```

## Examples

```bash
osapi client admission get --machine-id d530664edc1c431eb5ef1e9dfd6617a5
```

## Permission

It needs the `admission:read` permission. `osapi token roles list` lists the
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
