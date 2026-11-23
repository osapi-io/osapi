---
sidebar_position: 7
---

# Certificate

The certificate authorities (CAs) that a host trusts. `List` and `Get` read each
CA of the trust store on each platform. `Create`, `Update` and `Delete` change
only the CAs that osapi deploys, from PEM objects in the Object Store.

## Methods

| Method                              | Description                       |
| ----------------------------------- | --------------------------------- |
| `List(ctx, hostname)`               | List the CAs that the host trusts |
| `Get(ctx, hostname, name)`          | Get one CA                        |
| `Create(ctx, hostname, opts)`       | Add a CA to the trust store       |
| `Update(ctx, hostname, name, opts)` | Write a CA that osapi deployed    |
| `Delete(ctx, hostname, name)`       | Remove a CA that osapi deployed   |

## Request Types

| Type                    | Fields                             |
| ----------------------- | ---------------------------------- |
| `CertificateCreateOpts` | Name (required), Object (required) |
| `CertificateUpdateOpts` | Object (the deployed one if empty) |

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

Refer to
[`examples/sdk/client/certificate.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/certificate.go)
for a complete working example.

## Permissions

| Operation              | Permission          |
| ---------------------- | ------------------- |
| List                   | `certificate:read`  |
| Create, Update, Delete | `certificate:write` |

`List` and `Get` operate on Linux and on macOS. `Create`, `Update` and `Delete`
operate on Linux: on the Debian family, and on the RHEL family and Fedora. On
macOS, they return `status: skipped`.
