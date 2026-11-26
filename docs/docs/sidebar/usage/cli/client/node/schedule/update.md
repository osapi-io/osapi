# Update

Deploy a cron entry that osapi deployed again.

Writes a new object, or the entry's own, at the entry's path, with the fields
that the request gives and the entry's own for the rest. A file that has the
content does not change. An entry that osapi did not deploy is not found. The
entry stays in its directory: an interval does not change with update.

```bash
osapi client node schedule update [flags]
```

## Examples

```bash
# A new object
osapi client node schedule update --target web-01 --name osapitest --object osapitest-v2.cron

# A new cron expression, with the same object
osapi client node schedule update --target web-01 --name osapitest --schedule "30 2 * * *"

# A new object, cron expression and user
osapi client node schedule update --target web-01 --name osapitest --object osapitest-v2.cron --schedule "30 2 * * *" --user nobody

# A daily script with a new object
osapi client node schedule update --target web-01 --name osapitest --object osapitest-v2.sh
```

## Result

A new object, on one host:

```ansi
$ osapi client node schedule update --target web-01 --name osapitest --object osapitest-v2.cron
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mosapitest[0m

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node schedule update --target _all --name osapitest --object osapitest-v2.cron
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mosapitest[0m
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m  [1mosapitest[0m

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
      "name": "osapitest",
      "status": "ok"
    }
  ]
}
```

A new cron expression, with the same object, on one host:

```ansi
$ osapi client node schedule update --target web-01 --name osapitest --schedule "30 2 * * *"
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mosapitest[0m

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node schedule update --target _all --name osapitest --schedule "30 2 * * *"
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mosapitest[0m
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m  [1mosapitest[0m

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
      "name": "osapitest",
      "status": "ok"
    }
  ]
}
```

A new object, cron expression and user, on one host:

```ansi
$ osapi client node schedule update --target web-01 --name osapitest --object osapitest-v2.cron --schedule "30 2 * * *" --user nobody
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mosapitest[0m

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node schedule update --target _all --name osapitest --object osapitest-v2.cron --schedule "30 2 * * *" --user nobody
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mosapitest[0m
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m  [1mosapitest[0m

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
      "name": "osapitest",
      "status": "ok"
    }
  ]
}
```

A daily script with a new object, on one host:

```ansi
$ osapi client node schedule update --target web-01 --name osapitest --object osapitest-v2.sh
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mosapitest[0m

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node schedule update --target _all --name osapitest --object osapitest-v2.sh
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mosapitest[0m
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m  [1mosapitest[0m

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
      "name": "osapitest",
      "status": "ok"
    }
  ]
}
```

With the default settings, each example does this on the host:

| Example                                     | On the host                                                                                  |
| ------------------------------------------- | -------------------------------------------------------------------------------------------- |
| A new object                                | writes `/etc/cron.d/osapitest`: `30 2 * * * root /usr/local/bin/backup --full`               |
| A new cron expression, with the same object | writes `/etc/cron.d/osapitest`: `0 2 * * * root /usr/local/bin/backup`                       |
| A new object, cron expression and user      | writes `/etc/cron.d/osapitest`: `30 2 * * * root /usr/local/bin/backup --full`               |
| A daily script with a new object            | writes `/etc/cron.daily/osapitest`: `#!/bin/sh\n/usr/local/bin/backup\n/usr/local/bin/prune` |

The command fails in these cases:

| Call                                  | Flags                                                                                                                                                                                                                                                                                        | Fails with             |
| ------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------- |
| An entry that osapi did not deploy    | `--name osapitest --object osapitest.cron`                                                                                                                                                                                                                                                   | not found              |
| A cron expression that does not parse | `--name osapitest --schedule "every day"`                                                                                                                                                                                                                                                    | Schedule               |
| A name with a space                   | `--name "a b"`                                                                                                                                                                                                                                                                               | Name                   |
| A name that is an option              | `--name --now`                                                                                                                                                                                                                                                                               | not a valid entry name |
| A name over 64 characters             | `--name aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa`                                                                                                                                                                                                                   | Name                   |
| An object name over 255 characters    | `--name osapitest --object aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa` | Object                 |
| A user that is not an account name    | `--name osapitest --user "a b"`                                                                                                                                                                                                                                                              | User                   |
| A content type that is not one        | `--name osapitest --content-type zip`                                                                                                                                                                                                                                                        | ContentType            |

## Permission

It needs the `schedule:write` permission.
`osapi controller role get --name <role>` shows what a role grants.

## Flags

| Flag             | Description                                                                    | Default |
| ---------------- | ------------------------------------------------------------------------------ | ------- |
| `--content-type` | The content type, raw or template; the object's own when empty (raw, template) |         |
| `--name`         | The entry, such as backup (required)                                           |         |
| `--object`       | The object to deploy; the entry's own when empty                               |         |
| `--schedule`     | A new cron expression; the entry's own when empty                              |         |
| `-T, --target`   | Target: _all, _any, hostname, or label (group:web.dev)                         | `_all`  |
| `--user`         | The user the entry runs as; the entry's own when empty                         |         |
| `--vars`         | The vars of a template, as key=value                                           |         |

## Global flags

| Flag               | Description                                                                      | Default                  |
| ------------------ | -------------------------------------------------------------------------------- | ------------------------ |
| `-d, --debug`      | Enable or disable debug mode                                                     |                          |
| `-j, --json`       | Enable JSON output                                                               |                          |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml`  |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `https://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
