# Create

Add an authorized SSH key to a user.

Adds the key line to the end of ~/.ssh/authorized_keys of the user, and makes
the directory and the file if they are missing, with modes 700 and 600, owned by
the user. A key that the file has already, by its fingerprint, does not change
it. A line with no public key is refused. A user that the host does not have is
not found. It runs on the Debian family only, and elsewhere reports skipped.

```bash
osapi client node user sshkey create [flags]
```

## Examples

```bash
# Add a key
osapi client node user sshkey create --target web-01 --name osapitest --key 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJl85dcLw7lIdX+IyXEXT00CTCezv/2E1jvolleleo6Z osapitest@example'

# Add a second key
osapi client node user sshkey create --target web-01 --name osapitest --key 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGLPeEZQDbhOrSIx6gISjxXIcREV3rwL517pX6m9G5jc other@example'

# A key after a last line with no line break
osapi client node user sshkey create --target web-01 --name osapitest --key 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGLPeEZQDbhOrSIx6gISjxXIcREV3rwL517pX6m9G5jc other@example'

# A key the user has
osapi client node user sshkey create --target web-01 --name osapitest --key 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJl85dcLw7lIdX+IyXEXT00CTCezv/2E1jvolleleo6Z osapitest@example'
```

## Result

Add a key, on one host:

