# Fingerprint

Show the fingerprint of the controller's key.

osapi calculates the fingerprint from the Ed25519 public key of the controller,
in `controller.pki.key_dir` (`/etc/osapi/pki` by default). Agents use this
fingerprint to verify that jobs come from a trusted controller. If no key
exists, the command makes a new key pair. Then it shows the fingerprint.

```bash
osapi controller key fingerprint
```

## Examples

```bash
osapi controller key fingerprint
```

## Result

The command prints:

```ansi
$ osapi controller key fingerprint
[1;38;2;0;213;255mController key:[0m
  [38;2;10;143;176mFingerprint:[0m  SHA256:4b8e2a6c0d9f1e3a5c7b9d2f4a6c8e0b1d3f5a7c9e2b4d6f8a0c2e4b6d8f0a1c
  [38;2;10;143;176mKey dir:[0m      /etc/osapi/pki
```

With `--json`, the command prints:

```json
{
  "fingerprint": "SHA256:4b8e2a6c0d9f1e3a5c7b9d2f4a6c8e0b1d3f5a7c9e2b4d6f8a0c2e4b6d8f0a1c",
  "key_dir": "/etc/osapi/pki"
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
