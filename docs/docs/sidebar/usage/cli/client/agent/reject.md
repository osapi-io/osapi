# Reject

Reject a pending agent's PKI enrollment request:

```bash
$ osapi client agent reject --hostname web-01

  Status: Rejected
  Message: Agent web-01 enrollment rejected
```

When PKI is enabled, a new agent sends an enrollment request. An administrator
accepts or rejects it. This command rejects a pending agent, so the agent cannot
join the fleet. The command removes the enrollment entry of the agent from the
pending queue.

Use `agent list --pending` to see the agents that wait for approval.

## Flags

| Flag         | Description                             | Required |
| ------------ | --------------------------------------- | -------- |
| `--hostname` | Hostname of the pending agent to reject | Yes      |
| `--json`     | Output raw JSON                         | No       |
