# Status

Show the health of the NATS server.

The NATS server, with each of its sub-components, from the health status of the
controller.

```bash
osapi client nats status
```

## Examples

```bash
osapi client nats status
osapi client nats status --json
```

## Result

The command prints:

```ansi
$ osapi client nats status
[1;38;2;0;213;255mComponents (1):[0m
  [38;2;10;143;176mTYPE  [0m[38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mCONDITIONS  [0m[38;2;10;143;176mAGE    [0m[38;2;10;143;176mCPU %  [0m[38;2;10;143;176mMEM      [0m[38;2;10;143;176mVERSION  [0m[38;2;10;143;176mURL[0m
  [1mnats[0m  [1mosapi-01[0m  [38;2;103;234;147mok[0m      [38;2;138;138;138m-[0m           [38;2;138;138;138m51h3m[0m  0.30   20.0 MB  2.11.4   nats://localhost:4222

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

With `--json`, the API's response:

```json
{
  "total_items": 1,
  "items": [
    {
      "type": "nats",
      "hostname": "osapi-01",
      "status": "ok",
      "age": "51h3m",
      "cpu_percent": 0.3,
      "mem_bytes": 20971520,
      "version": "2.11.4",
      "url": "nats://localhost:4222"
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
