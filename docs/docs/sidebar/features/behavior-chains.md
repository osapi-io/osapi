---
sidebar_position: 28
---

# Behavior chains

Some behaviors use other behaviors to do their work. For example, `dns update`
writes a netplan file. It does not write the file itself. It calls `netplan`,
and `netplan` calls `managedfile`. This set of calls is the chain of the
behavior:

```text
dns → netplan → managedfile
```

This page tells what a chain is, what you see when a chain runs, and what occurs
when a step of a chain fails.

## One request, one job

A chain runs in one job, on one host. You send one API request, for example
`dns update`. The controller makes one job. The agent on the host runs
`dns update`, and `dns update` calls `netplan` and `managedfile` as Go calls on
the same host. There is no second API request and no second job.

Thus:

- You need only the permission of the behavior that you call. `dns update` needs
  `dns:write`. `netplan` and `managedfile` are internal behaviors. They have no
  route, no CLI command and no permission of their own.
- The result of the request is the result of `dns update`. The job keeps a
  record of each call that `dns update` made.
- A job to more than one host, such as `--target _all`, runs the chain on each
  host. Each host has its own chain record.

`osapi behavior chain list` shows each chain in this build.
[Chains](../reference/chains/chains.mdx) shows each chain, with each call that
its examples make.

## What you see

### The job

The CLI prints the job ID and the trace ID under the result:

```text
  1 host: 1 changed   job 550e8400-e29b-41d4-a716-446655440000   trace 26e448c9dbe1c4a8d216a98c675079c9
```

`osapi client job get --id <job>` shows the chain of each host on one line, in
the `Chain` field:

```text
  Chain:     web-01: (netplan get, netplan update (changed) → managedfile update (changed))
```

A call that changed the host has `(changed)`. A call that failed has `(failed)`.
The calls of one step that called more than one behavior are in parentheses.
`osapi client job list` shows the same line in its `CHAIN` column.

With `--json`, each item of `responses` has `steps`, the tree of the calls on
that host. Each step has the behavior, the method, `changed`, the error of a
failed step, the duration in milliseconds and the calls that it made:

```json
"steps": [
  { "behavior": "netplan", "method": "get", "duration_ms": 12 },
  {
    "behavior": "netplan",
    "method": "update",
    "changed": true,
    "duration_ms": 840,
    "steps": [
      { "behavior": "managedfile", "method": "update", "changed": true, "duration_ms": 3 }
    ]
  }
]
```

A step does not record its input or its output. The request and the result of
the job have the input and the output of the behavior that you called.

### The UI

Open the job in the job view of the
[management dashboard](management-dashboard.md). The **Chain** section shows the
tree of each host. Each row is one call, as `behavior.method`, under the call
that made it, with `changed`, or `failed:` and the error, and its duration.

### The trace

When [tracing](distributed-tracing.md) is on, each call of a chain is a span of
its own, with the name `behavior.<name>.<method>`. The span of a call is under
the span of the call that made it. The commands and the file changes of a call
are under its span:

```text
job.process
└─ behavior.dns.update
   ├─ behavior.netplan.get
   │  └─ exec netplan
   └─ behavior.netplan.update
      ├─ behavior.managedfile.update
      │  ├─ file.write
      │  └─ kv.put
      ├─ exec netplan
      └─ exec netplan
```

The two `exec netplan` spans under `behavior.netplan.update` are
`netplan generate` and `netplan apply`. A span does not record the arguments of
a command. `osapi client telemetry trace get --id <trace>` prints the tree. A
span of a step that failed has the error status.

### The audit log

The [audit entry](audit-logging.md) of the request has the `job_id` and the
`trace_id`. It does not have the chain. To see the chain of an audit entry, read
its job with `osapi client job get --id <job_id>`.
[osapi#1193](https://github.com/osapi-io/osapi/issues/1193) tracks the chain in
the audit entry.

## When a step fails

A chain does not roll back. A step that fails gives its error to the step that
called it. That step returns the error, and the job fails on that host with the
message. The writes that ran before the failure stay on the host.

The step tree records the failure. The agent records the tree also when the job
fails, so the failed step and the writes before it are in the job.

These are the failures of `dns update`:

| Failure                                 | The host after the failure                                                                                      | The job error                                                                                         |
| --------------------------------------- | --------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- |
| `managedfile update` cannot write       | No change. The write is atomic.                                                                                 | `write file: ...`                                                                                     |
| `netplan generate` refuses the new file | No change. `netplan` puts back the content that the file had, or removes a new file, with a `managedfile` call. | `netplan generate refused <path>, so its content was put back: ...`, or `..., so it was removed: ...` |
| `netplan apply` fails                   | The new file stays in `/etc/netplan`. netplan did not apply it.                                                 | `netplan apply: ...`                                                                                  |

Only a building block undoes its own change, as `netplan` does when
`netplan generate` refuses a file. The caller of a step does not undo the work
of the step.

For the last failure, the chain line of the job is:

```text
  Chain:     web-01: (netplan get, netplan update (failed) → managedfile update (changed))
```

`netplan update (failed)` is the step that failed.
`managedfile update (changed)` under it is the write that stayed on the host.

A retry with the same input does not apply the file now. `managedfile update`
finds the same content on disk and reports no change, so `netplan` does not run
`netplan apply` again, and the job reports unchanged.
[osapi#1194](https://github.com/osapi-io/osapi/issues/1194) tracks this. Until
it is fixed, correct the cause and run `netplan apply` on the host.

## What a chain is not

A chain is for work on one host, in one job. Work across hosts, or work in steps
where you look at the result of one step before you start the next, is a
workflow of
[osapi-orchestrator](https://github.com/osapi-io/osapi-orchestrator).

Only the author of a behavior makes a chain. The author writes the behavior in
Go and declares the behaviors that it uses with `behavior.Requires`. A user
cannot put behaviors together into a chain without Go. For multi-step work, a
user writes an osapi-orchestrator workflow. This is the current rule.
[Extensions](https://github.com/osapi-io/osapi/issues/1064) can change it.

The [Chains DSL](../development/dsl/chain.md) tells an author how to build a
chain.
