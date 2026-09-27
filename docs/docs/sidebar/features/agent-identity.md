---
sidebar_position: 26
sidebar_label: Agent Identity & PKI
---

# Agent Identity & PKI

OSAPI agents identify themselves using a persistent machine ID and support
optional PKI enrollment for cryptographic trust between agents and the
controller.

## Machine Identity

Every agent has two identity values:

- **Machine ID** -- permanent identifier from `/etc/machine-id` (Linux) or
  `IOPlatformUUID` (macOS). Used as the registry key and stable reference across
  hostname changes.
- **Hostname** -- mutable display name from the OS or `agent.hostname` config.
  Used for human-friendly targeting and display.

The agent resolves its machine ID once at startup and refuses to start if it
cannot be read. NATS subject routing uses the machine ID
(`jobs.query.host.<machineID>`), so consumers never need to resubscribe when the
hostname changes. The hostname is re-read on each heartbeat tick (10s) and
updated in the registry for display.

### CLI Targeting

Both hostname and machine ID work as targets. The controller resolves hostnames
to machine IDs for NATS subject routing automatically:

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

The `node list` and `node get` commands include the machine ID in their output
so operators can identify agents across hostname changes. The management
dashboard shows machine ID and fingerprint on agent cards, and pending agents
are managed via the Admin > Enrollment page.

## PKI Enrollment

When `pki.enabled` is true on both controller and agent, OSAPI uses Ed25519
keypairs for cryptographic agent identity and job signing.

### Enrollment Flow

The enrollment process follows a Salt-style accept/reject model:

1. **Generate keypair** -- on first start, the agent generates an Ed25519
   keypair and saves it to `agent.pki.key_dir` (default `/etc/osapi/pki`).
   Files: `agent.key` (mode 0600) and `agent.pub` (mode 0644).

2. **Request enrollment** -- the agent publishes an enrollment request to the
   controller via NATS containing its machine ID, hostname, public key, and
   SHA256 fingerprint.

3. **Pending state** -- the controller stores the request in a JetStream KV
   bucket. The agent enters pending state and waits.

4. **Admin accepts** -- an administrator reviews pending agents and accepts or
   rejects them via the CLI. On acceptance, the controller replies with its own
   public key.

5. **Ready** -- the agent saves the controller's public key as `controller.pub`
   and begins verifying job signatures. On subsequent restarts, the agent loads
   the controller key from disk and skips enrollment.

### Auto-Accept Mode

For development and testing, set `controller.pki.auto_accept: true`. The
controller automatically accepts all enrollment requests without admin
intervention. Do not use this in production.

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

When `controller.pki.enabled` is true, the controller signs every job payload
with its Ed25519 private key before storing it in the KV bucket. The payload is
wrapped in a `SignedEnvelope`:

```json
{
  "payload": "<raw job JSON>",
  "signature": "<Ed25519 signature bytes>",
  "fingerprint": "sha256:a1b2c3..."
}
```

Agents also sign their job responses with their own keypair before writing them
to the KV bucket, using the same envelope shape.

When PKI is disabled on either side, that side neither signs nor verifies --
jobs and responses are stored and processed as plain JSON, with no envelope
wrapping.

### Verification and failure behavior

When `agent.pki.enabled` is true, an agent verifies the envelope on every job it
receives before running the operation. Verification is fail-closed: the job is
rejected and never executed unless it fully passes. Three rejection reasons are
logged and recorded as the job's termination reason, so an operator can tell
them apart:

- **Not enrolled** -- the agent has no cached controller public key yet
  (enrollment is still pending, or was never completed). This is reported
  distinctly from a bad signature, since it means "not yet trusted" rather than
  "this job was tampered with."
- **Missing or malformed signature** -- the job data is not wrapped in a signed
  envelope, or the envelope is missing a required field.
