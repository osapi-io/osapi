# Create

Add a certificate authority to the trust store.

Writes the PEM of the object as a file with the osapi- prefix, such as
osapi-internal-ca.crt for internal-ca, in the family's local directory:
/usr/local/share/ca-certificates on the Debian family, /etc/pki/trust/anchors on
SUSE, /etc/pki/ca-trust/source/anchors on RHEL and Fedora. Then it rebuilds the
trust store: with update-ca-certificates on the Debian family and SUSE, and with
update-ca-trust on the others. Alpine reports skipped. A CA that osapi deployed
already with the same content does not change and reports changed: false. A file
of that name that osapi did not write is refused. macOS reports skipped.

```bash
osapi client node certificate create [flags]
```

## Examples

```bash
# Add a CA
osapi client node certificate create --target web-01 --name internal-ca --object internal-ca.pem

# Add a CA on the RHEL family
osapi client node certificate create --target web-01 --name internal-ca --object internal-ca.pem

# Add a CA on SUSE
osapi client node certificate create --target web-01 --name internal-ca --object internal-ca.pem

# A CA that osapi deployed with the same content
osapi client node certificate create --target web-01 --name internal-ca --object internal-ca.pem
```

## Result

Add a CA, on one host:

```ansi
$ osapi client node certificate create --target web-01 --name internal-ca --object internal-ca.pem
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1minternal-ca[0m

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node certificate create --target _all --name internal-ca --object internal-ca.pem
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1minternal-ca[0m
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m  [1minternal-ca[0m

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
      "name": "internal-ca",
      "status": "ok"
    }
  ]
}
```

Add a CA on the RHEL family, on one host:

```ansi
$ osapi client node certificate create --target web-01 --name internal-ca --object internal-ca.pem
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1minternal-ca[0m

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node certificate create --target _all --name internal-ca --object internal-ca.pem
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1minternal-ca[0m
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m  [1minternal-ca[0m

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
      "name": "internal-ca",
      "status": "ok"
    }
  ]
}
```

Add a CA on SUSE, on one host:

```ansi
$ osapi client node certificate create --target web-01 --name internal-ca --object internal-ca.pem
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1minternal-ca[0m

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node certificate create --target _all --name internal-ca --object internal-ca.pem
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1minternal-ca[0m
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m  [1minternal-ca[0m

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
      "name": "internal-ca",
      "status": "ok"
    }
  ]
}
```

A CA that osapi deployed with the same content, on one host:

```ansi
$ osapi client node certificate create --target web-01 --name internal-ca --object internal-ca.pem
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1minternal-ca[0m

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node certificate create --target _all --name internal-ca --object internal-ca.pem
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1minternal-ca[0m
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [1minternal-ca[0m

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
      "hostname": "web-01",
      "name": "internal-ca",
      "status": "ok"
    }
  ]
}
```

With the default settings, each example does this on the host:

| Example                     | On the host                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| --------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Add a CA                    | writes `/usr/local/share/ca-certificates/osapi-internal-ca.crt`: `-----BEGIN CERTIFICATE-----\nMIIBvjCCAWOgAwIBAgIUUZjQew/w2XTBbe8VtXbYJ09SIa8wCgYIKoZIzj0EAwIw\nKzEZMBcGA1UEAwwQb3NhcGkgZXhhbXBsZSBDQTEOMAwGA1UECgwFb3NhcGkwIBcN\nMjYxMTIyMjMzODAxWhgPMjEyNjEwMjkyMzM4MDFaMCsxGTAXBgNVBAMMEG9zYXBp\nIGV4YW1wbGUgQ0ExDjAMBgNVBAoMBW9zYXBpMFkwEwYHKoZIzj0CAQYIKoZIzj0D\nAQcDQgAEYw7VLFKb+bykrwpFhkfkV1zCWS1PpnaxA7eBrMBNHd8fWKjO6ZRTmufJ\nZw5Cer8Qa/k/CWWaX4rPZR/N3trzSaNjMGEwHQYDVR0OBBYEFJMZvsnO2uSZngzL\ns/2/sLvsxF9KMB8GA1UdIwQYMBaAFJMZvsnO2uSZngzLs/2/sLvsxF9KMA8GA1Ud\nEwEB/wQFMAMBAf8wDgYDVR0PAQH/BAQDAgIEMAoGCCqGSM49BAMCA0kAMEYCIQCM\nKPI9T2lEfDJdlv4h2AY9AhLdLk3PW9vqATbr4gi7PgIhAOqEAMpruGrDNfCo9n27\ny9NIOpsPtN8EAH9a6bt+vnPx\n-----END CERTIFICATE-----`; runs `update-ca-certificates` |
| Add a CA on the RHEL family | writes `/etc/pki/ca-trust/source/anchors/osapi-internal-ca.crt`: `-----BEGIN CERTIFICATE-----\nMIIBvjCCAWOgAwIBAgIUUZjQew/w2XTBbe8VtXbYJ09SIa8wCgYIKoZIzj0EAwIw\nKzEZMBcGA1UEAwwQb3NhcGkgZXhhbXBsZSBDQTEOMAwGA1UECgwFb3NhcGkwIBcN\nMjYxMTIyMjMzODAxWhgPMjEyNjEwMjkyMzM4MDFaMCsxGTAXBgNVBAMMEG9zYXBp\nIGV4YW1wbGUgQ0ExDjAMBgNVBAoMBW9zYXBpMFkwEwYHKoZIzj0CAQYIKoZIzj0D\nAQcDQgAEYw7VLFKb+bykrwpFhkfkV1zCWS1PpnaxA7eBrMBNHd8fWKjO6ZRTmufJ\nZw5Cer8Qa/k/CWWaX4rPZR/N3trzSaNjMGEwHQYDVR0OBBYEFJMZvsnO2uSZngzL\ns/2/sLvsxF9KMB8GA1UdIwQYMBaAFJMZvsnO2uSZngzLs/2/sLvsxF9KMA8GA1Ud\nEwEB/wQFMAMBAf8wDgYDVR0PAQH/BAQDAgIEMAoGCCqGSM49BAMCA0kAMEYCIQCM\nKPI9T2lEfDJdlv4h2AY9AhLdLk3PW9vqATbr4gi7PgIhAOqEAMpruGrDNfCo9n27\ny9NIOpsPtN8EAH9a6bt+vnPx\n-----END CERTIFICATE-----`; runs `update-ca-trust`        |
| Add a CA on SUSE            | writes `/etc/pki/trust/anchors/osapi-internal-ca.crt`: `-----BEGIN CERTIFICATE-----\nMIIBvjCCAWOgAwIBAgIUUZjQew/w2XTBbe8VtXbYJ09SIa8wCgYIKoZIzj0EAwIw\nKzEZMBcGA1UEAwwQb3NhcGkgZXhhbXBsZSBDQTEOMAwGA1UECgwFb3NhcGkwIBcN\nMjYxMTIyMjMzODAxWhgPMjEyNjEwMjkyMzM4MDFaMCsxGTAXBgNVBAMMEG9zYXBp\nIGV4YW1wbGUgQ0ExDjAMBgNVBAoMBW9zYXBpMFkwEwYHKoZIzj0CAQYIKoZIzj0D\nAQcDQgAEYw7VLFKb+bykrwpFhkfkV1zCWS1PpnaxA7eBrMBNHd8fWKjO6ZRTmufJ\nZw5Cer8Qa/k/CWWaX4rPZR/N3trzSaNjMGEwHQYDVR0OBBYEFJMZvsnO2uSZngzL\ns/2/sLvsxF9KMB8GA1UdIwQYMBaAFJMZvsnO2uSZngzLs/2/sLvsxF9KMA8GA1Ud\nEwEB/wQFMAMBAf8wDgYDVR0PAQH/BAQDAgIEMAoGCCqGSM49BAMCA0kAMEYCIQCM\nKPI9T2lEfDJdlv4h2AY9AhLdLk3PW9vqATbr4gi7PgIhAOqEAMpruGrDNfCo9n27\ny9NIOpsPtN8EAH9a6bt+vnPx\n-----END CERTIFICATE-----`; runs `update-ca-certificates`           |

The command fails in these cases:

| Call                            | Flags                                         | Fails with              |
| ------------------------------- | --------------------------------------------- | ----------------------- |
| A file that osapi did not write | `--name internal-ca --object internal-ca.pem` | osapi did not deploy it |
| A name that is an option        | `--name -r --object internal-ca.pem`          | not a valid name        |
| No object                       | `--name internal-ca`                          | required                |
| A name with a space             | `--name 'a b' --object internal-ca.pem`       | resource_name           |

## Permission

It needs the `certificate:write` permission.
`osapi controller role get --name <role>` shows what a role grants.

## Flags

| Flag           | Description                                                                                                 | Default |
| -------------- | ----------------------------------------------------------------------------------------------------------- | ------- |
| `--name`       | The name of the CA, such as internal-ca, which osapi writes as osapi-internal-ca.crt (required)             |         |
| `--object`     | The object of the Object Store that has the PEM of the CA, as osapi client file upload stored it (required) |         |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev)                                                      | `_all`  |

## Global flags

| Flag               | Description                                                                      | Default                  |
| ------------------ | -------------------------------------------------------------------------------- | ------------------------ |
| `-d, --debug`      | Enable or disable debug mode                                                     |                          |
| `-j, --json`       | Enable JSON output                                                               |                          |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml`  |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `https://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
