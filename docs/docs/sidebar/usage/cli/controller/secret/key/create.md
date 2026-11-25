# Create

Make the key of a nats vault.

The command writes a new random AES-256 key to the `key_file` of the vault in
`controller.secrets.vaults`, with the mode 0600. The vault encrypts each value
with this key. The command does not replace a key that exists, because each
value under it would then not decrypt. Use `rotate` for that. It prints the ID
of the key, not the key. The controller makes the key at start when the file is
not there, so run this only to make the key before that.

```bash
osapi controller secret key create [flags]
```

## Examples

```bash
osapi controller secret key create --vault default
```

## Result

The command prints:

```ansi
$ osapi controller secret key create --vault default
[1;38;2;0;213;255mVault key:[0m
  [38;2;10;143;176mName:[0m      default
  [38;2;10;143;176mKey file:[0m  /etc/osapi/secrets.key
  [38;2;10;143;176mKey ID:[0m    3f9a1c0e7b2d4a58
```

With `--json`, the command prints:

```json
{
  "name": "default",
  "key_file": "/etc/osapi/secrets.key",
  "key_id": "3f9a1c0e7b2d4a58"
}
```

## Flags

| Flag      | Description                                                                  | Default |
| --------- | ---------------------------------------------------------------------------- | ------- |
| `--vault` | The vault, as controller.secrets.vaults names it, such as default (required) |         |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
