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
osapi pki cert rotate --name nats --json   # with an overlap, on the controller's host
```

Each of the five is in the list, with a `state`: `ok`; `not made`, when that
component has not started on this host; or `unreadable`, with the reason, such
as a file that only root can read. A client config, such as one that points at a
controller on another host, reads no certificate: run the command on the host of
the server, or read the fleet through the API.

## The fleet, through the API

`osapi client pki cert` reads the certificates of each host from any machine
with a client config. The controller reports the certificates of its host, and
each agent reports its own in its registration, at each heartbeat:

```sh
osapi client pki cert list --json                     # each host's certificates
osapi client pki cert list --hostname web-01 --json   # one host's
osapi client pki cert get --name api --json           # the controller's host
osapi client pki cert get --name agent-metrics --hostname web-01 --json
```

Each item has the `hostname` of its host, and the fields of `osapi pki cert`.
Rotate stays a command on the host: it writes the files there.

`fips_approved` is true when FIPS 186-5 approves the key (RSA of 2048 bits or
more, ECDSA P-256, P-384 or P-521, Ed25519) and the signature (no MD5, no
SHA-1). It is about the certificate only. FIPS 140-3 mode is a setting of each
running process, not of a certificate, so this command does not show it.

A metrics certificate follows the host's name: a server that starts on a renamed
host makes a new one for its new name. Nothing pins it. The API's and NATS's
certificates keep the names that they were made with.

Rotate refuses a `cert_file` of the operator: replace that file. Each server
reads its certificate files again when they change, at its next handshake, so a
new certificate takes effect with no restart, also one of the operator's.

Each joined agent pins the NATS server's certificate, so `rotate --name nats`
rotates it with an overlap. It makes `nats.next.crt` beside the current one, and
writes a record to the state bucket that the controller's key signs. Each agent
adds the next certificate to the ones it trusts at its next heartbeat;
`osapi client agent get` shows it as `nats_cert_next`. When each admitted agent
trusts it, or `controller.pki.rotation_grace_period` ends, the controller
renames it into place. Run it on the controller's host: the controller's key
signs the record.

From any machine with a client config, the API starts the same rotation and
shows its progress, with each agent that the switch waits for:

```sh
osapi client pki nats rotate --json   # start it, or show the one in progress
osapi client pki nats get --json      # current, next, and each agent's has_next
```

A rotation of `api` changes the fingerprint that join tokens carry: a token made
before it no longer joins. A CLI on another host needs the new `api.crt` as
`controller.client.tls.ca_file`.
