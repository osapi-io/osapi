# Admit

Admit a pending machine.

Run `osapi pki key get --name agent` on the machine, and give the fingerprint
that it prints. The controller admits the machine only when the fingerprint is
the one of the key in its request.

```bash
osapi client admission admit [flags]
```

## Examples

```bash
osapi client admission admit --machine-id d530664edc1c431eb5ef1e9dfd6617a5 --fingerprint SHA256:1f53dea6...
```

## Result

The command prints:

```ansi
$ osapi client admission admit --machine-id d530664edc1c431eb5ef1e9dfd6617a5 --fingerprint SHA256:1f53dea6...
[38;2;103;234;147m✓[0m d530664edc1c431eb5ef1e9dfd6617a5 admitted

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

With `--json`, the command prints:

```json
{ "machine_id": "d530664edc1c431eb5ef1e9dfd6617a5", "result": "admitted" }
```

## Permission

It needs the `admission:write` permission.
`osapi controller role get --name <role>` shows what a role grants.

## Flags

| Flag            | Description                                                                          | Default |
| --------------- | ------------------------------------------------------------------------------------ | ------- |
| `--fingerprint` | The fingerprint that osapi pki key get --name agent prints on the machine (required) |         |
| `--machine-id`  | The machine ID of the node (required)                                                |         |

## Global flags

| Flag               | Description                                                                      | Default                  |
| ------------------ | -------------------------------------------------------------------------------- | ------------------------ |
| `-d, --debug`      | Enable or disable debug mode                                                     |                          |
| `-j, --json`       | Enable JSON output                                                               |                          |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml`  |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `https://localhost:8080` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
