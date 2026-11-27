---
sidebar_position: 2
title: 2. NATS
---

# 2. Set up NATS

Each osapi process talks through NATS. The controller sends each job on NATS,
and each agent takes its jobs and writes its results there. osapi also keeps its
state in JetStream: the jobs, the agent registry, the facts, the files, the
secrets and the audit log.

This page has three decisions: osapi's NATS or yours, the type of auth, and TLS.

## Select the type of auth

Each osapi process connects as a NATS user with a **role**. The server lets the
user use the subjects of its role, and nothing else. An agent cannot read a
secret, send a job, or make a stream.

| Role         | Who connects as it                                        |
| ------------ | --------------------------------------------------------- |
| `setup`      | `osapi nats setup`, and osapi's NATS server itself        |
| `controller` | the controller, and the local `osapi controller` commands |
| `agent`      | each agent                                                |
| `join`       | an agent with a join string, before its join              |

osapi makes the credentials of each role for you. Each secret goes to a file of
mode 0600, and none is on the command line.

| Type     | Command                           | Join over NATS | Use it when                                                                                |
| -------- | --------------------------------- | -------------- | ------------------------------------------------------------------------------------------ |
| Operator | `osapi nats auth operator create` | Yes            | Recommended. Each agent gets its own NATS user, which can expire and renew.                |
| Password | `osapi nats auth password create` | Yes            | Your server has static users. The agents share one user.                                   |
| Nkey     | `osapi nats auth nkey create`     | No             | Your server uses nkeys. Each agent needs its nkey in its config.                           |
| Callout  | your server's `auth_callout`      | Yes            | Your server asks osapi about each connection. A revocation applies at the next connection. |
| None     | nothing                           | Yes            | Development only.                                                                          |

A shared NATS user does not let one agent answer for another. The controller
admits each agent by its own key, and checks the signature of each answer
([Agent identity](../architecture/agent-identity.md)).

## osapi's NATS

Do these steps on the host of NATS and the controller.

1. Nothing to make for the auth. osapi's NATS server runs in operator mode by
   default. At its first start it makes an operator, an account for osapi, a
   credentials file for each role and two scoped signing keys for the join, in
   `nats-auth` of `nats.server.key_dir`. The controller and an agent on the same
   host connect with their role's credentials from there, with no setting, and
   the controller mints the users of a join with the signing keys. Keep a copy
   of `operator.nk` and `account.nk` offline: they sign the accounts.

   To make the credentials yourself, such as for a controller on another host,
   run `osapi nats auth operator create --out /etc/osapi/nats` and merge its
   `osapi-nats.yaml` into `/etc/osapi/osapi.yaml`. A server with
   `nats.server.auth.resolver_file` uses those, and makes nothing.

1. NATS serves TLS by default. With no `cert_file`, the server makes a
   self-signed certificate, `nats.crt`, in `nats.server.key_dir` at its first
   start. The controller and an agent on the same host trust it with no setting.
   `osapi pki cert get --name nats` shows it. Each joined agent pins it, so
   `osapi pki cert rotate --name nats`, on the controller's host, rotates it
   with an overlap and no restart:

   - It makes the next certificate, `nats.next.crt`, beside the current one, and
     writes a record to the state bucket that the controller's key signs.
   - At its next heartbeat, each agent adds the next certificate to the ones
     that it trusts, and `osapi client agent get` shows it as `nats_cert_next`.
   - When each admitted agent trusts it, or
     `controller.pki.rotation_grace_period` ends, the controller renames it into
     place. NATS serves it at its next handshake, and the agents' connections
     stay up.
   - An agent that was away for the whole overlap trusts the new certificate
     from the record at its next heartbeat, if it still has a connection.
     Otherwise, it joins again.

   A certificate of your own, in `cert_file`, is served the same way when you
   replace its files. At a handshake, NATS reads them again, at most once each
   second, and compares their content, so a copy that keeps the file times, such
   as `cp -p`, is served too. To serve one:

   ```yaml
   nats:
     server:
       tls:
         cert_file: /etc/osapi/tls/nats.crt
         key_file: /etc/osapi/tls/nats.key
   ```

   Set `controller.join.url` to `tls://`. Then the join string has the
   fingerprint of the certificate, and an agent refuses a server with a
   different certificate.

