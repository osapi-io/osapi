# Accept

Accept a pending agent's PKI enrollment request:

```bash
$ osapi client agent accept --hostname web-01

  Status: Accepted
  Message: Agent web-01 enrollment accepted
```

When PKI is enabled, new agents submit enrollment requests that must be
accepted before the agent can participate in the fleet. This command
approves a pending agent by hostname. Optionally, pass `--fingerprint`
to accept by the agent's public key fingerprint instead.

Use `agent list --pending` to see agents awaiting approval.

## Flags

| Flag            | Description                                  | Required |
| --------------- | -------------------------------------------- | -------- |
| `--hostname`    | Hostname of the pending agent to accept      | Yes      |
| `--fingerprint` | Accept by key fingerprint instead of hostname | No       |
| `--json`        | Output raw JSON                              | No       |

## Examples

```bash
# Accept by hostname
osapi client agent accept --hostname web-01

# Accept by fingerprint
osapi client agent accept --hostname web-01 \
  --fingerprint SHA256:ab12cd34...
```
