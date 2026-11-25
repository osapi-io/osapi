# Update

Write a certificate authority that osapi deployed again.

Writes the CA from the object, or from the one that it was deployed from when
the object is empty, and rebuilds the trust store when the file changed. A CA
that osapi did not deploy is not found. macOS reports skipped.

```bash
osapi client node certificate update [flags]
```

## Examples

```bash
# Write a CA from a new object
osapi client node certificate update --target web-01 --name internal-ca --object internal-ca-2.pem

# Write a CA from a new object on the RHEL family
osapi client node certificate update --target web-01 --name internal-ca --object internal-ca-2.pem

# Write a CA from a new object on SUSE
osapi client node certificate update --target web-01 --name internal-ca --object internal-ca-2.pem

# The object that it was deployed from
osapi client node certificate update --target web-01 --name internal-ca
```

## Result

Write a CA from a new object, on one host:

```ansi
$ osapi client node certificate update --target web-01 --name internal-ca --object internal-ca-2.pem
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1minternal-ca[0m

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node certificate update --target _all --name internal-ca --object internal-ca-2.pem
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

Write a CA from a new object on the RHEL family, on one host:

```ansi
$ osapi client node certificate update --target web-01 --name internal-ca --object internal-ca-2.pem
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1minternal-ca[0m

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node certificate update --target _all --name internal-ca --object internal-ca-2.pem
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

Write a CA from a new object on SUSE, on one host:

```ansi
$ osapi client node certificate update --target web-01 --name internal-ca --object internal-ca-2.pem
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1minternal-ca[0m

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node certificate update --target _all --name internal-ca --object internal-ca-2.pem
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

The object that it was deployed from, on one host:

```ansi
$ osapi client node certificate update --target web-01 --name internal-ca
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1minternal-ca[0m

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node certificate update --target _all --name internal-ca
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

| Example                                         | On the host                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| ----------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Write a CA from a new object                    | writes `/usr/local/share/ca-certificates/osapi-internal-ca.crt`: `-----BEGIN CERTIFICATE-----\nMIIBvjCCAWOgAwIBAgIUUZjQew/w2XTBbe8VtXbYJ09SIa8wCgYIKoZIzj0EAwIw\nKzEZMBcGA1UEAwwQb3NhcGkgZXhhbXBsZSBDQTEOMAwGA1UECgwFb3NhcGkwIBcN\nMjYxMTIyMjMzODAxWhgPMjEyNjEwMjkyMzM4MDFaMCsxGTAXBgNVBAMMEG9zYXBp\nIGV4YW1wbGUgQ0ExDjAMBgNVBAoMBW9zYXBpMFkwEwYHKoZIzj0CAQYIKoZIzj0D\nAQcDQgAEYw7VLFKb+bykrwpFhkfkV1zCWS1PpnaxA7eBrMBNHd8fWKjO6ZRTmufJ\nZw5Cer8Qa/k/CWWaX4rPZR/N3trzSaNjMGEwHQYDVR0OBBYEFJMZvsnO2uSZngzL\ns/2/sLvsxF9KMB8GA1UdIwQYMBaAFJMZvsnO2uSZngzLs/2/sLvsxF9KMA8GA1Ud\nEwEB/wQFMAMBAf8wDgYDVR0PAQH/BAQDAgIEMAoGCCqGSM49BAMCA0kAMEYCIQCM\nKPI9T2lEfDJdlv4h2AY9AhLdLk3PW9vqATbr4gi7PgIhAOqEAMpruGrDNfCo9n27\ny9NIOpsPtN8EAH9a6bt+vnPx\n-----END CERTIFICATE-----`; runs `update-ca-certificates` |
| Write a CA from a new object on the RHEL family | writes `/etc/pki/ca-trust/source/anchors/osapi-internal-ca.crt`: `-----BEGIN CERTIFICATE-----\nMIIBvjCCAWOgAwIBAgIUUZjQew/w2XTBbe8VtXbYJ09SIa8wCgYIKoZIzj0EAwIw\nKzEZMBcGA1UEAwwQb3NhcGkgZXhhbXBsZSBDQTEOMAwGA1UECgwFb3NhcGkwIBcN\nMjYxMTIyMjMzODAxWhgPMjEyNjEwMjkyMzM4MDFaMCsxGTAXBgNVBAMMEG9zYXBp\nIGV4YW1wbGUgQ0ExDjAMBgNVBAoMBW9zYXBpMFkwEwYHKoZIzj0CAQYIKoZIzj0D\nAQcDQgAEYw7VLFKb+bykrwpFhkfkV1zCWS1PpnaxA7eBrMBNHd8fWKjO6ZRTmufJ\nZw5Cer8Qa/k/CWWaX4rPZR/N3trzSaNjMGEwHQYDVR0OBBYEFJMZvsnO2uSZngzL\ns/2/sLvsxF9KMB8GA1UdIwQYMBaAFJMZvsnO2uSZngzLs/2/sLvsxF9KMA8GA1Ud\nEwEB/wQFMAMBAf8wDgYDVR0PAQH/BAQDAgIEMAoGCCqGSM49BAMCA0kAMEYCIQCM\nKPI9T2lEfDJdlv4h2AY9AhLdLk3PW9vqATbr4gi7PgIhAOqEAMpruGrDNfCo9n27\ny9NIOpsPtN8EAH9a6bt+vnPx\n-----END CERTIFICATE-----`; runs `update-ca-trust`        |
| Write a CA from a new object on SUSE            | writes `/etc/pki/trust/anchors/osapi-internal-ca.crt`: `-----BEGIN CERTIFICATE-----\nMIIBvjCCAWOgAwIBAgIUUZjQew/w2XTBbe8VtXbYJ09SIa8wCgYIKoZIzj0EAwIw\nKzEZMBcGA1UEAwwQb3NhcGkgZXhhbXBsZSBDQTEOMAwGA1UECgwFb3NhcGkwIBcN\nMjYxMTIyMjMzODAxWhgPMjEyNjEwMjkyMzM4MDFaMCsxGTAXBgNVBAMMEG9zYXBp\nIGV4YW1wbGUgQ0ExDjAMBgNVBAoMBW9zYXBpMFkwEwYHKoZIzj0CAQYIKoZIzj0D\nAQcDQgAEYw7VLFKb+bykrwpFhkfkV1zCWS1PpnaxA7eBrMBNHd8fWKjO6ZRTmufJ\nZw5Cer8Qa/k/CWWaX4rPZR/N3trzSaNjMGEwHQYDVR0OBBYEFJMZvsnO2uSZngzL\ns/2/sLvsxF9KMB8GA1UdIwQYMBaAFJMZvsnO2uSZngzLs/2/sLvsxF9KMA8GA1Ud\nEwEB/wQFMAMBAf8wDgYDVR0PAQH/BAQDAgIEMAoGCCqGSM49BAMCA0kAMEYCIQCM\nKPI9T2lEfDJdlv4h2AY9AhLdLk3PW9vqATbr4gi7PgIhAOqEAMpruGrDNfCo9n27\ny9NIOpsPtN8EAH9a6bt+vnPx\n-----END CERTIFICATE-----`; runs `update-ca-certificates`           |

The command fails in these cases:

| Call                           | Flags            | Fails with       |
| ------------------------------ | ---------------- | ---------------- |
| A CA that osapi did not deploy | `--name nothing` | not found        |
| A name that is an option       | `--name -r`      | not a valid name |
| A name with a space            | `--name "a b"`   | resource_name    |

## Permission

It needs the `certificate:write` permission.
`osapi controller role get --name <role>` shows what a role grants.

## Flags

| Flag           | Description                                                          | Default |
| -------------- | -------------------------------------------------------------------- | ------- |
| `--name`       | The CA that osapi deployed, such as internal-ca (required)           |         |
| `--object`     | The object that has the PEM; the one it was deployed from when empty |         |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev)               | `_all`  |

## Global flags

| Flag               | Description                                                 | Default                 |
| ------------------ | ----------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                |                         |
| `-j, --json`       | Enable JSON output                                          |                         |
| `-f, --osapi-file` | Path to config file                                         | `/etc/osapi/osapi.yaml` |
| `--url`            | Base URL the CLI client connects to (controller.client.url) | `http://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
