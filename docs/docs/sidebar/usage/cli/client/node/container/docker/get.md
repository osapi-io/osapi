# Get

Get one container.

Reads one container by its name, its ID or the start of its ID. A container that
the host does not have is not found.

```bash
osapi client node container docker get [flags]
```

## Examples

```bash
# A container
osapi client node container docker get --target web-01 --name osapitest
```

## Result

A container, on one host:

```ansi
$ osapi client node container docker get --target web-01 --name osapitest
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mID                                                                [0m[38;2;10;143;176mNAME       [0m[38;2;10;143;176mIMAGE        [0m[38;2;10;143;176mSTATE    [0m[38;2;10;143;176mSUMMARY       [0m[38;2;10;143;176mCREATED              [0m[38;2;10;143;176mHEALTH  [0m[38;2;10;143;176mPORTS                 [0m[38;2;10;143;176mMOUNTS              [0m[38;2;10;143;176mNETWORKS                          [0m[38;2;10;143;176mENV NAMES[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [38;2;138;138;138m3f4e8a1b2c5d6e7f8091a2b3c4d5e6f708192a3b4c5d6e7f8091a2b3c4d5e6f7[0m  [1mosapitest[0m  alpine:3.20  [38;2;103;234;147mrunning[0m  Up 2 minutes  [38;2;138;138;138m2026-11-22 10:00:00[0m          0.0.0.0:8080->80/tcp  /srv/data:/data:ro  bridge 172.17.0.2 via 172.17.0.1  PATH, APP_TOKEN

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node container docker get --target _all --name osapitest
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mID                                                                [0m[38;2;10;143;176mNAME       [0m[38;2;10;143;176mIMAGE        [0m[38;2;10;143;176mSTATE    [0m[38;2;10;143;176mSUMMARY       [0m[38;2;10;143;176mCREATED              [0m[38;2;10;143;176mHEALTH  [0m[38;2;10;143;176mPORTS                 [0m[38;2;10;143;176mMOUNTS              [0m[38;2;10;143;176mNETWORKS                          [0m[38;2;10;143;176mENV NAMES[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [38;2;138;138;138m3f4e8a1b2c5d6e7f8091a2b3c4d5e6f708192a3b4c5d6e7f8091a2b3c4d5e6f7[0m  [1mosapitest[0m  alpine:3.20  [38;2;103;234;147mrunning[0m  Up 2 minutes  [38;2;138;138;138m2026-11-22 10:00:00[0m          0.0.0.0:8080->80/tcp  /srv/data:/data:ro  bridge 172.17.0.2 via 172.17.0.1  PATH, APP_TOKEN
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [38;2;138;138;138m3f4e8a1b2c5d6e7f8091a2b3c4d5e6f708192a3b4c5d6e7f8091a2b3c4d5e6f7[0m  [1mosapitest[0m  alpine:3.20  [38;2;103;234;147mrunning[0m  Up 2 minutes  [38;2;138;138;138m2026-11-22 10:00:00[0m          0.0.0.0:8080->80/tcp  /srv/data:/data:ro  bridge 172.17.0.2 via 172.17.0.1  PATH, APP_TOKEN

  3 hosts: [38;2;103;234;147m2 ok[0m, [38;5;222m1 skipped[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m

[1;38;2;255;0;255mDetails:[0m
  [1mmac-01[0m  [38;5;222mskipped[0m  operation not supported on this OS family
```

With `--json`, the API's response:

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "results": [
    {
      "created": "2026-11-22T10:00:00Z",
      "env_names": ["PATH", "APP_TOKEN"],
      "hostname": "web-01",
      "id": "3f4e8a1b2c5d6e7f8091a2b3c4d5e6f708192a3b4c5d6e7f8091a2b3c4d5e6f7",
      "image": "alpine:3.20",
      "mounts": ["/srv/data:/data:ro"],
      "name": "osapitest",
      "networks": ["bridge 172.17.0.2 via 172.17.0.1"],
      "ports": ["0.0.0.0:8080-\u003e80/tcp"],
      "state": "running",
      "status": "ok",
      "summary": "Up 2 minutes"
    }
  ]
}
```

The command fails in these cases:

| Call                                    | Flags            | Fails with   |
| --------------------------------------- | ---------------- | ------------ |
| A container that the host does not have | `--name nothere` | not found    |
| A name that is an option                | `--name -f`      | container_id |

## Permission

It needs the `container:read` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Flags

| Flag           | Description                                            | Default |
| -------------- | ------------------------------------------------------ | ------- |
| `--name`       | The container's name or ID (required)                  |         |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev) | `_all`  |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../../output.md) tells how the result is printed, and what
`--json` gives.
