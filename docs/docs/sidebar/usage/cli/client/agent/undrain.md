# Undrain

Undrain an agent.

The agent subscribes to the NATS JetStream consumers again and changes back to
`Ready`.

```bash
osapi client agent undrain [flags]
```

## Examples

```bash
osapi client agent undrain --hostname web-01
```

## Result

The command prints:

```ansi
$ osapi client agent undrain --hostname web-01
[ok] web-01 ready

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

With `--json`, the command prints:

```json
{ "hostname": "web-01", "result": "ready" }
```

## Permission

It needs the `agent:write` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Flags

| Flag         | Description                                     | Default |
| ------------ | ----------------------------------------------- | ------- |
| `--hostname` | The hostname of the agent to undrain (required) |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
