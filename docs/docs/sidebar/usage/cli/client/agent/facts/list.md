# List

List the fact keys.

Lists each fact key that an input can refer to as @fact.&lt;key&gt;, with its
description.

```bash
osapi client agent facts list
```

## Examples

```bash
osapi client agent facts list
osapi client agent facts list --json
```

## Result

The command prints:

```ansi
$ osapi client agent facts list
[1;38;2;0;213;255mFact keys (3):[0m
  [38;2;10;143;176mREFERENCE                [0m[38;2;10;143;176mDESCRIPTION[0m
  @fact.hostname           The hostname of the agent.
  @fact.primary_interface  The interface of the default route.
  @fact.datacenter         A custom fact from the config of the agent.

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

With `--json`, the API's response:

```json
{
  "items": [
    {
      "key": "hostname",
      "description": "The hostname of the agent.",
      "builtin": true
    },
    {
      "key": "primary_interface",
      "description": "The interface of the default route.",
      "builtin": true
    },
    {
      "key": "datacenter",
      "description": "A custom fact from the config of the agent."
    }
  ],
  "total_items": 3
}
```

## Permission

It needs the `agent:read` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Global flags

| Flag               | Description                                                 | Default                 |
| ------------------ | ----------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                |                         |
| `-j, --json`       | Enable JSON output                                          |                         |
| `-f, --osapi-file` | Path to config file                                         | `/etc/osapi/osapi.yaml` |
| `--url`            | Base URL the CLI client connects to (controller.client.url) | `http://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
