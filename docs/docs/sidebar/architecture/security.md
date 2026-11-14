---
sidebar_position: 6
---

# Security

Four things decide what can happen on a host: who a request is from, what that
caller may do, which agent is allowed to answer, and what the agent itself may
run. Every request leaves a record.

## Who is asking

Every API request carries a JWT, signed with the controller's key
(`controller.api.security.signing_key`, HS256). `osapi token generate` makes one
with roles, a subject and optionally a fixed list of permissions;
`osapi token validate` shows what a token grants. The CLI sends the token in
`controller.client.security.bearer_token`. Only the liveness and readiness
probes, `/api/health` and `/api/health/ready`, answer without one.

## What they may do

Every operation needs one permission, such as `sysctl:read` or
`command:execute`. A token gets permissions from its roles, or carries its own
list instead:

- **Built-in roles**: `admin`, `write` and `read`. `osapi token roles` prints
  every permission and which role grants it, as the binary you run has them.
- **Custom roles**, defined in `controller.api.security.roles`, each a list of
  permissions. A custom role with a built-in name replaces it.
- **Behaviors** add their own permissions, `<name>:read` and `<name>:write`,
  which `admin` and `write` grant in full and `read` grants the read of.

Permissions are split by what an operation can do, not by area. Running an
arbitrary command (`command:execute`) and a shell (`command:shell`) are
separate, and only `admin` has either.

## Which agent may answer

Every agent is identified by its **machine ID**, not its hostname. Jobs are
routed to the machine ID, so renaming a host does not redirect its work.

With PKI on (`controller.pki.enabled` and `agent.pki.enabled`), an agent must be
accepted before it gets any work:

1. On first start it generates an Ed25519 keypair in `agent.pki.key_dir`.
2. It asks to enroll, and waits; `osapi client agent list --pending` lists it.
3. An administrator accepts it, by hostname or by key fingerprint, or the
   controller accepts every agent itself with `controller.pki.auto_accept`.
4. From then on the controller signs every job, the agent refuses any job that
   is not signed by the controller, and the agent signs its results and
   heartbeats. An agent whose heartbeat is not signed by its accepted key, or
   that claims a hostname it did not enroll as, stops being routed to.

Keys rotate with a grace period (`controller.pki.rotation_grace_period`) in
which both the old and the new key are accepted.

## What the agent may run

The agent runs as an unprivileged user unless told otherwise. With
`agent.privilege_escalation` on, it runs the commands that change the host
through `sudo`, and reads without it, and the sudoers drop-in lists exactly
which programs and arguments it may run. Linux capabilities can stand in for
sudo where they cover the operation. At startup a preflight checks every program
the agent will need sudo for, behaviors' included, and every capability, and the
agent refuses to start if one is missing.

Input that becomes a file path, a filename or a command argument is checked
twice: by the API before a job is queued, and by the provider on the host before
it is used, because a job is stored and run later. A secret reaches a command on
its standard input, never as an argument. Every command runs through one place,
with a timeout of at most ten minutes.

## Audit

Every authenticated request is recorded, whatever its outcome: who made it, with
which roles and permissions, the method and path, the status, the time it took
and a summary of what it asked for. Fields that carry secrets, such as
`password`, `token` and `secret`, are replaced with `[redacted]` in that
summary. The record is a NATS stream kept for `nats.audit.max_age`;
`osapi client audit list` and `osapi client audit export` read it.

## Design details

- [Permissions](https://github.com/osapi-io/specs/blob/main/components/osapi/permissions.md),
  [agent identity](https://github.com/osapi-io/specs/blob/main/components/osapi/agent-identity.md),
  [running commands](https://github.com/osapi-io/specs/blob/main/components/osapi/exec.md)
  and
  [audit](https://github.com/osapi-io/specs/blob/main/components/osapi/audit.md).
- [Authentication](../features/authentication.md),
  [Agent identity and PKI](../features/agent-identity.md),
  [Agent hardening](../features/agent-hardening.md) and
  [Audit logging](../features/audit-logging.md), for configuring each.
