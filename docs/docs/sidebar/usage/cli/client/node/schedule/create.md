# Create

Deploy a cron entry.

Gets the object from the object store and writes it as /etc/cron.d/backup for
backup, mode 0644, for a cron expression, or as /etc/cron.daily/backup, mode
0755, for a daily interval. An object of content type template, or content_type
template, is rendered with the agent's facts and the vars. A file that has the
content does not change. An entry of the name in a different directory does not
change: move it with delete and create. A file at the path that osapi did not
deploy is refused. An entry runs as its user, root included, so only admin can
deploy one.

```bash
osapi client node schedule create [flags]
```

## Examples

```bash
# An entry with a cron expression
osapi client node schedule create --target web-01 --name osapitest --object osapitest.cron --schedule "0 2 * * *" --user root

# A script that runs daily
osapi client node schedule create --target web-01 --name osapitest --object osapitest.sh --interval daily

# A template, rendered with its vars
osapi client node schedule create --target web-01 --name osapitest --object osapitest.tmpl --schedule "0 2 * * *" --content-type template --vars target=/srv

# An entry that has the content
osapi client node schedule create --target web-01 --name osapitest --object osapitest.cron --schedule "0 2 * * *" --user root

# An entry of the name in a different directory
osapi client node schedule create --target web-01 --name osapitest --object osapitest.sh --interval daily
```

## Result

An entry with a cron expression, on one host:

```ansi
$ osapi client node schedule create --target web-01 --name osapitest --object osapitest.cron --schedule "0 2 * * *" --user root
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mosapitest[0m

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node schedule create --target _all --name osapitest --object osapitest.cron --schedule "0 2 * * *" --user root
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

A script that runs daily, on one host:

```ansi
$ osapi client node schedule create --target web-01 --name osapitest --object osapitest.sh --interval daily
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mosapitest[0m

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node schedule create --target _all --name osapitest --object osapitest.sh --interval daily
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

A template, rendered with its vars, on one host:

```ansi
$ osapi client node schedule create --target web-01 --name osapitest --object osapitest.tmpl --schedule "0 2 * * *" --content-type template --vars target=/srv
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mosapitest[0m

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node schedule create --target _all --name osapitest --object osapitest.tmpl --schedule "0 2 * * *" --content-type template --vars target=/srv
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

An entry that has the content, on one host:

```ansi
$ osapi client node schedule create --target web-01 --name osapitest --object osapitest.cron --schedule "0 2 * * *" --user root
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mosapitest[0m

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node schedule create --target _all --name osapitest --object osapitest.cron --schedule "0 2 * * *" --user root
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mosapitest[0m
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [1mosapitest[0m

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
      "name": "osapitest",
      "status": "ok"
    }
  ]
}
```

An entry of the name in a different directory, on one host:

```ansi
$ osapi client node schedule create --target web-01 --name osapitest --object osapitest.sh --interval daily
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mosapitest[0m

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node schedule create --target _all --name osapitest --object osapitest.sh --interval daily
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mosapitest[0m
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [1mosapitest[0m

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
      "name": "osapitest",
      "status": "ok"
    }
  ]
}
```

With the default settings, each example does this on the host:

| Example                            | On the host                                                                 |
| ---------------------------------- | --------------------------------------------------------------------------- |
| An entry with a cron expression    | writes `/etc/cron.d/osapitest`: `0 2 * * * root /usr/local/bin/backup`      |
| A script that runs daily           | writes `/etc/cron.daily/osapitest`: `#!/bin/sh\n/usr/local/bin/backup`      |
| A template, rendered with its vars | writes `/etc/cron.d/osapitest`: `0 2 * * * root /usr/local/bin/backup /srv` |

The command fails in these cases:

| Call                                                  | Flags                                                                                                                                                                                                                                                                                                               | Fails with              |
| ----------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------- |
| An object the store does not have                     | `--name osapitest --object nosuch --schedule "0 2 * * *"`                                                                                                                                                                                                                                                           | not found               |
| A template that reads a var the request does not give | `--name osapitest --object osapitest.tmpl --schedule "0 2 * * *" --content-type template`                                                                                                                                                                                                                           | render template         |
| A file that osapi did not deploy                      | `--name osapitest --object osapitest.cron --schedule "0 2 * * *"`                                                                                                                                                                                                                                                   | osapi did not deploy it |
| No object                                             | `--name osapitest --schedule "0 2 * * *"`                                                                                                                                                                                                                                                                           | Object                  |
| Neither a cron expression nor an interval             | `--name osapitest --object osapitest.cron`                                                                                                                                                                                                                                                                          | Schedule                |
| A cron expression and an interval                     | `--name osapitest --object osapitest.cron --schedule "0 2 * * *" --interval daily`                                                                                                                                                                                                                                  | Schedule                |
| A cron expression that does not parse                 | `--name osapitest --object osapitest.cron --schedule "every day"`                                                                                                                                                                                                                                                   | Schedule                |
| An interval that is not one                           | `--name osapitest --object osapitest.cron --interval yearly`                                                                                                                                                                                                                                                        | Interval                |
| A name with a slash                                   | `--name a/b --object osapitest.cron --schedule "0 2 * * *"`                                                                                                                                                                                                                                                         | Name                    |
| A name that is an option                              | `--name --now --object osapitest.cron --schedule "0 2 * * *"`                                                                                                                                                                                                                                                       | not a valid entry name  |
| A name over 64 characters                             | `--name aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa --object osapitest.cron --schedule "0 2 * * *"`                                                                                                                                                                                           | Name                    |
| An object name over 255 characters                    | `--name osapitest --object aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa --schedule "0 2 * * *"` | Object                  |
| A user that is not an account name                    | `--name osapitest --object osapitest.cron --schedule "0 2 * * *" --user "a b"`                                                                                                                                                                                                                                      | User                    |
| A content type that is not one                        | `--name osapitest --object osapitest.cron --schedule "0 2 * * *" --content-type zip`                                                                                                                                                                                                                                | ContentType             |

## Permission

It needs the `schedule:write` permission.
`osapi controller role get --name <role>` shows what a role grants.

## Flags

| Flag             | Description                                                                                                                          | Default |
| ---------------- | ------------------------------------------------------------------------------------------------------------------------------------ | ------- |
| `--content-type` | The content type: raw, or template to render the object with the agent's facts and vars; the object's own when empty (raw, template) |         |
| `--interval`     | The interval of a script in /etc/cron.daily and the like; give it or schedule (hourly, daily, weekly, monthly)                       |         |
| `--name`         | The entry, and its file name, such as backup (required)                                                                              |         |
| `--object`       | The object that holds the entry, from osapi client file upload (required)                                                            |         |
| `--schedule`     | A cron expression, such as \*/5 \* \* \* \*; give it or interval                                                                     |         |
| `-T, --target`   | Target: _all, _any, hostname, or label (group:web.dev)                                                                               | `_all`  |
| `--user`         | The user the entry runs as                                                                                                           |         |
| `--vars`         | The vars of a template, as key=value                                                                                                 |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
