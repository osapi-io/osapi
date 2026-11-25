# List

List the join tokens, with their expiry and uses left.

```bash
osapi controller join token list
```

## Examples

```bash
osapi controller join token list
osapi controller join token list --json
```

## Result

The command prints:

```ansi
$ osapi controller join token list
[1;38;2;0;213;255mJoin tokens (1):[0m
  [38;2;10;143;176mID        [0m[38;2;10;143;176mEXPIRES              [0m[38;2;10;143;176mUSES LEFT  [0m[38;2;10;143;176mLABELS[0m
  [38;2;138;138;138m1a2b3c4d[0m  2026-11-25 13:00:00  1          group:web
```

With `--json`, the command prints:

```json
{
  "items": [
    {
      "id": "1a2b3c4d",
      "created_at": "2026-11-25T12:00:00Z",
      "expires_at": "2026-11-25T13:00:00Z",
      "uses_left": 1,
      "labels": {
        "group": "web"
      }
    }
  ]
}
```

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
