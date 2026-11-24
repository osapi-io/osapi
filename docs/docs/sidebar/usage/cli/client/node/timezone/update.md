# Update

Set the timezone of the host.

Sets the timezone with timedatectl set-timezone. A host that already has the
timezone does not change and reports changed: false. It runs on the Debian
family only, and elsewhere reports skipped.

```bash
osapi client node timezone update [flags]
```

## Examples

```bash
# Set a new timezone
osapi client node timezone update --target web-01 --timezone America/New_York

# A host that already has the timezone
osapi client node timezone update --target web-01 --timezone UTC
```

## Result

Set a new timezone, on one host:

```ansi
$ osapi client node timezone update --target web-01 --timezone America/New_York
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mTIMEZONE[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  America/New_York

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node timezone update --target _all --timezone America/New_York
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mTIMEZONE[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  America/New_York
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m  America/New_York

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
      "status": "ok",
      "timezone": "America/New_York"
    }
  ]
}
```

A host that already has the timezone, on one host:

```ansi
$ osapi client node timezone update --target web-01 --timezone UTC
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mTIMEZONE[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      UTC

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node timezone update --target _all --timezone UTC
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mTIMEZONE[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      UTC
  [1mweb-02[0m    [38;2;103;234;147mok[0m      UTC

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
      "status": "ok",
      "timezone": "UTC"
    }
  ]
}
```

With the default settings, each example does this on the host:

| Example            | On the host                                      |
| ------------------ | ------------------------------------------------ |
| Set a new timezone | runs `timedatectl set-timezone America/New_York` |

The command fails in these cases:

| Call                             | Flags                                                                          | Fails with                |
| -------------------------------- | ------------------------------------------------------------------------------ | ------------------------- |
| A name that is not a timezone    | `--timezone ../../etc/passwd`                                                  | not a valid timezone name |
| A name that starts with a dash   | `--timezone -UTC`                                                              | not a valid timezone name |
| A name longer than 64 characters | `--timezone aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa` | timezone                  |

## Permission

It needs the `timezone:write` permission.
`osapi controller role get --name <role>` shows what a role grants.

## Flags

| Flag           | Description                                                               | Default |
| -------------- | ------------------------------------------------------------------------- | ------- |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev)                    | `_all`  |
| `--timezone`   | The IANA name of the timezone, such as UTC or America/New_York (required) |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
