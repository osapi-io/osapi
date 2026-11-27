# Get

Show one key in full; the agent's key has the command that admits this machine.

It reads one key, as list does. The agent's signing key also has the command
that admits this machine, with its machine ID and the fingerprint that covers
the agent's two keys. A key that is not on this host is an error that names
osapi pki key create.

```bash
osapi pki key get [flags]
```

## Examples

```bash
osapi pki key get --name agent
```

## Result

The command prints:

```ansi
$ osapi pki key get --name agent
[1;38;2;0;213;255mKey:[0m
  [38;2;10;143;176mName:[0m         agent
  [38;2;10;143;176mKind:[0m         signing
  [38;2;10;143;176mComponent:[0m    agent
  [38;2;10;143;176mState:[0m        ok
  [38;2;10;143;176mFile:[0m         /etc/osapi/pki/agent/agent.key
  [38;2;10;143;176mAlgorithm:[0m    Ed25519
  [38;2;10;143;176mApproved:[0m     yes
  [38;2;10;143;176mFingerprint:[0m  SHA256:4b2a7c6d5e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b
  [38;2;10;143;176mTo admit:[0m     osapi client admission admit --machine-id 4c4c4544-0035-4b10-8031-b4c04f4d3332 --fingerprint SHA256:4b2a7c6d5e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b
```

With `--json`, the command prints:

```json
{
  "name": "agent",
  "kind": "signing",
  "component": "agent",
  "state": "ok",
  "file": "/etc/osapi/pki/agent/agent.key",
  "algorithm": "Ed25519",
  "approved": true,
  "fingerprint": "SHA256:4b2a7c6d5e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b",
  "admit": "osapi client admission admit --machine-id 4c4c4544-0035-4b10-8031-b4c04f4d3332 --fingerprint SHA256:4b2a7c6d5e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b"
}
```

## Flags

| Flag     | Description                                                                                             | Default |
| -------- | ------------------------------------------------------------------------------------------------------- | ------- |
| `--name` | The name of the key, as list shows it, such as controller, agent, jetstream or vault:default (required) |         |

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
