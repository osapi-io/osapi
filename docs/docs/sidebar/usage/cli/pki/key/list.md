# List

List the keys, with their state, algorithm and fingerprint or ID.

Run this on the host. It reads each key of osapi that the config names: the
controller's Ed25519 signing key in controller.pki.key_dir, the agent's Ed25519
signing key and X25519 encryption key in agent.pki.key_dir, the key of the
JetStream store in nats.server.jetstream.key_file, and the key of each vault of
type nats. A public key shows its fingerprint, and a secret key shows its ID:
never the key. A component that has not started on this host made no key: its
state is not made.

```bash
osapi pki key list
```

## Examples

```bash
osapi pki key list
```

## Result

The command prints:

```ansi
$ osapi pki key list
[1;38;2;0;213;255mKeys (5):[0m
  [38;2;10;143;176mNAME              [0m[38;2;10;143;176mKIND        [0m[38;2;10;143;176mCOMPONENT   [0m[38;2;10;143;176mSTATE  [0m[38;2;10;143;176mALGORITHM          [0m[38;2;10;143;176mKEY ID[0m
  [1mcontroller[0m        signing     controller  ok     Ed25519            [38;2;138;138;138m[0m
  [1magent[0m             signing     agent       ok     Ed25519            [38;2;138;138;138m[0m
  [1magent-encryption[0m  encryption  agent       ok     X25519             [38;2;138;138;138m[0m
  [1mjetstream[0m         encryption  nats        ok     ChaCha20-Poly1305  [38;2;138;138;138m7f3a9c1e5b2d8f40[0m
  [1mvault:default[0m     encryption  vault       ok     AES-256-GCM        [38;2;138;138;138mb8e14a2c9d3f6071[0m
```

With `--json`, the command prints:

```json
{
  "items": [
    {
      "name": "controller",
      "kind": "signing",
      "component": "controller",
      "state": "ok",
      "file": "/etc/osapi/pki/controller.key",
      "algorithm": "Ed25519",
      "fingerprint": "SHA256:9c1f2e3d4b5a69788796a5b4c3d2e1f0a9b8c7d6e5f4a3b2c1d0e9f8a7b6c5d4"
    },
    {
      "name": "agent",
      "kind": "signing",
      "component": "agent",
      "state": "ok",
      "file": "/etc/osapi/pki/agent/agent.key",
      "algorithm": "Ed25519",
      "fingerprint": "SHA256:4b2a7c6d5e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b",
      "admit": "osapi client admission admit --machine-id 4c4c4544-0035-4b10-8031-b4c04f4d3332 --fingerprint SHA256:4b2a7c6d5e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b"
    },
    {
      "name": "agent-encryption",
      "kind": "encryption",
      "component": "agent",
      "state": "ok",
      "file": "/etc/osapi/pki/agent/agent.xkey",
      "algorithm": "X25519",
      "public_key": "XA3NQJ6DPZ2Y7C4ULM5WRSVBKTEHGOIF63QXN4J2PDLW7YKZSMBUCRVE"
    },
    {
      "name": "jetstream",
      "kind": "encryption",
      "component": "nats",
      "state": "ok",
      "file": "/etc/osapi/nats/jetstream.key",
      "algorithm": "ChaCha20-Poly1305",
      "key_id": "7f3a9c1e5b2d8f40"
    },
    {
      "name": "vault:default",
      "kind": "encryption",
      "component": "vault",
      "state": "ok",
      "file": "/etc/osapi/secrets.key",
      "algorithm": "AES-256-GCM",
      "key_id": "b8e14a2c9d3f6071",
      "keys": 1
    }
  ]
}
```

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
