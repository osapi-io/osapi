---
sidebar_position: 7
---

# Certificate

CA certificate management on target hosts. Certificates are deployed as PEM
files from the Object Store and installed into the system trust store.

## Methods

| Method                              | Description                      |
| ----------------------------------- | -------------------------------- |
| `List(ctx, hostname)`               | List all CA certificates         |
| `Create(ctx, hostname, opts)`       | Deploy a custom CA certificate   |
| `Update(ctx, hostname, name, opts)` | Redeploy a custom CA certificate |
| `Delete(ctx, hostname, name)`       | Remove a custom CA certificate   |

## Request Types

| Type                    | Fields                             |
| ----------------------- | ---------------------------------- |
| `CertificateCreateOpts` | Name (required), Object (required) |
| `CertificateUpdateOpts` | Object (required)                  |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// List all CA certificates
resp, err := c.Certificate.List(ctx, "web-01")
for _, r := range resp.Data.Results {
    for _, cert := range r.Certificates {
        fmt.Printf("%s source=%s\n", cert.Name, cert.Source)
    }
}

// Create a custom CA certificate
resp, err := c.Certificate.Create(ctx, "web-01",
    client.CertificateCreateOpts{
        Name:   "internal-ca",
        Object: "internal-ca",
    })

// Update a certificate with a new object
resp, err := c.Certificate.Update(ctx, "web-01", "internal-ca",
    client.CertificateUpdateOpts{
        Object: "internal-ca-v2",
    })

// Delete a certificate
resp, err := c.Certificate.Delete(ctx, "web-01", "internal-ca")
```

## Example

See
[`examples/sdk/client/certificate.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/certificate.go)
for a complete working example.

## Permissions

| Operation              | Permission          |
| ---------------------- | ------------------- |
| List                   | `certificate:read`  |
| Create, Update, Delete | `certificate:write` |

Certificate management is supported on the Debian OS family (Ubuntu, Debian,
Raspbian). On unsupported platforms (Darwin, generic Linux), operations return
`status: skipped`.
