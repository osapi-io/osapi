# Get

Get one certificate authority.

Reads one CA from the same list, by the name that list gives it. A CA that the
host does not trust is not found. A name with a slash, such as a system CA of
the Debian family, is in list only.

```bash
osapi client node certificate get [flags]
```

## Examples

```bash
# A CA that osapi deployed, by the name it was deployed with
osapi client node certificate get --target web-01 --name internal-ca

# A CA that osapi deployed on the RHEL family
osapi client node certificate get --target web-01 --name internal-ca

# A CA that osapi deployed on SUSE
osapi client node certificate get --target web-01 --name internal-ca

# A CA of a Mac
osapi client node certificate get --target web-01 --name 'osapi example CA'
```

## Result

A CA that osapi deployed, by the name it was deployed with, on one host:

```ansi
$ osapi client node certificate get --target web-01 --name internal-ca
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME         [0m[38;2;10;143;176mPATH                                                    [0m[38;2;10;143;176mSOURCE  [0m[38;2;10;143;176mENABLED  [0m[38;2;10;143;176mSUBJECT                      [0m[38;2;10;143;176mFINGERPRINT                                                       [0m[38;2;10;143;176mNOT AFTER             [0m[38;2;10;143;176mMANAGED  [0m[38;2;10;143;176mOBJECT[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1minternal-ca[0m  /usr/local/share/ca-certificates/osapi-internal-ca.crt  local   yes      CN=osapi example CA,O=osapi  6513d3f115f766cdffa3922e5ee47c4cccc98475cbe6471c5d32ddd4104c2312  2126-10-29T23:38:01Z  yes      internal-ca.pem

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node certificate get --target _all --name internal-ca
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME         [0m[38;2;10;143;176mPATH                                                    [0m[38;2;10;143;176mSOURCE  [0m[38;2;10;143;176mENABLED  [0m[38;2;10;143;176mSUBJECT                      [0m[38;2;10;143;176mFINGERPRINT                                                       [0m[38;2;10;143;176mNOT AFTER             [0m[38;2;10;143;176mMANAGED  [0m[38;2;10;143;176mOBJECT[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1minternal-ca[0m  /usr/local/share/ca-certificates/osapi-internal-ca.crt  local   yes      CN=osapi example CA,O=osapi  6513d3f115f766cdffa3922e5ee47c4cccc98475cbe6471c5d32ddd4104c2312  2126-10-29T23:38:01Z  yes      internal-ca.pem
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [1minternal-ca[0m  /usr/local/share/ca-certificates/osapi-internal-ca.crt  local   yes      CN=osapi example CA,O=osapi  6513d3f115f766cdffa3922e5ee47c4cccc98475cbe6471c5d32ddd4104c2312  2126-10-29T23:38:01Z  yes      internal-ca.pem

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
      "enabled": true,
      "fingerprint": "6513d3f115f766cdffa3922e5ee47c4cccc98475cbe6471c5d32ddd4104c2312",
      "hostname": "web-01",
      "managed": true,
      "name": "internal-ca",
      "not_after": "2126-10-29T23:38:01Z",
      "object": "internal-ca.pem",
      "path": "/usr/local/share/ca-certificates/osapi-internal-ca.crt",
      "source": "local",
      "status": "ok",
      "subject": "CN=osapi example CA,O=osapi"
    }
  ]
}
```

A CA that osapi deployed on the RHEL family, on one host:

```ansi
$ osapi client node certificate get --target web-01 --name internal-ca
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME         [0m[38;2;10;143;176mPATH                                                    [0m[38;2;10;143;176mSOURCE  [0m[38;2;10;143;176mENABLED  [0m[38;2;10;143;176mSUBJECT                      [0m[38;2;10;143;176mFINGERPRINT                                                       [0m[38;2;10;143;176mNOT AFTER             [0m[38;2;10;143;176mMANAGED  [0m[38;2;10;143;176mOBJECT[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1minternal-ca[0m  /etc/pki/ca-trust/source/anchors/osapi-internal-ca.crt  local   yes      CN=osapi example CA,O=osapi  6513d3f115f766cdffa3922e5ee47c4cccc98475cbe6471c5d32ddd4104c2312  2126-10-29T23:38:01Z  yes      internal-ca.pem

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node certificate get --target _all --name internal-ca
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME         [0m[38;2;10;143;176mPATH                                                    [0m[38;2;10;143;176mSOURCE  [0m[38;2;10;143;176mENABLED  [0m[38;2;10;143;176mSUBJECT                      [0m[38;2;10;143;176mFINGERPRINT                                                       [0m[38;2;10;143;176mNOT AFTER             [0m[38;2;10;143;176mMANAGED  [0m[38;2;10;143;176mOBJECT[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1minternal-ca[0m  /etc/pki/ca-trust/source/anchors/osapi-internal-ca.crt  local   yes      CN=osapi example CA,O=osapi  6513d3f115f766cdffa3922e5ee47c4cccc98475cbe6471c5d32ddd4104c2312  2126-10-29T23:38:01Z  yes      internal-ca.pem
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [1minternal-ca[0m  /etc/pki/ca-trust/source/anchors/osapi-internal-ca.crt  local   yes      CN=osapi example CA,O=osapi  6513d3f115f766cdffa3922e5ee47c4cccc98475cbe6471c5d32ddd4104c2312  2126-10-29T23:38:01Z  yes      internal-ca.pem

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
      "enabled": true,
      "fingerprint": "6513d3f115f766cdffa3922e5ee47c4cccc98475cbe6471c5d32ddd4104c2312",
      "hostname": "web-01",
      "managed": true,
      "name": "internal-ca",
      "not_after": "2126-10-29T23:38:01Z",
      "object": "internal-ca.pem",
      "path": "/etc/pki/ca-trust/source/anchors/osapi-internal-ca.crt",
      "source": "local",
      "status": "ok",
      "subject": "CN=osapi example CA,O=osapi"
    }
  ]
}
```

A CA that osapi deployed on SUSE, on one host:

```ansi
$ osapi client node certificate get --target web-01 --name internal-ca
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME         [0m[38;2;10;143;176mPATH                                          [0m[38;2;10;143;176mSOURCE  [0m[38;2;10;143;176mENABLED  [0m[38;2;10;143;176mSUBJECT                      [0m[38;2;10;143;176mFINGERPRINT                                                       [0m[38;2;10;143;176mNOT AFTER             [0m[38;2;10;143;176mMANAGED  [0m[38;2;10;143;176mOBJECT[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1minternal-ca[0m  /etc/pki/trust/anchors/osapi-internal-ca.crt  local   yes      CN=osapi example CA,O=osapi  6513d3f115f766cdffa3922e5ee47c4cccc98475cbe6471c5d32ddd4104c2312  2126-10-29T23:38:01Z  yes      internal-ca.pem

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node certificate get --target _all --name internal-ca
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME         [0m[38;2;10;143;176mPATH                                          [0m[38;2;10;143;176mSOURCE  [0m[38;2;10;143;176mENABLED  [0m[38;2;10;143;176mSUBJECT                      [0m[38;2;10;143;176mFINGERPRINT                                                       [0m[38;2;10;143;176mNOT AFTER             [0m[38;2;10;143;176mMANAGED  [0m[38;2;10;143;176mOBJECT[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1minternal-ca[0m  /etc/pki/trust/anchors/osapi-internal-ca.crt  local   yes      CN=osapi example CA,O=osapi  6513d3f115f766cdffa3922e5ee47c4cccc98475cbe6471c5d32ddd4104c2312  2126-10-29T23:38:01Z  yes      internal-ca.pem
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [1minternal-ca[0m  /etc/pki/trust/anchors/osapi-internal-ca.crt  local   yes      CN=osapi example CA,O=osapi  6513d3f115f766cdffa3922e5ee47c4cccc98475cbe6471c5d32ddd4104c2312  2126-10-29T23:38:01Z  yes      internal-ca.pem

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
      "enabled": true,
      "fingerprint": "6513d3f115f766cdffa3922e5ee47c4cccc98475cbe6471c5d32ddd4104c2312",
      "hostname": "web-01",
      "managed": true,
      "name": "internal-ca",
      "not_after": "2126-10-29T23:38:01Z",
      "object": "internal-ca.pem",
      "path": "/etc/pki/trust/anchors/osapi-internal-ca.crt",
      "source": "local",
      "status": "ok",
      "subject": "CN=osapi example CA,O=osapi"
    }
  ]
}
```