1. As an option, encrypt the store on the disk:

   ```bash
   osapi pki key create --name jetstream -f /etc/osapi/osapi.yaml
   ```

   It writes the key to `nats.server.jetstream.key_file`. The server encrypts
   the store with AES-256-GCM. Keep a copy of the key apart from the backup of
   `nats.server.store_dir`: without it, the server cannot read the store and
   does not start. osapi does not rotate this key.

   A store that an osapi before osapi-io/osapi#1528 encrypted uses
   ChaCha20-Poly1305, and the server does not open it with AES-256-GCM. Stop
   NATS and the controller, remove `nats.server.store_dir`, and start them
   again. The store holds the admissions, the vault, the jobs and the audit log,
   so each agent joins again and each vault value is written again.

1. Start NATS. It makes osapi's streams and buckets when it starts:

   ```bash
   osapi nats start
   ```

## Your NATS

This part is for the administrator of a NATS server that osapi does not run.
Your server needs:

- **JetStream**, on the account of osapi.
- **A user for each role**, with the subjects of the role.
  `osapi nats permissions get --role <role>` prints them for your config and
  namespace, and the flags of `nsc add user`.
- **The streams and buckets of osapi.** `osapi nats setup` makes them, or
  changes them to match the config. Run it one time, and again after an upgrade
  that changes a stream or a bucket.

### When someone else runs the server

osapi cannot add an operator to a server that it does not run: the operator is
in the server's own config. It needs an account, and a key that lets it make the
users of that account. Ask the administrator of the server for one of these, one
time. After it, a host that joins or leaves needs nothing from them.

