# Get

Get detailed information about a specific agent by hostname:

```bash
$ osapi client node get --hostname web-01

  Hostname: web-01                       Status: Ready
  Labels: group:web.dev.us-east
  OS: Ubuntu 24.04
  Uptime: 6 days, 3 hours, 54 minutes
  Age: 3d 4h
  Last Seen: 3s ago
  Load: 1.74, 1.79, 1.94 (1m, 5m, 15m)
  Memory: 32.0 GB total, 19.2 GB used, 12.8 GB free
```

This command reads directly from the agent heartbeat registry — no job is
created. The data comes from the agent's most recent heartbeat write.

| Field     | Description                           |
| --------- | ------------------------------------- |
| Hostname  | Agent's configured or OS hostname     |
| Status    | `Ready` if present in registry        |
| Labels    | Key-value labels from agent config    |
| OS        | Distribution and version              |
| Uptime    | System uptime reported by the agent   |
| Age       | Time since the agent process started  |
| Last Seen | Time since the last heartbeat refresh |
| Load      | 1-, 5-, and 15-minute load averages   |
| Memory    | Total, used, and free RAM             |

:::tip node get vs. node status

`node get` shows lightweight data from the heartbeat registry (instant, no job).
`node status` runs a full system inspection via the job system (includes disk
usage, deeper metrics).

:::

## Flags

| Flag         | Description                       | Required |
| ------------ | --------------------------------- | -------- |
| `--hostname` | Hostname of the agent to retrieve | Yes      |
