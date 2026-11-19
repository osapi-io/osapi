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

## Permission

It needs the `agent:write` permission. `osapi token roles list` lists the roles
that grant it.

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
