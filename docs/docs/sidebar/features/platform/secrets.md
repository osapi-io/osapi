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

The flag of a secret does not take the value. The shell history keeps a flag,
and `ps` shows it to each user of the host. Give `-`, and send the value on
standard input:

```bash
osapi client node user password update --target web-01 \
    --name deploy --password - < password.txt
```

The flag also takes a reference, `secret://<vault>/<key>`. No vault exists yet,
so the controller refuses each reference with 400. The flag of a map of secrets,
such as `--env` of `osapi client node docker create`, takes only
`NAME=secret://<vault>/<key>`. Thus until a vault exists, give the environment
of a container in the request body of the API, not from the CLI.

The audit log records a reference as it is, so an entry shows which secret a
request used. It records `[redacted]` in place of a value.

The settings that are secrets, such as the JWT signing keys and the NATS
passwords, are in `osapi.yaml` or in the environment. `osapi config list` and
`osapi config get` print them as `********`. Refer to
[Configuration](../../usage/configuration.md).

The controller accepts tokens from more than one signing key. Each token has the
`id` of its key in its `kid` header. To change the key, put a new key first in
`controller.api.security.signing_keys`, then remove the old key when its tokens
expire. [Authentication](authentication.md#change-the-signing-key) has the
steps.

The controller can have vaults, which keep secrets encrypted in NATS.
[Configuration](../../usage/configuration.md#vaults) tells how to add a vault
and make its key with `osapi controller vault key create`. A signing key and the
bearer token can be a reference to a key in a vault, such as
`secret://default/jwt-signing-key`.
[References in the config](../../usage/configuration.md#references-in-the-config)
tells which settings take one.

## What comes next

The `secret` commands, `secret://` references in a behavior input and values
sealed to the agent that uses them are in progress, in
[osapi #1065](https://github.com/osapi-io/osapi/issues/1065). This page tells
how to use each part when it is built.
