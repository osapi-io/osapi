# List

List the images of the host.

Reads each image through gohai's docker collector, which runs docker images. A
host whose docker does not answer reports skipped.

```bash
osapi client node image list [flags]
```

## Examples

```bash
# A host with an image
osapi client node image list --target web-01
```

## Result

A host with an image, on one host:

```ansi
$ osapi client node image list --target web-01
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mID            [0m[38;2;10;143;176mREPOSITORY  [0m[38;2;10;143;176mTAG   [0m[38;2;10;143;176mSIZE[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [38;2;138;138;138m91ef0af61f39[0m  alpine      3.20  7.8MB

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node image list --target _all
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mID            [0m[38;2;10;143;176mREPOSITORY  [0m[38;2;10;143;176mTAG   [0m[38;2;10;143;176mSIZE[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [38;2;138;138;138m91ef0af61f39[0m  alpine      3.20  7.8MB
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [38;2;138;138;138m91ef0af61f39[0m  alpine      3.20  7.8MB

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
      "hostname": "web-01",
      "id": "91ef0af61f39",
      "repository": "alpine",
      "size": "7.8MB",
      "status": "ok",
      "tag": "3.20"
    }
  ]
}
```

## Permission

It needs the `image:read` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Flags

| Flag           | Description                                            | Default |
| -------------- | ------------------------------------------------------ | ------- |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev) | `_all`  |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
