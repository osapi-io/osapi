# Update

Set the members of a group.

Sets the members with gpasswd -M, in place of the members that the group has.
The same members, in any order, do not change it and report changed: false. A
group that the host does not have is not found. It runs on the Debian family
only, and elsewhere reports skipped.

```bash
osapi client node group update [flags]
```

## Examples

```bash
# Set the members
osapi client node group update --target web-01 --name osapitest --members root

# The members it has
osapi client node group update --target web-01 --name developers --members bob --members alice
```

## Result

Set the members, on one host:

```ansi
$ osapi client node group update --target web-01 --name osapitest --members root
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mosapitest[0m

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node group update --target _all --name osapitest --members root
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

The members it has, on one host:

```ansi
$ osapi client node group update --target web-01 --name developers --members bob --members alice
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mdevelopers[0m

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node group update --target _all --name developers --members bob --members alice
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mdevelopers[0m
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [1mdevelopers[0m

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
      "name": "developers",
      "status": "ok"
    }
  ]
}
```

With the default settings, each example does this on the host:

| Example         | On the host                         |
| --------------- | ----------------------------------- |
| Set the members | runs `gpasswd -M root -- osapitest` |

The command fails in these cases:

| Call                                | Flags                                                           | Fails with   |
| ----------------------------------- | --------------------------------------------------------------- | ------------ |
| A group that the host does not have | `--name nobody`                                                 | not found    |
| A name that is an option            | `--name -r`                                                     | account_name |
| A name longer than 32 characters    | `--name aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa`                      | max          |
| A member that is an option          | `--name developers --members -r`                                | account_name |
| A member longer than 32 characters  | `--name developers --members aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa` | max          |

## Permission

It needs the `group:write` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Flags

| Flag           | Description                                                    | Default |
| -------------- | -------------------------------------------------------------- | ------- |
| `--members`    | Its members, which replace the members it has; none when empty |         |
| `--name`       | The group, such as developers (required)                       |         |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev)         | `_all`  |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
