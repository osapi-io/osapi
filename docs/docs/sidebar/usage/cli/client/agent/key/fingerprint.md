# Fingerprint

Show the local agent key fingerprint.

osapi calculates the fingerprint from the Ed25519 public key of the agent, in
the PKI key directory (`/etc/osapi/pki` by default). Enrollment requests include
this value. You can use it with `agent accept --fingerprint` to accept an agent
by its fingerprint. If no key exists, the command makes a new key pair. Then it
shows the fingerprint.

```bash
osapi client agent key fingerprint
```

## Examples

```bash
osapi client agent key fingerprint
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
