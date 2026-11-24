# Secrets

A secret is a value such as an API token, a password or a private key. This page
tells what osapi does with a secret now.

## What osapi does now

A behavior input that holds a secret, such as the `--env` of
`osapi client node docker create`, stays out of the audit log and the traces.
When `controller.pki.enabled` is true, it also stays out of NATS: the controller
seals the value to each agent that the target names, and only those agents can
open it. [Sealed values](#sealed-values) has how. When PKI is off, no agent has
an encryption key. Then the job data in NATS has the value as plain text, so do
not give a long-lived token to a behavior input.

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

A reference keeps the value out of the request, the shell history and the audit
log. The controller then seals the value in the job, as it seals a value that
you give.

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

The controller has vaults, which keep secrets encrypted in NATS. With no config,
it has the vault `default`, and it makes the key of that vault at its first
start. [Configuration](../../usage/configuration.md#vaults) tells how to add a
vault. A signing key and the bearer token can be a reference to a key in a
vault, such as `secret://default/jwt-signing-key`.
[References in the config](../../usage/configuration.md#references-in-the-config)
tells which settings take one.

## Write and read a secret

`osapi client secret` writes the value of a key, and reads the metadata of each
key: its version, when it changed and who changed it. No command and no API
answer gives a value. Give the value on standard input:

```bash
osapi client secret update --vault default --key registry-token --value - < token.txt
osapi client secret list --vault default
osapi client secret vault list
```

Each write is a new version of the key. `osapi client secret delete` deletes a
key. The Secrets page of the UI, in the Admin menu, lists the keys of a vault,
writes a value with a password field and deletes a key.

The permissions:

| Permission     | Built-in roles                            | Lets a token                           |
| -------------- | ----------------------------------------- | -------------------------------------- |
| `secret:read`  | `admin`, `operator`, `automation`, `read` | list the vaults and the keys of each   |
| `secret:write` | `admin`, `automation`                     | write and delete a key                 |
| `secret:use`   | `admin`, `automation`                     | give a reference to a secret in a call |

A reference as the value of `osapi client secret update` resolves too, so it
copies the value of that key. It needs `secret:use`, as each reference does.

On the Plans page of the UI, a secret field of a step has a key button that
lists the reference of each key. A plan holds only references: Export YAML
writes `<secret>` for a value that you typed.

## Sealed values

When `controller.pki.enabled` is true, the controller seals each secret of a job
before it stores the job. `osapi client job get` then shows each secret as
`sealed:` and ciphertext, and the job has a `sealed_keys` map with one entry for
each agent that can open it. The job data in JetStream and in the job KV never
has the value.

The controller seals the job to each agent that the target names when it queues
the job, and that has an encryption key:

- An agent that joins after that has no key in the job. It fails the job with
  the reason `the job was not sealed to this agent`, and does not run it. Run
  the job again.
- An agent that was admitted before agents had an encryption key has no key in
  the job either. Enroll it again, as
  [Agent identity](agent-identity.md#an-agent-admitted-before-the-encryption-key)
  tells.

A password is not sealed: the controller hashes it, and the job has only the
hash.

## What comes next

More parts of the secrets are in progress, in
[osapi #1065](https://github.com/osapi-io/osapi/issues/1065). This page tells
how to use each part when it is built.
