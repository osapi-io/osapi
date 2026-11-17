# Accept

Accept a pending agent enrollment.

When PKI is enabled, a new agent sends an enrollment request. An administrator
must accept it before the agent can do work. This command accepts a pending
agent by hostname. To accept by the public key fingerprint of the agent, add
`--fingerprint`. Use `agent pending` to see the agents that wait for approval.

```bash
osapi client agent accept [flags]
```

## Examples

```bash
osapi client agent accept --hostname web-01
osapi client agent accept --hostname web-01 --fingerprint SHA256:ab12cd34...
```

## Permission

It needs the `agent:write` permission. `osapi token roles` lists the roles that
grant it.

## Flags

| Flag            | Description                                             | Default |
| --------------- | ------------------------------------------------------- | ------- |
| `--fingerprint` | Accept by fingerprint instead of hostname.              |         |
| `--hostname`    | The hostname of the pending agent to accept. (required) |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
