# Get

Get one installed package.

Reads one package from the same list. A package that is not installed is not
found.

```bash
osapi client node package get [flags]
```

## Examples

```bash
# An installed package
osapi client node package get --target web-01 --name curl

# An installed formula
osapi client node package get --target web-01 --name jq

# An installed RPM
osapi client node package get --target web-01 --name curl
```

## Result

The command fails in these cases:

| Call                              | Flags                                                                                                                                                                                                                                                                     | Fails with               |
| --------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------ |
| A package that is not installed   | `--name tree`                                                                                                                                                                                                                                                             | not found                |
| A name that is an option          | `--name -oAPT::Get=x`                                                                                                                                                                                                                                                     | not a valid package name |
| A name longer than 255 characters | `--name aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa` | not a valid package name |

## Permission

It needs the `package:read` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Flags

| Flag           | Description                                            | Default |
| -------------- | ------------------------------------------------------ | ------- |
| `--name`       | The package, such as curl (required)                   |         |
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