| Ask for                                                                                           | Then                                                                            |
| ------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------- |
| An account for osapi with JetStream, and two scoped signing keys with the agent and join subjects | [Operator mode](#operator-mode): osapi makes the NATS user of each agent itself |
| An `auth_callout` block that names osapi's issuer key, and an nkey callout user                   | [Auth callout](#auth-callout): osapi decides each connection to its account     |
| A join user and an agent user, with the subjects of each role                                     | [Static users](#static-users): the agents share one user                        |

`osapi nats permissions get --role <role>` prints the subjects to give them.
Each way also needs the controller's user and setup's user in the same account.

### Operator mode

```bash
osapi nats auth operator create --out /etc/osapi/nats --print
```

1. Load `resolver.conf` into the config of your server, with `include`. It has
   the operator, the system account and osapi's account with JetStream.
1. Set `max_control_line: 32KB` in the config of your server. A user JWT has the
   subjects of its role, and the controller's JWT is about 7 KB.
1. Merge `osapi-nats.yaml` into the controller's config, without its
   `nats.server` part. It has the controller's credentials and the two scoped
   signing keys. The controller makes the NATS user of each agent with them, and
   the template of each key limits what the controller can give.
1. Run `osapi nats setup`.

To use an account that you have already, add two scoped signing keys to it with
`nsc edit signing-key`. Give them the subjects that
`osapi nats permissions get --role agent` and `--role join` print. Then set
`controller.join.scoped: true`, `signing_key_file` (the key of the agents) and
`join_signing_key_file`.

### Static users

```bash
osapi nats auth password create --out /etc/osapi/nats --print
```

1. Load the `authorization` block of `nats-server.conf` into the config of your
   server. It has a user for each role.
1. Give the controller its user, and the two users of the join, from the same
   file:

   ```yaml
   controller:
     nats:
       auth:
         type: user_pass
         username: controller
         password: <from nats-server.conf>
     join:
       url: tls://nats.example.com:4222
       user:
         username: join
         password: <from nats-server.conf>
       agent:
         username: agent
         password: <from nats-server.conf>
   ```

1. Run `osapi nats setup` as the setup user.

### Nkeys

```bash
osapi nats auth nkey create --out /etc/osapi/nats --print
```

1. Load the `authorization` block of `nats-server.conf` into the config of your
   server. It has an nkey for the setup, controller and agent roles.
1. Give the controller `controller.nk`, and run `osapi nats setup` with
   `setup.nk`, as `osapi-nats.yaml` shows.

A join over NATS needs a password or a JWT for its bootstrap user. With nkeys,
give each agent `agent.nk` in its config. Then it joins over HTTPS, or an admin
admits it ([4. Join hosts](join.md)).

### Auth callout

With auth callout (NATS 2.10 or later), your server sends each new connection to
the controller, and the controller decides it:

- A connection with a join token that admits now gets the join role, for 10
  minutes at most.
- A connection with the NATS user key of an admitted agent, which signs the
  server's nonce with it, gets the agent role.
- The controller refuses each other connection. It refuses a revoked agent at
  its next connection.

Each decision is an entry of the audit log, `nats.connect`.

1. Make the issuer key and the nkeys:

   ```bash
   osapi nats auth operator create --out /etc/osapi/nats   # account.nk is the issuer
   osapi nats auth nkey create --out /etc/osapi/nats-nkeys
   ```

   The second command makes an nkey for setup, for the controller and for the
   agents. With auth callout each agent connects with its own key, so no agent
   uses `agent.nk`: use it as the callout user.

1. Give your server an account for the callout user, an account for osapi with
   JetStream, and the `auth_callout` block. The callout user, the controller and
   setup are in `auth_users`, so the callout does not decide them:

   ```text
   accounts {
     AUTH  { users: [ { nkey: "<callout user, U...>" } ] }
     OSAPI {
       jetstream: enabled
       users: [ { nkey: "<controller, U...>" }, { nkey: "<setup, U...>" } ]
     }
   }
   authorization {
     auth_callout {
       issuer: "<public key of account.nk, A...>"
       auth_users: [ "<callout user>", "<controller>", "<setup>" ]
       account: AUTH
     }
   }
   ```

   Make the callout user an nkey. A server offers each connection a nonce only
   when it has an nkey user, and an agent proves its key by signing the nonce.

1. Give the controller the callout:

   ```yaml
   controller:
     join:
       url: tls://nats.example.com:4222
       agent_ttl: 1h
       callout:
         issuer_key_file: /etc/osapi/nats/account.nk
         account: OSAPI
         auth:
           type: nkey
           nkey_file: /etc/osapi/nats-nkeys/agent.nk
   ```

   `agent_ttl` is how long an agent's connection lasts before the server asks
   the controller again. Thus it is the longest time that a revoked agent stays
   connected.

1. Run `osapi nats setup` as the setup user, and make a join token. The join
   string has no user: the agent connects with its token.

### TLS to your server

Give the controller the CA of your server:

```yaml
controller:
  nats:
    tls:
      enabled: true
      ca_file: /etc/osapi/tls/ca.crt
```

An agent that joins needs no CA. Its join string starts with `tls://` and has
the fingerprint of the server's certificate, which the agent pins. An agent that
you configure by hand needs `agent.nats.tls`, with the same keys.

## Namespace

Each subject, stream and bucket of osapi starts with the namespace, `osapi` by
default. Two fleets on one server use two namespaces. Set the same value in
`nats.server.namespace`, `controller.nats.namespace` and `agent.nats.namespace`.
A join string carries it to each agent.

## Check it

```bash
osapi config validate -f /etc/osapi/osapi.yaml
osapi nats permissions get --role agent     # the subjects that an agent may use
```

After the controller starts, `osapi client health status` shows NATS and each
stream.

The credentials of `operator create` have the subjects of the version that made
them, and so has the template of the agents' signing key in `resolver.conf`.
After an upgrade that adds a subject, such as `join.rotate` for the rotation of
an agent's keys, sign them again with the same keys:

```bash
osapi nats auth operator update --out /etc/osapi/nats
```

Then restart the NATS server, which reads `resolver.conf` at its start, and the
controller, for its new credentials. The keys do not change, so each agent keeps
its NATS user and no agent joins again. Until then, the server refuses each
subject that the old credentials do not have, and its log has a
`Permissions Violation` for each.
[NATS subjects and roles](../architecture/nats.md) tells why each role has its
subjects.

Next: [3. Controller](controller.md).
