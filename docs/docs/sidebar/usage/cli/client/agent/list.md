# List

List active agents.

This command reads the agent heartbeat registry directly. It makes no job. Each
agent writes a heartbeat every 10 seconds, with a 30-second TTL. When an agent
stops its heartbeat, it goes from the list.

```bash
osapi client agent list
```

## Examples

```bash
osapi client agent list
```

## Result

The command prints:

```ansi
$ osapi client agent list
[1;38;2;0;213;255mAgents (2):[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mMACHINE ID                            [0m[38;2;10;143;176mSTATE     [0m[38;2;10;143;176mCONDITIONS  [0m[38;2;10;143;176mLABELS     [0m[38;2;10;143;176mAGE     [0m[38;2;10;143;176mLOAD            [0m[38;2;10;143;176mOS                   [0m[38;2;10;143;176mVERSION[0m
  [1mweb-01[0m    [38;2;138;138;138m4c4c4544-0035-4b10-8031-b4c04f4d3332[0m  [38;2;103;234;147mReady[0m     [38;2;138;138;138m-[0m           group:web  [38;2;138;138;138m2d 3h[0m   0.32 0.28 0.25  Ubuntu 24.04 debian  v1.4.0
  [1mmac-01[0m    [38;2;138;138;138m[0m                                      [38;5;222mDraining[0m  [38;2;138;138;138m-[0m                      [38;2;138;138;138m1h 30m[0m  1.12 0.94 0.88  macOS 15.5 darwin    v1.4.0

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

With `--json`, the API's response:

```json
{
  "items": [
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
    },
    {
      "hostname": "mac-01",
      "state": "Draining",
      "started_at": "2026-01-15T08:30:00Z",
      "load_average": {
        "1min": 1.12,
        "5min": 0.94,
        "15min": 0.88
      },
      "os_info": {
        "distribution": "macOS",
        "version": "15.5",
        "family": "darwin"
      },
      "version": "v1.4.0",
      "registered_at": "2026-01-15T09:59:52Z"
    }
  ],
  "total_items": 2
}
```

## Permission

It needs the `agent:read` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Notes

The `--json` output has more system facts that the agent collects: architecture,
kernel version, FQDN, CPU count, network interfaces, service manager and package
manager. The table view does not show these fields.

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
