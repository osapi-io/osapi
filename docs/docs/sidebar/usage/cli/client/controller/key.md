# Key

Commands for the controller key. The `key` subcommands operate on the PKI key
pair of the local controller.

## fingerprint

Show the local controller's public key fingerprint:

```bash
$ osapi client controller key fingerprint

  Fingerprint: SHA256:ef78ab90cd12...
```

osapi calculates the fingerprint from the Ed25519 public key of the controller,
in the PKI key directory (`/etc/osapi/pki` by default). Agents use this
fingerprint to verify that jobs come from a trusted controller.

If no key exists, the command makes a new key pair. Then it shows the
fingerprint.

### Flags

| Flag     | Description     | Required |
| -------- | --------------- | -------- |
| `--json` | Output raw JSON | No       |
