# Status

System status and component health.

The table has each component that registered: its status, its active conditions,
its uptime, and the CPU and memory of its process.

```bash
osapi client health status [flags]
```

## Examples

```bash
osapi client health status
osapi client health status --component agent
osapi client health status --json
```

## Result

The command prints:

```ansi
$ osapi client health status
[1;38;2;0;213;255mController:[0m
  [38;2;10;143;176mStatus:[0m   [38;2;103;234;147mok[0m
  [38;2;10;143;176mVersion:[0m  v1.4.0
  [38;2;10;143;176mUptime:[0m   51h3m

[1;38;2;255;0;255mChecks (2):[0m
  [38;2;176;48;176mNAME  [0m[38;2;176;48;176mSTATUS[0m
  [1mnats[0m  [38;2;103;234;147mok[0m
  [1mkv[0m    [38;2;103;234;147mok[0m

[1;38;2;255;209;0mComponents (3):[0m
  [38;2;176;143;0mTYPE        [0m[38;2;176;143;0mHOSTNAME  [0m[38;2;176;143;0mSTATUS  [0m[38;2;176;143;0mCONDITIONS  [0m[38;2;176;143;0mAGE    [0m[38;2;176;143;0mCPU %  [0m[38;2;176;143;0mMEM      [0m[38;2;176;143;0mVERSION  [0m[38;2;176;143;0mADDRESS       [0m[38;2;176;143;0mURL[0m
  [1mcontroller[0m  [1mosapi-01[0m  [38;2;103;234;147mok[0m      [38;2;138;138;138m-[0m           [38;2;138;138;138m51h3m[0m  1.20   50.0 MB  v1.4.0   0.0.0.0:8080
  [1mnats[0m        [1mosapi-01[0m  [38;2;103;234;147mok[0m      [38;2;138;138;138m-[0m           [38;2;138;138;138m51h3m[0m                  2.11.4                 nats://localhost:4222
  [1magent[0m       [1mweb-01[0m    [38;2;103;234;147mok[0m      [38;2;138;138;138m-[0m           [38;2;138;138;138m51h2m[0m  0.40   30.0 MB  v1.4.0

[1;38;2;69;174;23mJobs:[0m
  [38;2;58;143;20mTotal:[0m       128
  [38;2;58;143;20mSubmitted:[0m   3
  [38;2;58;143;20mProcessing:[0m  1
  [38;2;58;143;20mCompleted:[0m   121
  [38;2;58;143;20mFailed:[0m      3
  [38;2;58;143;20mDLQ:[0m         0

[1;38;2;107;92;255mStreams (1):[0m
  [38;2;89;76;216mNAME  [0m[38;2;89;76;216mMESSAGES  [0m[38;2;89;76;216mBYTES     [0m[38;2;89;76;216mCONSUMERS[0m
  [1mJOBS[0m  128       256.0 KB  2

[1;38;2;0;213;255mKV buckets (2):[0m
  [38;2;10;143;176mNAME            [0m[38;2;10;143;176mKEYS  [0m[38;2;10;143;176mBYTES[0m
  [1mjob-queue[0m       128   512.0 KB
  [1magent-registry[0m  2     4.0 KB

[1;38;2;255;0;255mObject stores (1):[0m
  [38;2;176;48;176mNAME          [0m[38;2;176;48;176mSIZE[0m
  [1mfile-objects[0m  1.0 MB

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

With `--json`, the API's response:

```json
{
  "status": "ok",
  "version": "v1.4.0",
  "uptime": "51h3m",
  "checks": [
    {
      "name": "nats",
      "status": "ok"
    },
    {
      "name": "kv",
      "status": "ok"
    }
  ],
  "registry": [
    {
      "type": "controller",
      "hostname": "osapi-01",
      "status": "ok",
      "age": "51h3m",
      "cpu_percent": 1.2,
      "mem_bytes": 52428800,
      "version": "v1.4.0",
      "address": "0.0.0.0:8080"
    },
    {
      "type": "nats",
      "hostname": "osapi-01",
      "status": "ok",
      "age": "51h3m",
      "version": "2.11.4",
      "url": "nats://localhost:4222"
    },
    {
      "type": "agent",
      "hostname": "web-01",
      "status": "ok",
      "age": "51h2m",
      "cpu_percent": 0.4,
      "mem_bytes": 31457280,
      "version": "v1.4.0"
    }
  ],
  "jobs": {
    "total": 128,
    "submitted": 3,
    "processing": 1,
    "completed": 121,
    "failed": 3,
    "dlq": 0
  },
  "streams": [
    {
      "name": "JOBS",
      "messages": 128,
      "bytes": 262144,
      "consumers": 2
    }
  ],
  "kv_buckets": [
    {
      "name": "job-queue",
      "keys": 128,
      "bytes": 524288
    },
    {
      "name": "agent-registry",
      "keys": 2,
      "bytes": 4096
    }
  ],
  "object_stores": [
    {
      "name": "file-objects",
      "size": 1048576
    }
  ],
  "consumers": 2
}
```

## Permission

It needs the `health:read` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Notes

Each sub-component is under its component, as a tree, with its address and its
URL. The checks tell if the controller can reach NATS and the KV buckets. The
output also has the counts of the job queue, the streams, the KV buckets and the
Object Stores. The command needs a token. --component shows one type of
component: controller, nats or agent. Use `osapi client agent list` for the
labels and the OS of each agent. When a check fails, the overall status is
`degraded`, and the check shows the error.

## Flags

| Flag          | Description                                                                                                                    | Default |
| ------------- | ------------------------------------------------------------------------------------------------------------------------------ | ------- |
| `--component` | Show only the components of this type: controller, nats or agent. The other parts of the status stay (controller, nats, agent) |         |

## Global flags

| Flag               | Description                                                 | Default                 |
| ------------------ | ----------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                |                         |
| `-j, --json`       | Enable JSON output                                          |                         |
| `-f, --osapi-file` | Path to config file                                         | `/etc/osapi/osapi.yaml` |
| `--url`            | Base URL the CLI client connects to (controller.client.url) | `http://localhost:8080` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
