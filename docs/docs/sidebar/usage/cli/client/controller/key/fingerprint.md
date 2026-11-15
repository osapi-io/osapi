# Fingerprint

Show the local controller key fingerprint.

osapi calculates the fingerprint from the Ed25519 public key of the controller,
in the PKI key directory (`/etc/osapi/pki` by default). Agents use this
fingerprint to verify that jobs come from a trusted controller. If no key
exists, the command makes a new key pair. Then it shows the fingerprint.

```bash
osapi client controller key fingerprint
```

## Examples

```bash
osapi client controller key fingerprint
```

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
