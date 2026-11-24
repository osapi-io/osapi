# Secrets

A secret is a value such as an API token, a password or a private key. This page
tells what osapi does with a secret now.

## What osapi does now

A behavior input that holds a secret, such as the `--env` of
`osapi client node docker create`, stays out of the audit log and the traces.
The value is still in the job data in NATS. Thus do not give a long-lived token
to a behavior input.

A password is the exception. `osapi client node user create` and
`osapi client node user password update` send the password to the controller,
which hashes it before it queues the job. The job and NATS have only the hash.

The settings that are secrets, such as the JWT signing key and the NATS
passwords, are in `osapi.yaml` or in the environment. The controller masks them
when it logs its config. Refer to [Configuration](../../usage/configuration.md).

## What comes next

Vaults, `secret://` references and values sealed to the agent that uses them are
in progress, in [osapi #1065](https://github.com/osapi-io/osapi/issues/1065).
This page tells how to use each part when it is built.
