---
title: TLS certificate management
status: backlog
created: 2026-02-15
updated: 2026-02-15
---

## Objective

Add certificate management for the appliance's own TLS certificates and
CA trust store. An appliance serving HTTPS needs to manage its certs.

## API Endpoints

```
GET    /certificate          - List installed certificates
GET    /certificate/{id}     - Get certificate details (not private key)
POST   /certificate          - Upload/install certificate + key
DELETE /certificate/{id}     - Remove certificate
POST   /certificate/csr      - Generate a CSR

GET    /certificate/ca       - List trusted CA certificates
POST   /certificate/ca       - Add CA certificate to trust store
DELETE /certificate/ca/{id}  - Remove CA certificate
```

## Operations

- `certificate.list.get`, `certificate.status.get` (query)
- `certificate.install.execute`, `certificate.delete.execute` (modify)
- `certificate.csr.create` (modify)
- `certificate.ca.list.get` (query)
- `certificate.ca.add.execute`, `certificate.ca.remove.execute` (modify)

## Provider

- `internal/provider/security/certificate/`
- Parse PEM files, read x509 metadata (subject, issuer, expiry, SANs)
- Manage system CA trust store (`update-ca-certificates`)

## Notes

- Never expose private keys via GET endpoints
- Certificate expiry monitoring is valuable for alerting
- Scopes: `certificate:read`, `certificate:write`
- Consider ACME/Let's Encrypt integration as future enhancement
