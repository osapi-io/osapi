# Status

Show the health of the controller.

Each controller, with each of its sub-components, from the health status of the
controller.

```bash
osapi client controller status
```

## Examples

```bash
osapi client controller status
osapi client controller status --json
```

## Result

The command prints:

```ansi
$ osapi client controller status
[1;38;2;0;213;255mComponents (1):[0m
  [38;2;10;143;176mTYPE        [0m[38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mCONDITIONS  [0m[38;2;10;143;176mAGE    [0m[38;2;10;143;176mCPU %  [0m[38;2;10;143;176mMEM      [0m[38;2;10;143;176mVERSION  [0m[38;2;10;143;176mADDRESS[0m
  [1mcontroller[0m  [1mosapi-01[0m  [38;2;103;234;147mok[0m      [38;2;138;138;138m-[0m           [38;2;138;138;138m51h3m[0m  1.20   50.0 MB  v1.4.0   0.0.0.0:8080
  [38;2;138;138;138m├─ [0m[1mapi[0m      [1m[0m          [38;2;103;234;147mok[0m      [38;2;138;138;138m-[0m           [38;2;138;138;138m[0m
  [38;2;138;138;138m└─ [0m[1mmetrics[0m  [1m[0m          [38;2;103;234;147mok[0m      [38;2;138;138;138m-[0m           [38;2;138;138;138m[0m

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

With `--json`, the API's response:

```json
{
  "total_items": 1,
  "items": [
    {
      "type": "controller",
      "hostname": "osapi-01",
      "status": "ok",
      "age": "51h3m",
      "cpu_percent": 1.2,
      "mem_bytes": 52428800,
      "version": "v1.4.0",
      "address": "0.0.0.0:8080",
      "components": [
        {
          "type": "api",
          "status": "ok"
        },
        {
          "type": "metrics",
          "status": "ok"
        }
      ]
    }
  ]
}
```

## Permission

It needs the `health:read` permission. `osapi controller role get --name <role>`
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
