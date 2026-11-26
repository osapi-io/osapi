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

1. Make the credentials:

   ```bash
   osapi nats auth operator create --out /etc/osapi/nats
   ```

   It writes an operator, an account for osapi, a credentials file for each
   role, two scoped signing keys for the join, and `osapi-nats.yaml`. Keep
   `operator.nk` and `account.nk` offline: they sign the accounts.

1. Merge `osapi-nats.yaml` into `/etc/osapi/osapi.yaml`. It sets the auth of the
   server, of the controller and of the join.

1. Turn on TLS when the network is not yours. Without TLS, a job, the output of
   a command and the facts cross the network as plain text:

   ```yaml
   nats:
     server:
       tls:
         enabled: true
         cert_file: /etc/osapi/tls/nats.crt
         key_file: /etc/osapi/tls/nats.key
   ```

   Then the join string starts with `tls://`, and it has the fingerprint of the
   certificate. An agent refuses a server with a different certificate.

1. As an option, encrypt the store on the disk:

   ```bash
   osapi nats key create -f /etc/osapi/osapi.yaml
   ```

   It writes the key to `nats.server.jetstream.key_file`. Keep a copy of the key
   apart from the backup of `nats.server.store_dir`: without it, the server
   cannot read the store and does not start. osapi does not rotate this key.

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
them. After an upgrade that adds a subject, make them again.
[NATS subjects and roles](../architecture/nats.md) tells why each role has its
subjects.

Next: [3. Controller](controller.md).
