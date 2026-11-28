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
  embedded NATS server encrypts its store with AES-256-GCM, an approved
  algorithm. But the server makes the GCM with a nonce that it selects, which
  only mode refuses. Thus the server does not make an encrypted stream in only
  mode (osapi-io/osapi#1549). A store with no `key_file` is not encrypted, and
  operates in only mode.

The `go-fips-test` recipe runs the tests of the seal, the NATS vault and the
agent keys in only mode. CI runs it in the `fips140=only` job.

## The crypto of each process

FIPS 140-3 mode is a setting of each process, not of the fleet. `GODEBUG` sets
it when the process starts. Thus each process reports its own crypto: the
controller, the NATS server and each agent. The report goes in each heartbeat.

| Where                                      | What it shows                                                      |
| ------------------------------------------ | ------------------------------------------------------------------ |
| `osapi client pki tls list`                | The TLS section of each component, one line each, with its summary |
| `osapi client pki tls get --component <c>` | One section in full: the reasons, certificates, listeners, crypto  |
| `osapi client health status`               | The FIPS column of the controller and NATS                         |
| `osapi client agent get --hostname <name>` | The FIPS line and the crypto of the agent                          |
| The dashboard and the agent card in the UI | The TLS summary, which opens the section                           |
| `osapi pki list` and `osapi pki key list`  | Whether FIPS 140-3 approves each key                               |

The report of a process has:

- **The Go version and the module version**, from `crypto/fips140.Version()`.
- **The mode:** `off`, `on`, or `only`.
- **Each TLS listener**, such as `api`, `nats` or `metrics`: its lowest TLS
  version, its suites and its groups. Each listener and each client of osapi
  uses only approved suites and groups, also out of FIPS 140-3 mode: the AES-GCM
  suites, and P-256, P-384, P-521 and P-256 with ML-KEM-768. Thus a listener is
  approved in each mode.

  crypto/tls has no setting for the suites of TLS 1.3. Out of FIPS 140-3 mode,
  it can select ChaCha20-Poly1305 for TLS 1.3, which is not approved: when the
  client prefers it, offers no AES-GCM, or when the CPU has no AES in hardware.
  For such a client, a listener of osapi offers TLS 1.2 only, with the approved
  suites. A client that has no approved suite or group cannot connect.

- **Each algorithm that the process uses**, with `approved` and `only_mode`. An
  approved algorithm can still fail in only mode, as NATS authentication does.
  `only_mode` says so, and `reason` names the issue.
- **The gaps:** each reason that the process is not in FIPS 140-3 mode with only
  approved algorithms.

The FIPS line is the mode, then what is not approved, by kind. With nothing that
is not approved, it is the mode alone (`off`), or the mode with `approved`:

```text
off
on · approved
off · 1 TLS listener not approved
only · 2 TLS listeners and 1 algorithm not approved
```

## The TLS section of a component

`osapi client pki tls` shows one section for each component: the controller, the
NATS server and each agent. The UI shows the same section when you click the
summary on the dashboard or on the card of an agent.

```bash
osapi client pki tls list
osapi client pki tls get --component nats
osapi client pki tls get --component agent --hostname web-01
```

A section has these parts:

- **The summary:** the FIPS line and the days left of the certificate that
  expires first, such as `FIPS off · cert 62d`. The summary goes yellow when a
  certificate has 30 days or less, and red when a certificate is expired.
- **FIPS 140-3:** the mode, and each reason as a sentence, with what closes it.
  A reason can name an issue. In FIPS 140-3 only mode, the NATS server refuses
  two algorithms that FIPS 140-3 approves. The nkeys library signs NATS JWTs
  (osapi-io/osapi#1527), and nats-server makes its own GCM nonce for the
  JetStream store (osapi-io/osapi#1549). Thus the section of the NATS server
  shows both as reasons until those issues close.
- **Certificates:** each certificate that the component serves, from
  `pki cert list`. Each has its subject, its names, its issuer, its validity and
  the days left. Its expiry is `ok`, `soon` (30 days or less), `expired`, or
  `unknown` when the component could not read it.
- **Listeners:** each TLS listener, with its lowest version, its suites and its
  groups, and whether FIPS 140-3 approves them.
- **Versions and algorithms:** the Go version, the version of the Go
  Cryptographic Module, and each algorithm with whether FIPS 140-3 approves it.

A test holds `only_mode` to what the code does. Each algorithm that claims to
operate in only mode is in a package that the `fips140=only` job tests. For NATS
authentication and the JetStream store, the job runs the same call that nkeys
and nats-server make, and expects it to fail. When a fix upstream makes that
call operate, the test fails until the report says so.
