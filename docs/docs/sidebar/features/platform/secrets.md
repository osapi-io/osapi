# Secrets

A secret is a value such as an API token, a password or a private key. With
secrets, you keep the value in a vault, and a request names it with a reference.
The value goes only to the hosts that use it, sealed to the key of each agent.

:::caution Planned

Secrets are not built. This page tells how you will use them, from the plan of
[osapi #1065](https://github.com/osapi-io/osapi/issues/1065). The commands, the
settings and the permissions on this page do not exist in a release yet.
[Secrets](../../architecture/secrets.md) in the architecture section has the
issue of each part.

:::

## What you can do now

A behavior input that holds a secret, such as the `--env` of
`osapi client node docker create`, stays out of the audit log and the traces.
The value is still in the job data in NATS. Thus do not give a long-lived token
to a behavior input until sealed values are built
([osapi #1247](https://github.com/osapi-io/osapi/issues/1247)).

## How it will work

1. You configure a vault on the controller.
1. You set a secret in the vault, from standard input.
1. A request uses a reference to the secret, such as
   `secret://default/registry-token`, in place of the value.
1. The controller gets the value from the vault, and checks that the caller may
   use it.
1. The controller seals the value to the key of each target agent, and queues
   the job.
1. The agent opens the value, gives it to the behavior, and drops it.

The API never returns the value of a secret. A list shows names, and a get shows
the metadata of a key.

## Set a vault

A vault is in `controller.secrets.vaults` in `osapi.yaml`. osapi has one vault
type, `nats`. It keeps each value encrypted in a JetStream KV bucket:

```yaml
controller:
  secrets:
    vaults:
      - name: default
        type: nats
        key_file: /etc/osapi/secrets.key
```

Make the vault key on the controller host:

```bash
osapi controller vault key create --name default
```

To give the secrets of a vault to some hosts only, add `targets`, such as
`targets: [group:prod]`. The controller does not send a secret of that vault to
an agent outside them. A vault with no targets goes to each host.

A different store, such as HashiCorp Vault, will be a vault type that an
extension adds ([osapi #1250](https://github.com/osapi-io/osapi/issues/1250)).

## Set a secret

Give the value on standard input. Do not put it in an argument, because the
shell history keeps arguments:

```bash
osapi client secret update --vault default --key registry-token < token.txt
```

Show the vaults, the keys of a vault, and the metadata of a key:

```bash
osapi client secret vault list
osapi client secret list --vault default
osapi client secret get --vault default --key registry-token
```

`get` shows the version of the key, when it changed and who changed it. It does
not show the value. `osapi client secret delete` removes a key.

## Use a secret in a behavior input

Put the reference in the input that takes the secret:

```bash
osapi client node docker create --target group:web --name app \
    --image registry.example.com/app:1.4 \
    --env REGISTRY_TOKEN=secret://default/registry-token
```

The controller resolves the reference before it queues the job. A reference to a
vault or a key that does not exist gives `400`. The audit log, the traces and
the chain records show the reference, not the value.

A secret flag of the CLI takes a reference, or `-` to read the value from
standard input. It does not take a value in the argument.

## Change a secret

Run `osapi client secret update` again. A job that the controller queues after
the change gets the new value. A host that has the old value, in a file or a
container, keeps it until a job runs on it again. The audit log shows which
hosts used the secret.

To change the vault key, run this command on the controller host. It adds a new
key, encrypts each value again, and then removes the old key:

```bash
osapi controller vault key rotate --name default
```

## Permissions

| Permission     | What it permits                                    |
| -------------- | -------------------------------------------------- |
| `secret:read`  | List the vaults and the keys, and get the metadata |
| `secret:write` | Set and delete a secret                            |
| `secret:use`   | Use a reference in a request                       |

The roles get `secret:read` and `secret:write` by the rule for an endpoint
group, as [Authentication](authentication.md#roles-and-permissions) tells. Only
`admin` and `automation` get `secret:use`. To use a secret is to read it,
because a caller that puts a secret into a file on a host can read the file.

A request that uses a reference must also pass two more limits. Each target must
be in the target scope of the token, and in the targets of the vault. A refusal
gives `403`.

## Related

- [Secrets](../../architecture/secrets.md): the design, and the status of each
  part
- [Authentication](authentication.md): tokens, roles and the target scope
- [Audit logging](audit-logging.md): what each entry records
