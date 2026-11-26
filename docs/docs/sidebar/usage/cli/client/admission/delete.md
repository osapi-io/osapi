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

## Result

The command prints:

```ansi
$ osapi client admission delete --machine-id d530664edc1c431eb5ef1e9dfd6617a5
[38;2;103;234;147m✓[0m d530664edc1c431eb5ef1e9dfd6617a5 deleted

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

With `--json`, the command prints:

```json
{ "machine_id": "d530664edc1c431eb5ef1e9dfd6617a5", "result": "deleted" }
```

## Permission

It needs the `admission:write` permission.
`osapi controller role get --name <role>` shows what a role grants.

## Flags

| Flag           | Description                           | Default |
| -------------- | ------------------------------------- | ------- |
| `--machine-id` | The machine ID of the node (required) |         |

## Global flags

| Flag               | Description                                                                      | Default                  |
| ------------------ | -------------------------------------------------------------------------------- | ------------------------ |
| `-d, --debug`      | Enable or disable debug mode                                                     |                          |
| `-j, --json`       | Enable JSON output                                                               |                          |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml`  |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `https://localhost:8080` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
