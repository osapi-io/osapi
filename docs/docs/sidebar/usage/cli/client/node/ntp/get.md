# Get

Get the sync status of chrony and its sources.

Reads chronyc tracking, for whether chrony is synchronized, its stratum, its
offset from NTP time and the source that it follows, and chronyc sources, for
each source that it has.

```bash
osapi client node ntp get [flags]
```

## Examples

```bash
# A host synchronized to time.cloudflare.com
osapi client node ntp get --target web-01

# A host that chrony does not synchronize
osapi client node ntp get --target web-01
```

## Result

A host synchronized to time.cloudflare.com, on one host:

```ansi
$ osapi client node ntp get --target web-01
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mSYNCHRONIZED  [0m[38;2;10;143;176mSTRATUM  [0m[38;2;10;143;176mOFFSET         [0m[38;2;10;143;176mCURRENT SOURCE       [0m[38;2;10;143;176mSERVERS[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      yes           3        +0.000003422s  time.cloudflare.com  time.cloudflare.com, 0.pool.ntp.org

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node ntp get --target _all
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mSYNCHRONIZED  [0m[38;2;10;143;176mSTRATUM  [0m[38;2;10;143;176mOFFSET         [0m[38;2;10;143;176mCURRENT SOURCE       [0m[38;2;10;143;176mSERVERS[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      yes           3        +0.000003422s  time.cloudflare.com  time.cloudflare.com, 0.pool.ntp.org
  [1mweb-02[0m    [38;2;103;234;147mok[0m      yes           3        +0.000003422s  time.cloudflare.com  time.cloudflare.com, 0.pool.ntp.org

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
      "current_source": "time.cloudflare.com",
      "hostname": "web-01",
      "offset": "+0.000003422s",
      "servers": ["time.cloudflare.com", "0.pool.ntp.org"],
      "status": "ok",
      "stratum": 3,
      "synchronized": true
    }
  ]
}
```

A host that chrony does not synchronize, on one host:

```ansi
$ osapi client node ntp get --target web-01
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mSYNCHRONIZED  [0m[38;2;10;143;176mSTRATUM  [0m[38;2;10;143;176mOFFSET         [0m[38;2;10;143;176mCURRENT SOURCE  [0m[38;2;10;143;176mSERVERS[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      no                     -0.000001834s

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node ntp get --target _all
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mSYNCHRONIZED  [0m[38;2;10;143;176mSTRATUM  [0m[38;2;10;143;176mOFFSET         [0m[38;2;10;143;176mCURRENT SOURCE  [0m[38;2;10;143;176mSERVERS[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      no                     -0.000001834s
  [1mweb-02[0m    [38;2;103;234;147mok[0m      no                     -0.000001834s

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
      "offset": "-0.000001834s",
      "status": "ok",
      "synchronized": false
    }
  ]
}
```

## Permission

It needs the `ntp:read` permission. `osapi controller role get --name <role>`
shows what a role grants.

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

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
