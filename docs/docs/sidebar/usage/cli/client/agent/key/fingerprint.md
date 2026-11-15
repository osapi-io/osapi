# Fingerprint

Show the local agent's public key fingerprint:

```bash
$ osapi client agent key fingerprint

  Fingerprint: SHA256:ab12cd34ef56...
```

osapi calculates the fingerprint from the Ed25519 public key of the agent, in
the PKI key directory (`/etc/osapi/pki` by default). Enrollment requests include
this value. You can use it with `agent accept --fingerprint` to accept an agent
by its fingerprint.

If no key exists, the command makes a new key pair. Then it shows the
fingerprint.

## Flags

| Flag     | Description     | Required |
| -------- | --------------- | -------- |
| `--json` | Output raw JSON | No       |
