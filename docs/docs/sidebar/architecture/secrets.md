---
sidebar_position: 6.5
---

# Secrets

A secret is a value that only the hosts that use it can read, such as an API
token or a private key. This page tells how osapi keeps a secret now.

## What osapi has now

Each kind of secret has its own solution:

- The JWT signing keys and the NATS passwords are values in `osapi.yaml` or in
  an environment variable. `osapi config list` and `osapi config get` print them
  as `********`.
- The PKI keys are files in `key_dir`.
- A behavior input that is a secret, such as a password or the environment of a
  container, stays out of the audit log, the spans and the chain records.

A behavior input that is a secret is not kept out of NATS. The input of a
behavior is the data of its job, and JetStream keeps that data. Thus a token in
the environment of a container is plain text to each NATS user that can read the
stream. A password is the exception: the controller hashes it before it queues
the job, so the job has only the hash.

## The store on the disk

The embedded NATS server encrypts each stream and bucket in its `store_dir` when
`nats.server.jetstream.key_file` names a key. `osapi nats key create` makes the
key on the NATS host. Thus a person who reads the disk or a backup of the store
does not read the jobs, the audit log or the agent registry. The key does not
protect a value from a NATS user, because the server decrypts each message for a
user that can read the stream.
[Encryption at rest](../usage/configuration.md#encryption-at-rest) has the
setting, and what happens when the key is lost.

## The JWT signing keys

`controller.api.security.signing_keys` is a list of keys, each with an `id` and
a `key`. The first key signs each new token, and puts its `id` in the `kid`
header of the token. Each key verifies the tokens with its `id`. The controller
refuses a token with no `kid`, or with a `kid` that the list does not have.

Thus a signing key changes with no outage. Put a new key first, and restart the
controller. The tokens of the old key still work. Remove the old key when no
client sends a token that it signed. A token expires after three months.
[Authentication](../features/platform/authentication.md#change-the-signing-key)
has the steps.

## Vaults

A vault is a store of secrets on the controller, with a name and a type.
`controller.secrets.vaults` names each vault. The controller opens each one when
it starts, and does not start when it cannot open one, such as when its key is
not there.

`internal/secret/vault` has the `Vault` interface (`Name`, `Get`, `Put`,
`Delete` and `List`) and a registry of types. A type registers its name, a
constructor and the Go struct of its `config` block. `osapi config validate`
checks each block against the struct of its type, and the struct gives the
settings in the config table. Thus a different store is a type that plugs in.

osapi has one type, `nats`, in `internal/secret/vault/natsvault`:

- It encrypts each value with AES-256-GCM under the vault key, with a new nonce
  each time. The vault and the key of the value are the additional data, so a
  value that somebody copies to a different key does not decrypt.
- It keeps each value in the KV bucket of `nats.secrets`, `osapi-secrets` with
  the namespace `osapi`, under `<vault>.<key>`. The bucket keeps the earlier
  versions of a key, as `nats.secrets.history` sets.
- The record of a value has its version, when it changed and who changed it, as
  plain text. A NATS user that reads the bucket sees these, and not the value.
- The vault key comes from `key_file` or `key_env`. The vault reads it at each
  call, so a rotation applies with no restart.

`osapi controller vault key create` writes a key, and
`osapi controller vault key rotate` adds a new key, encrypts each value again
and removes the old key. A rotation does not encrypt the earlier versions in the
history again, so they do not decrypt after it.
[Configuration](../usage/configuration.md#vaults) has the settings.

## What comes next

Vaults, secret references and values sealed to the agent that uses them are in
progress, in [osapi #1065](https://github.com/osapi-io/osapi/issues/1065). This
page tells each part when it is built.
