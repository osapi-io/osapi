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
  [38;2;176;48;176mFIPS:[0m               off
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

[1;38;2;69;174;23mCrypto:[0m
  [38;2;58;143;20mGo:[0m         go1.27.3
  [38;2;58;143;20mModule:[0m     v1.0.0
  [38;2;58;143;20mMode:[0m       off
  [38;2;58;143;20mOnly mode:[0m  no
  [38;2;58;143;20mGaps:[0m       FIPS 140-3 mode is off: set GODEBUG=fips140=on, metrics: crypto/tls may negotiate a suite or group that is not approved

[1;38;2;107;92;255mListeners (1):[0m
  [38;2;89;76;216mNAME     [0m[38;2;89;76;216mTLS  [0m[38;2;89;76;216mMIN VERSION  [0m[38;2;89;76;216mSUITES                                                [0m[38;2;89;76;216mGROUPS              [0m[38;2;89;76;216mAPPROVED[0m
  [1mmetrics[0m  yes  TLS 1.2      TLS_AES_128_GCM_SHA256, TLS_CHACHA20_POLY1305_SHA256  crypto/tls default  no

[1;38;2;0;213;255mAlgorithms (3):[0m
  [38;2;10;143;176mNAME                                  [0m[38;2;10;143;176mALGORITHM                             [0m[38;2;10;143;176mAPPROVED  [0m[38;2;10;143;176mONLY MODE  [0m[38;2;10;143;176mREASON[0m
  [1mjob, join and admission signatures[0m    Ed25519                               yes       yes
  [1msecrets sealed to agents[0m              ECDH P-256, HKDF-SHA256, AES-256-GCM  yes       yes
  [1mNATS operator, account and user JWTs[0m  Ed25519 (nkeys)                       yes       no         nkeys v0.4.17 makes a key with ed25519.GenerateKey from a seed, which FIPS 140…

[1;38;2;255;0;255mMemory:[0m

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
  [38;2;176;48;176mTotal:[0m  8.0 GB
  [38;2;176;48;176mFree:[0m   5.0 GB
  [38;2;176;48;176mUsed:[0m   3.0 GB
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
  "fips": "off",
  "crypto": {
    "go": "go1.27.3",
    "module": "v1.0.0",
    "mode": "off",
    "enforceable": false,
    "listeners": [
      {
        "name": "metrics",
        "tls": true,
        "min_version": "TLS 1.2",
        "suites": ["TLS_AES_128_GCM_SHA256", "TLS_CHACHA20_POLY1305_SHA256"],
        "groups": ["crypto/tls default"],
        "approved": false
      }
    ],
    "algorithms": [
      {
        "name": "job, join and admission signatures",
        "algorithm": "Ed25519",
        "approved": true,
        "only_mode": true
      },
      {
        "name": "secrets sealed to agents",
        "algorithm": "ECDH P-256, HKDF-SHA256, AES-256-GCM",
        "approved": true,
        "only_mode": true
      },
      {
        "name": "NATS operator, account and user JWTs",
        "algorithm": "Ed25519 (nkeys)",
        "approved": true,
        "only_mode": false,
        "reason": "nkeys v0.4.17 makes a key with ed25519.GenerateKey from a seed, which FIPS 140-3 only mode refuses (osapi#1527)"
      }
    ],
    "gaps": [
      "FIPS 140-3 mode is off: set GODEBUG=fips140=on",
      "metrics: crypto/tls may negotiate a suite or group that is not approved"
    ]
  },
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

| Flag               | Description                                                                      | Default                  |
| ------------------ | -------------------------------------------------------------------------------- | ------------------------ |
| `-d, --debug`      | Enable or disable debug mode                                                     |                          |
| `-j, --json`       | Enable JSON output                                                               |                          |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml`  |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `https://localhost:8080` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
