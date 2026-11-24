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

The flag also takes a reference, `secret://<vault>/<key>`. The flag of a map of
secrets, such as `--env` of `osapi client node docker create`, takes only
`NAME=secret://<vault>/<key>`:

```bash
osapi client node docker create --target web-01 --name app --image app:1 \
    --env API_TOKEN=secret://default/api-token
```

The controller reads the value of each reference from the vault that it names,
before it queues the job. It refuses the request when one of these is false:

| Condition                                               | Refusal |
| ------------------------------------------------------- | ------- |
| The vault exists in `controller.secrets.vaults`.        | 400     |
| The token has the permission `secret:use`.              | 403     |
| The target is in the `targets` of the vault, if it has. | 403     |
| The vault has the key.                                  | 400     |
| The value obeys the rules of its field.                 | 400     |

The target scope of the token applies too, as to each request. A vault with
`targets` gives its secrets to one host or to a label, such as `group:prod`. A
target such as `_all` is not in them, so give the host or the label.

The value is in the job data in NATS, as a value that you give is. Thus a
reference keeps the value out of the request, the shell history and the audit
log, but not out of NATS.

The audit log records a reference as it is, so an entry shows which secret a
request used, on which target. It records `[redacted]` in place of a value.

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

The `secret` commands and values sealed to the agent that uses them are in
progress, in [osapi #1065](https://github.com/osapi-io/osapi/issues/1065). This
page tells how to use each part when it is built.
