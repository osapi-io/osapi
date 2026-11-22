---
sidebar_position: 26
sidebar_label: Agent Identity & PKI
---

# Agent identity and PKI

An osapi agent identifies itself with a permanent machine ID. As an option, PKI
enrollment gives cryptographic trust between the agents and the controller.

## Machine identity

Each agent has two identity values:

- **Machine ID.** A permanent identifier from `/etc/machine-id` (Linux) or
  `IOPlatformUUID` (macOS). It is the registry key. It does not change when the
  hostname changes.
- **Hostname.** A display name from the OS or from the `agent.hostname` config.
  It can change. Use it to target hosts and to read the output.

The agent reads its machine ID one time, when it starts. If it cannot read the
machine ID, it does not start. NATS subjects use the machine ID
(`jobs.query.host.<machineID>`). Thus the consumers do not subscribe again when
the hostname changes. The agent reads the hostname again at each heartbeat (10s)
and updates it in the registry.

### CLI targeting

A hostname and a machine ID both work as targets. The controller changes a
hostname to the machine ID for the NATS subject automatically:

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

The output of `node list` and `node get` has the machine ID. Thus an operator
can identify an agent when its hostname changes. The management dashboard shows
the machine ID and the fingerprint on the agent cards. Use the Admin >
Enrollment page to manage pending agents.

## PKI enrollment

When `pki.enabled` is true on the controller and on the agent, osapi uses
Ed25519 key pairs. They give each agent a cryptographic identity, and they sign
the jobs.

### Enrollment flow

Enrollment uses an accept and reject model, like Salt:

1. **The agent makes a key pair.** When the agent starts for the first time, it
   makes an Ed25519 key pair. It saves the key pair in `agent.pki.key_dir`
   (`/etc/osapi/pki` by default). The files are `agent.key` (mode 0600) and
   `agent.pub` (mode 0644).

   The user that runs the agent must own the directory. The agent does not
   replace a key that it cannot read, or half of a key pair. It writes the error
   to the log and stops. A new key is not the key that the controller accepted.
   Thus a host with a replaced key gets no work until you enroll it again. If
   you change the user of the agent, also change the owner of the directory
   (`chown -R osapi /etc/osapi/pki`).

1. **The agent sends an enrollment request.** It publishes a request to the
   controller through NATS. The request has its machine ID, hostname, public key
   and SHA256 fingerprint.

1. **The agent waits.** The controller stores the request in a JetStream KV
   bucket. The agent goes into the pending state and waits. **It publishes its
   request again at each heartbeat** until the controller accepts it.

   The first request goes over core NATS, which does not guarantee delivery. If
   the controller was down or a publish failed, the controller did not know the
   agent, and the agent stayed pending. A pending agent does not take jobs. Thus
   that state stayed until somebody restarted the process. The repeated request
   also fixes a controller that lost its pending entries. It is safe, because
   the controller keys acceptance by machine ID. The controller writes over the
   pending entry. It does not add one entry for each attempt.

1. **An administrator accepts the agent.** An administrator examines the pending
   agents, and accepts or rejects them with the CLI. When the controller accepts
   an agent, it sends back its own public key.

1. **The agent is ready.** The agent saves the public key of the controller as
   `controller.pub`, and starts to verify job signatures. At the next restarts,
   the agent loads the controller key from disk and does not enroll again.

### Auto-admit mode

For development and tests, set `controller.pki.auto_admit: true`. Then the
controller admits each machine that asks to join, without an administrator. Do
not use this in production.

### CLI commands

```bash
# On the machine: show its machine ID and key fingerprint
osapi agent admission get

# List the machines that wait for a decision
osapi client admission list --state pending

# Admit a machine, with the fingerprint that the machine printed
osapi client admission admit --machine-id <machine-id> \
  --fingerprint SHA256:a1b2c3...

# Reject a machine, with the reason
osapi client admission reject --machine-id <machine-id> --reason "unknown host"

# Revoke an admitted machine that is not trusted: its key stops verifying at
# once, and a record shows it if it asks again
osapi client admission revoke --machine-id <machine-id>

# Delete a machine that goes away on purpose, such as before a rebuild: its key
# and each record of it go, and it is new if it asks again
osapi client admission delete --machine-id <machine-id>

# Show the local controller key fingerprint
osapi controller key fingerprint
```

