# List

List each command of this binary, and the layer that made it.

The layer is `command` for a command of a local group, `endpoint` for a command
of an endpoint group, `behavior` for a command of a behavior, and `hand-written`
for a command in `cmd/` that no definition makes yet.

```bash
osapi cli command list
```

## Examples

```bash
osapi cli command list
osapi cli command list --json
```

## Result

The command prints:

```ansi
$ osapi cli command list
[1;38;2;0;213;255mCommands (4):[0m
  [38;2;10;143;176mCOMMAND                         [0m[38;2;10;143;176mLAYER     [0m[38;2;10;143;176mDESCRIPTION[0m
  osapi token generate            command   Generate a new token
  osapi token validate            command   Validate a token for authenticity and claims
  osapi client agent list         endpoint  List active agents
  osapi client node hostname get  behavior  Get the hostname
  [38;2;10;143;176mCommand:[0m       2
  [38;2;10;143;176mEndpoint:[0m      1
  [38;2;10;143;176mBehavior:[0m      1
  [38;2;10;143;176mHand written:[0m  0
```

With `--json`, the command prints:

```json
{
  "items": [
    {
      "command": "osapi token generate",
      "layer": "command",
      "group": "token",
      "doc": "Generate a new token"
    },
    {
      "command": "osapi token validate",
      "layer": "command",
      "group": "token",
      "doc": "Validate a token for authenticity and claims"
    },
    {
      "command": "osapi client agent list",
      "layer": "endpoint",
      "group": "agent",
      "doc": "List active agents"
    },
    {
      "command": "osapi client node hostname get",
      "layer": "behavior",
      "group": "hostname",
      "doc": "Get the hostname"
    }
  ],
  "command": 2,
  "endpoint": 1,
  "behavior": 1,
  "hand_written": 0
}
```

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
