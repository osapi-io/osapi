# Get

Get the memory usage of the host.

Reads the total, the free and the used memory, in bytes. Used is the total less
what is available, which includes the cache, as free(1) reports it.

```bash
osapi client node memory get [flags]
```

## Examples

```bash
# A host with 8 GiB, 2 GiB of it free
osapi client node memory get --target web-01
```

## Result

A host with 8 GiB, 2 GiB of it free, on one host:

```ansi
$ osapi client node memory get --target web-01
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mMEMORY[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      8.0 GB 2.0 GB 4.0 GB

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node memory get --target _all
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mMEMORY[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      8.0 GB 2.0 GB 4.0 GB
  [1mweb-02[0m    [38;2;103;234;147mok[0m      8.0 GB 2.0 GB 4.0 GB

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
      "memory": {
        "free": 2147483648,
        "total": 8589934592,
        "used": 4294967296
      },
      "status": "ok"
    }
  ]
}
```

## Permission

It needs the `memory:read` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Flags

| Flag           | Description                                            | Default |
| -------------- | ------------------------------------------------------ | ------- |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev) | `_all`  |

## Global flags

| Flag               | Description                                                                      | Default                  |
| ------------------ | -------------------------------------------------------------------------------- | ------------------------ |
| `-d, --debug`      | Enable or disable debug mode                                                     |                          |
| `-j, --json`       | Enable JSON output                                                               |                          |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml`  |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `https://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
