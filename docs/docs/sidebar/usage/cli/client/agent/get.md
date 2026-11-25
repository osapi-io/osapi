# Get

Get agent details.

This command reads the agent heartbeat registry directly. It makes no job. The
data comes from the last heartbeat of the agent. `node status` runs a full
inspection through the job system, with disk usage and more metrics.

```bash
osapi client agent get [flags]
```

## Examples

```bash
osapi client agent get --hostname web-01
```

## Result

The command prints:

```ansi
$ osapi client agent get --hostname web-01
[1;38;2;0;213;255mAgent:[0m
  [38;2;176;48;176mHostname:[0m           web-01
  [38;2;176;48;176mMachine ID:[0m         4c4c4544-0035-4b10-8031-b4c04f4d3332
  [38;2;176;48;176mState:[0m              [38;2;103;234;147mReady[0m
  [38;2;176;48;176mLabels:[0m             group:web
  [38;2;176;48;176mAge:[0m                2d 3h
  [38;2;176;48;176mVersion:[0m            v1.4.0
  [38;2;176;48;176mLast seen:[0m          8s
  [38;2;176;48;176mUptime:[0m             2 days, 3 hours
  [38;2;176;48;176mArchitecture:[0m       amd64
  [38;2;176;48;176mKernel version:[0m     6.8.0-51-generic
  [38;2;176;48;176mCPU count:[0m          4
  [38;2;176;48;176mFQDN:[0m               web-01.example.com
  [38;2;176;48;176mService manager:[0m    systemd
  [38;2;176;48;176mPackage manager:[0m    apt
  [38;2;176;48;176mPrimary interface:[0m  eth0

[1;38;2;255;0;255mLoad:[0m
  [38;2;176;143;0m1 min:[0m   0.32
  [38;2;176;143;0m5 min:[0m   0.28
  [38;2;176;143;0m15 min:[0m  0.25

[1;38;2;255;209;0mOS:[0m
  [38;2;58;143;20mDistribution:[0m  Ubuntu
  [38;2;58;143;20mVersion:[0m       24.04
  [38;2;58;143;20mFamily:[0m        debian

[1;38;2;69;174;23mMemory:[0m

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
  [38;2;58;143;20mTotal:[0m  8.0 GB
  [38;2;58;143;20mFree:[0m   5.0 GB
  [38;2;58;143;20mUsed:[0m   3.0 GB
```

With `--json`, the API's response:

```json
{
  "hostname": "web-01",
  "machine_id": "4c4c4544-0035-4b10-8031-b4c04f4d3332",
  "state": "Ready",
  "labels": {
    "group": "web"
  },
  "started_at": "2026-01-13T07:00:00Z",
  "load_average": {
    "1min": 0.32,
    "5min": 0.28,
    "15min": 0.25
  },
  "os_info": {
    "distribution": "Ubuntu",
    "version": "24.04",
    "family": "debian"
  },
  "version": "v1.4.0",
  "registered_at": "2026-01-15T09:59:52Z",
  "uptime": "2 days, 3 hours",
  "memory": {
    "total": 8589934592,
    "free": 5368709120,
    "used": 3221225472
  },
  "architecture": "amd64",
  "kernel_version": "6.8.0-51-generic",
  "cpu_count": 4,
  "fqdn": "web-01.example.com",
  "service_mgr": "systemd",
  "package_mgr": "apt",
  "primary_interface": "eth0"
}
```

## Permission

It needs the `agent:read` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Flags

| Flag         | Description                                      | Default |
| ------------ | ------------------------------------------------ | ------- |
| `--hostname` | The hostname of the agent to retrieve (required) |         |

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `http://localhost:8080` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
