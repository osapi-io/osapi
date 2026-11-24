# List

List the authorized SSH keys of a user.

Reads ~/.ssh/authorized_keys of the user through gohai, with the type, the
SHA256 fingerprint, the comment and the options of each key. A user with no file
has no keys. A user that the host does not have is not found.

```bash
osapi client node user sshkey list [flags]
```

## Examples

```bash
# A user with keys
osapi client node user sshkey list --target web-01 --name osapitest

# A user with no keys
osapi client node user sshkey list --target web-01 --name osapitest

# A host key that gohai cannot read
osapi client node user sshkey list --target web-01 --name osapitest
```

## Result

A user with keys, on one host:

```ansi
$ osapi client node user sshkey list --target web-01 --name osapitest
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mTYPE         [0m[38;2;10;143;176mFINGERPRINT                                         [0m[38;2;10;143;176mCOMMENT            [0m[38;2;10;143;176mOPTIONS[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mssh-ed25519[0m  SHA256:mnpIO1El26FsWnOeLzppADcNXgXRNHh9MPNPLPvSCwk  osapitest@example
  [38;2;138;138;138mweb-01[0m    [38;2;103;234;147mok[0m      [1mssh-ed25519[0m  SHA256:a5CeY6Lh3ZXkJOqGyuwb24kQeVjgag1gc9ujzv++nhE  other@example      from="10.0.0.0/8"

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node user sshkey list --target _all --name osapitest
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mTYPE         [0m[38;2;10;143;176mFINGERPRINT                                         [0m[38;2;10;143;176mCOMMENT            [0m[38;2;10;143;176mOPTIONS[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mssh-ed25519[0m  SHA256:mnpIO1El26FsWnOeLzppADcNXgXRNHh9MPNPLPvSCwk  osapitest@example
  [38;2;138;138;138mweb-01[0m    [38;2;103;234;147mok[0m      [1mssh-ed25519[0m  SHA256:a5CeY6Lh3ZXkJOqGyuwb24kQeVjgag1gc9ujzv++nhE  other@example      from="10.0.0.0/8"
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [1mssh-ed25519[0m  SHA256:mnpIO1El26FsWnOeLzppADcNXgXRNHh9MPNPLPvSCwk  osapitest@example
  [38;2;138;138;138mweb-02[0m    [38;2;103;234;147mok[0m      [1mssh-ed25519[0m  SHA256:a5CeY6Lh3ZXkJOqGyuwb24kQeVjgag1gc9ujzv++nhE  other@example      from="10.0.0.0/8"

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
      "comment": "osapitest@example",
      "fingerprint": "SHA256:mnpIO1El26FsWnOeLzppADcNXgXRNHh9MPNPLPvSCwk",
      "hostname": "web-01",
      "status": "ok",
      "type": "ssh-ed25519"
    },
    {
      "comment": "other@example",
      "fingerprint": "SHA256:a5CeY6Lh3ZXkJOqGyuwb24kQeVjgag1gc9ujzv++nhE",
      "hostname": "web-01",
      "options": "from=\"10.0.0.0/8\"",
      "status": "ok",
      "type": "ssh-ed25519"
    }
  ]
}
```

A user with no keys, on one host:

```ansi
$ osapi client node user sshkey list --target web-01 --name osapitest
[1;38;2;0;213;255mResults:[0m
  [38;2;138;138;138mNone.[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node user sshkey list --target _all --name osapitest
[1;38;2;0;213;255mResults:[0m

  1 host: [38;5;222m1 skipped[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m

[1;38;2;255;0;255mDetails:[0m
  [1mmac-01[0m  [38;5;222mskipped[0m  operation not supported on this OS family
```

With `--json`, the API's response:

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "results": []
}
```

A host key that gohai cannot read, on one host:

```ansi
$ osapi client node user sshkey list --target web-01 --name osapitest
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mTYPE         [0m[38;2;10;143;176mFINGERPRINT                                         [0m[38;2;10;143;176mCOMMENT            [0m[38;2;10;143;176mOPTIONS[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mssh-ed25519[0m  SHA256:mnpIO1El26FsWnOeLzppADcNXgXRNHh9MPNPLPvSCwk  osapitest@example

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node user sshkey list --target _all --name osapitest
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mTYPE         [0m[38;2;10;143;176mFINGERPRINT                                         [0m[38;2;10;143;176mCOMMENT            [0m[38;2;10;143;176mOPTIONS[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mssh-ed25519[0m  SHA256:mnpIO1El26FsWnOeLzppADcNXgXRNHh9MPNPLPvSCwk  osapitest@example
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [1mssh-ed25519[0m  SHA256:mnpIO1El26FsWnOeLzppADcNXgXRNHh9MPNPLPvSCwk  osapitest@example

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
      "comment": "osapitest@example",
      "fingerprint": "SHA256:mnpIO1El26FsWnOeLzppADcNXgXRNHh9MPNPLPvSCwk",
      "hostname": "web-01",
      "status": "ok",
      "type": "ssh-ed25519"
    }
  ]
}
```

The command fails in these cases:

| Call                               | Flags                                      | Fails with |
| ---------------------------------- | ------------------------------------------ | ---------- |
| A user that the host does not have | `--name nobody`                            | not found  |
| A name that is not a user name     | `--name "Bad;name"`                        | Name       |
| A name over 32 characters          | `--name aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa` | max        |

## Permission

It needs the `user:read` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Flags

| Flag           | Description                                            | Default |
| -------------- | ------------------------------------------------------ | ------- |
| `--name`       | The user, such as alice (required)                     |         |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev) | `_all`  |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../../output.md) tells how the result is printed, and what
`--json` gives.
