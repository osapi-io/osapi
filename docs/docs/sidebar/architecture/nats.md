---
sidebar_position: 8.5
---

# Running osapi on your NATS

This page is for the NATS administrator: the person who runs the NATS
infrastructure that osapi uses. (In NATS, the operator is the root of the JWTs
in operator mode, so this page calls the person the administrator.) It tells
what osapi needs from a NATS server, what to provision for each kind of server,
and how a fleet whose hosts come and go connects.

You do not need this page when osapi runs its own embedded server
(`osapi nats start`, or `osapi start`). That server makes its streams and
buckets at start, and needs nothing from you.

## What osapi needs from a NATS server

- **JetStream.** osapi keeps its state in JetStream: the jobs and their results,
  the agent registry, the facts, the uploaded files, the secrets and the audit
  log. [State and storage](state.md) lists each stream and bucket, and its size
  settings. The object store of the uploaded files is 10 GiB by default
  (`nats.objects.max_bytes`).
- **Its own account, or its own users.** The isolation of osapi is the account,
  or the users that you give it. osapi does not define a smaller set of
  permissions for each of its processes yet (osapi-io/osapi#1415). Thus give
  osapi an account of its own, or users that can use the subjects and the
  JetStream API of its namespace.
- **A namespace.** Each subject, stream and bucket of osapi starts with its
  namespace, `osapi` by default. Two deployments, such as staging and
  production, share a server with two namespaces
  ([More than one deployment on one NATS](deployment.md#more-than-one-deployment-on-one-nats)).
- **Its streams and buckets.** `osapi nats setup` makes or updates each one to
  match the config, with the credentials of `controller.nats`. Run it one time,
  and again after a change to a stream or a bucket. A user that runs it needs
  the JetStream API to make streams, buckets and object stores. An agent makes
  the object store bucket of the uploaded files when it is not there.
- **TLS, if the network is not yours.** `nats.server.tls` is the embedded
  server's. For your server, set `controller.nats.tls` and `agent.nats.tls`
  ([TLS](../usage/configuration.md)). Without TLS, a job's payload, a command's
  output and the facts cross the network in plain text; the controller still
  signs each job and verifies each answer.

## What each kind of server needs

Each osapi process authenticates with its own `nats.auth`: `none`, `user_pass`,
`nkey` or `creds`.

| Your NATS server      | What you provision                                                                                     | What each agent uses        |
| --------------------- | ------------------------------------------------------------------------------------------------------ | --------------------------- |
| Operator mode         | An account for osapi with JetStream on, and a user for the controller and for the agents (creds files) | `creds`: a credentials file |
| Static users or nkeys | A user or an nkey for the controller, and one for the agents                                           | `user_pass` or `nkey`       |
| No authentication     | Nothing. Use this only for development.                                                                | `none`                      |

The agents can share one user. A NATS user is not the identity of an agent: the
controller admits each agent by its own key, and verifies each answer that it
signs ([Agent identity](../features/platform/agent-identity.md)). Thus a shared
user does not let one agent answer for another.

## Hosts that come and go

You do not make a NATS user for each host. The agents share their user, so a
fleet that scales up and down needs nothing from you at each host. What each
host needs:

- **A machine ID of its own.** An agent's identity is `/etc/machine-id` on
  Linux. An image that ships a filled-in machine ID gives each clone the same
  identity, and a second clone replaces the first in admission. Ship the image
  with an empty `/etc/machine-id`, so that systemd makes one at the first boot.
- **Labels, not hostnames.** A host whose hostname is random is targeted by its
  labels, such as `--target role:web` with `agent.labels`.
- **Admission.** An agent asks to be admitted when it first starts, and an admin
  admits it ([Agent identity](../features/platform/agent-identity.md)). For
  hosts that come up on their own, an agent can join with a join token, which
  admits it at once.
- **Cleanup.** `controller.pki.delete_after`, such as `24h`, deletes a machine
  that has not reported for that time. Thus a host that went away leaves no
  record.

## Planned: agents that join over NATS

This part is planned work (osapi-io/osapi#1313, osapi-io/osapi#1416). Today an
agent needs its NATS settings and credentials in its own config. With the
planned join, a host needs only a join string. The agent connects to NATS with a
bootstrap credential that can only ask to join, the controller admits it, and
gives it its own credential and its config over NATS. An agent then needs to
reach NATS, and nothing else, which fits a site behind a NATS leaf node or a
firewall that lets only NATS through.

What you provision then depends on your server:

| Your NATS server   | What you give osapi                                                            |
| ------------------ | ------------------------------------------------------------------------------ |
| Operator mode      | An account, and a scoped signing key whose template is the agent's permissions |
| Auth callout       | An `auth_callout` block that names osapi's issuer key (NATS 2.10 or later)     |
| Static users       | A join user that can send only the join request, and an agent user             |
| osapi's own server | Nothing                                                                        |

With a scoped signing key, the operator's key stays with you: the template sets
the most that osapi can give an agent. The design is
[agent join](https://github.com/osapi-io/specs/blob/main/components/osapi/agent-join.md)
in the specs, and this page changes as each part is built.
