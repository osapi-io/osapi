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

## A secret is a type

An input field that holds a secret has the type `secret.Value` of
`internal/secret`, in the behavior, endpoint and command DSLs. A map of secrets
is `map[string]secret.Value`. No tag makes a secret. A tag with a typo makes an
ordinary field, and nothing reports it. A type cannot have a typo.

A `secret.Value` holds a reference, `secret://<vault>/<key>`, or a value. Its
`String`, `Format`, `MarshalJSON` and `LogValue` give the reference, or
`[redacted]` for a value. Only `Reveal` gives the value. The DSLs do these
things with the type:

- The spec makes it a string of format `password`, `writeOnly`, with
  `x-osapi-secret: true`. Its description tells what it takes.
- The audit redacts each field of the type by name, and keeps a reference.
- A span never records it, and a chain record has the reference or `[redacted]`.
- The CLI flag takes a reference, or `-` to read the value from standard input.
  It refuses a value.
- `validate` refuses one with `telemetry:"attr"`, one in the path or the query
  of a route, and a chain that gives one to a field of a different type.
- `TestSecretsAreTyped` fails a `string` input field whose name is on the audit
  denylist.

No vault exists yet, so the route of each method refuses a reference with 400.
The CLI sends the request, and the controller queues the job, with
`secret.Plain`, which gives each value.

A behavior input that is a secret is not kept out of NATS. The input of a
behavior is the data of its job, and JetStream keeps that data. Thus a token in
the environment of a container is plain text in JetStream. A password is the
exception: the controller hashes it before it queues the job, so the job has
only the hash.

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

## What comes next

Vaults, secret references and values sealed to the agent that uses them are in
progress, in [osapi #1065](https://github.com/osapi-io/osapi/issues/1065). This
page tells each part when it is built.
