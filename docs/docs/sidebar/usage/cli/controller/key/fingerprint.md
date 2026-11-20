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

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
