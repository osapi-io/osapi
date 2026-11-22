# Create

Install a package.

Installs the package with apt-get install, at the version when one is given. A
package that is installed, at that version when one is given, does not change
and reports changed: false. It runs on the Debian family only, and elsewhere
reports skipped.

```bash
osapi client node package create [flags]
```

## Examples

```bash
# Install a package
osapi client node package create --target web-01 --name tree
```

## Result

Each example prints this:

Install a package:

```ansi
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mtree[0m

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

With the default settings, each example does this on the host:

| Example           | On the host                       |
| ----------------- | --------------------------------- |
| Install a package | runs `apt-get install -y -- tree` |

The command fails in these cases:

| Call                                 | Flags                                                                                                                                                                                                                                                                                    | Fails with                  |
| ------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------- |
| A name that is an option             | `--name -oAPT::Get=x`                                                                                                                                                                                                                                                                    | not a valid package name    |
| A name longer than 255 characters    | `--name aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa`                | not a valid package name    |
| A version that is an option          | `--name tree --version -x`                                                                                                                                                                                                                                                               | not a valid package version |
| A version longer than 255 characters | `--name tree --version 1111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111` | not a valid package version |

## Permission

It needs the `package:write` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Flags

| Flag           | Description                                                            | Default |
| -------------- | ---------------------------------------------------------------------- | ------- |
| `--name`       | The package to install, such as curl (required)                        |         |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev)                 | `_all`  |
| `--version`    | The version to install, such as 8.5.0-2ubuntu10; the newest when empty |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
