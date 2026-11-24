# Rotate

Rotate the key of a nats vault.

The command adds a new key to the `key_file` of the vault, and keeps the old key
for reads. Then it connects to NATS as the controller does, encrypts each value
of the vault again with the new key, and removes the old key from the file. A
controller that runs reads the file at each call, so it needs no restart. If the
rotation stops, the file keeps the old key, and a second rotation completes it.
The earlier versions of a value stay under the old key, and do not decrypt after
the rotation.

```bash
osapi controller vault key rotate [flags]
```

## Examples

```bash
osapi controller vault key rotate --name default
```

## Result

The command prints:

```ansi
$ osapi controller vault key rotate --name default
[1;38;2;0;213;255mVault key:[0m
  [38;2;10;143;176mName:[0m      default
  [38;2;10;143;176mKey file:[0m  /etc/osapi/secrets.key
  [38;2;10;143;176mKey ID:[0m    a07c5e91d3b8f246
  [38;2;10;143;176mValues:[0m    12
```

With `--json`, the command prints:

```json
{
  "name": "default",
  "key_file": "/etc/osapi/secrets.key",
  "key_id": "a07c5e91d3b8f246",
  "values": 12
}
```

## Flags

| Flag     | Description                                                                  | Default |
| -------- | ---------------------------------------------------------------------------- | ------- |
| `--name` | The vault, as controller.secrets.vaults names it, such as default (required) |         |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
