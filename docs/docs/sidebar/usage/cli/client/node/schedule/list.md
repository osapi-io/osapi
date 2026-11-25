# List

List the cron entries that osapi deployed.

Lists each file in /etc/cron.d and in /etc/cron.hourly, daily, weekly and
monthly that osapi deployed, with its object, its cron expression or interval,
and its user. A cron file that osapi did not deploy is not listed.

```bash
osapi client node schedule list [flags]
```

## Examples

```bash
# An entry that osapi deployed
osapi client node schedule list --target web-01

# A daily script that osapi deployed
osapi client node schedule list --target web-01

# A host where osapi deployed none
osapi client node schedule list --target web-01
```

## Result

An entry that osapi deployed, on one host:

```ansi
$ osapi client node schedule list --target web-01
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME       [0m[38;2;10;143;176mOBJECT          [0m[38;2;10;143;176mSCHEDULE   [0m[38;2;10;143;176mINTERVAL  [0m[38;2;10;143;176mUSER  [0m[38;2;10;143;176mSOURCE[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mosapitest[0m  osapitest.cron  0 2 * * *            root  cron.d

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node schedule list --target _all
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME       [0m[38;2;10;143;176mOBJECT          [0m[38;2;10;143;176mSCHEDULE   [0m[38;2;10;143;176mINTERVAL  [0m[38;2;10;143;176mUSER  [0m[38;2;10;143;176mSOURCE[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mosapitest[0m  osapitest.cron  0 2 * * *            root  cron.d
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [1mosapitest[0m  osapitest.cron  0 2 * * *            root  cron.d

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
      "name": "osapitest",
      "object": "osapitest.cron",
      "schedule": "0 2 * * *",
      "source": "cron.d",
      "status": "ok",
      "user": "root"
    }
  ]
}
```

A daily script that osapi deployed, on one host:

```ansi
$ osapi client node schedule list --target web-01
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME       [0m[38;2;10;143;176mOBJECT        [0m[38;2;10;143;176mSCHEDULE  [0m[38;2;10;143;176mINTERVAL  [0m[38;2;10;143;176mUSER  [0m[38;2;10;143;176mSOURCE[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mosapitest[0m  osapitest.sh            daily           daily

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node schedule list --target _all
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME       [0m[38;2;10;143;176mOBJECT        [0m[38;2;10;143;176mSCHEDULE  [0m[38;2;10;143;176mINTERVAL  [0m[38;2;10;143;176mUSER  [0m[38;2;10;143;176mSOURCE[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mosapitest[0m  osapitest.sh            daily           daily
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [1mosapitest[0m  osapitest.sh            daily           daily

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
      "interval": "daily",
      "name": "osapitest",
      "object": "osapitest.sh",
      "source": "daily",
      "status": "ok"
    }
  ]
}
```

A host where osapi deployed none, on one host:

```ansi
$ osapi client node schedule list --target web-01
[1;38;2;0;213;255mResults:[0m
  [38;2;138;138;138mNone.[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node schedule list --target _all
[1;38;2;0;213;255mResults:[0m

  1 host: [38;5;222m1 skipped[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m

[1;38;2;255;0;255mDetails:[0m
  [1mmac-01[0m  [38;5;222mskipped[0m  operation not supported on this OS family
```

With `--json`, the API's response:

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "results": []
}
```

## Permission

It needs the `schedule:read` permission.
`osapi controller role get --name <role>` shows what a role grants.

## Flags

| Flag           | Description                                            | Default |
| -------------- | ------------------------------------------------------ | ------- |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev) | `_all`  |

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `http://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
