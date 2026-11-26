---
sidebar_position: 5
title: Run the fleet
---

# Run the fleet

This page tells how to see the fleet, and how to change it after the setup.

## See the fleet

```bash
osapi client agent list                    # each agent: state, conditions, labels, version
osapi client agent get --hostname web-01   # one agent, with its timeline
osapi client admission list                # each admitted host, online or not
osapi client health status                 # the controller, NATS, each stream and each agent
```

`agent list` has two fields for the trust of each agent:

| Field        | It means                                                           |
| ------------ | ------------------------------------------------------------------ |
| `key_stored` | The controller has the key that it admitted for this agent.        |
| `verified`   | That key signed the last registration, and it claims its own host. |

`key_stored: true` with `verified: false` means that data arrived that did not
pass the check. Examine it.

An agent reports **conditions** at each heartbeat, such as `MemoryPressure`,
`DiskPressure` and `DrainStuck`. A condition shows only while it holds.
[Agent lifecycle](../architecture/agent-lifecycle.md#node-conditions) lists each
one with its limit.

The timeline of an agent records what occurred to it: `admitted`, `started`,
`stopped`, `drain`, `undrain`, `hostname_changed`, `revoked`. The dashboard of
the UI shows each agent as a card.

## Change the labels of an agent

The controller holds a config for each agent, by its machine ID. Thus you change
the labels of an agent with no edit on the host:

```bash
osapi client agent config update --hostname web-01 --labels group=db
osapi client agent config get --hostname web-01
```

The agent reads its config at each heartbeat. Thus the change applies in one
heartbeat, 10 seconds. The labels of the config replace the labels of
`agent.labels` with the same keys. The agent then consumes jobs on the subjects
of its new labels, and writes `config_applied` to its timeline. Each update
increments the version of the config. `agent list` shows the version that each
agent applied, as `config_version`.

The config stays with the machine ID. Thus an agent that starts again, or that
gets a new hostname, keeps it. An update needs the admin role.

## Drain a host

A drained agent takes no new job, and finishes the jobs that it has:

```bash
osapi client agent drain --hostname web-01
osapi client agent undrain --hostname web-01
```

The state goes from `Draining` to `Cordoned` when the jobs stop.

## Rename a host

```bash
osapi client node hostname update --target web-01 --name web-02
```

The machine ID and the key are the identity of the host, and the hostname
follows them. The host stays admitted and keeps its timeline. The controller
refuses a hostname that a different admitted host has.

## Upgrade the fleet

1. Upgrade the controller first. An agent newer than its controller can send
   what the controller does not know.
1. Upgrade each agent. `agent list` shows the version of each agent. The UI
   shows the version on a card only when it is not the controller's.
1. In operator mode, make the NATS credentials again when the notes of the
   release say that a role has a new subject: `osapi nats auth operator create`.

## Renew and revoke the NATS user of an agent

In operator mode, the controller gives each agent its own NATS user, a JWT. With
`controller.join.agent_ttl`, such as `24h`, the JWT expires. The agent renews it
over NATS when a third of its life is left. The controller renews only a host
that is admitted, for a request that its key signed.

Thus `osapi client admission revoke` removes the host from NATS when its JWT
expires. Select the TTL by how fast a revocation must apply, and how long a host
can be off: a host that is off for longer than the TTL cannot renew, and must
join again. An empty `agent_ttl` gives a JWT that does not expire.

## Rotate the controller's key

1. Put a new key pair in `controller.pki.key_dir`.
1. Restart the controller. It signs with the new key.
1. During `controller.pki.rotation_grace_period` (`24h`), each agent accepts the
   old key and the new key, and gets the new key.
1. After the grace period, only the new key is accepted.

To change the key that signs the API tokens, see
[3. Controller](controller.md#1-make-a-signing-key).

## Remove a host

[4. Join hosts](join.md#remove-a-host) tells the difference between `revoke` and
`delete`.
