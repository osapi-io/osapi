---
sidebar_position: 3
---

# Working with an agent

osapi is built to be driven by an AI agent such as Claude Code or Codex. The
agent never logs in to a host. It runs osapi's CLI, which turns each request
into a typed operation the controller checks before any host sees it.

## Set it up

1. Run osapi and make a token, as [Get started](../intro.md#get-started) shows.
2. In the repository the agent works in, run [`osapi init`](cli/init.md). It
   installs two skills and points the tool's instructions file at them.
3. Start the agent in that repository, with the token in its environment as
   `OSAPI_API_CLIENT_SECURITY_BEARER_TOKEN`.

| Skill              | The agent uses it to                                                                                 |
| ------------------ | ---------------------------------------------------------------------------------------------------- |
| `osapi`            | Read and change hosts: kernel parameters, packages, services, users, files, DNS, NTP and the rest    |
| `osapi-operations` | Run osapi itself: start it, check its configuration, make and check tokens, find why it is unhealthy |

Run `osapi init` again after upgrading osapi, so the skills match the binary.

## What to ask

Ask for the outcome, not the commands:

- _"Set vm.swappiness to 10 on every host in group:web and tell me which ones
  changed."_
- _"Which hosts have package updates waiting?"_
- _"Point eth0 on web-01 at 1.1.1.1 for DNS and show me before and after."_
- _"Find every host under disk or memory pressure."_
- _"Why is the agent on db-02 not picking up jobs?"_

## How the agent knows what to run

It does not guess. Before it runs an operation it reads its contract from the
binary:

```bash
osapi behavior list --json                                 # every behavior and its methods
osapi behavior get --name sysctl --method update --json    # one method's input and output schema
osapi help --json                                          # every command, flag and error shape
```

Then it runs the command with `--json` and reads each host's result: `ok` with
whether the host `changed`, `failed` with the error, or `skipped` where the
operation does not run on that platform. A write run again reports
`changed: false`, so the agent can tell a change from a no-op.

## What it can do is what the token allows

The agent can do exactly what its token's roles and permissions allow, and the
controller checks every request. Give it the least it needs:

```bash
osapi token generate -r read -u agent@example.com                        # look, never change
osapi token generate -p sysctl:read -p sysctl:write -u agent@example.com # one area only
```

`command:execute` and `command:shell` let a token run commands it supplies,
through the `command` domain. Leave them out of an agent's token to keep it to
typed operations. Every request, the agent's included, is in the
[audit log](cli/client/audit/audit.mdx).

## What it will not do

There is no shell behind the skills. When osapi has no operation for something,
the agent says so rather than working around it. Adding one is a behavior to
build:
[Building a behavior with an agent](../development/building-with-an-agent.md).
