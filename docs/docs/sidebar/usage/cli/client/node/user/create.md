# Create

Add a user.

Adds the user with useradd, with its home directory, and sets its password with
chpasswd when one is given. The controller hashes the password, so the job and
NATS have only the hash. A user that the host has already does not change and
reports changed: false. When chpasswd fails, the user is removed again with
userdel, so a second run adds it with its password. Its home directory goes with
it only when useradd made it, so a home that was there before stays. It runs on
the Debian family only, and elsewhere reports skipped.

```bash
osapi client node user create [flags]
```

## Examples

```bash
# Add a user
osapi client node user create --target web-01 --name osapitest

# Add a user with a shell, groups and a password
osapi client node user create --target web-01 --name osapitest --shell /bin/sh --groups root --password -

# Add a system user with its IDs and its home
osapi client node user create --target web-01 --name osapitest --uid 1999 --gid 100 --home /home/osapitest --system

# A user that the host has
osapi client node user create --target web-01 --name osapitest
```

## Result

Add a user, on one host:

```ansi
$ osapi client node user create --target web-01 --name osapitest
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mosapitest[0m

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node user create --target _all --name osapitest
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

Add a user with a shell, groups and a password, on one host:

```ansi
$ osapi client node user create --target web-01 --name osapitest --shell /bin/sh --groups root --password -
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mosapitest[0m

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node user create --target _all --name osapitest --shell /bin/sh --groups root --password -
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

Add a system user with its IDs and its home, on one host:

```ansi
$ osapi client node user create --target web-01 --name osapitest --uid 1999 --gid 100 --home /home/osapitest --system
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mosapitest[0m

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node user create --target _all --name osapitest --uid 1999 --gid 100 --home /home/osapitest --system
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

A user that the host has, on one host:

```ansi
$ osapi client node user create --target web-01 --name osapitest
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mosapitest[0m

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node user create --target _all --name osapitest
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

| Example                                        | On the host                                                                      |
| ---------------------------------------------- | -------------------------------------------------------------------------------- |
| Add a user                                     | runs `useradd --create-home -- osapitest`                                        |
| Add a user with a shell, groups and a password | runs `useradd --create-home -s /bin/sh -G root -- osapitest`; runs `chpasswd -e` |
| Add a system user with its IDs and its home    | runs `useradd --create-home -u 1999 -g 100 -d /home/osapitest -r -- osapitest`   |

The command fails in these cases:

| Call                              | Flags                                                         | Fails with   |
| --------------------------------- | ------------------------------------------------------------- | ------------ |
| A name that is an option          | `--name -r`                                                   | account_name |
| A name longer than 32 characters  | `--name aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa`                    | max          |
| A negative user ID                | `--name osapitest --uid -1`                                   | min          |
| A negative group ID               | `--name osapitest --gid -1`                                   | min          |
| A home that is not absolute       | `--name osapitest --home home`                                | startswith   |
| A home with a line break          | `--name osapitest --home $'/home/a\nb'`                       | no_linebreak |
| A shell that is not absolute      | `--name osapitest --shell bash`                               | startswith   |
| A shell with a line break         | `--name osapitest --shell $'/bin/a\nb'`                       | no_linebreak |
| A group that is an option         | `--name osapitest --groups -r`                                | account_name |
| A group longer than 32 characters | `--name osapitest --groups aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa` | max          |
| A password with a line break      | `--name osapitest --password -`                               | no_linebreak |
| A hash that is not a crypt hash   | `--name osapitest`                                            | crypt hash   |

## Permission

It needs the `user:write` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Flags

| Flag           | Description                                                                                                                                             | Default |
| -------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- |
| `--gid`        | Its primary group ID; a group of its own when it is 0                                                                                                   |         |
| `--groups`     | The groups it is a member of, besides its primary group                                                                                                 |         |
| `--home`       | Its home directory; /home/&lt;name&gt; when empty                                                                                                       |         |
| `--name`       | The user to add, such as alice (required)                                                                                                               |         |
| `--password`   | Its password, which osapi hashes on the controller; none when empty (a secret: secret://&lt;vault&gt;/&lt;key&gt;, or - to read it from standard input) |         |
| `--shell`      | Its login shell, such as /bin/bash; the system's default when empty                                                                                     |         |
| `--system`     | Add it as a system user, with an ID from the system range                                                                                               |         |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev)                                                                                                  | `_all`  |
| `--uid`        | Its numeric user ID; the system assigns one when it is 0                                                                                                |         |

## Global flags

| Flag               | Description                                                                      | Default                  |
| ------------------ | -------------------------------------------------------------------------------- | ------------------------ |
| `-d, --debug`      | Enable or disable debug mode                                                     |                          |
| `-j, --json`       | Enable JSON output                                                               |                          |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml`  |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `https://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
