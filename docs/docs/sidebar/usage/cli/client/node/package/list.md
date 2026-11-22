# List

List the packages installed on the host.

Lists each installed package with its version: dpkg on the Debian family, rpm on
the other Linux families, and Homebrew on macOS.

```bash
osapi client node package list [flags]
```

## Examples

```bash
# A Debian host
osapi client node package list --target web-01

# A Fedora host
osapi client node package list --target web-01

# A Mac
osapi client node package list --target web-01
```

## Result

Each example prints this:

A Debian host:

```ansi
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME  [0m[38;2;10;143;176mVERSION          [0m[38;2;10;143;176mARCHITECTURE  [0m[38;2;10;143;176mMANAGER  [0m[38;2;10;143;176mDESCRIPTION                                              [0m[38;2;10;143;176mSIZE[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mcurl[0m  8.5.0-2ubuntu10  amd64         dpkg     command line tool for transferring data with URL syntax  500.0 KB

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

A Fedora host:

```ansi
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME  [0m[38;2;10;143;176mVERSION       [0m[38;2;10;143;176mARCHITECTURE  [0m[38;2;10;143;176mMANAGER  [0m[38;2;10;143;176mDESCRIPTION                                      [0m[38;2;10;143;176mSIZE[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mcurl[0m  8.6.0-7.fc40  x86_64        rpm      A utility for getting files from remote servers  771.5 KB

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

A Mac:

```ansi
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME  [0m[38;2;10;143;176mVERSION  [0m[38;2;10;143;176mARCHITECTURE  [0m[38;2;10;143;176mMANAGER  [0m[38;2;10;143;176mDESCRIPTION  [0m[38;2;10;143;176mSIZE[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mjq[0m    1.7.1                  brew

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

## Permission

It needs the `package:read` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Flags

| Flag           | Description                                            | Default |
| -------------- | ------------------------------------------------------ | ------- |
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
