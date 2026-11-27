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

| Where                                      | What it shows                               |
| ------------------------------------------ | ------------------------------------------- |
| `osapi client health status`               | The FIPS column of the controller and NATS  |
| `osapi client health status --json`        | The `crypto` of each component              |
| `osapi client agent get --hostname <name>` | The FIPS line and the crypto of the agent   |
| The dashboard and the agent card in the UI | The FIPS line, with each gap in its tooltip |
| `osapi pki list` and `osapi pki key list`  | Whether FIPS 140-3 approves each key        |

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
  approved algorithms. The FIPS line is the mode, then `approved` when there is
  no gap, or the number of gaps.

A test holds `only_mode` to what the code does. Each algorithm that claims to
operate in only mode is in a package that the `fips140=only` job tests. For NATS
authentication and the JetStream store, the job runs the same call that nkeys
and nats-server make, and expects it to fail. When a fix upstream makes that
call operate, the test fails until the report says so.
