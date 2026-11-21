# Delete

Remove a package.

Removes the package with apt-get remove. A package that is not installed does
not change and reports changed: false. It runs on the Debian family only, and
elsewhere reports skipped.

```bash
osapi client node package delete [flags]
```

## Examples

```bash
# Remove a package
osapi client node package delete --target web-01 --name tree

# A package that is not installed
osapi client node package delete --target web-01 --name tree
```

## Result

With the default settings, each example does this on the host:

| Example          | On the host                      |
| ---------------- | -------------------------------- |
| Remove a package | runs `apt-get remove -y -- tree` |

The command fails in these cases:

| Call                              | Flags                                                                                                                                                                                                                                                                     | Fails with               |
| --------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------ |
| A name that is an option          | `--name -oAPT::Get=x`                                                                                                                                                                                                                                                     | not a valid package name |
| A name longer than 255 characters | `--name aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa` | not a valid package name |

## Permission

It needs the `package:write` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Flags

| Flag           | Description                                            | Default |
| -------------- | ------------------------------------------------------ | ------- |
| `--name`       | The package to remove, such as curl (required)         |         |
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
