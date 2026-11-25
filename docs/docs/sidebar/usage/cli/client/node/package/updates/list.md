# List

List the packages the index has newer versions of.

Runs apt list --upgradable: each package with its installed version and the
version the index has. Refresh the index first for what the repositories have
now.

```bash
osapi client node package updates list [flags]
```

## Examples

```bash
# Two updates
osapi client node package updates list --target web-01
```

## Result

Two updates, on one host:

```ansi
$ osapi client node package updates list --target web-01
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME     [0m[38;2;10;143;176mCURRENT VERSION  [0m[38;2;10;143;176mNEW VERSION[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mcurl[0m     8.5.0-2ubuntu10  8.5.0-2ubuntu10.1
  [38;2;138;138;138mweb-01[0m    [38;2;103;234;147mok[0m      [1mopenssl[0m  3.0.13-0ubuntu3  3.0.13-0ubuntu3.1

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node package updates list --target _all
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME     [0m[38;2;10;143;176mCURRENT VERSION  [0m[38;2;10;143;176mNEW VERSION[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mcurl[0m     8.5.0-2ubuntu10  8.5.0-2ubuntu10.1
  [38;2;138;138;138mweb-01[0m    [38;2;103;234;147mok[0m      [1mopenssl[0m  3.0.13-0ubuntu3  3.0.13-0ubuntu3.1
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [1mcurl[0m     8.5.0-2ubuntu10  8.5.0-2ubuntu10.1
  [38;2;138;138;138mweb-02[0m    [38;2;103;234;147mok[0m      [1mopenssl[0m  3.0.13-0ubuntu3  3.0.13-0ubuntu3.1

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
      "current_version": "8.5.0-2ubuntu10",
      "hostname": "web-01",
      "name": "curl",
      "new_version": "8.5.0-2ubuntu10.1",
      "status": "ok"
    },
    {
      "current_version": "3.0.13-0ubuntu3",
      "hostname": "web-01",
      "name": "openssl",
      "new_version": "3.0.13-0ubuntu3.1",
      "status": "ok"
    }
  ]
}
```

## Permission

It needs the `package:read` permission.
`osapi controller role get --name <role>` shows what a role grants.

## Flags

| Flag           | Description                                            | Default |
| -------------- | ------------------------------------------------------ | ------- |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev) | `_all`  |

## Global flags

| Flag               | Description                                                 | Default                 |
| ------------------ | ----------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                |                         |
| `-j, --json`       | Enable JSON output                                          |                         |
| `-f, --osapi-file` | Path to config file                                         | `/etc/osapi/osapi.yaml` |
| `--url`            | Base URL the CLI client connects to (controller.client.url) | `http://localhost:8080` |

[Output](../../../../output.md) tells how the result is printed, and what
`--json` gives.
