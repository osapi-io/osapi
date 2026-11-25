# Create

Make a join token, and print it with the join string.

The token admits an agent at its join, until it expires or its uses are gone.
The controller keeps a hash of it, not the token: this command prints the token
one time. The join string has the controller's URL (`controller.api.url`) and
the fingerprint of its certificate, so an agent checks the controller before it
sends anything. It needs `controller.api.tls.enabled`.

```bash
osapi controller join token create [flags]
```

## Examples

```bash
osapi controller join token create
osapi controller join token create --expires 24h --uses 10 --labels group=web
```

## Result

The command prints:

```ansi
$ osapi controller join token create
[1;38;2;0;213;255mJoin token:[0m
  [38;2;10;143;176mID:[0m           1a2b3c4d
  [38;2;10;143;176mToken:[0m        osapi_jt_1a2b3c4d.vC3m9...
  [38;2;10;143;176mJoin string:[0m  osapi://osapi.example.com:8080?fp=SHA256%3A9f86d08...&token=osapi_jt_1a2b3c4d.vC3m9...
  [38;2;10;143;176mExpires:[0m      2026-11-25 13:00:00
  [38;2;10;143;176mUses:[0m         1
```

With `--json`, the command prints:

```json
{
  "id": "1a2b3c4d",
  "token": "osapi_jt_1a2b3c4d.vC3m9...",
  "join": "osapi://osapi.example.com:8080?fp=SHA256%3A9f86d08...\u0026token=osapi_jt_1a2b3c4d.vC3m9...",
  "expires_at": "2026-11-25T13:00:00Z",
  "uses": 1
}
```

## Flags

| Flag        | Description                                                         | Default |
| ----------- | ------------------------------------------------------------------- | ------- |
| `--expires` | How long the token admits agents, as a Go duration such as 1h       | `1h`    |
| `--labels`  | Labels that each agent that joins with the token gets, as key=value |         |
| `--uses`    | How many agents the token admits                                    | `1`     |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
