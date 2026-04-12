# Reject

Reject a pending agent's PKI enrollment request:

```bash
$ osapi client agent reject --hostname web-01

  Status: Rejected
  Message: Agent web-01 enrollment rejected
```

When PKI is enabled, new agents submit enrollment requests that must be accepted
or rejected by an administrator. This command rejects a pending agent,
preventing it from joining the fleet. The agent's enrollment entry is removed
from the pending queue.

Use `agent list --pending` to see agents awaiting approval.

## Flags

| Flag         | Description                             | Required |
| ------------ | --------------------------------------- | -------- |
| `--hostname` | Hostname of the pending agent to reject | Yes      |
| `--json`     | Output raw JSON                         | No       |
