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

An installed package, on one host:

```ansi
$ osapi client node package get --target web-01 --name curl
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME  [0m[38;2;10;143;176mVERSION          [0m[38;2;10;143;176mARCHITECTURE  [0m[38;2;10;143;176mMANAGER  [0m[38;2;10;143;176mDESCRIPTION                                              [0m[38;2;10;143;176mSIZE[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mcurl[0m  8.5.0-2ubuntu10  amd64         dpkg     command line tool for transferring data with URL syntax  500.0 KB

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node package get --target _all --name curl
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME  [0m[38;2;10;143;176mVERSION          [0m[38;2;10;143;176mARCHITECTURE  [0m[38;2;10;143;176mMANAGER  [0m[38;2;10;143;176mDESCRIPTION                                              [0m[38;2;10;143;176mSIZE[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mcurl[0m  8.5.0-2ubuntu10  amd64         dpkg     command line tool for transferring data with URL syntax  500.0 KB
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [1mcurl[0m  8.5.0-2ubuntu10  amd64         dpkg     command line tool for transferring data with URL syntax  500.0 KB

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
      "architecture": "amd64",
      "description": "command line tool for transferring data with URL syntax",
      "hostname": "web-01",
      "manager": "dpkg",
      "name": "curl",
      "size": 512000,
      "status": "ok",
      "version": "8.5.0-2ubuntu10"
    }
  ]
}
```

An installed formula, on one host:

```ansi
$ osapi client node package get --target web-01 --name jq
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME  [0m[38;2;10;143;176mVERSION  [0m[38;2;10;143;176mARCHITECTURE  [0m[38;2;10;143;176mMANAGER  [0m[38;2;10;143;176mDESCRIPTION  [0m[38;2;10;143;176mSIZE[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mjq[0m    1.7.1                  brew

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node package get --target _all --name jq
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME  [0m[38;2;10;143;176mVERSION  [0m[38;2;10;143;176mARCHITECTURE  [0m[38;2;10;143;176mMANAGER  [0m[38;2;10;143;176mDESCRIPTION  [0m[38;2;10;143;176mSIZE[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mjq[0m    1.7.1                  brew
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [1mjq[0m    1.7.1                  brew

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
      "hostname": "web-01",
      "manager": "brew",
      "name": "jq",
      "status": "ok",
      "version": "1.7.1"
    }
  ]
}
```

An installed RPM, on one host:

```ansi
$ osapi client node package get --target web-01 --name curl
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME  [0m[38;2;10;143;176mVERSION       [0m[38;2;10;143;176mARCHITECTURE  [0m[38;2;10;143;176mMANAGER  [0m[38;2;10;143;176mDESCRIPTION  [0m[38;2;10;143;176mSIZE[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mcurl[0m  8.6.0-7.fc40  x86_64        rpm

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node package get --target _all --name curl
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME  [0m[38;2;10;143;176mVERSION       [0m[38;2;10;143;176mARCHITECTURE  [0m[38;2;10;143;176mMANAGER  [0m[38;2;10;143;176mDESCRIPTION  [0m[38;2;10;143;176mSIZE[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mcurl[0m  8.6.0-7.fc40  x86_64        rpm
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [1mcurl[0m  8.6.0-7.fc40  x86_64        rpm

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
      "architecture": "x86_64",
      "hostname": "web-01",
      "manager": "rpm",
      "name": "curl",
      "status": "ok",
      "version": "8.6.0-7.fc40"
    }
  ]
}
```

The command fails in these cases:

| Call                              | Flags                                                                                                                                                                                                                                                                     | Fails with               |
| --------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------ |
| A package that is not installed   | `--name tree`                                                                                                                                                                                                                                                             | not found                |
| A name that is an option          | `--name -oAPT::Get=x`                                                                                                                                                                                                                                                     | not a valid package name |
| A name longer than 255 characters | `--name aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa` | not a valid package name |

## Permission

It needs the `package:read` permission.
`osapi controller role get --name <role>` shows what a role grants.

## Flags

| Flag           | Description                                            | Default |
| -------------- | ------------------------------------------------------ | ------- |
| `--name`       | The package, such as curl (required)                   |         |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev) | `_all`  |

## Global flags

| Flag               | Description                                                                      | Default                  |
| ------------------ | -------------------------------------------------------------------------------- | ------------------------ |
| `-d, --debug`      | Enable or disable debug mode                                                     |                          |
| `-j, --json`       | Enable JSON output                                                               |                          |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml`  |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `https://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
