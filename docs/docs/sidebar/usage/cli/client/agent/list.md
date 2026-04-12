# List

List active agents in the fleet with status, labels, age, and system metrics:

```bash
$ osapi client agent list

  Active Agents (3):

  MACHINE ID                            HOSTNAME  STATUS    CONDITIONS               LABELS                 AGE     LOAD (1m)  OS
  a1b2c3d4-e5f6-7890-abcd-ef1234567890  web-01    Ready     HighLoad,MemoryPressure  group:web.dev.us-east  3d 4h   4.12       Ubuntu 24.04
  b2c3d4e5-f6a7-8901-bcde-f12345678901  web-02    Ready     -                        group:web.dev.us-west  12h 5m  0.45       Ubuntu 24.04
  c3d4e5f6-a7b8-9012-cdef-123456789012  db-01     Cordoned  DiskPressure             -                      5d 2h   1.22       Ubuntu 24.04
```

This command reads directly from the agent heartbeat registry -- no job is
created. Each agent writes a heartbeat every 10 seconds with a 30-second TTL.
Agents that stop heartbeating disappear from the list automatically.

| Column     | Source                                                          |
| ---------- | --------------------------------------------------------------- |
| MACHINE ID | Permanent identifier from `/etc/machine-id` or macOS UUID       |
| HOSTNAME   | Agent's configured or OS hostname                               |
| STATUS     | Scheduling state: `Ready`, `Pending`, `Draining`, or `Cordoned` |
| CONDITIONS | Active node conditions (MemoryPressure, HighLoad, DiskPressure) |
| LABELS     | Key-value labels from agent config                              |
| AGE        | Time since the agent process started                            |
| LOAD (1m)  | 1-minute load average from heartbeat                            |
| OS         | Distribution and version from heartbeat                         |

:::tip Full facts in JSON output

`--json` output includes additional system facts collected by the agent:
architecture, kernel version, FQDN, CPU count, network interfaces, service
manager, and package manager. These fields are not shown in the table view.

:::

## Pending Agents

When PKI is enabled, list agents awaiting enrollment with `--pending`:

```bash
$ osapi client agent list --pending

  Pending Agents (2):

  MACHINE ID    HOSTNAME  FINGERPRINT           REQUESTED
  abc123...     web-03    SHA256:ab12cd34ef...   5m ago
  def456...     web-04    SHA256:ef56ab78cd...   2m ago
```

Use `agent accept` and `agent reject` to manage pending enrollment requests.

---

Use `agent get --hostname X` for detailed information about a specific agent, or
`node status` for deep system metrics gathered via the job system.
