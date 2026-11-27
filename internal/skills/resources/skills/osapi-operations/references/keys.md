# Keys

`osapi pki key` reads, makes and rotates the keys of osapi on this host. Run it
on the host; it reads the config there.

```sh
osapi pki key list --json                         # each key: kind, algorithm, state, fingerprint or ID
osapi pki key get --name agent --json             # the agent's fingerprint, and the command that admits it
osapi pki key create --name jetstream --json      # make a key that is not there
osapi pki key rotate --name vault:default --json  # a new vault key; each value is encrypted again
osapi pki list --json                             # each key and each certificate, in one list
osapi client pki key get --json                   # the controller's key, and its rotation: the agents that the switch waits for
osapi client pki key rotate --json                # rotate the controller's key, from any host, with no new admission
```

The keys are `controller` (Ed25519, signs jobs), `agent` (Ed25519) and
`agent-encryption` (ECDH P-256) of the agent, `jetstream` (the key of the NATS
store, ChaCha20-Poly1305), and `vault:<name>` (AES-256-GCM) for each vault of
type nats. A public key shows its fingerprint; a secret key shows its ID, the
first 16 hex digits of its SHA-256, never the key. The agent's fingerprint
covers both of its keys, as its admission does.

Each key is in the list, with a `state`: `ok`; `not made`, when its component
has not made it on this host; `off`, for a store with no `key_file`; or
`unreadable`, with the reason.

`create` refuses a key that is there. The JetStream key does not rotate: a new
one makes the store unreadable. The controller and the agent make their keys at
their first start.

`rotate --name agent` sends the agent's new keys to the controller, signed with
the old key, with no new admission. `rotate --name controller` on the
controller's host, or `osapi client pki key rotate` from any host, starts the
rotation of the controller's key. The agents take the next key at their next
heartbeat. The controller switches to it when each admitted agent that reported
within `controller.pki.rotation_grace_period` has it, or when that period ends.
`osapi client pki key get --json` shows `next` and, for each agent, `has_next`.
An agent that is away for the whole period needs a new admission.
