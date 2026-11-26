# List

List the users of the host.

Lists each user with its IDs, home, shell, groups and whether it is locked,
through gohai: from /etc/passwd, /etc/group and /etc/shadow on Linux, and from
the directory service with dscl on macOS. The groups are those that list the
user as a member, besides its primary group.

```bash
osapi client node user list [flags]
```

## Examples

```bash
# A host with users
osapi client node user list --target web-01
```

## Result

A host with users, on one host:

```ansi
$ osapi client node user list --target web-01
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME   [0m[38;2;10;143;176mUID   [0m[38;2;10;143;176mGID   [0m[38;2;10;143;176mHOME         [0m[38;2;10;143;176mSHELL      [0m[38;2;10;143;176mGROUPS      [0m[38;2;10;143;176mLOCKED[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1malice[0m  1000  1000  /home/alice  /bin/bash  developers  no
  [38;2;138;138;138mweb-01[0m    [38;2;103;234;147mok[0m      [1mroot[0m   0     0     /root        /bin/bash  none        yes

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node user list --target _all
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME   [0m[38;2;10;143;176mUID   [0m[38;2;10;143;176mGID   [0m[38;2;10;143;176mHOME         [0m[38;2;10;143;176mSHELL      [0m[38;2;10;143;176mGROUPS      [0m[38;2;10;143;176mLOCKED[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1malice[0m  1000  1000  /home/alice  /bin/bash  developers  no
  [38;2;138;138;138mweb-01[0m    [38;2;103;234;147mok[0m      [1mroot[0m   0     0     /root        /bin/bash  none        yes
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [1malice[0m  1000  1000  /home/alice  /bin/bash  developers  no
  [38;2;138;138;138mweb-02[0m    [38;2;103;234;147mok[0m      [1mroot[0m   0     0     /root        /bin/bash  none        yes

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
      "gid": 1000,
      "groups": ["developers"],
      "home": "/home/alice",
      "hostname": "web-01",
      "locked": false,
      "name": "alice",
      "shell": "/bin/bash",
      "status": "ok",
      "uid": 1000
    },
    {
      "gid": 0,
      "home": "/root",
      "hostname": "web-01",
      "locked": true,
      "name": "root",
      "shell": "/bin/bash",
      "status": "ok",
      "uid": 0
    }
  ]
}
```

## Permission

It needs the `user:read` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Flags

| Flag           | Description                                            | Default |
| -------------- | ------------------------------------------------------ | ------- |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev) | `_all`  |

## Global flags

| Flag               | Description                                                                      | Default                  |
| ------------------ | -------------------------------------------------------------------------------- | ------------------------ |
| `-d, --debug`      | Enable or disable debug mode                                                     |                          |
| `-j, --json`       | Enable JSON output                                                               |                          |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml`  |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `https://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
