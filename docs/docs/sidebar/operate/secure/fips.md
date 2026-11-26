---
sidebar_label: FIPS 140-3 Mode
---

# FIPS 140-3 mode

FIPS 140-3 is the US standard for cryptographic modules. Go has its own
cryptographic module, the Go Cryptographic Module. A Go program in FIPS 140-3
mode uses only that module for the algorithms of the standard library. An
approved algorithm is an algorithm that FIPS 140-3 approves, such as AES-GCM,
ECDH on P-256, SHA-256 and HKDF. FIPS 140-3 does not approve X25519,
ChaCha20-Poly1305, MD5, or SHA-1 for a signature.

This page tells which parts of osapi use only approved algorithms, how to get
the FIPS build, and what does not operate in FIPS 140-3 only mode now. FIPS
140-3 mode is not a validation of osapi. A validation applies to a module, not
to a program that uses it.

## The seal and the agent keys

The seal of job secrets and the encryption key of each agent use only approved
algorithms from the Go standard library:

| Part                       | Algorithm                                |
| -------------------------- | ---------------------------------------- |
| The encryption key         | ECDH on P-256, in `agent.ekey`           |
| The key-encryption key     | HKDF-SHA256 of the ECDH secret           |
| The job key and each value | AES-256-GCM, with a random 12-byte nonce |
| The fingerprint            | SHA-256                                  |
| The signing key            | Ed25519                                  |

[Secrets](../../architecture/secrets.md#the-wire-format) has the wire format.

Before osapi-io/osapi#1382, the encryption key was an X25519 key in
`agent.xkey`, and the seal used NaCl `box` and XChaCha20-Poly1305. An agent that
you upgrade makes a new P-256 key and removes `agent.xkey`. Its fingerprint
changes, so it must join again.
[Agent identity](../../architecture/agent-identity.md#an-agent-admitted-with-an-x25519-key)
has the procedure. Until it joins again, a job with a secret for that agent
fails at the controller.

## The FIPS build

Each release has two binaries for each platform: `osapi` and `osapi-fips`. The
two binaries have the same code. The release builds `osapi-fips` with
`GOFIPS140=v1.0.0`:

- It links the frozen Go Cryptographic Module v1.0.0, the version that the Go
  team sent to the CMVP for validation.
- It starts in FIPS 140-3 mode (`GODEBUG=fips140=on`). In this mode, the
  standard library uses only the module, and does its self-tests at startup.

To see the settings of a binary, run `go version -m osapi-fips`. It shows
`GOFIPS140=v1.0.0` and `DefaultGODEBUG=fips140=on`.

To build it from the source:

```bash
GOFIPS140=v1.0.0 go build -o osapi-fips .
```

## FIPS 140-3 only mode

With `GODEBUG=fips140=only`, Go refuses each algorithm that is not approved,
with an error or a panic. Do not use this mode for osapi now. These parts of
osapi do not operate in it:

- **NATS authentication.** osapi signs each NATS JWT with an nkeys Ed25519 key.
  nkeys v0.4.17 makes the key from its seed with a reader that is not
  `crypto/rand.Reader`, and only mode refuses that. Thus the controller and the
  agents cannot authenticate to NATS (osapi-io/osapi#1527).
- **The JetStream store encryption.** With `nats.server.jetstream.key_file`, the
  embedded NATS server encrypts its store with ChaCha20-Poly1305, which only
  mode refuses. The server does not make a stream (osapi-io/osapi#1528).
- **The NATS vault.** It uses AES-GCM with a nonce that it selects, which only
  mode refuses (osapi-io/osapi#1529).
- **Facts from gohai.** The SSH host key facts and the shard of a host use MD5
  (osapi-io/gohai#357).

The `go-fips-test` recipe runs the tests of the seal and of the agent keys in
only mode. CI runs it in the `fips140=only` job.
