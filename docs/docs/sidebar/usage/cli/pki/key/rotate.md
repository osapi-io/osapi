# Rotate

Rotate the agent's keys or the key of a vault.

For the agent, it makes two new keys beside the old ones and sends them to the
controller over NATS, as the agent connects, signed with the old signing key and
the new one. The controller takes them in place of the admitted keys, with no
new admission, and keeps the old signing key valid for
controller.pki.rotation_grace_period. Then it moves the new keys in place, and a
running agent reads them at its next heartbeat. A refusal removes the new keys.
When the controller does not answer, the new keys stay, and a second rotation
sends them again.

For a vault of type nats, it adds a new key to its key_file, encrypts each value
of the vault again with it, then drops the old key. It connects to NATS as the
controller does. A running controller reads the new key at its next call. A
rotation that stops leaves the old key in the file, and a second rotation
completes it.

It refuses each other key, with what to do instead: a new JetStream key makes
the store unreadable, and a new controller key needs each agent admitted again.

```bash
osapi pki key rotate [flags]
```

## Examples

```bash
osapi pki key rotate --name vault:default
```

## Result

The command prints:

```ansi
$ osapi pki key rotate --name vault:default
[1;38;2;0;213;255mRotated:[0m
  [38;2;176;48;176mValues:[0m  12

[1;38;2;255;0;255mKey:[0m
  [38;2;176;48;176mName:[0m       vault:default
  [38;2;176;48;176mKind:[0m       encryption
  [38;2;176;48;176mComponent:[0m  vault
  [38;2;176;48;176mState:[0m      ok
  [38;2;176;48;176mFile:[0m       /etc/osapi/secrets.key
  [38;2;176;48;176mAlgorithm:[0m  AES-256-GCM
  [38;2;176;48;176mKey ID:[0m     b8e14a2c9d3f6071
  [38;2;176;48;176mKeys:[0m       1
```

With `--json`, the command prints:

```json
{
  "key": {
    "name": "vault:default",
    "kind": "encryption",
    "component": "vault",
    "state": "ok",
    "file": "/etc/osapi/secrets.key",
    "algorithm": "AES-256-GCM",
    "key_id": "b8e14a2c9d3f6071",
    "keys": 1
  },
  "values": 12
}
```

## Flags

| Flag     | Description                                                                                   | Default |
| -------- | --------------------------------------------------------------------------------------------- | ------- |
| `--name` | The name of the key, as list shows it: agent, or a vault key such as vault:default (required) |         |

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
