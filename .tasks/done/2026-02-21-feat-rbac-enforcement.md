---
title: RBAC enforcement with fine-grained permissions
status: done
created: 2026-02-21
updated: 2026-02-21
---

## Objective

Today OSAPI has three flat roles (`read`, `write`, `admin`) with a simple
hierarchy baked into `authtoken.RoleHierarchy`. Scopes are declared in
OpenAPI specs and enforced by `scopeMiddleware`. This works for basic
access control but falls short for multi-tenant or enterprise deployments
where:

- Different users need different permission sets (operator can restart
  services but not modify DNS; network admin can change DNS but not view
  jobs).
- Audit logging (see `2026-02-15-feat-audit-logging.md`) needs to know
  _which user_ performed an action, not just that _a valid token_ was
  used.
- Customers will want to map OSAPI permissions to their existing AD/LDAP
  groups or SSO provider roles.

The goal is to introduce a capability-based RBAC model — inspired by
Kubernetes RBAC and Linux capabilities — where permissions are granular,
composable into roles, and the JWT `sub` claim identifies the acting
user for audit purposes.

## Current State

- JWT claims: `{ roles: ["admin"], sub: "john@dewey.ws", ... }`
- `RoleHierarchy` maps role → implied scopes: `admin → [read, write, admin]`
- OpenAPI specs declare required scope per endpoint:
  `security: [{ BearerAuth: [read] }]`
- `scopeMiddleware` checks if any token role implies the required scope
- `hasScope()` does the matching; 403 on failure
- No user identity is passed to handlers or logged beyond the JWT claims
- `unauthenticatedOperations` map exempts health probes

## Design

### Permissions (capabilities)

Define granular permissions modeled after `resource:verb` (Kubernetes
style). Each OpenAPI endpoint maps to exactly one permission:

```
system:read          GET /system/hostname, GET /system/status
network:read         GET /network/dns/{interface}
network:write        PUT /network/dns, POST /network/ping
job:read             GET /job, GET /job/{id}, GET /job/stats, GET /job/workers
job:write            POST /job, DELETE /job/{id}, POST /job/{id}/retry
health:read          GET /health/status
audit:read           GET /audit, GET /audit/{id}  (future)
```

Health probes (`GET /health`, `GET /health/ready`) remain unauthenticated.

### Roles

Roles are named collections of permissions. Built-in roles provide
backward compatibility:

```
admin     → [system:read, network:read, network:write, job:read,
             job:write, health:read, audit:read]
operator  → [system:read, network:read, job:read, job:write, health:read]
read      → [system:read, network:read, job:read, health:read]
```

The `write` role maps to `operator` for backward compatibility — existing
tokens with `roles: ["write"]` continue to work.

Custom roles can be defined in config for deployments that need
non-standard groupings:

```yaml
api:
  server:
    security:
      roles:
        network-admin:
          - network:read
          - network:write
        job-viewer:
          - job:read
```

### JWT Claims

Extend `CustomClaims` to carry the user identity and optionally
fine-grained permissions:

```go
type CustomClaims struct {
    Roles       []string `json:"roles" validate:"required,dive"`
    Permissions []string `json:"permissions,omitempty"`
    jwt.RegisteredClaims
}
```

- `sub` (Subject): identifies the user — required for audit logging
- `roles`: maps to built-in or custom role definitions
- `permissions`: optional override — if present, these are used directly
  instead of resolving from roles. This allows an upstream IdP (AD,
  Okta, Keycloak) to embed exact permissions in the token without OSAPI
  needing to know about the IdP's group structure.

### Middleware Changes

`scopeMiddleware` changes:

1. Resolve effective permissions: if `claims.Permissions` is non-empty,
   use it directly; otherwise expand `claims.Roles` through the role →
   permission mapping.
2. Check if the required permission (from OpenAPI spec) is in the
   effective set.
3. Inject user identity into the request context so handlers and audit
   logging can access it:
   ```go
   ctx.Set("auth.subject", claims.Subject)
   ctx.Set("auth.roles", claims.Roles)
   ctx.Set("auth.permissions", effectivePermissions)
   ```

### OpenAPI Spec Changes

Update security schemes in each `api.yaml` to use the new permission
names as scopes:

```yaml
components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT

# Per endpoint:
security:
  - BearerAuth:
      - system:read
```

