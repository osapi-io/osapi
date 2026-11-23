// ---------------------------------------------------------------------------
// Permissions — generated from Go, in permissions.gen.ts
// ---------------------------------------------------------------------------

import {
  ALL_PERMISSIONS,
  ROLE_PERMISSIONS,
  type Permission,
} from "./permissions.gen";

export { ALL_PERMISSIONS, type Permission };

// ---------------------------------------------------------------------------
// Roles — named sets of permissions with hierarchy
// ---------------------------------------------------------------------------

export interface RoleDefinition {
  name: string;
  label: string;
  description: string;
  /** The JWT role value this maps to */
  jwtRole: string;
  permissions: Set<Permission>;
}

// The built-in roles of the controller, one for each kind of person or
// program that uses osapi. Their permissions come from Go
// (ROLE_PERMISSIONS), so the UI shows what a token of the role can do. The
// last is the one with the least, which the UI falls back to.
const role = (
  name: keyof typeof ROLE_PERMISSIONS,
  description: string,
): RoleDefinition => ({
  name,
  label: name,
  description,
  jwtRole: name,
  permissions: new Set(ROLE_PERMISSIONS[name]),
});

export const ROLES: RoleDefinition[] = [
  role(
    "admin",
    "Everything, including admission, tokens, shells and the audit log",
  ),
  role(
    "operator",
    "Reads, and actions such as a restart, a signal or a reboot",
  ),
  role(
    "automation",
    "Reads, and changes to the configuration, with no actions",
  ),
  role("auditor", "The audit log, traces, jobs and health, read only"),
  role("read", "Reads each thing that is not admin only"),
  role("observer", "Health, status and agents only, for dashboards"),
];

export function getRoleByName(name: string): RoleDefinition | undefined {
  return ROLES.find((r) => r.name === name);
}

export function getRoleByJwtRole(jwtRole: string): RoleDefinition | undefined {
  return ROLES.find((r) => r.jwtRole === jwtRole);
}

export function hasPermission(
  role: RoleDefinition,
  permission: Permission,
): boolean {
  return role.permissions.has(permission);
}

export function hasAllPermissions(
  role: RoleDefinition,
  permissions: Permission[],
): boolean {
  return permissions.every((p) => role.permissions.has(p));
}

export function hasAnyPermission(
  role: RoleDefinition,
  permissions: Permission[],
): boolean {
  return permissions.some((p) => role.permissions.has(p));
}

// ---------------------------------------------------------------------------
// A step's permission
// ---------------------------------------------------------------------------

/**
 * Whether a role that has the permissions that can accepts may run a step
 * that needs permission. A step with no permission that the UI knows is
 * refused, not allowed: the UI cannot tell what the API will ask for.
 */
export function permitted(
  permission: string | undefined,
  can: (permission: Permission) => boolean,
): boolean {
  return (
    permission !== undefined &&
    (ALL_PERMISSIONS as readonly string[]).includes(permission) &&
    can(permission as Permission)
  );
}
