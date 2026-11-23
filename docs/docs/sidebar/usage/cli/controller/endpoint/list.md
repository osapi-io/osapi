# List

List the endpoint groups this binary has.

```bash
osapi controller endpoint list
```

## Examples

```bash
osapi controller endpoint list
osapi controller endpoint list --json
```

## Result

The command prints:

```ansi
$ osapi controller endpoint list
[1;38;2;0;213;255mEndpoint groups (2):[0m
  [38;2;10;143;176mNAME   [0m[38;2;10;143;176mMETHODS                            [0m[38;2;10;143;176mDESCRIPTION[0m
  [1magent[0m  status, list, get, drain, undrain  Manage the agents in the fleet: discovery, drain and PKI enrollment
  [1maudit[0m  list, export, get                  Read the audit log
```

With `--json`, the command prints:

```json
{
  "items": [
    {
      "name": "agent",
      "methods": ["status", "list", "get", "drain", "undrain"],
      "doc": "Manage the agents in the fleet: discovery, drain and PKI enrollment"
    },
    {
      "name": "audit",
      "methods": ["list", "export", "get"],
      "doc": "Read the audit log"
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

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
