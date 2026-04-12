# Key

Agent key management commands. The `key` subcommand group provides
operations on the local agent's PKI key pair.

## fingerprint

Show the local agent's public key fingerprint:

```bash
$ osapi client agent key fingerprint

  Fingerprint: SHA256:ab12cd34ef56...
```

The fingerprint is derived from the agent's Ed25519 public key stored
in the PKI key directory (default `/etc/osapi/pki`). This value is
included in enrollment requests and can be used with
`agent accept --fingerprint` for fingerprint-based acceptance.

If no key exists yet, the command generates a new key pair
automatically before displaying the fingerprint.

### Flags

| Flag     | Description    | Required |
| -------- | -------------- | -------- |
| `--json` | Output raw JSON | No       |
