# Delete

Remove a file that osapi deployed.

Removes the file and keeps its record, marked undeployed, so status reports
unmanaged. A path that osapi did not deploy, or deleted already, does not change
and reports changed: false. The object stays in the object store.

```bash
osapi client node file delete [flags]
```

## Examples

```bash
# Remove a deployed file
osapi client node file delete --target web-01 --path /etc/osapi/file-example.conf

# A path that osapi did not deploy
osapi client node file delete --target web-01 --path /etc/osapi/file-example.conf
```

## Result

Remove a deployed file, on one host:

```ansi
$ osapi client node file delete --target web-01 --path /etc/osapi/file-example.conf
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mPATH[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  /etc/osapi/file-example.conf

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node file delete --target _all --path /etc/osapi/file-example.conf
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mPATH[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  /etc/osapi/file-example.conf
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m  /etc/osapi/file-example.conf

  3 hosts: [38;2;103;234;147m2 changed[0m, [38;5;222m1 skipped[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m

[1;38;2;255;0;255mDetails:[0m
  [1mmac-01[0m  [38;5;222mskipped[0m  operation not supported on this OS family
```

With `--json`, the API's response:

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "results": [
    {
      "changed": true,
      "hostname": "web-01",
      "path": "/etc/osapi/file-example.conf",
      "status": "ok"
    }
  ]
}
```

A path that osapi did not deploy, on one host:

```ansi
$ osapi client node file delete --target web-01 --path /etc/osapi/file-example.conf
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mPATH[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      /etc/osapi/file-example.conf

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node file delete --target _all --path /etc/osapi/file-example.conf
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mPATH[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      /etc/osapi/file-example.conf
  [1mweb-02[0m    [38;2;103;234;147mok[0m      /etc/osapi/file-example.conf

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
      "changed": false,
      "hostname": "web-01",
      "path": "/etc/osapi/file-example.conf",
      "status": "ok"
    }
  ]
}
```

With the default settings, each example does this on the host:

| Example                | On the host                            |
| ---------------------- | -------------------------------------- |
| Remove a deployed file | removes `/etc/osapi/file-example.conf` |

The command fails in these cases:

| Call            | Flags          | Fails with |
| --------------- | -------------- | ---------- |
| A relative path | `--path etc/x` | startswith |

## Permission

It needs the `file:write` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Flags

| Flag           | Description                                            | Default |
| -------------- | ------------------------------------------------------ | ------- |
| `--path`       | Absolute path of the file on the host (required)       |         |
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
