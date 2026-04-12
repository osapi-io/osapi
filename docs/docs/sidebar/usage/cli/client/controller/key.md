# Key

Controller key management commands. The `key` subcommand group provides
operations on the local controller's PKI key pair.

## fingerprint

Show the local controller's public key fingerprint:

```bash
$ osapi client controller key fingerprint

  Fingerprint: SHA256:ef78ab90cd12...
```

The fingerprint is derived from the controller's Ed25519 public key stored in
the PKI key directory (default `/etc/osapi/pki`). Agents use this fingerprint to
verify that jobs originate from a trusted controller.

If no key exists yet, the command generates a new key pair automatically before
displaying the fingerprint.

### Flags

| Flag     | Description     | Required |
| -------- | --------------- | -------- |
| `--json` | Output raw JSON | No       |