## Job signing

When `controller.pki.enabled` is true, the controller signs each job payload
with its Ed25519 private key before it stores the job in the KV bucket. The
payload is in a `SignedEnvelope`:

```json
{
  "payload": "<raw job JSON>",
  "signature": "<Ed25519 signature bytes>",
  "fingerprint": "sha256:a1b2c3..."
}
```

Agents also sign their job responses with their own key pair before they write
them to the KV bucket. They use the same envelope.

When PKI is disabled on one side, that side does not sign and does not verify.
osapi stores and processes the jobs and the responses as plain JSON, without an
envelope.

### Verification and failures

When `agent.pki.enabled` is true, the agent verifies the envelope of each job
before it runs the operation. Verification fails closed: the agent rejects the
job and does not run it unless the job passes all checks. The agent writes one
of three reasons to the log and records it as the termination reason of the job.
Thus an operator can tell them apart:

- **Not enrolled.** The agent has no controller public key yet, because
  enrollment is pending or did not complete. osapi reports this separately from
  a bad signature, because it means "not trusted yet", not "somebody changed
  this job".
- **Missing or malformed signature.** The job data is not in a signed envelope,
  or the envelope has no value in a required field.
- **Invalid signature.** The signature of the envelope does not verify with the
  controller key of the agent, or with the previous key during a rotation grace
  period. Refer to [Key rotation](#key-rotation).

The controller also verifies the signature on the job response of an agent
before it uses it as a result. A response that fails verification is a failed or
missing result, never a successful result.

When PKI is disabled, none of this applies. There is no envelope to verify, so
osapi rejects nothing because of a signature.

## Registration signing

An agent sends more than job responses. At each heartbeat, it writes its own
entry in the agent registry again. Targeting reads that entry. It reads the
hostname that a job names and the labels that a label target selects. It also
reads the scheduling state, which decides if the agent can get work.

When `agent.pki.enabled` is true, the agent signs the routing fields of its
registration with its own key pair. The signature includes:

| Field         | Why the agent signs it                                              |
| ------------- | ------------------------------------------------------------------- |
| `machine_id`  | It selects the stored key that checks the signature                 |
| `hostname`    | A job names it. A false one sends the work of that host elsewhere   |
| `fingerprint` | The fleet view shows it as the key of this agent                    |
| `state`       | osapi refuses work to a `Pending` agent. A false `Ready` stops that |
| `labels`      | A label target uses them. A false label gets jobs                   |

The other fields of a registration, for example uptime, load, memory and
conditions, only report data. They do not route work, and the agent does not
sign them. The signature is next to these fields, not in the signed bytes.

### What the controller checks

When the controller has a key for the agent, it verifies each registration
before it uses the registration for a decision:

1. The machine ID selects the stored record. The controller checks a
   registration that names a different agent with the key of **that** agent. It
   fails unless that agent signed it.
1. The signature must verify with the key that the controller recorded when it
   accepted the enrollment of the agent.
1. The hostname must be the hostname that the controller recorded at acceptance.
   An accepted agent cannot answer for a host that it did not enroll as.

When a registration fails a check, the controller does not return an error to
the agent. The agent continues its heartbeats and stays in the list, but its
registration does not decide anything. Target resolution, label matching and
machine-ID targets do not see it. Work for that hostname continues to go to the
agent that enrolled with it.

When more than one agent claims a hostname, the result is always the same: the
lowest machine ID wins. The order of the registry list does not decide.

### Check the state of a fleet

`osapi client agent list` and `GET /agent` report two separate fields for each
agent:

| Field        | Meaning                                                           |
| ------------ | ----------------------------------------------------------------- |
| `key_stored` | The controller has the key of this agent, from acceptance         |
| `verified`   | That key signed the last registration, and it claims its own host |

They answer different questions. `key_stored: false` means that the agent did
not enroll after the key store started. This is normal during a rollout. To fix
it, enroll the agent again. `key_stored: true` with `verified: false` means that
the controller can check this agent, and the data that arrived did not pass the
check. Examine it. Do not wait.

## Key rotation

The controller can change its Ed25519 key pair. During a grace period (`24h` by
default), agents accept signatures from the old key and from the new key of the
controller. You can configure the grace period.

To rotate the key:

1. Make a new controller key pair. Replace the files in
   `controller.pki.key_dir`.
1. Restart the controller. It loads the new key and starts to sign with it.
1. The enrolled agents still have the old public key of the controller. The
   `VerifyWithGrace` method checks the signature with the current key and with
   the previous key of the controller.
1. During the grace period (`controller.pki.rotation_grace_period`), the agents
   get the new public key of the controller in an updated enrollment response.
   Then they change to the new key.
1. After the grace period, osapi accepts only the new key.

## Rollout

Verification on the agent fails closed. Thus the order is important when you
enable `agent.pki.enabled` on a fleet:

1. Enable `controller.pki.enabled` first. Make sure that the agents complete
   enrollment and show as accepted (`osapi client agent list`).
1. Then enable `agent.pki.enabled`. An agent that did not complete enrollment
   has no public key of the controller. If you enable PKI on it before
   enrollment completes, it rejects each job as "not enrolled" until enrollment
   completes.

An upgrade does not change these settings. An agent that the controller accepted
before the key store started has no stored key. Thus it reports
`key_stored: false`, and its registration does not decide anything when the
controller verifies. Nothing starts to refuse work at a time that nobody
selected. Do these steps in this order:

1. Enable `controller.pki.enabled`. The controller starts to record the key of
   each agent when it accepts the enrollment.
1. Read `osapi client agent list`. Each agent that still shows
   `key_stored: false` did not enroll after the store started.
1. Enroll those agents again. Watch the field change as the controller accepts
   each one.
1. When all of the fleet reports `key_stored: true` and `verified: true`, enable
   `agent.pki.enabled`.

If you do step 4 first, the fleet refuses work: the agents verify the jobs
before the controller can verify the agents.

## Configuration

### Agent PKI

<!-- osapi:config agent.pki -->

| Key                 | Type   | Default          | Environment               | Description                                           |
| ------------------- | ------ | ---------------- | ------------------------- | ----------------------------------------------------- |
| `agent.pki.enabled` | bool   | `false`          | `OSAPI_AGENT_PKI_ENABLED` | Enable PKI enrollment and job signature verification. |
| `agent.pki.key_dir` | string | `/etc/osapi/pki` | `OSAPI_AGENT_PKI_KEY_DIR` | Directory for agent keypair storage.                  |

<!-- /osapi:config -->

### Controller PKI

<!-- osapi:config controller.pki -->

| Key                                    | Type   | Default          | Environment                                  | Description                                                                                                       |
| -------------------------------------- | ------ | ---------------- | -------------------------------------------- | ----------------------------------------------------------------------------------------------------------------- |
| `controller.pki.enabled`               | bool   | `false`          | `OSAPI_CONTROLLER_PKI_ENABLED`               | Enable PKI enrollment and job signing.                                                                            |
| `controller.pki.key_dir`               | string | `/etc/osapi/pki` | `OSAPI_CONTROLLER_PKI_KEY_DIR`               | Directory for controller keypair storage.                                                                         |
| `controller.pki.auto_admit`            | bool   | `false`          | `OSAPI_CONTROLLER_PKI_AUTO_ADMIT`            | Admit each machine that asks to join, with no operator (dev/test only).                                           |
| `controller.pki.rotation_grace_period` | string | `24h`            | `OSAPI_CONTROLLER_PKI_ROTATION_GRACE_PERIOD` | Grace period during key rotation where both old and new keys are accepted. Go duration format.                    |
| `controller.pki.delete_after`          | string | none             | `OSAPI_CONTROLLER_PKI_DELETE_AFTER`          | Delete an admitted machine that has not reported for this long, such as 720h. Empty is never. Go duration format. |

<!-- /osapi:config -->

## What does not change

- **JWT authentication.** The HMAC-SHA256 JWT tokens for API authentication do
  not change. PKI is an additional trust layer between the controller and the
  agents. It does not replace JWT.
- **NATS transport.** NATS connections still use their own authentication
  (`none`, `user_pass`, `nkey` or `creds`). PKI operates on the job payloads,
  not on the transport.
