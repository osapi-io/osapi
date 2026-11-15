# Drain

Drain an agent.

The agent changes from `Ready` to `Draining`. When all its running jobs are
complete, the state is `Cordoned`. The agent continues to run and to send
heartbeats. It only stops the intake of new work from the job queue. Use
`agent undrain` to accept jobs again.

```bash
osapi client agent drain [flags]
```

## Examples

```bash
osapi client agent drain --hostname web-01
```

## Flags

| Flag         | Description                               | Default |
| ------------ | ----------------------------------------- | ------- |
| `--hostname` | Hostname of the agent to drain (required) |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
