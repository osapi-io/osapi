# Create

Make a key that is not on this host.

It makes the key with the name, with mode 0600: the signing key of the
controller, the two keys of the agent, a random key for the JetStream store, or
a random AES-256 key for a vault. It refuses a key that is there, because a new
key in its place is a rotation: a JetStream store and the values of a vault do
not decrypt with it, and an admission does not match it. The controller and the
agent make their keys at their first start, so run this to make a key before
then, such as the JetStream key before NATS starts. Keep a copy of a JetStream
or a vault key apart from the backup of what it encrypts.

```bash
osapi pki key create [flags]
```

## Examples

```bash
osapi pki key create --name jetstream
```

## Result

The command prints:

```ansi
$ osapi pki key create --name jetstream
[1;38;2;0;213;255mKey:[0m
  [38;2;10;143;176mName:[0m       jetstream
  [38;2;10;143;176mKind:[0m       encryption
  [38;2;10;143;176mComponent:[0m  nats
  [38;2;10;143;176mState:[0m      ok
  [38;2;10;143;176mFile:[0m       /etc/osapi/nats/jetstream.key
  [38;2;10;143;176mAlgorithm:[0m  ChaCha20-Poly1305
  [38;2;10;143;176mKey ID:[0m     7f3a9c1e5b2d8f40
```

With `--json`, the command prints:

```json
{
  "name": "jetstream",
  "kind": "encryption",
  "component": "nats",
  "state": "ok",
  "file": "/etc/osapi/nats/jetstream.key",
  "algorithm": "ChaCha20-Poly1305",
  "key_id": "7f3a9c1e5b2d8f40"
}
```

## Flags

| Flag     | Description                                                                          | Default |
| -------- | ------------------------------------------------------------------------------------ | ------- |
| `--name` | The name of the key, as list shows it, such as jetstream or vault:default (required) |         |

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
