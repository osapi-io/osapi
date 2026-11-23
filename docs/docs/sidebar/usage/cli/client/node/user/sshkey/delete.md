# Delete

Remove an authorized SSH key of a user.

Removes each line of the key with the fingerprint from ~/.ssh/authorized_keys of
the user, and keeps the other lines as they are. A key that the file does not
have, a user with no file, or a user that the host does not have, changes
nothing. It runs on the Debian family only, and elsewhere reports skipped.

```bash
osapi client node user sshkey delete [flags]
```

## Examples

```bash
# Remove a key
osapi client node user sshkey delete --target web-01 --name osapitest --fingerprint SHA256:mnpIO1El26FsWnOeLzppADcNXgXRNHh9MPNPLPvSCwk

# A key the user does not have
osapi client node user sshkey delete --target web-01 --name osapitest --fingerprint SHA256:a5CeY6Lh3ZXkJOqGyuwb24kQeVjgag1gc9ujzv++nhE

# A user that the host does not have
osapi client node user sshkey delete --target web-01 --name nobody --fingerprint SHA256:mnpIO1El26FsWnOeLzppADcNXgXRNHh9MPNPLPvSCwk
```

## Result

Remove a key, on one host:

```ansi
$ osapi client node user sshkey delete --target web-01 --name osapitest --fingerprint SHA256:mnpIO1El26FsWnOeLzppADcNXgXRNHh9MPNPLPvSCwk
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mFINGERPRINT[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  SHA256:mnpIO1El26FsWnOeLzppADcNXgXRNHh9MPNPLPvSCwk

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node user sshkey delete --target _all --name osapitest --fingerprint SHA256:mnpIO1El26FsWnOeLzppADcNXgXRNHh9MPNPLPvSCwk
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mFINGERPRINT[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  SHA256:mnpIO1El26FsWnOeLzppADcNXgXRNHh9MPNPLPvSCwk
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m  SHA256:mnpIO1El26FsWnOeLzppADcNXgXRNHh9MPNPLPvSCwk

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
      "fingerprint": "SHA256:mnpIO1El26FsWnOeLzppADcNXgXRNHh9MPNPLPvSCwk",
      "hostname": "web-01",
      "status": "ok"
    }
  ]
}
```

A key the user does not have, on one host:

```ansi
$ osapi client node user sshkey delete --target web-01 --name osapitest --fingerprint SHA256:a5CeY6Lh3ZXkJOqGyuwb24kQeVjgag1gc9ujzv++nhE
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mFINGERPRINT[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      SHA256:a5CeY6Lh3ZXkJOqGyuwb24kQeVjgag1gc9ujzv++nhE

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node user sshkey delete --target _all --name osapitest --fingerprint SHA256:a5CeY6Lh3ZXkJOqGyuwb24kQeVjgag1gc9ujzv++nhE
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mFINGERPRINT[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      SHA256:a5CeY6Lh3ZXkJOqGyuwb24kQeVjgag1gc9ujzv++nhE
  [1mweb-02[0m    [38;2;103;234;147mok[0m      SHA256:a5CeY6Lh3ZXkJOqGyuwb24kQeVjgag1gc9ujzv++nhE

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
      "fingerprint": "SHA256:a5CeY6Lh3ZXkJOqGyuwb24kQeVjgag1gc9ujzv++nhE",
      "hostname": "web-01",
      "status": "ok"
    }
  ]
}
```

A user that the host does not have, on one host:

```ansi
$ osapi client node user sshkey delete --target web-01 --name nobody --fingerprint SHA256:mnpIO1El26FsWnOeLzppADcNXgXRNHh9MPNPLPvSCwk
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mFINGERPRINT[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      SHA256:mnpIO1El26FsWnOeLzppADcNXgXRNHh9MPNPLPvSCwk

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node user sshkey delete --target _all --name nobody --fingerprint SHA256:mnpIO1El26FsWnOeLzppADcNXgXRNHh9MPNPLPvSCwk
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mFINGERPRINT[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      SHA256:mnpIO1El26FsWnOeLzppADcNXgXRNHh9MPNPLPvSCwk
  [1mweb-02[0m    [38;2;103;234;147mok[0m      SHA256:mnpIO1El26FsWnOeLzppADcNXgXRNHh9MPNPLPvSCwk

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
      "fingerprint": "SHA256:mnpIO1El26FsWnOeLzppADcNXgXRNHh9MPNPLPvSCwk",
      "hostname": "web-01",
      "status": "ok"
    }
  ]
}
```

With the default settings, each example does this on the host:

| Example      | On the host                                                                                                                                     |
| ------------ | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| Remove a key | writes `/home/osapitest/.ssh/authorized_keys`: `ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGLPeEZQDbhOrSIx6gISjxXIcREV3rwL517pX6m9G5jc other@example` |

The command fails in these cases:

| Call                             | Flags                                                                                                                                     | Fails with   |
| -------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- | ------------ |
| A fingerprint that is not SHA256 | `--name osapitest --fingerprint MD5:aa`                                                                                                   | Fingerprint  |
| A fingerprint with a line break  | `--name osapitest --fingerprint "SHA256:mnpIO1El26FsWnOeLzppADcNXgXRNHh9MPNPLPvSCwk\nSHA256:a5CeY6Lh3ZXkJOqGyuwb24kQeVjgag1gc9ujzv++nhE"` | no_linebreak |
| A name that is not a user name   | `--name "Bad;name" --fingerprint SHA256:mnpIO1El26FsWnOeLzppADcNXgXRNHh9MPNPLPvSCwk`                                                      | Name         |
| A name over 32 characters        | `--name aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa --fingerprint SHA256:mnpIO1El26FsWnOeLzppADcNXgXRNHh9MPNPLPvSCwk`                               | max          |

## Permission

It needs the `user:write` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Flags

| Flag            | Description                                                    | Default |
| --------------- | -------------------------------------------------------------- | ------- |
| `--fingerprint` | The SHA256 fingerprint of the key, as list shows it (required) |         |
| `--name`        | The user, such as alice (required)                             |         |
| `-T, --target`  | Target: _all, _any, hostname, or label (group:web.dev)         | `_all`  |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../../output.md) tells how the result is printed, and what
`--json` gives.
