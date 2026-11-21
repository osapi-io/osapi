---
sidebar_position: 2
title: Core concepts
---

# Core concepts

These are the parts of osapi. Each section tells what the part is and where to
read more. The last sections tell what is not built yet.

## Behavior

A behavior is one thing that osapi manages on a host, for example kernel
parameters. Somebody defines it one time, as typed Go. osapi makes these parts
from the definition:

- the API route
- the CLI command
- the permission
- the docs page
- the tests

The examples of a behavior are also its tests. Thus the docs page shows what
really runs.

```bash
osapi behavior list                  # each behavior in this build
osapi behavior get --name sysctl     # its methods, inputs, platforms and permissions
```

Refer to [Behaviors](reference/behaviors/behaviors.mdx).

## Method

A behavior has methods. A method is a read or a write:

- A **read** changes nothing.
- A **write** reports `changed: true` when it changed the host. When you run it
  again with the same input, it reports `changed: false`.

A method has a typed input and a typed output. The agent reads them with
`osapi behavior get --json` before it makes a request.

## Building block

A building block is an internal behavior. Only other behaviors use it, on the
same host and in the same job. It has no route and no command.

For example, sysctl requires `managedfile`. `managedfile` writes files and keeps
a record of each file. Thus sysctl finds drift, and has no code of its own for
it.

## Job

osapi does not call a host. It puts a job on a queue. The agent on each target
host takes the job, runs it and writes the result. Thus a slow host does not
block the others. osapi keeps each job, and the audit log records each request.

Refer to [Job system](features/job-system.md).

## Target

`--target` tells which hosts get a job:

| Target      | Means                     |
| ----------- | ------------------------- |
| a hostname  | that host                 |
| `_any`      | one host, any host        |
| `_all`      | each host                 |
| `group:web` | each host with that label |

Each host answers for itself: `ok`, `failed` or `skipped`. A host gives
`skipped` when the behavior does not run on its platform.

## Controller and agent

The controller serves the API and puts jobs on the queue. An agent runs on each
host and does the work. `osapi start` runs the queue, the controller and an
agent in one process, for a test.

Each agent has an identity. With PKI on, the controller signs each job, and an
agent signs each answer. Refer to [Agent identity](features/agent-identity.md).

## Skill

A skill tells an agent how to use osapi. `osapi init` installs two skills:

- `osapi`, which reads and changes hosts
- `osapi-operations`, which runs osapi itself

The skills tell the agent to ask the binary what exists. They do not keep a copy
of the docs. Refer to [Working with an agent](usage/agents.md).

## Least privilege

The goal is that the agent never runs as root. Each behavior declares the steps
that need root. An admin installs a sudoers file for each behavior that the host
needs. The agent never writes its own grants.

Parts of this are built. Command exec and docker exec are off until you set
`agent.command.enabled: true`. The other parts are in
[the plan](https://github.com/osapi-io/specs/blob/main/components/osapi/least-privilege.md).

## Not built yet: workflows

A workflow is a named, ordered set of behavior calls across hosts, in a file.
osapi checks the file before it runs: each step names a real method, and each
input passes the checks of that method. A step can run only where an earlier
step changed something. A second run of the same file reports `changed: false`.

For example: set a kernel parameter on the web hosts, then restart the service
that reads it, but only on the hosts where the parameter changed.

Refer to
[the workflows design](https://github.com/osapi-io/specs/blob/main/components/osapi/workflows.md).

## Not built yet: behaviors on one system

Now each behavior runs on a managed host. A behavior can also run on one system,
for example the controller. Such a behavior can call a remote API, for example
the AWS SDK. A workflow can then use its output: get the instances from AWS,
then configure each one.

## Not built yet: scheduled workflows

A workflow file can declare a schedule, for example a cron expression. osapi
then runs the workflow at those times.
