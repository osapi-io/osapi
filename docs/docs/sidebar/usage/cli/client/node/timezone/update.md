# Update

Set the timezone of the host.

Sets the timezone with timedatectl set-timezone. A host that already has the
timezone does not change and reports changed: false. It runs on the Debian
family only, and elsewhere reports skipped.

```bash
osapi client node timezone update [flags]
```

## Examples

```bash
# Set a new timezone
osapi client node timezone update --target web-01 --timezone America/New_York

# A host that already has the timezone
osapi client node timezone update --target web-01 --timezone UTC
```

## Result

With the default settings, each example does this on the host:

| Example            | On the host                                      |
| ------------------ | ------------------------------------------------ |
| Set a new timezone | runs `timedatectl set-timezone America/New_York` |

The command fails in these cases:

| Call                             | Flags                                                                          | Fails with                |
| -------------------------------- | ------------------------------------------------------------------------------ | ------------------------- |
| A name that is not a timezone    | `--timezone ../../etc/passwd`                                                  | not a valid timezone name |
| A name that starts with a dash   | `--timezone -UTC`                                                              | not a valid timezone name |
| A name longer than 64 characters | `--timezone aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa` | timezone                  |

## Permission

It needs the `timezone:write` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Flags

| Flag           | Description                                                               | Default |
| -------------- | ------------------------------------------------------------------------- | ------- |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev)                    | `_all`  |
| `--timezone`   | The IANA name of the timezone, such as UTC or America/New_York (required) |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