```ansi
$ osapi client node user sshkey create --target web-01 --name osapitest --key 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJl85dcLw7lIdX+IyXEXT00CTCezv/2E1jvolleleo6Z osapitest@example'
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mFINGERPRINT[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  SHA256:mnpIO1El26FsWnOeLzppADcNXgXRNHh9MPNPLPvSCwk

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node user sshkey create --target _all --name osapitest --key 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJl85dcLw7lIdX+IyXEXT00CTCezv/2E1jvolleleo6Z osapitest@example'
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

Add a second key, on one host:

```ansi
$ osapi client node user sshkey create --target web-01 --name osapitest --key 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGLPeEZQDbhOrSIx6gISjxXIcREV3rwL517pX6m9G5jc other@example'
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mFINGERPRINT[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  SHA256:a5CeY6Lh3ZXkJOqGyuwb24kQeVjgag1gc9ujzv++nhE

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node user sshkey create --target _all --name osapitest --key 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGLPeEZQDbhOrSIx6gISjxXIcREV3rwL517pX6m9G5jc other@example'
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mFINGERPRINT[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  SHA256:a5CeY6Lh3ZXkJOqGyuwb24kQeVjgag1gc9ujzv++nhE
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m  SHA256:a5CeY6Lh3ZXkJOqGyuwb24kQeVjgag1gc9ujzv++nhE

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
      "fingerprint": "SHA256:a5CeY6Lh3ZXkJOqGyuwb24kQeVjgag1gc9ujzv++nhE",
      "hostname": "web-01",
      "status": "ok"
    }
  ]
}
```

A key after a last line with no line break, on one host:

```ansi
$ osapi client node user sshkey create --target web-01 --name osapitest --key 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGLPeEZQDbhOrSIx6gISjxXIcREV3rwL517pX6m9G5jc other@example'
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mFINGERPRINT[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  SHA256:a5CeY6Lh3ZXkJOqGyuwb24kQeVjgag1gc9ujzv++nhE

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node user sshkey create --target _all --name osapitest --key 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGLPeEZQDbhOrSIx6gISjxXIcREV3rwL517pX6m9G5jc other@example'
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mFINGERPRINT[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  SHA256:a5CeY6Lh3ZXkJOqGyuwb24kQeVjgag1gc9ujzv++nhE
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m  SHA256:a5CeY6Lh3ZXkJOqGyuwb24kQeVjgag1gc9ujzv++nhE

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
      "fingerprint": "SHA256:a5CeY6Lh3ZXkJOqGyuwb24kQeVjgag1gc9ujzv++nhE",
      "hostname": "web-01",
      "status": "ok"
    }
  ]
}
```

A key the user has, on one host:

```ansi
$ osapi client node user sshkey create --target web-01 --name osapitest --key 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJl85dcLw7lIdX+IyXEXT00CTCezv/2E1jvolleleo6Z osapitest@example'
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mFINGERPRINT[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      SHA256:mnpIO1El26FsWnOeLzppADcNXgXRNHh9MPNPLPvSCwk

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node user sshkey create --target _all --name osapitest --key 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJl85dcLw7lIdX+IyXEXT00CTCezv/2E1jvolleleo6Z osapitest@example'
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

| Example                                    | On the host                                                                                                                                                                                                                                                                                         |
| ------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Add a key                                  | writes `/home/osapitest/.ssh/authorized_keys`: `ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJl85dcLw7lIdX+IyXEXT00CTCezv/2E1jvolleleo6Z osapitest@example`; runs `chown -R 1001:1001 /home/osapitest/.ssh`                                                                                                 |
| Add a second key                           | writes `/home/osapitest/.ssh/authorized_keys`: `ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJl85dcLw7lIdX+IyXEXT00CTCezv/2E1jvolleleo6Z osapitest@example\nssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGLPeEZQDbhOrSIx6gISjxXIcREV3rwL517pX6m9G5jc other@example`; runs `chown -R 1001:1001 /home/osapitest/.ssh` |
| A key after a last line with no line break | writes `/home/osapitest/.ssh/authorized_keys`: `ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJl85dcLw7lIdX+IyXEXT00CTCezv/2E1jvolleleo6Z osapitest@example\nssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGLPeEZQDbhOrSIx6gISjxXIcREV3rwL517pX6m9G5jc other@example`; runs `chown -R 1001:1001 /home/osapitest/.ssh` |

The command fails in these cases:

| Call                               | Flags                                                                                                                                                                                                                          | Fails with            |
| ---------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | --------------------- |
| A line with no public key          | `--name osapitest --key 'not a key'`                                                                                                                                                                                           | not an SSH public key |
| A key that is not base64           | `--name osapitest --key 'ssh-ed25519 !!!'`                                                                                                                                                                                     | not an SSH public key |
| A user that the host does not have | `--name nobody --key 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJl85dcLw7lIdX+IyXEXT00CTCezv/2E1jvolleleo6Z osapitest@example'`                                                                                                     | not found             |
| A name that is not a user name     | `--name 'Bad;name' --key 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJl85dcLw7lIdX+IyXEXT00CTCezv/2E1jvolleleo6Z osapitest@example'`                                                                                                 | Name                  |
| A name over 32 characters          | `--name aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa --key 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJl85dcLw7lIdX+IyXEXT00CTCezv/2E1jvolleleo6Z osapitest@example'`                                                                          | max                   |
| A key with a line break            | `--name osapitest --key $'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJl85dcLw7lIdX+IyXEXT00CTCezv/2E1jvolleleo6Z osapitest@example\nssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGLPeEZQDbhOrSIx6gISjxXIcREV3rwL517pX6m9G5jc other@example'` | no_linebreak          |

## Permission

It needs the `user:write` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Flags

| Flag           | Description                                                                                             | Default |
| -------------- | ------------------------------------------------------------------------------------------------------- | ------- |
| `--key`        | The public key line, such as ssh-ed25519 AAAA... alice@laptop, with options before it if any (required) |         |
| `--name`       | The user, such as alice (required)                                                                      |         |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev)                                                  | `_all`  |

## Global flags

| Flag               | Description                                                                      | Default                  |
| ------------------ | -------------------------------------------------------------------------------- | ------------------------ |
| `-d, --debug`      | Enable or disable debug mode                                                     |                          |
| `-j, --json`       | Enable JSON output                                                               |                          |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml`  |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `https://localhost:8080` |

[Output](../../../../output.md) tells how the result is printed, and what
`--json` gives.
