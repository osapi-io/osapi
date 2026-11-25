# List

List the vaults of the controller.

The list shows each vault of controller.secrets.vaults. A vault that does not
answer is a backend that the controller cannot reach now.

```bash
osapi client secret vault list
```

## Examples

```bash
osapi client secret vault list
osapi client secret vault list --json
```

## Result

The command prints:

```ansi
$ osapi client secret vault list
[1;38;2;0;213;255mVaults (2):[0m
  [38;2;10;143;176mNAME     [0m[38;2;10;143;176mTYPE  [0m[38;2;10;143;176mTARGETS     [0m[38;2;10;143;176mANSWERS[0m
  [1mdefault[0m  [1mnats[0m              yes
  [1mprod[0m     [1mnats[0m  group:prod  yes

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

With `--json`, the API's response:

```json
{
  "items": [
    {
      "name": "default",
      "type": "nats",
      "targets": [],
      "answers": true
    },
    {
      "name": "prod",
      "type": "nats",
      "targets": ["group:prod"],
      "answers": true
    }
  ],
  "total_items": 2
}
```

## Permission

It needs the `secret:read` permission. `osapi controller role get --name <role>`
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
