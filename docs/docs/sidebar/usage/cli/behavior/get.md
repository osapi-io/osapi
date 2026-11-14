# Get

Show one behavior in full: where it is served, its platforms, the programs it
runs with sudo, the behaviors it uses and is used by, its methods with their
routes and permissions, its types, settings and telemetry.

```bash
$ osapi behavior get --name sysctl

Behavior:
  Name:         sysctl
  Description:  Kernel parameters, written to sysctl.d files and applied
  Placement:    node
  Internal:     no
  API:          /api/next
  CLI:          osapi client next
  Platforms:    debian
  Sudo:         sysctl

Uses:
  NAME         DESCRIPTION
  managedfile  Files osapi writes on a host and keeps a record of

Used by:
  None.

Methods (5):
  NAME    VERB   ROUTE                                          PERMISSION    DESCRIPTION
  list    read   GET /api/next/node/{hostname}/sysctl           sysctl:read   List the kernel parameters osapi manages
  get     read   GET /api/next/node/{hostname}/sysctl/{key}     sysctl:read   Get a kernel parameter's current value
  create  write  POST /api/next/node/{hostname}/sysctl          sysctl:write  Manage a kernel parameter and apply it
  update  write  PUT /api/next/node/{hostname}/sysctl/{key}     sysctl:write  Change a managed kernel parameter and apply it
  delete  write  DELETE /api/next/node/{hostname}/sysctl/{key}  sysctl:write  Stop managing a kernel parameter

Types (5):
  NAME    INPUT                         OUTPUT
  list    ListInput                     []Entry { key, value }
  get     GetInput { key* }             Entry { key, value }
  create  CreateInput { key*, value* }  MutationResult { key, changed }
  update  UpdateInput { key*, value* }  MutationResult { key, changed }
  delete  DeleteInput { key* }          MutationResult { key, changed }

Settings (1):
  KEY                   AGENT FLAG    TYPE    DEFAULT
  behaviors.sysctl.dir  --sysctl-dir  string  /etc/sysctl.d

Telemetry:
  Span:             behavior.sysctl.<method>
  Metrics:          osapi_behavior_calls_total, osapi_behavior_duration_seconds, osapi_behavior_changes_total
  Attributes:       osapi.behavior.name, osapi.behavior.method, osapi.behavior.verb, osapi.behavior.outcome, osapi.behavior.changed
  Recorded inputs:  key (get, create, update, delete)
```

## Uses and used by

**Uses** lists the behaviors this one is built on. The agent builds their
providers first, and this behavior's methods call theirs directly, on the same
host inside the same job. If a behavior it uses cannot run on a platform, this
one reports `skipped` there too. **Used by** lists the behaviors built on this
one.

A behavior with `Internal: yes` exists only to be built on: it has no route, no
command and no permission, and only the behaviors under Used by call it.

```bash
$ osapi behavior get --name managedfile

Behavior:
  Name:         managedfile
  Description:  Files osapi writes on a host and keeps a record of
  Placement:    node
  Internal:     yes
  API:          none
  CLI:          none
  Platforms:    debian
  Sudo:         none

Uses:
  None.

Used by:
  NAME    DESCRIPTION
  sysctl  Kernel parameters, written to sysctl.d files and applied
...
```

## One method

`--method` shows one method in full: its fields with their constraints, its
output and its examples.

```bash
$ osapi behavior get --name sysctl --method update

Method:
  Name:         update
  Verb:         write
  Description:  Change a managed kernel parameter and apply it
  Detail:       Fails if osapi does not manage the parameter yet; create it first. Setting the value it already has reports changed: false.
  Route:        PUT /api/next/node/{hostname}/sysctl/{key}
  Permission:   sysctl:write
  Command:      osapi client next node sysctl update

Input (UpdateInput):
  FIELD  TYPE    REQUIRED  CONSTRAINTS                                         DESCRIPTION
  key    string  true      pattern ^[a-zA-Z0-9][a-zA-Z0-9._-]*$, min length 1  Parameter name, such as net.ipv4.ip_forward
  value  string  true      pattern ^[^\r\n]*$                                  Value to set

Output (MutationResult):
  FIELD    TYPE     DESCRIPTION
  key      string   Parameter name
  changed  boolean  Whether the host changed

Examples (1):
  EXAMPLE           INPUT                        OUTPUT
  Lower swappiness  key=vm.swappiness, value=10  key=vm.swappiness, changed=true
```

`--json` prints the same as one object, with `uses` and `used_by` as lists of
names.

## Flags

| Flag       | Description                                                      |
| ---------- | ---------------------------------------------------------------- |
| `--name`   | The behavior, from `osapi behavior list` (required)              |
| `--method` | One method to show in full: its fields, constraints and examples |
