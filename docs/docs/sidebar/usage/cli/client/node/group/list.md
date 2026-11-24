# List

List the groups of the host.

Lists each group with its ID and its members, by name, from /etc/group through
gohai, on every platform. The members are the users that /etc/group names; a
user whose primary group it is is not listed again.

```bash
osapi client node group list [flags]
```

## Examples

```bash
# A host with a group
osapi client node group list --target web-01
```

## Result

A host with a group, on one host:

```ansi
$ osapi client node group list --target web-01
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME        [0m[38;2;10;143;176mGID   [0m[38;2;10;143;176mMEMBERS[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mdevelopers[0m  1500  alice, bob
  [38;2;138;138;138mweb-01[0m    [38;2;103;234;147mok[0m      [1mroot[0m        0     none

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node group list --target _all
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME        [0m[38;2;10;143;176mGID   [0m[38;2;10;143;176mMEMBERS[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mdevelopers[0m  1500  alice, bob
  [38;2;138;138;138mweb-01[0m    [38;2;103;234;147mok[0m      [1mroot[0m        0     none
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [1mdevelopers[0m  1500  alice, bob
  [38;2;138;138;138mweb-02[0m    [38;2;103;234;147mok[0m      [1mroot[0m        0     none

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
    },
    {
      "gid": 0,
      "hostname": "web-01",
      "name": "root",
      "status": "ok"
    }
  ]
}
```

## Permission

It needs the `group:read` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Flags

| Flag           | Description                                            | Default |
| -------------- | ------------------------------------------------------ | ------- |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev) | `_all`  |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
