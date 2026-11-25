# Update

Change a user.

Changes what the input gives and the user does not have already, with usermod:
its shell, its home, which usermod moves the files to, its groups, which replace
the groups it has, and whether it is locked. The same values report changed:
false. A user that the host does not have is not found. It runs on the Debian
family only, and elsewhere reports skipped.

```bash
osapi client node user update [flags]
```

## Examples

```bash
# Set the shell
osapi client node user update --target web-01 --name osapitest --shell /bin/bash

# The shell it has
osapi client node user update --target web-01 --name osapitest --shell /bin/bash

# Move the home, set the groups and lock it
osapi client node user update --target web-01 --name osapitest --home /home/osapitest2 --groups root --lock locked

# Unlock it
osapi client node user update --target web-01 --name osapitest --lock unlocked
```

## Result

Set the shell, on one host:

```ansi
$ osapi client node user update --target web-01 --name osapitest --shell /bin/bash
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mosapitest[0m

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node user update --target _all --name osapitest --shell /bin/bash
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

The shell it has, on one host:

```ansi
$ osapi client node user update --target web-01 --name osapitest --shell /bin/bash
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mosapitest[0m

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node user update --target _all --name osapitest --shell /bin/bash
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

Move the home, set the groups and lock it, on one host:

```ansi
$ osapi client node user update --target web-01 --name osapitest --home /home/osapitest2 --groups root --lock locked
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mosapitest[0m

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node user update --target _all --name osapitest --home /home/osapitest2 --groups root --lock locked
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

Unlock it, on one host:

```ansi
$ osapi client node user update --target web-01 --name osapitest --lock unlocked
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mosapitest[0m

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node user update --target _all --name osapitest --lock unlocked
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

| Example                                   | On the host                                                   |
| ----------------------------------------- | ------------------------------------------------------------- |
| Set the shell                             | runs `usermod -s /bin/bash -- osapitest`                      |
| Move the home, set the groups and lock it | runs `usermod -d /home/osapitest2 -m -G root -L -- osapitest` |
| Unlock it                                 | runs `usermod -U -- osapitest`                                |

The command fails in these cases:

| Call                                  | Flags                                                     | Fails with   |
| ------------------------------------- | --------------------------------------------------------- | ------------ |
| A user that the host does not have    | `--name nobody`                                           | not found    |
| A name that is an option              | `--name -r`                                               | account_name |
| A name longer than 32 characters      | `--name aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa`                | max          |
| A shell that is not absolute          | `--name alice --shell bash`                               | startswith   |
| A shell with a line break             | `--name alice --shell "/bin/a\nb"`                        | no_linebreak |
| A home that is not absolute           | `--name alice --home home`                                | startswith   |
| A home with a line break              | `--name alice --home "/home/a\nb"`                        | no_linebreak |
| A group that is an option             | `--name alice --groups -r`                                | account_name |
| A group longer than 32 characters     | `--name alice --groups aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa` | max          |
| A lock that is not locked or unlocked | `--name alice --lock maybe`                               | enum         |

## Permission

It needs the `user:write` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Flags

| Flag           | Description                                                                                              | Default |
| -------------- | -------------------------------------------------------------------------------------------------------- | ------- |
| `--groups`     | The groups it is a member of, which replace the groups it has                                            |         |
| `--home`       | Its home directory, which usermod moves the files to                                                     |         |
| `--lock`       | Whether to refuse logins with its password: locked refuses them, unlocked allows them (locked, unlocked) |         |
| `--name`       | The user, such as alice (required)                                                                       |         |
| `--shell`      | Its login shell, such as /bin/bash                                                                       |         |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev)                                                   | `_all`  |

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `http://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
