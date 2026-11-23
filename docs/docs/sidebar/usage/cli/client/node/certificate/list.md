# List

List the certificate authorities that the host trusts.

Lists each CA of the trust store through gohai: the distribution's and the local
ones on Linux, from the family's directories, and the system keychains on macOS,
each with its subject, fingerprint and expiry. A CA that osapi deployed reads by
the name that it was deployed with, not its file's osapi- name, and is managed.

```bash
osapi client node certificate list [flags]
```

## Examples

```bash
# A Debian host with a system CA
osapi client node certificate list --target web-01

# A CA that osapi deployed reads by its name
osapi client node certificate list --target web-01

# A host of the RHEL family
osapi client node certificate list --target web-01

# A CA that osapi deployed on SUSE
osapi client node certificate list --target web-01

# A Mac
osapi client node certificate list --target web-01
```

## Result

A Debian host with a system CA, on one host:

```ansi
$ osapi client node certificate list --target web-01
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME                  [0m[38;2;10;143;176mPATH                                                 [0m[38;2;10;143;176mSOURCE  [0m[38;2;10;143;176mENABLED  [0m[38;2;10;143;176mSUBJECT                      [0m[38;2;10;143;176mFINGERPRINT                                                       [0m[38;2;10;143;176mNOT AFTER             [0m[38;2;10;143;176mMANAGED  [0m[38;2;10;143;176mOBJECT[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mmozilla/Example_Root[0m  /usr/share/ca-certificates/mozilla/Example_Root.crt  system  yes      CN=osapi example CA,O=osapi  6513d3f115f766cdffa3922e5ee47c4cccc98475cbe6471c5d32ddd4104c2312  2126-10-29T23:38:01Z  no       -

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node certificate list --target _all
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME                  [0m[38;2;10;143;176mPATH                                                 [0m[38;2;10;143;176mSOURCE  [0m[38;2;10;143;176mENABLED  [0m[38;2;10;143;176mSUBJECT                      [0m[38;2;10;143;176mFINGERPRINT                                                       [0m[38;2;10;143;176mNOT AFTER             [0m[38;2;10;143;176mMANAGED  [0m[38;2;10;143;176mOBJECT[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mmozilla/Example_Root[0m  /usr/share/ca-certificates/mozilla/Example_Root.crt  system  yes      CN=osapi example CA,O=osapi  6513d3f115f766cdffa3922e5ee47c4cccc98475cbe6471c5d32ddd4104c2312  2126-10-29T23:38:01Z  no       -
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [1mmozilla/Example_Root[0m  /usr/share/ca-certificates/mozilla/Example_Root.crt  system  yes      CN=osapi example CA,O=osapi  6513d3f115f766cdffa3922e5ee47c4cccc98475cbe6471c5d32ddd4104c2312  2126-10-29T23:38:01Z  no       -

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
      "name": "mozilla/Example_Root",
      "not_after": "2126-10-29T23:38:01Z",
      "path": "/usr/share/ca-certificates/mozilla/Example_Root.crt",
      "source": "system",
      "status": "ok",
      "subject": "CN=osapi example CA,O=osapi"
    }
  ]
}
```

A CA that osapi deployed reads by its name, on one host:

```ansi
$ osapi client node certificate list --target web-01
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME                  [0m[38;2;10;143;176mPATH                                                    [0m[38;2;10;143;176mSOURCE  [0m[38;2;10;143;176mENABLED  [0m[38;2;10;143;176mSUBJECT                      [0m[38;2;10;143;176mFINGERPRINT                                                       [0m[38;2;10;143;176mNOT AFTER             [0m[38;2;10;143;176mMANAGED  [0m[38;2;10;143;176mOBJECT[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mmozilla/Example_Root[0m  /usr/share/ca-certificates/mozilla/Example_Root.crt     system  yes      CN=osapi example CA,O=osapi  6513d3f115f766cdffa3922e5ee47c4cccc98475cbe6471c5d32ddd4104c2312  2126-10-29T23:38:01Z  no       -
  [38;2;138;138;138mweb-01[0m    [38;2;103;234;147mok[0m      [1minternal-ca[0m           /usr/local/share/ca-certificates/osapi-internal-ca.crt  local   yes      CN=osapi example CA,O=osapi  6513d3f115f766cdffa3922e5ee47c4cccc98475cbe6471c5d32ddd4104c2312  2126-10-29T23:38:01Z  yes      internal-ca.pem

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node certificate list --target _all
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME                  [0m[38;2;10;143;176mPATH                                                    [0m[38;2;10;143;176mSOURCE  [0m[38;2;10;143;176mENABLED  [0m[38;2;10;143;176mSUBJECT                      [0m[38;2;10;143;176mFINGERPRINT                                                       [0m[38;2;10;143;176mNOT AFTER             [0m[38;2;10;143;176mMANAGED  [0m[38;2;10;143;176mOBJECT[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mmozilla/Example_Root[0m  /usr/share/ca-certificates/mozilla/Example_Root.crt     system  yes      CN=osapi example CA,O=osapi  6513d3f115f766cdffa3922e5ee47c4cccc98475cbe6471c5d32ddd4104c2312  2126-10-29T23:38:01Z  no       -
  [38;2;138;138;138mweb-01[0m    [38;2;103;234;147mok[0m      [1minternal-ca[0m           /usr/local/share/ca-certificates/osapi-internal-ca.crt  local   yes      CN=osapi example CA,O=osapi  6513d3f115f766cdffa3922e5ee47c4cccc98475cbe6471c5d32ddd4104c2312  2126-10-29T23:38:01Z  yes      internal-ca.pem
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [1mmozilla/Example_Root[0m  /usr/share/ca-certificates/mozilla/Example_Root.crt     system  yes      CN=osapi example CA,O=osapi  6513d3f115f766cdffa3922e5ee47c4cccc98475cbe6471c5d32ddd4104c2312  2126-10-29T23:38:01Z  no       -
  [38;2;138;138;138mweb-02[0m    [38;2;103;234;147mok[0m      [1minternal-ca[0m           /usr/local/share/ca-certificates/osapi-internal-ca.crt  local   yes      CN=osapi example CA,O=osapi  6513d3f115f766cdffa3922e5ee47c4cccc98475cbe6471c5d32ddd4104c2312  2126-10-29T23:38:01Z  yes      internal-ca.pem

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
      "name": "mozilla/Example_Root",
      "not_after": "2126-10-29T23:38:01Z",
      "path": "/usr/share/ca-certificates/mozilla/Example_Root.crt",
      "source": "system",
      "status": "ok",
      "subject": "CN=osapi example CA,O=osapi"
    },
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

A host of the RHEL family, on one host:

```ansi
$ osapi client node certificate list --target web-01
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME                  [0m[38;2;10;143;176mPATH                                                     [0m[38;2;10;143;176mSOURCE  [0m[38;2;10;143;176mENABLED  [0m[38;2;10;143;176mSUBJECT                      [0m[38;2;10;143;176mFINGERPRINT                                                       [0m[38;2;10;143;176mNOT AFTER             [0m[38;2;10;143;176mMANAGED  [0m[38;2;10;143;176mOBJECT[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1manchors/Example_Root[0m  /usr/share/pki/ca-trust-source/anchors/Example_Root.pem  system  yes      CN=osapi example CA,O=osapi  6513d3f115f766cdffa3922e5ee47c4cccc98475cbe6471c5d32ddd4104c2312  2126-10-29T23:38:01Z  no       -

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node certificate list --target _all
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME                  [0m[38;2;10;143;176mPATH                                                     [0m[38;2;10;143;176mSOURCE  [0m[38;2;10;143;176mENABLED  [0m[38;2;10;143;176mSUBJECT                      [0m[38;2;10;143;176mFINGERPRINT                                                       [0m[38;2;10;143;176mNOT AFTER             [0m[38;2;10;143;176mMANAGED  [0m[38;2;10;143;176mOBJECT[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1manchors/Example_Root[0m  /usr/share/pki/ca-trust-source/anchors/Example_Root.pem  system  yes      CN=osapi example CA,O=osapi  6513d3f115f766cdffa3922e5ee47c4cccc98475cbe6471c5d32ddd4104c2312  2126-10-29T23:38:01Z  no       -
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [1manchors/Example_Root[0m  /usr/share/pki/ca-trust-source/anchors/Example_Root.pem  system  yes      CN=osapi example CA,O=osapi  6513d3f115f766cdffa3922e5ee47c4cccc98475cbe6471c5d32ddd4104c2312  2126-10-29T23:38:01Z  no       -

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
      "name": "anchors/Example_Root",
      "not_after": "2126-10-29T23:38:01Z",
      "path": "/usr/share/pki/ca-trust-source/anchors/Example_Root.pem",
      "source": "system",
      "status": "ok",
      "subject": "CN=osapi example CA,O=osapi"
    }
  ]
}
```

A CA that osapi deployed on SUSE, on one host:

```ansi
$ osapi client node certificate list --target web-01
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME         [0m[38;2;10;143;176mPATH                                          [0m[38;2;10;143;176mSOURCE  [0m[38;2;10;143;176mENABLED  [0m[38;2;10;143;176mSUBJECT                      [0m[38;2;10;143;176mFINGERPRINT                                                       [0m[38;2;10;143;176mNOT AFTER             [0m[38;2;10;143;176mMANAGED  [0m[38;2;10;143;176mOBJECT[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1minternal-ca[0m  /etc/pki/trust/anchors/osapi-internal-ca.crt  local   yes      CN=osapi example CA,O=osapi  6513d3f115f766cdffa3922e5ee47c4cccc98475cbe6471c5d32ddd4104c2312  2126-10-29T23:38:01Z  yes      internal-ca.pem

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node certificate list --target _all
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

A Mac, on one host:

```ansi
$ osapi client node certificate list --target web-01
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME              [0m[38;2;10;143;176mPATH                                                       [0m[38;2;10;143;176mSOURCE  [0m[38;2;10;143;176mENABLED  [0m[38;2;10;143;176mSUBJECT                      [0m[38;2;10;143;176mFINGERPRINT                                                       [0m[38;2;10;143;176mNOT AFTER             [0m[38;2;10;143;176mMANAGED  [0m[38;2;10;143;176mOBJECT[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mosapi example CA[0m  /System/Library/Keychains/SystemRootCertificates.keychain  system  yes      CN=osapi example CA,O=osapi  6513d3f115f766cdffa3922e5ee47c4cccc98475cbe6471c5d32ddd4104c2312  2126-10-29T23:38:01Z  no       -

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node certificate list --target _all
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

## Permission

It needs the `certificate:read` permission.
`osapi token roles get --name <role>` shows what a role grants.

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