A CA of a Mac, on one host:

```ansi
$ osapi client node certificate get --target web-01 --name 'osapi example CA'
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME              [0m[38;2;10;143;176mPATH                                                       [0m[38;2;10;143;176mSOURCE  [0m[38;2;10;143;176mENABLED  [0m[38;2;10;143;176mSUBJECT                      [0m[38;2;10;143;176mFINGERPRINT                                                       [0m[38;2;10;143;176mNOT AFTER             [0m[38;2;10;143;176mMANAGED  [0m[38;2;10;143;176mOBJECT[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mosapi example CA[0m  /System/Library/Keychains/SystemRootCertificates.keychain  system  yes      CN=osapi example CA,O=osapi  6513d3f115f766cdffa3922e5ee47c4cccc98475cbe6471c5d32ddd4104c2312  2126-10-29T23:38:01Z  no       -

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node certificate get --target _all --name 'osapi example CA'
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME              [0m[38;2;10;143;176mPATH                                                       [0m[38;2;10;143;176mSOURCE  [0m[38;2;10;143;176mENABLED  [0m[38;2;10;143;176mSUBJECT                      [0m[38;2;10;143;176mFINGERPRINT                                                       [0m[38;2;10;143;176mNOT AFTER             [0m[38;2;10;143;176mMANAGED  [0m[38;2;10;143;176mOBJECT[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mosapi example CA[0m  /System/Library/Keychains/SystemRootCertificates.keychain  system  yes      CN=osapi example CA,O=osapi  6513d3f115f766cdffa3922e5ee47c4cccc98475cbe6471c5d32ddd4104c2312  2126-10-29T23:38:01Z  no       -
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [1mosapi example CA[0m  /System/Library/Keychains/SystemRootCertificates.keychain  system  yes      CN=osapi example CA,O=osapi  6513d3f115f766cdffa3922e5ee47c4cccc98475cbe6471c5d32ddd4104c2312  2126-10-29T23:38:01Z  no       -

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
      "enabled": true,
      "fingerprint": "6513d3f115f766cdffa3922e5ee47c4cccc98475cbe6471c5d32ddd4104c2312",
      "hostname": "web-01",
      "managed": false,
      "name": "osapi example CA",
      "not_after": "2126-10-29T23:38:01Z",
      "path": "/System/Library/Keychains/SystemRootCertificates.keychain",
      "source": "system",
      "status": "ok",
      "subject": "CN=osapi example CA,O=osapi"
    }
  ]
}
```

The command fails in these cases:

| Call                                        | Flags                                                                                                                                                                                                                                                                     | Fails with |
| ------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------- |
| A CA that the host does not trust           | `--name nothing`                                                                                                                                                                                                                                                          | not found  |
| A name with a %, which an encoded slash has | `--name mozilla%2FExample_Root`                                                                                                                                                                                                                                           | ca_name    |
| A name of 256 characters                    | `--name aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa` | max        |

## Permission

It needs the `certificate:read` permission.
`osapi controller role get --name <role>` shows what a role grants.

## Flags

| Flag           | Description                                                                                                                                                                                    | Default |
| -------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- |
| `--name`       | The CA, as list names it, such as internal-ca, or ISRG Root X1 on macOS; a name with a slash, such as mozilla/ISRG_Root_X1, is in list only, because a path segment cannot have one (required) |         |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev)                                                                                                                                         | `_all`  |

## Global flags

| Flag               | Description                                                                      | Default                  |
| ------------------ | -------------------------------------------------------------------------------- | ------------------------ |
| `-d, --debug`      | Enable or disable debug mode                                                     |                          |
| `-j, --json`       | Enable JSON output                                                               |                          |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml`  |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `https://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
