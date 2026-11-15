# Reject

Reject a pending agent enrollment.

When PKI is enabled, a new agent sends an enrollment request. An administrator
accepts or rejects it. This command rejects a pending agent, so the agent cannot
join the fleet. The command removes the enrollment entry of the agent from the
pending queue. Use `agent list --pending` to see the agents that wait for
approval.

```bash
osapi client agent reject [flags]
```

## Examples

```bash
osapi client agent reject --hostname web-01
```

## Flags

| Flag         | Description                                        | Default |
| ------------ | -------------------------------------------------- | ------- |
| `--hostname` | Hostname of the pending agent to reject (required) |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
