# Accept

Accept a pending agent's PKI enrollment request:

```bash
$ osapi client agent accept --hostname web-01

  Status: Accepted
  Message: Agent web-01 enrollment accepted
```

When PKI is enabled, a new agent sends an enrollment request. An administrator
must accept it before the agent can do work. This command accepts a pending
agent by hostname. To accept by the public key fingerprint of the agent, add
`--fingerprint`.

Use `agent list --pending` to see the agents that wait for approval.

## Flags

| Flag            | Description                                   | Required |
| --------------- | --------------------------------------------- | -------- |
| `--hostname`    | Hostname of the pending agent to accept       | Yes      |
| `--fingerprint` | Accept by key fingerprint instead of hostname | No       |
| `--json`        | Output raw JSON                               | No       |

## Examples

```bash
# Accept by hostname
osapi client agent accept --hostname web-01

# Accept by fingerprint
osapi client agent accept --hostname web-01 \
  --fingerprint SHA256:ab12cd34...
```
