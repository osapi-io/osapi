# Admit

Admit a pending machine.

Run `osapi agent admission` on the machine, and give the fingerprint that it
prints. The controller admits the machine only when the fingerprint is the one
of the key in its request.

```bash
osapi client admission admit [flags]
```

## Examples

```bash
osapi client admission admit --machine-id d530664edc1c431eb5ef1e9dfd6617a5 --fingerprint SHA256:1f53dea6...
```

## Permission

It needs the `admission:write` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Flags

| Flag            | Description                                                                 | Default |
| --------------- | --------------------------------------------------------------------------- | ------- |
| `--fingerprint` | The fingerprint that osapi agent admission prints on the machine (required) |         |
| `--machine-id`  | The machine ID of the node (required)                                       |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