- **Invalid signature** -- the envelope's signature does not verify against the
  agent's cached controller key (or the previous key, during a rotation grace
  period; see [Key Rotation](#key-rotation)).

The controller likewise verifies the signature on an agent's job response before
treating it as a result. A response that fails verification is treated as a
failed or missing result, never as a successful one.

When PKI is disabled, none of the above applies -- there is no envelope to
verify, so nothing is rejected on signature grounds.

## Registration Signing

An agent's job responses are not the only thing it sends. Every heartbeat it
rewrites its own entry in the agent registry, and that entry is what targeting
reads: the hostname a job names, the labels a label target selects, and the
scheduling state that decides whether the agent is eligible for work at all.

When `agent.pki.enabled` is true, the agent signs the routing fields of its
registration with its own keypair. The signature covers:

| Field         | Why it is signed                                               |
| ------------- | -------------------------------------------------------------- |
| `machine_id`  | Selects which stored key the signature is checked against      |
| `hostname`    | A job names it, so a forged one redirects that host's work     |
| `fingerprint` | Reported as this agent's key in the fleet view                 |
| `state`       | A `Pending` agent is refused work; a forged `Ready` lifts that |
| `labels`      | A label target routes on them, so a forged label attracts jobs |

Everything else in a registration -- uptime, load, memory, conditions -- is
reporting rather than routing, and is not signed. The signature is carried
beside these fields, never inside the signed bytes.

### What the controller checks

When the controller holds a key for the agent, it verifies each registration
before letting it decide anything:

1. The machine ID selects the stored record. A registration naming another agent
   is checked against **that** agent's key, and fails unless it was signed by
   it.
2. The signature must verify against the key recorded when the agent's
   enrollment was accepted.
3. The hostname must match the one recorded at acceptance. An accepted agent is
   not entitled to answer for a host it did not enrol as.

A registration that fails any of these is not an error returned to the agent.
The agent keeps heartbeating and stays listed; it simply stops being
authoritative. It is invisible to target resolution, to label matching, and to
machine-ID targeting, and work aimed at that hostname continues to reach the
agent that enrolled under it.

Where more than one agent claims a hostname, resolution is deterministic: the
lowest machine ID wins, rather than whichever entry the registry happened to
list first.

### Seeing where a fleet stands

`osapi client agent list` and `GET /agent` report two separate fields per agent:

| Field        | Meaning                                                                |
| ------------ | ---------------------------------------------------------------------- |
| `key_stored` | The controller holds this agent's key, recorded at acceptance          |
| `verified`   | What it last registered was signed by that key and claims its own host |

They answer different questions. `key_stored: false` means the agent has not
enrolled since the key store existed -- expected during a rollout, and fixed by
re-enrolling it. `key_stored: true` with `verified: false` means the controller
can check this agent and what arrived did not check out, which is worth
investigating rather than waiting out.

## Key Rotation

The controller can rotate its Ed25519 keypair. During a configurable grace
period (default `24h`), agents accept signatures from both the old and new
controller keys.

The rotation flow:

1. Generate a new controller keypair (replace files in
   `controller.pki.key_dir`).
2. Restart the controller. It loads the new key and begins signing with it.
3. Agents that have already enrolled still hold the old controller public key.
   The `VerifyWithGrace` method checks the signature against both the current
   and previous controller keys.
4. During the grace period (`controller.pki.rotation_grace_period`), agents
   receive the new controller public key via an updated enrollment response and
   transition to the new key.
5. After the grace period, only the new key is accepted.

## Rollout

Because agent-side verification is fail-closed, enabling `agent.pki.enabled` on
a fleet is order-sensitive:

1. Enable `controller.pki.enabled` first, and confirm agents complete enrollment
   and show as accepted (`osapi client agent list`).
2. Only then enable `agent.pki.enabled`. An agent that has not completed
   enrollment holds no controller public key, so if PKI is enabled on it before
   enrollment finishes, every job it receives is rejected as "not enrolled"
   until enrollment completes.

Neither switch flips as a consequence of upgrading. An agent accepted before the
key store existed has no stored key, so it reports `key_stored: false` and is
not authoritative once the controller is verifying. Nothing starts refusing work
at a moment nobody chose, and the order to follow is:

1. Enable `controller.pki.enabled`. The controller begins recording each agent's
   key as its enrollment is accepted.
2. Read `osapi client agent list`. Every agent still showing `key_stored: false`
   has not enrolled since the store existed.
3. Re-enrol those agents, and watch the field flip as each is accepted.
4. Once the fleet reports `key_stored: true` and `verified: true` throughout,
   enable `agent.pki.enabled`.

Doing step 4 first is what produces a fleet that refuses work: the agents are
verifying jobs before the controller can verify them back.

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

| Field                   | Type   | Default          | Description                              |
| ----------------------- | ------ | ---------------- | ---------------------------------------- |
| `enabled`               | bool   | `false`          | Activate PKI enrollment and job signing  |
| `key_dir`               | string | `/etc/osapi/pki` | Directory for controller keypair         |
| `auto_accept`           | bool   | `false`          | Auto-accept agent enrollments (dev/test) |
| `rotation_grace_period` | string | `24h`            | Both keys accepted during rotation       |

## What Is Not Changed

- **JWT authentication** -- HMAC-SHA256 JWT tokens for API authentication are
  unchanged. PKI is an additional trust layer between the controller and agents,
  not a replacement for JWT.
- **NATS transport** -- NATS connections still use their own auth (`none`,
  `user_pass`, or `nkey`). PKI operates at the job payload level, not the
  transport level.
