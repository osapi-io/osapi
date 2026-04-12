---
sidebar_position: 26
sidebar_label: Agent Identity & PKI
---

# Agent Identity & PKI

OSAPI agents identify themselves using a persistent machine ID and
support optional PKI enrollment for cryptographic trust between agents
and the controller.

## Machine Identity

Every agent has two identity values:

- **Machine ID** -- permanent identifier from `/etc/machine-id` (Linux)
  or `IOPlatformUUID` (macOS). Used as the registry key and stable
  reference across hostname changes.
- **Hostname** -- mutable display name from the OS or `agent.hostname`
  config. Used for human-friendly targeting and display.

The agent resolves its machine ID once at startup and refuses to start
if it cannot be read. NATS subject routing uses the machine ID
(`jobs.query.host.<machineID>`), so consumers never need to
resubscribe when the hostname changes. The hostname is re-read on
each heartbeat tick (10s) and updated in the registry for display.

### CLI Targeting

Both hostname and machine ID work as targets. The controller resolves
hostnames to machine IDs for NATS subject routing automatically:

```bash
# Target by hostname
osapi client node hostname --target web-01

# Target by machine ID
osapi client node hostname --target a1b2c3d4e5f6

# Broadcast to all agents
osapi client node hostname --target _all

# Target by label
osapi client node hostname --target group:web.dev
```

The `node list` and `node get` commands include the machine ID in their
output so operators can identify agents across hostname changes.

## PKI Enrollment

When `pki.enabled` is true on both controller and agent, OSAPI uses
Ed25519 keypairs for cryptographic agent identity and job signing.

### Enrollment Flow

The enrollment process follows a Salt-style accept/reject model:

1. **Generate keypair** -- on first start, the agent generates an
   Ed25519 keypair and saves it to `agent.pki.key_dir` (default
   `/etc/osapi/pki`). Files: `agent.key` (mode 0600) and `agent.pub`
   (mode 0644).

2. **Request enrollment** -- the agent publishes an enrollment request
   to the controller via NATS containing its machine ID, hostname,
   public key, and SHA256 fingerprint.

3. **Pending state** -- the controller stores the request in a
   JetStream KV bucket. The agent enters pending state and waits.

4. **Admin accepts** -- an administrator reviews pending agents and
   accepts or rejects them via the CLI. On acceptance, the controller
   replies with its own public key.

5. **Ready** -- the agent saves the controller's public key as
   `controller.pub` and begins verifying job signatures. On subsequent
   restarts, the agent loads the controller key from disk and skips
   enrollment.

### Auto-Accept Mode

For development and testing, set `controller.pki.auto_accept: true`.
The controller automatically accepts all enrollment requests without
admin intervention. Do not use this in production.

### CLI Commands

```bash
# List pending enrollment requests
osapi client agent list --pending

# Accept a pending agent by hostname
osapi client agent accept --hostname web-01

# Accept by fingerprint (for verification)
osapi client agent accept --hostname web-01 \
  --fingerprint sha256:a1b2c3...

# Reject a pending agent
osapi client agent reject --hostname web-01

# Show the local agent key fingerprint
osapi client agent key fingerprint

# Show the local controller key fingerprint
osapi client controller key fingerprint
```

## Job Signing

When PKI is enabled, the controller signs every job payload with its
Ed25519 private key before storing it in the KV bucket. The payload is
wrapped in a `SignedEnvelope`:

```json
{
  "payload": "<raw job JSON>",
  "signature": "<Ed25519 signature bytes>",
  "fingerprint": "sha256:a1b2c3..."
}
```

Agents verify the signature using the controller's public key (received
during enrollment) before processing. If verification fails, the agent
rejects the job. When PKI is disabled, jobs are stored and processed
without signatures -- the envelope wrapping is skipped entirely.

## Key Rotation

The controller can rotate its Ed25519 keypair. During a configurable
grace period (default `24h`), agents accept signatures from both the
old and new controller keys.

The rotation flow:

1. Generate a new controller keypair (replace files in
   `controller.pki.key_dir`).
2. Restart the controller. It loads the new key and begins signing
   with it.
3. Agents that have already enrolled still hold the old controller
   public key. The `VerifyWithGrace` method checks the signature
   against both the current and previous controller keys.
4. During the grace period (`controller.pki.rotation_grace_period`),
   agents receive the new controller public key via an updated
   enrollment response and transition to the new key.
5. After the grace period, only the new key is accepted.

## Configuration

### Agent PKI

```yaml
agent:
  pki:
    # Enable PKI enrollment and job signature verification.
    enabled: false
    # Directory for agent keypair storage.
    key_dir: /etc/osapi/pki
```

| Field     | Type   | Default          | Description                            |
| --------- | ------ | ---------------- | -------------------------------------- |
| `enabled` | bool   | `false`          | Activate PKI enrollment and job verify |
| `key_dir` | string | `/etc/osapi/pki` | Directory for `agent.key`, `agent.pub` |

### Controller PKI

```yaml
controller:
  pki:
    # Enable PKI enrollment and job signing.
    enabled: false
    # Directory for controller keypair storage.
    key_dir: /etc/osapi/pki
    # Automatically accept all enrollment requests (dev only).
    auto_accept: false
    # Grace period for key rotation (Go duration).
    rotation_grace_period: 24h
```

| Field                    | Type   | Default          | Description                                  |
| ------------------------ | ------ | ---------------- | -------------------------------------------- |
| `enabled`                | bool   | `false`          | Activate PKI enrollment and job signing      |
| `key_dir`                | string | `/etc/osapi/pki` | Directory for controller keypair             |
| `auto_accept`            | bool   | `false`          | Auto-accept agent enrollments (dev/test)     |
| `rotation_grace_period`  | string | `24h`            | Both keys accepted during rotation           |

## What Is Not Changed

- **JWT authentication** -- HMAC-SHA256 JWT tokens for API
  authentication are unchanged. PKI is an additional trust layer
  between the controller and agents, not a replacement for JWT.
- **NATS transport** -- NATS connections still use their own auth
  (`none`, `user_pass`, or `nkey`). PKI operates at the job payload
  level, not the transport level.
