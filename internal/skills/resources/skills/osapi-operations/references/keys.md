# Keys

`osapi pki key` reads, makes and rotates the keys of osapi on this host. Run it
on the host; it reads the config there.

```sh
osapi pki key list --json                         # each key: kind, algorithm, state, fingerprint or ID
osapi pki key get --name agent --json             # the agent's fingerprint, and the command that admits it
osapi pki key create --name jetstream --json      # make a key that is not there
osapi pki key rotate --name vault:default --json  # a new vault key; each value is encrypted again
osapi pki list --json                             # each key and each certificate, in one list
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

`create` refuses a key that is there. Only a vault key rotates: a new JetStream
key makes the store unreadable, a new controller key needs each agent admitted
again, and a new agent key needs this machine admitted again. The controller and
the agent make their keys at their first start.
