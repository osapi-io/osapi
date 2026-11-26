---
sidebar_position: 4
title: A fleet on your NATS
---

# A fleet on your NATS

This page is for the administrator of a NATS server that osapi does not run.
osapi makes the credentials of each of its roles. You load the server's part
into your server, and give osapi the rest.

| Role         | Who connects as it                                      |
| ------------ | ------------------------------------------------------- |
| `setup`      | `osapi nats setup`, which makes the streams and buckets |
| `controller` | the controller                                          |
| `agent`      | each agent                                              |
| `join`       | an agent with a join string, before its join            |

`osapi nats permissions get --role <role>` prints the subjects of a role.

## Operator mode

```bash
osapi nats auth operator create --out /etc/osapi/nats --print
```

1. Load `resolver.conf` into the config of your server, with `include`. It has
   the operator, the system account and osapi's account with JetStream.
2. Set `max_control_line: 32KB` in the config of your server: a user JWT has the
   subjects of its role, and the controller's is about 7 KB.
3. Merge `osapi-nats.yaml` into the controller's config, without its
   `nats.server` part. It sets the controller's credentials and
   `controller.join.signing_key_file`, which the controller mints the users of a
   join with.
4. Make the streams and buckets:

   ```bash
   osapi nats setup
   ```

5. Make a join token, and start the agents with the join string
   ([A fleet on osapi's NATS](fleet.md#3-make-a-join-token)).

A scoped signing key, whose template limits what osapi can give, is
osapi-io/osapi#1432.

## Static users

```bash
osapi nats auth password create --out /etc/osapi/nats --print
```

1. Load the `authorization` block of `nats-server.conf` into the config of your
   server. It has a user for each role, with the permissions of the role.
2. Give the controller its user, and the users of the join, from the same file:

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

3. Run `osapi nats setup` with the setup user, then make a join token.

## Nkeys

```bash
osapi nats auth nkey create --out /etc/osapi/nats --print
```

1. Load the `authorization` block of `nats-server.conf` into the config of your
   server. It has an nkey for the setup, controller and agent roles, with the
   permissions of each.
2. Give the controller `controller.nk` and the agents `agent.nk`, as
   `osapi-nats.yaml` shows, and run `osapi nats setup` with `setup.nk`.

A join over NATS needs a password or a JWT for its bootstrap user: use the
static users or operator mode for it. With osapi's own server, `osapi-nats.yaml`
sets `nats.server.auth.nkey_roles`, which limits each nkey to its role.

## Auth callout

The join with auth callout is osapi-io/osapi#1416. Until then, give each role
the subjects that `osapi nats permissions get` prints.

[Running osapi on your NATS](architecture/nats.md) is the reference for each
setting.
