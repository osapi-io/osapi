---
title: Audit and ensure validation on all API endpoints
status: done
created: 2026-02-15
updated: 2026-02-16
---

## Objective

Ensure all API endpoints have proper input validation where necessary. With the
migration to strict-server, validation that was previously handled by manual
`validator.Struct()` calls in non-strict handlers needs to be verified in the
new strict-server handler implementations.

## Endpoints to Audit

### Network API

- `POST /network/ping` - Validate `address` field is required + valid IP
  (currently validated via `validator.New()` in handler)
- `PUT /network/dns` - Validate servers (IP), search_domains (hostname),
  interface_name (required, alphanum) (currently validated in handler)
- `GET /network/dns/{interfaceName}` - Validate interfaceName path param
  (alphanum) — note: with strict-server, path params come as typed fields; need
  to verify validation still works

### System API

- `GET /system/status` - No body validation needed (no params)
- `GET /system/hostname` - No body validation needed (no params)

### Job API

- `POST /job` - Validate operation and target_hostname are present
- `GET /job` - No required validation (status is optional filter)
- `GET /job/{id}` - Path param validation
- `DELETE /job/{id}` - Path param validation
- `GET /job/status` - No validation needed

## Acceptance Criteria

- All endpoints with user input have validation
- Validation errors return appropriate 400 responses
- Tests cover validation error cases
- `x-oapi-codegen-extra-tags` with `validate:` tags are present in OpenAPI specs
  where applicable

## Outcome

All endpoints with user input now have validation with comprehensive test
coverage at both handler and HTTP integration levels:

- **POST /job**: Added `validate` tags to OpenAPI spec, added
  `validator.New()` + `validate.Struct()` in handler, added handler-level and
  HTTP integration tests for each validation point (operation required,
  target_hostname required)
- **POST /network/ping**: Already had validation; strengthened test assertions
  to check error reason (Address/required, Address/ip); added HTTP integration
  tests
- **PUT /network/dns**: Already had validation; strengthened test assertions to
  check error reason (InterfaceName/required); added HTTP integration tests
  covering all validation points (InterfaceName/required,
  InterfaceName/alphanum, Servers/ip, SearchDomains/hostname)

## Notes

The network API validation was preserved during the Phase 4 migration by:

1. Keeping `x-oapi-codegen-extra-tags` with `validate:` in the OpenAPI spec
2. Using `validator.New()` + `validate.Struct()` in handlers for request bodies
3. Adding 400 response types back to the OpenAPI spec
