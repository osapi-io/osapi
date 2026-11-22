# Status

Show the health of each agent.

Each agent, with its process CPU and memory and each of its sub-components, from
the health status of the controller. agent list shows the host of each agent.

```bash
osapi client agent status
```

## Examples

```bash
osapi client agent status
osapi client agent status --json
```

## Result

The command prints:

```ansi
$ osapi client agent status
[1;38;2;0;213;255mComponents (2):[0m
  [38;2;10;143;176mTYPE          [0m[38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mCONDITIONS  [0m[38;2;10;143;176mAGE    [0m[38;2;10;143;176mCPU %  [0m[38;2;10;143;176mMEM      [0m[38;2;10;143;176mVERSION[0m
  [1magent[0m         [1mweb-01[0m    [38;2;103;234;147mok[0m      [38;2;138;138;138m-[0m           [38;2;138;138;138m51h2m[0m  0.40   30.0 MB  v1.4.0
  [38;2;138;138;138m├─ [0m[1mheartbeat[0m  [1m[0m          [38;2;103;234;147mok[0m      [38;2;138;138;138m-[0m           [38;2;138;138;138m[0m
  [38;2;138;138;138m└─ [0m[1mfacts[0m      [1m[0m          [38;2;103;234;147mok[0m      [38;2;138;138;138m-[0m           [38;2;138;138;138m[0m
  [1magent[0m         [1mmac-01[0m    [38;2;103;234;147mok[0m      [38;2;138;138;138m-[0m           [38;2;138;138;138m1h30m[0m  0.70   27.0 MB  v1.4.0

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

With `--json`, the API's response:

```json
{
  "total_items": 2,
  "items": [
    {
      "type": "agent",
      "hostname": "web-01",
      "status": "ok",
      "age": "51h2m",
      "cpu_percent": 0.4,
      "mem_bytes": 31457280,
      "version": "v1.4.0",
      "components": [
        {
          "type": "heartbeat",
          "status": "ok"
        },
        {
          "type": "facts",
          "status": "ok"
        }
      ]
    },
    {
      "type": "agent",
      "hostname": "mac-01",
      "status": "ok",
      "age": "1h30m",
      "cpu_percent": 0.7,
      "mem_bytes": 28311552,
      "version": "v1.4.0"
    }
  ]
}
```

## Permission

It needs the `health:read` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
