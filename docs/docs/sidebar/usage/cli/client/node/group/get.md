# Get

Get one group.

Reads one group from the same list. A group that the host does not have is not
found.

```bash
osapi client node group get [flags]
```

## Examples

```bash
# A group
osapi client node group get --target web-01 --name developers
```

## Result

A group, on one host:

```ansi
$ osapi client node group get --target web-01 --name developers
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME        [0m[38;2;10;143;176mGID   [0m[38;2;10;143;176mMEMBERS[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mdevelopers[0m  1500  alice, bob

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node group get --target _all --name developers
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME        [0m[38;2;10;143;176mGID   [0m[38;2;10;143;176mMEMBERS[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mdevelopers[0m  1500  alice, bob
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [1mdevelopers[0m  1500  alice, bob

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
      "gid": 1500,
      "hostname": "web-01",
      "members": ["alice", "bob"],
      "name": "developers",
      "status": "ok"
    }
  ]
}
```

The command fails in these cases:

| Call                                | Flags                                      | Fails with   |
| ----------------------------------- | ------------------------------------------ | ------------ |
| A group that the host does not have | `--name nobody`                            | not found    |
| A name that is an option            | `--name -r`                                | account_name |
| A name longer than 32 characters    | `--name aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa` | max          |

## Permission

It needs the `group:read` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Flags

| Flag           | Description                                            | Default |
| -------------- | ------------------------------------------------------ | ------- |
| `--name`       | The group, such as developers (required)               |         |
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
