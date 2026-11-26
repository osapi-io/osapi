---
name: osapi-operations
description: >-
  Run osapi itself: start the NATS server, controller and agent together or
  one at a time, read and check its configuration, create and check tokens,
  and find out why it is unhealthy. Use when asked to set up, start, stop,
  configure, upgrade or troubleshoot osapi, or when a command fails because
  osapi is not reachable or a token is refused. To change hosts through osapi, use the osapi skill.
---

# osapi operations

osapi has three components in one binary:

- an embedded NATS server that holds the job queue
- the controller, which serves the REST API
- an agent on each host, which runs the jobs

This skill runs these components. To change a host through them, use the osapi
skill.

## Rules

1. **Ask the binary, not memory.** `osapi help <command> --json` gives every
   command's flags. `osapi config list --json` gives every setting with its
   value, source, type, default, environment variable and description.
   `osapi dsl endpoint get --name <group> --json` gives the routes and
   permissions of the controller's own operations, such as the agent group. They
   all change between releases.
2. **Check the config before you start osapi.** Run `osapi config validate`
   after any change to the file or the environment. It runs the same checks
   `start` does and starts nothing.
3. **Read the config in effect, not the file.** An `OSAPI_` environment variable
   wins over the file. `osapi config list --json` shows each value and whether
   it came from the environment, the file or a default.
4. **Never print a secret.** `config list` and `config get` mask passwords, keys
   and tokens. Do not read them out of the file or the environment to show
   anyone.
5. **Pass `--json`** to every command, and match on a failure's `code`.

## Route

| Need                                               | Read                                                     |
| -------------------------------------------------- | -------------------------------------------------------- |
| Starting everything, or one component              | [references/components.md](references/components.md)     |
| What a setting does, its value, a bad config       | [references/config.md](references/config.md)             |
| Creating or checking a token, a refused request    | [references/tokens.md](references/tokens.md)             |
| A certificate, its expiry or algorithm, a rotation | [references/certificates.md](references/certificates.md) |
| Is it healthy, why a host is missing, a stuck job  | [references/health.md](references/health.md)             |
| Joining a host, its service, removing a host       | [references/join.md](references/join.md)                 |
| -------------------------------------------------  | ----------------------------------------------------     |
| Setting up NATS, the controller, a config          | [references/setup.md](references/setup.md)               |
| The agents, the audit log, files, secrets, traces  | [references/fleet.md](references/fleet.md)               |
