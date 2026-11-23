# Revoke

Revoke an admitted machine.

The controller removes the key of the machine. Its responses and its
registration stop verifying at once, so a target does not find it.

```bash
osapi client admission revoke [flags]
```

## Examples

```bash
osapi client admission revoke --machine-id d530664edc1c431eb5ef1e9dfd6617a5
```

## Result

The command prints:

```ansi
$ osapi client admission revoke --machine-id d530664edc1c431eb5ef1e9dfd6617a5
[ok] d530664edc1c431eb5ef1e9dfd6617a5 revoked

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

With `--json`, the command prints:

```json
{ "machine_id": "d530664edc1c431eb5ef1e9dfd6617a5", "result": "revoked" }
```

## Permission

It needs the `admission:write` permission. `osapi token roles get --name <role>`
shows what a role grants.

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
