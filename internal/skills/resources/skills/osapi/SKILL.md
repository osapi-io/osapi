---
name: osapi
description: >-
  Operate Linux hosts through osapi: read and change kernel parameters,
  packages, services, users, files, DNS, NTP and the rest on one host, a group,
  or the whole fleet. Use when asked to inspect or configure a host osapi
  manages, to run an osapi command, or to find out what osapi can do. Do not use
  to change the code of osapi.
---

# osapi

osapi manages Linux hosts through typed operations. Each change is a job a
host's agent runs, checked against a schema before it is queued, safe to run
twice, and recorded. Work through the `osapi` CLI, never by logging in to a
host.

## Rules

1. **Ask the binary, not memory.** Before running a command, run
   `osapi help <command...> --json` for its exact flags. Before building a
   request, run
   `osapi dsl behavior method get --name <behavior> --method <method> --json`
   for its input and output schema. For an operation on osapi itself, such as
   the agents or the jobs, run `osapi dsl endpoint get --name <group> --json`.
   Flags and fields change between releases.
2. **Always pass `--json`.** Output is the API's response body. On failure,
   stderr is one document, `{"error", "code"}`. Match on `code`, never on the
   message. Exit 2 means that the call was incorrect. Fix the call. Do not try
   it again unchanged.
3. **Read before you change.** Before an update or a delete, run the matching
   `get` or `list` on the same target. Make sure that the host is in the state
   that you expect.
4. **Check every host's result.** A command against many hosts returns one
   result per host, each with its own `status`: `ok`, `failed`, `skipped` or
   `timeout`. A `skipped` host does not support the operation; it is not an
   error and not a success.
5. **Give a secret as a reference.** A secret field, such as a password or the
   `--env` of a container, takes `secret://<vault>/<key>`. Put the value in a
   vault first, from standard input:
   `osapi client secret update --vault default --key registry-token --value - < token.txt`.
   Then give `secret://default/registry-token`.
   `osapi client secret list --vault default --json` gives the keys. Never put a
   value in a flag, a plan or a file: the shell history and the disk keep it.

## Route

| Need                                                        | Read                                                       |
| ----------------------------------------------------------- | ---------------------------------------------------------- |
| What osapi can do, and the shape of a request               | [references/behaviors.md](references/behaviors.md)         |
| The agents, the jobs, the audit log: osapi's own operations | [references/endpoints.md](references/endpoints.md)         |
| Choosing hosts: one, any, all, or a label                   | [references/targets.md](references/targets.md)             |
| A command failed                                            | [references/errors.md](references/errors.md)               |
| What a host's facts are read with: gohai's collectors       | `osapi agent collector list --json`, then `get --name <c>` |
| Calling osapi from a Go program, not the CLI                | the `osapi-sdk` skill, and `osapi sdk spec get --json`     |
