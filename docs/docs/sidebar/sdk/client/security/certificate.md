---
sidebar_position: 7
---

# Certificate

The certificate authorities (CAs) that a host trusts. `List` and `Get` read each
CA of the trust store on each platform. `Create`, `Update` and `Delete` change
only the CAs that osapi deploys, from PEM objects in the Object Store.

## Methods

| Method                                                              | Returns                                   | Permission          | Description                                             |
| ------------------------------------------------------------------- | ----------------------------------------- | ------------------- | ------------------------------------------------------- |
| `c.Certificate.List(ctx, hostname)`                                 | `Response[Collection[CertificateEntry]]`  | `certificate:read`  | List the certificate authorities that the host trusts   |
| `c.Certificate.Create(ctx, hostname, input CertificateCreateInput)` | `Response[Collection[CertificateResult]]` | `certificate:write` | Add a certificate authority to the trust store          |
| `c.Certificate.Get(ctx, hostname, name)`                            | `Response[Collection[CertificateEntry]]`  | `certificate:read`  | Get one certificate authority                           |
| `c.Certificate.Update(ctx, hostname, input CertificateUpdateInput)` | `Response[Collection[CertificateResult]]` | `certificate:write` | Write a certificate authority that osapi deployed again |
| `c.Certificate.Delete(ctx, hostname, name)`                         | `Response[Collection[CertificateResult]]` | `certificate:write` | Remove a certificate authority that osapi deployed      |

## Types

The fields of each type are the properties of its schema in the API's spec.

### CertificateEntry

| Field         | Type     | Description                                                                               |
| ------------- | -------- | ----------------------------------------------------------------------------------------- |
| `Enabled`     | `bool`   | Whether the host trusts it: false for a system CA that the trust store turns off.         |
| `Error`       | `string` | Error message if the agent failed.                                                        |
| `Fingerprint` | `string` | The SHA-256 of the certificate, in hex.                                                   |
| `Hostname`    | `string` | Hostname of the agent that produced this result.                                          |
| `Managed`     | `bool`   | Whether osapi deployed it, and its writes change it.                                      |
| `Name`        | `string` | The CA: the name it was deployed with for a CA that osapi deployed, else its file's name. |
| `NotAfter`    | `string` | When it expires.                                                                          |
| `Object`      | `string` | The object that osapi deployed it from.                                                   |
| `Path`        | `string` | The file, or on macOS the keychain, that has it.                                          |
| `Source`      | `string` | System for a CA of the distribution, local for one an administrator added.                |
| `Status`      | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout.    |
| `Subject`     | `string` | Its subject, such as CN=ISRG Root X1,O=Internet Security Research Group,C=US.             |

### CertificateCreateInput

| Field    | Type     | Description                                                                                       |
| -------- | -------- | ------------------------------------------------------------------------------------------------- |
| `Name`   | `string` | The name of the CA, such as internal-ca, which osapi writes as osapi-internal-ca.crt.             |
| `Object` | `string` | The object of the Object Store that has the PEM of the CA, as osapi client file upload stored it. |

### CertificateResult

| Field      | Type     | Description                                                                            |
| ---------- | -------- | -------------------------------------------------------------------------------------- |
| `Changed`  | `bool`   | Whether the host changed.                                                              |
| `Error`    | `string` | Error message if the agent failed.                                                     |
| `Hostname` | `string` | Hostname of the agent that produced this result.                                       |
| `Name`     | `string` | The CA.                                                                                |
| `Status`   | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |

### CertificateUpdateInput

| Field    | Type     | Description                                                            |
| -------- | -------- | ---------------------------------------------------------------------- |
| `Name`   | `string` | The name of the path: the CA that osapi deployed, such as internal-ca. |
| `Object` | `string` | The object that has the PEM; the one it was deployed from when empty.  |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// List the CAs that the host trusts
resp, err := c.Certificate.List(ctx, "web-01")
// Each result is one CA on one host.
for _, r := range resp.Data.Results {
    fmt.Printf("%s: %s source=%s managed=%v\n", r.Hostname, r.Name, r.Source, r.Managed)
}

// Get one CA
resp, err := c.Certificate.Get(ctx, "web-01", "internal-ca")

// Create a custom CA certificate
resp, err := c.Certificate.Create(ctx, "web-01",
    client.CertificateCreateInput{
        Name:   "internal-ca",
        Object: "internal-ca",
    })

// Update a certificate with a new object
resp, err := c.Certificate.Update(ctx, "web-01", client.CertificateUpdateInput{
    Name: "internal-ca",
    Object: "internal-ca-v2",
})

// Delete a certificate
resp, err := c.Certificate.Delete(ctx, "web-01", "internal-ca")
```

## Example

Refer to
[`examples/sdk/client/certificate.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/certificate.go)
for a complete working example.

## Permissions

The Permission column of [Methods](#methods) gives the permission of each
method.

`List` and `Get` operate on Linux and on macOS. `Create`, `Update` and `Delete`
operate on Linux: on the Debian family, and on the RHEL family and Fedora. On
macOS, they return `status: skipped`.
