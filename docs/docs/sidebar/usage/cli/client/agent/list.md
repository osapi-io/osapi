# List

List active agents in the fleet with status, labels, age, and system metrics:

```bash
$ osapi client agent list

  Active Agents (2):

  HOSTNAME  STATUS  LABELS                 AGE     LOAD (1m)  OS
  web-01    Ready   group:web.dev.us-east  3d 4h   1.78       Ubuntu 24.04
  web-02    Ready   group:web.dev.us-west  12h 5m  0.45       Ubuntu 24.04
```

This command reads directly from the agent heartbeat registry -- no job is
created. Each agent writes a heartbeat every 10 seconds with a 30-second TTL.
Agents that stop heartbeating disappear from the list automatically.

| Column    | Source                                  |
| --------- | --------------------------------------- |
| HOSTNAME  | Agent's configured or OS hostname       |
| STATUS    | `Ready` if present in registry          |
| LABELS    | Key-value labels from agent config      |
| AGE       | Time since the agent process started    |
| LOAD (1m) | 1-minute load average from heartbeat    |
| OS        | Distribution and version from heartbeat |

Use `agent get --hostname X` for detailed information about a specific agent, or
`node status` for deep system metrics gathered via the job system.
