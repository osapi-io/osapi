---
sidebar_position: 9
---

# UI architecture

osapi has an embedded management dashboard. It is a React single-page
application. The same host and port as the REST API serve it. It shows the fleet
health, the agents and the jobs. It shows the behaviors that osapi manages, and
it runs plans: steps that are methods of those behaviors. You do not need to
deploy a different frontend.

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

Six built-in roles have permissions. A rule gives each role its permissions from
the verbs of the permissions.
[Roles and permissions](../features/platform/authentication.md#roles-and-permissions)
tells the rule. In the UI, each role can do these things:

| Role       | JWT role     | In the UI                                                                                           |
| ---------- | ------------ | --------------------------------------------------------------------------------------------------- |
| Admin      | `admin`      | Each method and each page of the Admin menu. Only admin can switch the role of the session          |
| Operator   | `operator`   | Each read, and each action, such as a restart, a signal or a reboot. No change to the configuration |
| Automation | `automation` | Each read, and each change to the configuration. No action                                          |
| Auditor    | `auditor`    | The Audit Log, Jobs, Roles and Traces pages, and the health. It changes nothing                     |
| Read       | `read`       | Each read that is not admin only. A read of a log, a file or a process can show a secret            |
| Observer   | `observer`   | The health, the status of each host, and the agents with their facts, for a dashboard. Nothing more |

A permission with the verb `shell` or `exec` is admin only. Some other
permissions are admin only too, such as `command:execute`, `user:write`,
`schedule:write`, `admission:write` and `agent:write`. `audit:read` goes to
`admin` and `auditor` only.

The Audit Log, Admission, Jobs and Roles links of the Admin menu need
`audit:read`. The Traces link needs `trace:read`, which each role but `observer`
has. The Traces page shows the span tree of a trace by its ID.

A permission has the form `resource:verb`, as in the Go model of osapi, for
example `agent:read`, `file:write`, `command:execute` and `docker:execute`. The
UI's list of permissions and the permissions of each role are in
`ui/src/lib/permissions.gen.ts`, which `go generate ./internal/behaviors/all`
writes from the Go model. Do not edit it. The permission of each method comes
from the catalog of the behaviors, `ui/src/lib/catalog.gen.ts`, which the same
command writes. The UI shows a method that your role cannot run as faint, with a
lock icon, and you cannot add it to a plan. The controller makes the same check
on each request.

## Pages

The routes are in `ui/src/App.tsx`. The navbar has Dashboard, Behaviors and
Plans. The Admin menu has Audit Log, Admission, Jobs, Roles and Traces. A page
that needs a permission that the session does not have shows that the permission
is missing.

### Dashboard (`/`)

An overview of the fleet health:

- cards with summary numbers
- the health of the controller and the NATS server, with hostnames and resource
  use
- the counts of JetStream streams and consumers
- the use of the KV stores and the object store
- agent cards with the status, conditions and labels, and drain and undrain
  actions

### Behaviors (`/behaviors`)

The catalog of what osapi manages on a host, by area:

- a section for each area, such as Networking or Security, in the order of the
  SDK and the docs site
- a row for each behavior: its title, its name, what it manages, the behaviors
  that it calls, and a mark for each method: `R` for a read, `W` for a write and
  `A` for an action
- a faint mark for a method that your role cannot run, and a lock on a behavior
  whose methods your role cannot run
- a filter, which takes the name of a behavior or a method, such as `dns update`
- the internal behaviors last. Only other behaviors call them, so a plan cannot
  run them.

### A behavior (`/behaviors/<name>`)

One behavior, as `osapi behavior get` and `osapi behavior chain get` print it:
its area, its platforms, and the behaviors that it calls and that call it. Each
method has:

- its input and output types, with each required field marked `*`, and its route
- its permission, and whether your role has it
- the platforms that it runs on
- its chain: the calls that it makes to other behaviors, such as `dns.update`,
  then `netplan.update`, then `managedfile.update`
- its examples, with their input and output
- **Add to plan**, which adds the method to the plan as a step

A link such as `/behaviors/dns#update` opens the page at the method.

### Plans (`/plans/new`)

A plan is a list of steps that run in order. Each step is a method of a
behavior, with its target and its input. This is the shape of a step of a
workflow file: `run`, `target` and `with`.

- The list at the left has each area, then each behavior, then the methods of
  the behavior. Click a method to add it as a step. The filter opens each
  behavior that matches.
- A step card has the title of its behavior and its method, such as
  `DNS · Update`, the input and output types, and the behaviors that it calls.
- A read starts with the target `_all`. A write or an action starts with no
  target. It is pending, and Run does not run it, until you pick its hosts.
- The form shows the doc of each field. A field with fixed values is a select,
  and a list is a field of tags.
- Above the steps, a preflight shows each permission that the plan needs and
  whether your role has it. It also shows each host where a method does not run,
  such as a Debian-only write on a macOS host, and the steps that wait for input
  or a target.
- **Run** runs each step that can run, in order, and continues after a step
  fails. Each result shows its job and its trace, the rows that the method
  returned, and the chain of calls on each host.
- **Export YAML** shows the plan as a workflow file. The UI does not save a
  plan. Copy the YAML to keep it.
- The steps of the Object Store area, such as an upload, run on the controller
  and take no target.

The plan stays while you move between pages. Thus you can add methods from the
page of each behavior, then open Plans.

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
job of a step's result on the Plans page.

### Roles (`/admin/roles`)

An RBAC reference: the current session with a role badge, a table of the role
definitions, the full permission matrix and the permission of each step of a
plan.

### Traces (`/admin/traces`)

The span tree of a trace, read through the controller. Type the trace ID that a
command prints after its job ID, or open `/admin/traces/:id`. The page needs
`trace:read`.

### Sign in

JWT authentication: a field to paste a token, with validation. The page reads
the roles from the JWT claims, and shows a CLI hint for `osapi token generate`.
The page shows when the session has no token, on each route.
