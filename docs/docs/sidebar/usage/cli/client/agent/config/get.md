# Get

Read the config of an agent.

Shows the labels that the controller gives the agent, and the version of the
config. The agent applies each version at once.

```bash
osapi client agent config get [flags]
```

## Examples

```bash
osapi client agent config get --hostname web-01
```

## Result

The command prints:

```ansi
$ osapi client agent config get --hostname web-01
[1;38;2;0;213;255mAgent config:[0m

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
  [38;2;10;143;176mVersion:[0m   2
  [38;2;10;143;176mLabels:[0m    dc:east, group:web
  [38;2;10;143;176mMax jobs:[0m  4
  [38;2;10;143;176mUpdated:[0m   2026-11-25 12:00:00
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
  "updated_at": "2026-11-25T12:00:00Z"
}
```

## Permission

It needs the `agent:read` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Flags

| Flag         | Description                          | Default |
| ------------ | ------------------------------------ | ------- |
| `--hostname` | The hostname of the agent (required) |         |

## Global flags

| Flag               | Description                                                                      | Default                  |
| ------------------ | -------------------------------------------------------------------------------- | ------------------------ |
| `-d, --debug`      | Enable or disable debug mode                                                     |                          |
| `-j, --json`       | Enable JSON output                                                               |                          |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml`  |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `https://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
