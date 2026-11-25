# Create

Make a join token, and print it with the join string.

The token admits an agent at its join, until it expires or its uses are gone.
The controller keeps a hash of it, not the token: this command prints the token
one time.

By default the agent joins over NATS, and needs to reach NATS only. The join
string has the address of NATS (`controller.join.url`, or else
`nats.server.url`), the bootstrap user, which can only send the join request,
and for a `tls://` server the fingerprint of its certificate, which the agent
pins. The controller gives the agent its own NATS user at the join.

With `--transport https`, the agent joins over the controller's API: the join
string has `controller.api.url` and the fingerprint of the API's certificate. It
needs `controller.api.tls.enabled`.

```bash
osapi controller join token create [flags]
```

## Examples

```bash
osapi controller join token create
osapi controller join token create --expires 24h --uses 10 --labels group=web
osapi controller join token create --transport https
```

## Result

The command prints:

```ansi
$ osapi controller join token create
[1;38;2;0;213;255mJoin token:[0m
  [38;2;10;143;176mID:[0m           1a2b3c4d
  [38;2;10;143;176mToken:[0m        osapi_jt_1a2b3c4d.vC3m9...
  [38;2;10;143;176mJoin string:[0m  tls://nats.example.com:4222?fp=SHA256%3A9f86d08...&ns=osapi&pass=...&token=osapi_jt_1a2b3c4d.vC3m9...&user=join
  [38;2;10;143;176mExpires:[0m      2026-11-25 13:00:00
  [38;2;10;143;176mUses:[0m         1
```

With `--json`, the command prints:

```json
{
  "id": "1a2b3c4d",
  "token": "osapi_jt_1a2b3c4d.vC3m9...",
  "join": "tls://nats.example.com:4222?fp=SHA256%3A9f86d08...\u0026ns=osapi\u0026pass=...\u0026token=osapi_jt_1a2b3c4d.vC3m9...\u0026user=join",
  "expires_at": "2026-11-25T13:00:00Z",
  "uses": 1
}
```

## Flags

| Flag          | Description                                                                          | Default |
| ------------- | ------------------------------------------------------------------------------------ | ------- |
| `--expires`   | How long the token admits agents, as a Go duration such as 1h                        | `1h`    |
| `--labels`    | Labels that each agent that joins with the token gets, as key=value                  |         |
| `--transport` | How the agent joins: nats, over the NATS server, or https, over the controller's API | `nats`  |
| `--uses`      | How many agents the token admits                                                     | `1`     |

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
