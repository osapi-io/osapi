# Update

Set the labels, the max jobs and the log level of an agent.

The labels replace the labels of the agent's config file with the same keys. Max
jobs replaces its max_jobs, and the log level replaces the level of its debug
setting; with neither, the agent keeps the file's. The agent applies them at
once, with no restart. A host that joins with a join token gets the labels of
the token.

```bash
osapi client agent config update [flags]
```

## Examples

```bash
osapi client agent config update --hostname web-01 --labels group=web --labels dc=east --max-jobs 4 --log-level debug
```

## Result

The command prints:

```ansi
$ osapi client agent config update --hostname web-01 --labels group=web --labels dc=east --max-jobs 4 --log-level debug
[1;38;2;0;213;255mAgent config:[0m

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
  [38;2;10;143;176mVersion:[0m    2
  [38;2;10;143;176mLabels:[0m     dc:east, group:web
  [38;2;10;143;176mMax jobs:[0m   4
  [38;2;10;143;176mLog level:[0m  debug
  [38;2;10;143;176mUpdated:[0m    2026-11-25 12:00:00
```

With `--json`, the API's response:

```json
{
  "version": 2,
  "labels": {
    "dc": "east",
    "group": "web"
  },
  "max_jobs": 4,
  "log_level": "debug",
  "updated_at": "2026-11-25T12:00:00Z"
}
```

## Permission

It needs the `agent:write` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Flags

| Flag          | Description                                                                                                                               | Default |
| ------------- | ----------------------------------------------------------------------------------------------------------------------------------------- | ------- |
| `--hostname`  | The hostname of the agent (required)                                                                                                      |         |
| `--labels`    | The labels of the agent, as key=value. They replace the labels of the agent's config file with the same keys. Empty removes each          |         |
| `--log-level` | The level that the agent logs at: debug, info, warn or error. Empty keeps the level of the agent's config file (debug, info, warn, error) |         |
| `--max-jobs`  | The most jobs that the agent runs at the same time. It replaces max_jobs of the agent's config file; empty keeps the file's               |         |

## Global flags

| Flag               | Description                                                                      | Default                  |
| ------------------ | -------------------------------------------------------------------------------- | ------------------------ |
| `-d, --debug`      | Enable or disable debug mode                                                     |                          |
| `-j, --json`       | Enable JSON output                                                               |                          |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml`  |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `https://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
