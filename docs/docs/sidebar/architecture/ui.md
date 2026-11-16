---
sidebar_position: 9
---

# UI architecture

osapi has an embedded management dashboard. It is a React single-page
application. The same host and port as the REST API serve it. It shows the fleet
health, the agents and the jobs, and it lets you put operations together from
blocks. You do not need to deploy a different frontend.

This page tells what an operator controls: the setting that disables the UI, the
roles that control what it shows, and the pages. The
[specifications repository](https://github.com/osapi-io/specs/blob/main/components/osapi/ui.md)
tells how the UI is built, embedded, structured and generated.

## Configuration

To disable the UI, set `controller.ui.enabled: false` in `osapi.yaml`. Then the
controller does not register the SPA handler, and it serves only the REST API.

```yaml
controller:
  ui:
    enabled: true # default: true
```

The UI uses the same host and port as the REST API (`controller.api.port`). Thus
you need no more network configuration.

## Authentication and authorization

The UI uses the same JWT authentication as the rest of osapi.
`osapi token generate` makes the tokens. A token has a `roles` claim with a list
of role names (`admin`, `write`, `read`).

### RBAC model

Three built-in roles have permissions in levels:

| Role     | JWT value | Permissions                            |
| -------- | --------- | -------------------------------------- |
| Admin    | `admin`   | All permissions, `audit:read` included |
| Operator | `write`   | Read, write and execute, but not audit |
| Viewer   | `read`    | Read only                              |

Each role can read traces with `trace:read`. The Traces page in the Admin menu
shows the span tree of a trace by its ID.

A permission has the form `resource:verb`, as in the Go model of osapi, for
example `agent:read`, `file:write`, `command:execute` and `docker:execute`.
`BLOCK_PERMISSIONS` (`ui/src/lib/permissions.ts`) gives the permission that each
Configure block needs. The UI shows a block without permission in gray, with a
lock icon.

## Pages

### Dashboard (`/`)

An overview of the fleet health:

- cards with summary numbers
- the health of the controller and the NATS server, with hostnames and resource
  use
- the counts of JetStream streams and consumers
- the use of the KV stores and the object store
- agent cards with the status, conditions and labels, and drain and undrain
  actions

### Configure (`/configure`)

A builder that puts operations together from blocks:

- a sidebar with block categories (Schedule, File, Docker, Command, DNS,
  Network)
- blocks that the RBAC permissions control
- a target picker for each block (`_all`, `_any`, a hostname, labels)
- apply in sequence, with a spinner for each block
- the results

### Roles (`/roles`)

An RBAC reference: the current session with a role badge, a table of the role
definitions, the full permission matrix and the table of block permissions.

### Enrollment (`/admin/enrollment`)

PKI enrollment management:

- The page lists the pending agents, with the machine ID, the fingerprint and
  the time of the request. Each agent has Accept and Reject buttons. When no
  agent is pending, the page shows an empty state.
- Below them, the page lists the enrolled agents: the agents whose key the
  controller stores. Each has a Remove button. To confirm, click it a second
  time. Remove sends the machine ID of the card. Thus it removes that agent,
  also when another agent has the same hostname.
- The command bar has accept, reject and remove commands, by hostname.

### SignIn

JWT authentication: a field to paste a token, with validation. The page reads
the roles from the JWT claims, and shows a CLI hint for `osapi token generate`.