oapi-codegen will pass `["system:read"]` to the context, and
`scopeMiddleware` checks it against the resolved permission set.

### Backward Compatibility

- Existing tokens with `roles: ["admin"]` or `roles: ["read"]` continue
  to work — the role-to-permission mapping handles expansion.
- The old `read`/`write`/`admin` scope names in OpenAPI specs are
  replaced with `resource:verb` permissions, but the middleware resolves
  old role names through the mapping.
- `osapi token generate --roles admin` still works. Add
  `--permissions system:read,job:write` flag for fine-grained tokens.

### External IdP Integration (future)

The design supports upstream identity providers without code changes:

- **OIDC/OAuth2**: IdP issues JWTs with `roles` or `permissions` claims
  that OSAPI validates. Map IdP groups → OSAPI roles in config.
- **AD/LDAP**: A gateway (e.g., Keycloak, Dex) translates AD groups to
  JWT claims. OSAPI only sees the JWT.
- **Custom**: Any system that can produce a signed JWT with the expected
  claims structure works. Configure the signing key or JWKS endpoint.

This task does NOT implement IdP integration — it builds the permission
model that makes integration straightforward later.

## Implementation Steps

### Step 1: Define permissions and role mappings

- Create `internal/authtoken/permissions.go` with permission constants
  and the role → permission mapping.
- Update `RoleHierarchy` to map to permissions instead of the old flat
  scopes.
- Add `Permissions` field to `CustomClaims`.

### Step 2: Update OpenAPI specs

- Replace `read`/`write` scopes with `resource:verb` permissions in
  every `api.yaml` under `internal/api/*/gen/`.
- Regenerate with `just generate`.

### Step 3: Update middleware

- Update `scopeMiddleware` to resolve permissions from roles or use
  direct permissions from claims.
- Inject user identity (`sub`, roles, permissions) into Echo context.
- Update `hasScope` → `hasPermission`.

### Step 4: Update token CLI

- Add `--permissions` flag to `osapi token generate`.
- Update `osapi token validate` to show resolved permissions.
- Validate permission strings against known permissions.

### Step 5: Custom roles in config

- Add `api.server.security.roles` config section.
- Merge custom roles with built-in roles at startup.
- Validate that custom role permissions reference known permissions.

### Step 6: Integration tests

- Test each endpoint with tokens that have exactly the right permission
  (should succeed) and tokens missing it (should 403).
- Test backward compatibility: old `roles: ["admin"]` tokens still work.
- Test `permissions` claim override bypasses role resolution.
- Test custom roles from config.
- Test that user identity (`sub`) is available in request context.

### Step 7: Documentation

- Update `docs/docs/sidebar/configuration.md` with new roles config.
- Add RBAC section to architecture docs explaining the permission model.
- Update token generation docs with `--permissions` examples.

## Notes

- Keep the built-in roles simple — three is enough for most deployments.
  Custom roles handle the edge cases.
- Permission strings use `resource:verb` not `resource.verb` to avoid
  confusion with NATS subject dots.
- The `admin` role should always be a superset of all permissions.
- This is a prerequisite for audit logging — once RBAC is in place,
  audit middleware can read `auth.subject` from context.
- Consider adding a `deny` list later for temporary permission
  revocation without regenerating tokens.

## Outcome

Implemented fine-grained RBAC with `resource:verb` permissions across all
domains. Key changes:

- Created `internal/authtoken/permissions.go` with permission constants,
  `DefaultRolePermissions` map, `ResolvePermissions()`, and
  `HasPermission()`.
- Added `Permissions []string` to `CustomClaims` and updated `Generate()`
  signature.
- Added `CustomRole` struct and `Roles` config to `ServerSecurity`.
- Updated all 4 OpenAPI specs (`system`, `network`, `job`, `health`) to
  use `resource:verb` scopes and regenerated code.
- Rewrote `scopeMiddleware` to resolve permissions from roles/claims and
  inject user identity (`auth.subject`, `auth.roles`) into context.
- Added `--permissions/-p` flag to `osapi token generate` CLI.
- Updated `osapi token validate` to show resolved effective permissions.
- Added RBAC integration tests to all 13 authenticated endpoints across
  system (2), network (3), job (7), and health (1) domains.
- Renamed existing integration tests to `*Validation()` pattern.
- Documented JWT signing trust model, token structure, permission
  resolution, and custom roles in architecture and configuration docs.
