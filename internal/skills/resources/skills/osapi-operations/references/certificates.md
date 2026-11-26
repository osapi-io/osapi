# Certificates

The API, the NATS server and the three metrics servers serve TLS by default.
With no `cert_file`, each makes a self-signed ECDSA P-256 certificate in the key
directory of its component at its first start: `api.crt` and
`controller-metrics.crt` in `controller.pki.key_dir`, `nats.crt` and
`nats-metrics.crt` in `nats.server.key_dir`, and `agent-metrics.crt` in
`agent.pki.key_dir`. Run the commands on the host; they read the config there.

```sh
osapi pki cert list --json                 # each certificate, expiry, key, FIPS
osapi pki cert get --name nats --json      # names, fingerprint, signature
osapi pki cert rotate --name api --json    # a new self-signed certificate
osapi pki cert rotate --name nats --force  # cuts off each joined agent
```

`fips` is true when FIPS 186-5 approves the key (RSA of 2048 bits or more, ECDSA
P-256, P-384 or P-521, Ed25519) and the signature (no MD5, no SHA-1).
`fips_mode` in the list is true when the binary runs its crypto in FIPS 140-3
mode, with `GODEBUG=fips140=on`.

Rotate refuses a `cert_file` of the operator: replace that file. The server
serves a new certificate at its next start. Each join string pins the NATS
server's certificate, and each joined agent trusts it, so after
`rotate --name nats --force` make new join tokens and join each agent again. A
CLI on another host needs the new `api.crt` as `controller.client.tls.ca_file`.
