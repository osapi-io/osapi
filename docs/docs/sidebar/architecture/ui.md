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
of role names, such as `admin`, `operator` or `read`.

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
example `agent:read`, `file:write`, `command:execute` and `docker:execute`. The
UI's list of permissions and the permissions of each role are in
`ui/src/lib/permissions.gen.ts`, which `go generate ./internal/behaviors/all`
writes from the Go model. Do not edit it. `BLOCK_PERMISSIONS`
(`ui/src/lib/permissions.ts`) gives the permission that each Configure block
needs, and TypeScript refuses a permission that the generated list does not
have. The UI shows a block without permission in gray, with a lock icon.

## Pages

The routes are in `ui/src/App.tsx`. The navbar has Dashboard and Configure. The
Admin menu has Audit Log, Admission, Jobs, Roles and Traces. A page that needs a
permission that the session does not have shows that the permission is missing.

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

- The Add Block panel has a tab for each area that has blocks, in the order of
  `area.All` in Go: Services, Software, Config, Node, Logs, Networking,
  Security, Containers, Object Store, Command and Hardware. The Audit and Fleet
  areas have no blocks, so they have no tab.
- In a tab, the blocks of each behavior are in a category with the title of the
  behavior, such as Package updates. A behavior under another behavior, such as
  User SSH key, comes after its parent.
- The RBAC permissions control the blocks. A block that the session cannot run
  is gray, with a lock icon.
- Each block has a target picker (`_all`, `_any`, a hostname, labels).
- Apply runs the blocks in sequence, with a spinner for each block, and shows
  the results. A result links to its job and its trace.

### Audit Log (`/admin/audit`)

The history of the API requests: who sent each request, the operation, the
status and the duration. Filters select by user, operation or path, and status,
and the reads show only when you ask for them. An entry opens in full, with the
chain of calls that its job made on each host, and links to its job and its
trace. The page needs `audit:read`, which only the admin role has.

### Admission (`/admin/admission`)

The machines that ask to join the fleet, in four sections: Pending, Admitted,
Rejected and Revoked. You admit or reject a pending machine, and revoke an
admitted machine.

### Jobs (`/admin/jobs`)

The job queue and its history, with a status filter. You can retry or delete a
job. `/admin/jobs/:id` shows the detail of one job above the list, such as the
job of a block's result on the Configure page.

### Roles (`/admin/roles`)

An RBAC reference: the current session with a role badge, a table of the role
definitions, the full permission matrix and the table of block permissions.
`/roles` goes to this page.

### Traces (`/admin/traces`)

The span tree of a trace, read through the controller. Type the trace ID that a
command prints after its job ID, or open `/admin/traces/:id`. The page needs
`trace:read`.

### Sign in

JWT authentication: a field to paste a token, with validation. The page reads
the roles from the JWT claims, and shows a CLI hint for `osapi token generate`.
The page shows when the session has no token, on each route.
