---
sidebar_position: 6.5
---

# Secrets

A secret is a value that only the hosts that use it can read, such as an API
token or a private key. This page tells how osapi keeps a secret now.

## What osapi has now

Each kind of secret has its own solution:

- The JWT signing key and the NATS passwords are values in `osapi.yaml` or in an
  environment variable. The controller masks them before it logs its config.
- The PKI keys are files in `key_dir`.
- A behavior input that is a secret, such as a password or the environment of a
  container, stays out of the audit log, the spans and the chain records.

A behavior input that is a secret is not kept out of NATS. The input of a
behavior is the data of its job, and JetStream keeps that data. Thus a token in
the environment of a container is plain text in JetStream. A password is the
exception: the controller hashes it before it queues the job, so the job has
only the hash.

## What comes next

Vaults, secret references and values sealed to the agent that uses them are in
progress, in [osapi #1065](https://github.com/osapi-io/osapi/issues/1065). This
page tells each part when it is built.
