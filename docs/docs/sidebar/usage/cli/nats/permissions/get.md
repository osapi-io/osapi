# Get

Print the NATS subjects that one osapi process needs.

It prints the least that the role can publish and subscribe to, with the streams
and buckets of the config and the namespace of `controller.nats`. Give these to
the NATS user of the role on a NATS server that osapi does not run. `nsc` is the
flags of `nsc add user` or `nsc edit user`, for a server in operator mode. The
embedded server applies them itself to a user with a `role` in
`nats.server.auth.users`.

```bash
osapi nats permissions get [flags]
```

## Examples

```bash
osapi nats permissions get --role agent
osapi nats permissions get --role controller --json
```

## Result

The command prints:

```ansi
$ osapi nats permissions get --role agent
[1;38;2;0;213;255mNATS permissions:[0m
  [38;2;10;143;176mRole:[0m       agent
  [38;2;10;143;176mNamespace:[0m  osapi
  [38;2;10;143;176mPublish:[0m    $JS.API.INFO, $KV.osapi-agent-registry.>, osapi.enroll.request
  [38;2;10;143;176mSubscribe:[0m  _INBOX.>, osapi.enroll.response.>
  [38;2;10;143;176mnsc:[0m        --allow-pub "$JS.API.INFO,$KV.osapi-agent-registry.>,osapi.enroll.request" --allow-sub "_INBOX.>,osapi.enroll.response.>"
```

With `--json`, the command prints:

```json
{
  "role": "agent",
  "namespace": "osapi",
  "publish": [
    "$JS.API.INFO",
    "$KV.osapi-agent-registry.\u003e",
    "osapi.enroll.request"
  ],
  "subscribe": ["_INBOX.\u003e", "osapi.enroll.response.\u003e"],
  "nsc": "--allow-pub \"$JS.API.INFO,$KV.osapi-agent-registry.\u003e,osapi.enroll.request\" --allow-sub \"_INBOX.\u003e,osapi.enroll.response.\u003e\""
}
```

## Flags

| Flag     | Description                                              | Default |
| -------- | -------------------------------------------------------- | ------- |
| `--role` | The osapi process: controller, agent or setup (required) |         |

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
