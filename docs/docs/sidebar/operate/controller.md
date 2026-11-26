---
sidebar_position: 3
title: 3. Controller
---

# 3. Set up the controller

The controller serves the API and the UI. It sends each job, checks each answer,
and admits each host. Do these steps on the host of the controller.

## 1. Make a signing key

The controller signs each API token with a key. Make one of 32 characters or
more:

```yaml
controller:
  api:
    security:
      signing_keys:
        - id: '2026-11'
          key: <openssl rand -hex 32>
```

The first key of the list signs. Each key of the list verifies the tokens that
have its `id`. Thus you change the key with no break: add the new key first, and
remove the old key after each old token expires
([Authentication](secure/authentication.md#change-the-signing-key)).

## 2. TLS for the API

The API serves HTTPS by default. With no `cert_file`, the controller makes a
self-signed certificate, `api.crt`, in `controller.pki.key_dir` at its first
start. To serve a certificate of your own, such as one from your CA:

```yaml
controller:
  api:
    tls:
      cert_file: /etc/osapi/tls/api.crt
      key_file: /etc/osapi/tls/api.key
```

The CLI on the controller's host trusts the self-signed `api.crt` with no
setting. A CLI on another host needs it as `controller.client.tls.ca_file`, or a
certificate of a CA that it trusts. A join over HTTPS pins the certificate.

## 3. Start the controller

```bash
osapi controller start
```

At its first start the controller makes its key pair in `controller.pki.key_dir`
(`/etc/osapi/pki`), and the key of the `default` secrets vault
([Secrets vault](#secrets-vault)). Run it as a service, as
[4. Join hosts](join.md#run-the-agent-as-a-service) shows for the agent, with
`controller start` in place of `agent start`.

## 4. Make an admin token

```bash
osapi controller token generate -r admin -u admin@example.com
```

Give the token to the CLI of each person, in their config:

```yaml
controller:
  client:
    url: https://controller.example.com:8080
    security:
      bearer_token: <the token>
```

Give each person the role that their work needs, not `admin`. `-t group:web`
limits a token to some hosts. `osapi controller role list` shows each role and
its permissions. [Authentication](secure/authentication.md) tells how a token is
checked.

## 5. Open the UI

The controller serves the UI at the address of its API. Sign in with a token.
The UI shows only the controls that the roles of the token permit.
`controller.ui.enabled: false` turns it off.

## Secrets vault

A vault keeps secrets for the controller, such as a registry password that a job
uses. A config with no `controller.secrets.vaults` has one vault, `default`. Its
key is in `/etc/osapi/secrets.key`, and the controller makes it at its first
start, with mode 0600. Keep a copy of the key: without it, no value in the vault
decrypts.

```bash
osapi controller secret key create --vault default   # make the key before the first start
osapi controller secret key rotate --vault default   # change the key
```

A secret setting of the config can be a reference to a vault, in place of the
value, such as `secret://default/jwt-signing-key`. A NATS password and the key
of a vault cannot be a reference, because osapi needs them before it can read a
vault. [Secrets](secure/secrets.md) tells how to write and read a secret.

<!-- osapi:config controller.secrets nats.secrets -->

| Key                                    | Type   | Default                                                                                                          | Environment                       | Description                                                                                                                                |
| -------------------------------------- | ------ | ---------------------------------------------------------------------------------------------------------------- | --------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------ |
| `controller.secrets.vaults`            | list   | `[{"Name":"default","Type":"nats","Targets":null,"KeyFile":"/etc/osapi/secrets.key","KeyEnv":"","Config":null}]` | `OSAPI_CONTROLLER_SECRETS_VAULTS` | The vaults that a reference such as secret://default/registry-token names, each with name, type, targets, key_file or key_env, and config. |
| `nats.secrets.bucket`                  | string | `secrets`                                                                                                        | `OSAPI_NATS_SECRETS_BUCKET`       | KV bucket for the encrypted values of each nats vault.                                                                                     |
| `nats.secrets.history`                 | int    | `10`                                                                                                             | `OSAPI_NATS_SECRETS_HISTORY`      | How many versions of each secret the bucket keeps, from 1 to 64.                                                                           |
| `nats.secrets.storage`                 | string | `file`                                                                                                           | `OSAPI_NATS_SECRETS_STORAGE`      | Storage backend: 'file' or 'memory'.                                                                                                       |
| `nats.secrets.replicas`                | int    | `1`                                                                                                              | `OSAPI_NATS_SECRETS_REPLICAS`     | Number of KV replicas.                                                                                                                     |
| `controller.secrets.vaults[].name`     | string | none                                                                                                             | none                              | The name that a reference gives, such as default in secret://default/registry-token.                                                       |
| `controller.secrets.vaults[].type`     | string | none                                                                                                             | none                              | The type of the vault, such as nats.                                                                                                       |
| `controller.secrets.vaults[].targets`  | list   | none                                                                                                             | none                              | The targets that the secrets of the vault can go to, such as group:prod. Empty is each host.                                               |
| `controller.secrets.vaults[].key_file` | string | none                                                                                                             | none                              | For a type that encrypts, such as nats: the file of the vault key. osapi controller secret key create writes it.                           |
| `controller.secrets.vaults[].key_env`  | string | none                                                                                                             | none                              | For a type that encrypts, such as nats: the environment variable of the vault key, in place of key_file.                                   |
| `controller.secrets.vaults[].config`   | map    | none                                                                                                             | none                              | The settings of the type. The struct of the type gives its keys.                                                                           |

<!-- /osapi:config -->

## Check it

```bash
osapi client health status         # the controller, NATS, and each stream and bucket
osapi controller key fingerprint   # the key that each agent gets at its join
```

Then open the UI and sign in.

Next: [4. Join hosts](join.md).
