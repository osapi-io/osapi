---
name: osapi-operations
description: >
  Run osapi itself: start the NATS server, controller and agent together or
  one at a time, read and check its configuration, create and check tokens,
  and find out why it is unhealthy. Use when asked to set up, start, stop,
  configure, upgrade or troubleshoot osapi, or when a command fails because
  osapi is not reachable or a token is refused. For changing hosts through
  osapi, use the osapi skill instead.
---

# osapi operations

osapi is three components in one binary: an embedded NATS server that holds the
job queue, the controller that serves the REST API, and an agent on each host
that runs jobs. This skill is for running them. Changing a host through them is
the osapi skill.

## Rules

1. **Ask the binary, not memory.** `osapi help <command> --json` gives every
   command's flags. `osapi config list --json` gives every setting with its
   value, source, type, default, environment variable and description. Both
   change between releases.
2. **Check the config before starting.** Run `osapi config validate` after any
   change to the file or the environment. It runs the same checks `start` does
   and starts nothing.
3. **Read the config in effect, not the file.** An `OSAPI_` environment variable
   wins over the file. `osapi config list --json` shows each value and whether
   it came from the environment, the file or a default.
4. **Never print a secret.** `config list` and `config get` mask passwords,
   signing keys and tokens. Do not read them out of the file or the environment
   to show anyone.
5. **Pass `--json`** to every command, and match on a failure's `code`.

## Route

| Need                                             | Read                                                 |
| ------------------------------------------------ | ---------------------------------------------------- |
| Starting everything, or one component            | [references/components.md](references/components.md) |
| What a setting does, its value, a bad config     | [references/config.md](references/config.md)         |
| Creating or checking a token, a refused request  | [references/tokens.md](references/tokens.md)         |
| Is it healthy, why a host is missing, a stuck job | [references/health.md](references/health.md)        |
