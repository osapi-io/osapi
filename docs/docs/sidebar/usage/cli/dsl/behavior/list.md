# List

List the behaviors this binary has.

An internal behavior has no API or CLI of its own. Only other behaviors use it.

```bash
osapi dsl behavior list
```

## Examples

```bash
osapi dsl behavior list
osapi dsl behavior list --json
```

## Result

The command prints:

```ansi
$ osapi dsl behavior list
[1;38;2;0;213;255mBehaviors (8):[0m
  [38;2;10;143;176mNAME         [0m[38;2;10;143;176mUNDER  [0m[38;2;10;143;176mOF    [0m[38;2;10;143;176mMETHODS                            [0m[38;2;10;143;176mDESCRIPTION                                               [0m[38;2;10;143;176mINTERNAL[0m
  [1mhostname[0m                  get, update                        The hostname of the host
  [1mmanagedfile[0m               get, update, delete                The files that osapi writes and tracks                    yes
  [1mdns[0m                       get, update, delete                The DNS resolvers of an interface
  [1msource[0m       log          list                               The sources that write to the systemd journal
  [1msysctl[0m                    list, get, create, update, delete  Kernel parameters, written to sysctl.d files and applied
  [1muser[0m                      list, get, create, update, delete  The users of the host
  [1msshkey[0m              user  list, create, delete               The authorized SSH keys of a user
  [1mpassword[0m            user  update                             The password of a user
```

With `--json`, the command prints:

```json
{
  "items": [
    {
      "name": "hostname",
      "methods": ["get", "update"],
      "doc": "The hostname of the host"
    },
    {
      "name": "managedfile",
      "methods": ["get", "update", "delete"],
      "doc": "The files that osapi writes and tracks",
      "internal": true
    },
    {
      "name": "dns",
      "methods": ["get", "update", "delete"],
      "doc": "The DNS resolvers of an interface"
    },
    {
      "name": "source",
      "under": "log",
      "methods": ["list"],
      "doc": "The sources that write to the systemd journal"
    },
    {
      "name": "sysctl",
      "methods": ["list", "get", "create", "update", "delete"],
      "doc": "Kernel parameters, written to sysctl.d files and applied"
    },
    {
      "name": "user",
      "methods": ["list", "get", "create", "update", "delete"],
      "doc": "The users of the host"
    },
    {
      "name": "sshkey",
      "of": "user",
      "methods": ["list", "create", "delete"],
      "doc": "The authorized SSH keys of a user"
    },
    {
      "name": "password",
      "of": "user",
      "methods": ["update"],
      "doc": "The password of a user"
    }
  ]
}
```

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
