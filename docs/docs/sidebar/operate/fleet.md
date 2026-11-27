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

## Change the labels and the settings of an agent

The controller holds a config for each agent, by its machine ID. Thus you change
the labels and the settings of an agent with no edit on the host:

```bash
osapi client agent config update --hostname web-01 --labels group=db --max-jobs 4 \
  --log-level debug --facts-interval 5m
osapi client agent config get --hostname web-01
```

The agent reads its config at each heartbeat. Thus the change applies in one
heartbeat, 10 seconds. The labels of the config replace the labels of
`agent.labels` with the same keys. The agent then consumes jobs on the subjects
of its new labels, and writes `config_applied` to its timeline. Max jobs, from 1
to 1000, replaces `agent.max_jobs`: the agent then runs at most that number of
jobs at the same time. The log level, debug, info, warn or error, replaces the
level of `--debug` or `debug`: the agent logs at it at once, with no restart.
The facts interval, a Go duration such as `5m`, replaces `agent.facts.interval`:
the agent collects its facts at it from then on. A setting that an update does
not give takes the value of the agent's config file again. Each update sets all
of the config, so give the labels again when you change only a setting. Each
update increments the version of the config. `agent list` shows the version that
each agent applied, as `config_version`.

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

`osapi client admission revoke`, `reject` and `delete` cut the host's NATS user
from the server at once. The controller adds the user to the revocations of
osapi's account, with the files that `osapi nats auth operator create` writes
and `controller.join.revocation` names. The agent logs
`nats: authentication revoked`. To join again, the host needs a new token or a
new admission, and it gets a new NATS user.

A setup from before these files has no `account.jwt`. osapi's own NATS server
adds the files at its next start. For another server, run
`osapi nats auth operator update`, and merge its `osapi-nats.yaml`. Until then,
the controller logs a warning, and a revoked host stays on NATS until its JWT
expires.

Select the TTL by how long a host can be off: a host that is off for longer than
the TTL cannot renew, and must join again. An empty `agent_ttl` gives a JWT that
does not expire.

## Rotate an agent's keys

The agent sends the rotation on the NATS subject `join.rotate`. In operator
mode, a setup that an older osapi made does not have that subject: run
[`osapi nats auth operator update`](nats.md#check-it) first.

Run this on the agent's host, with the agent's config:

```bash
sudo osapi pki key rotate --name agent
```

1. The command makes two new keys in `agent.pki.key_dir`, beside the old keys,
   as `agent.next.key`, `agent.next.pub` and `agent.next.ekey`.
1. It sends the new public keys to the controller over NATS, as the agent
   connects. The old signing key signs the request, and the new key signs it
   too.
1. The controller checks the old signature against the admission, and that the
   machine is admitted. Then it puts the new keys in the admission. No new
   admission is necessary.
1. The command moves the new keys in place of the old keys. The agent reads them
   at its next heartbeat.
1. During `controller.pki.rotation_grace_period` (`24h`), the controller also
   accepts the old signing key, for a result that the agent signed before it
   read the new keys. The agent also opens a job that the controller sealed to
   its old encryption key.

The controller refuses a revoked machine, and keys that it did not admit. Then
the command removes the new keys, and the agent keeps its keys. When the
controller does not answer, the new keys stay. Run the command again: it sends
the same keys. The audit log records each rotation as `agent.key.rotate`.

## Rotate the controller's key

The controller's key rotates with no restart and no new admission of the agents:

```bash
osapi client pki key rotate
```

1. The controller makes the next key, and gives it to the agents in a record
   that its current key signs.
1. Each agent takes the next key at its next heartbeat.
1. The controller switches to the next key when each admitted agent has it, or
   when `controller.pki.rotation_grace_period` (`24h`) ends. Normally that takes
   one or two heartbeats.
1. Each agent then takes the new key at its next heartbeat. The old key still
   verifies the jobs that it signed before the switch.

Watch the progress:

```bash
osapi client pki key get
```

It shows the next key, and each agent that the switch waits for, with whether it
has the next key. On the controller's host,
`osapi pki key rotate --name controller` does the same as
`client pki key rotate`. A rotation in progress is not started again. An agent
that is away for the whole grace period needs a new admission. Refer to
[Key rotation](../architecture/agent-identity.md#the-controllers-key).

To change the key that signs the API tokens, see
[3. Controller](controller.md#1-make-a-signing-key).

## Remove a host

[4. Join hosts](join.md#remove-a-host) tells the difference between `revoke` and
`delete`.
