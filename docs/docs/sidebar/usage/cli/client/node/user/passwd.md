# Passwd

Set the password of a user.

Sets the password with chpasswd -e, from its hash on standard input. The
controller hashes the password, so the job and NATS have only the hash. Each run
makes a new hash, so it always reports changed: true. A user that the host does
not have is not found. It runs on the Debian family only, and elsewhere reports
skipped.

```bash
osapi client node user passwd [flags]
```

## Examples

```bash
# Set a password
osapi client node user passwd --target web-01 --name osapitest --password osapi-test-password
```

## Result

Set a password, on one host:

```ansi
$ osapi client node user passwd --target web-01 --name osapitest --password osapi-test-password
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mosapitest[0m

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node user passwd --target _all --name osapitest --password osapi-test-password
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

| Example        | On the host        |
| -------------- | ------------------ |
| Set a password | runs `chpasswd -e` |

The command fails in these cases:

| Call                               | Flags                                      | Fails with        |
| ---------------------------------- | ------------------------------------------ | ----------------- |
| No password                        | `--name alice`                             | give the password |
| A user that the host does not have | `--name nobody`                            | not found         |
| A hash that is not a crypt hash    | `--name alice`                             | crypt hash        |
| A name that is an option           | `--name -r`                                | account_name      |
| A name longer than 32 characters   | `--name aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa` | max               |
| A password with a line break       | `--name alice --password "a\nb"`           | no_linebreak      |

## Permission

It needs the `user:execute` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Flags

| Flag           | Description                                                                                           | Default |
| -------------- | ----------------------------------------------------------------------------------------------------- | ------- |
| `--name`       | The user, such as alice (required)                                                                    |         |
| `--password`   | The new password, which osapi hashes on the controller (secret: kept out of the audit log and traces) |         |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev)                                                | `_all`  |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
