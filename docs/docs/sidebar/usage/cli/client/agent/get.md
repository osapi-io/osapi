# Get

Get detailed information about a specific agent by hostname:

```bash
$ osapi client agent get --hostname web-01

  Hostname: web-01                       Machine ID: a1b2c3d4e5f6
  Fingerprint: SHA256:4feebecfc58b7fcd...  Status: Ready
  State: Draining
  Labels: group:web.dev.us-east
  OS: Ubuntu 24.04
  Uptime: 6 days, 3 hours, 54 minutes
  Age: 3d 4h
  Last Seen: 3s ago
  Load: 1.74, 1.79, 1.94 (1m, 5m, 15m)
  Memory: 32.0 GB total, 19.2 GB used, 12.8 GB free
  Architecture: amd64
  Kernel: 6.8.0-51-generic
  FQDN: web-01.example.com
  CPUs: 8
  Service Mgr: systemd
  Package Mgr: apt
  Interfaces:
    eth0: 10.0.1.10 (IPv4), fe80::1 (IPv6), MAC 00:1a:2b:3c:4d:5e
    lo: 127.0.0.1 (IPv4), ::1 (IPv6)

  Conditions:
    TYPE              STATUS  REASON                                     SINCE
    MemoryPressure    true    memory 94% used (15.1/16.0 GB)             2m ago
    HighLoad          true    load 4.12, threshold 4.00 for 2 CPUs       5m ago
    DiskPressure      false

  Timeline:
    TIMESTAMP              EVENT      HOSTNAME  MESSAGE
    2026-03-05 10:00:00    drain      web-01    Drain initiated
    2026-03-05 10:05:23    cordoned   web-01    All jobs completed
```

This command reads directly from the agent heartbeat registry -- no job is
created. The data comes from the agent's most recent heartbeat write.

| Field        | Description                                               |
| ------------ | --------------------------------------------------------- |
| Hostname     | Agent's configured or OS hostname                         |
| Machine ID   | Permanent identifier from `/etc/machine-id` or macOS UUID |
| Fingerprint  | SHA256 hash of agent's PKI public key (when PKI enabled)  |
| Status       | `Ready` if present in registry                            |
| State        | Scheduling state: `Pending`, `Draining`, or `Cordoned`    |
| Labels       | Key-value labels from agent config                        |
| OS           | Distribution and version                                  |
| Uptime       | System uptime reported by the agent                       |
| Age          | Time since the agent process started                      |
| Last Seen    | Time since the last heartbeat refresh                     |
| Load         | 1-, 5-, and 15-minute load averages                       |
| Memory       | Total, used, and free RAM                                 |
| Architecture | CPU architecture (e.g., amd64)                            |
| Kernel       | OS kernel version                                         |
| FQDN         | Fully qualified domain name                               |
| CPUs         | Number of logical CPUs                                    |
| Service Mgr  | Init system (e.g., systemd)                               |
| Package Mgr  | Package manager (e.g., apt)                               |
| Interfaces   | Network interfaces with IPv4, IPv6, MAC, and family       |
| Conditions   | Node conditions table (type, status, reason, since)       |
| Timeline     | State transition events (timestamp, event, hostname)      |

:::tip agent get vs. node status

`agent get` shows lightweight data from the heartbeat registry (instant, no
job). `node status` runs a full system inspection via the job system (includes
disk usage, deeper metrics).

:::

## Flags

| Flag         | Description                       | Required |
| ------------ | --------------------------------- | -------- |
| `--hostname` | Hostname of the agent to retrieve | Yes      |
