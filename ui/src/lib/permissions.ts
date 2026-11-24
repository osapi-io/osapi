// ---------------------------------------------------------------------------
// Permissions and roles, from the catalog that the controller serves
// ---------------------------------------------------------------------------

import type { CatalogRole, Permission } from "@/sdk/gen/schemas";

export type { Permission };

// ---------------------------------------------------------------------------
// Roles — named sets of permissions
// ---------------------------------------------------------------------------

export interface RoleDefinition {
  name: string;
  label: string;
  description: string;
  /** The JWT role value this maps to */
  jwtRole: string;
  permissions: Set<Permission>;
}

/**
 * The built-in roles of the controller, one for each kind of person or
 * program that uses osapi, in the order and with the descriptions and the
 * permissions that the catalog gives. Thus the UI shows what a token of the
 * role can do. The last is the one with the least, which the UI falls back
 * to.
 */
export function rolesOf(roles: readonly CatalogRole[]): RoleDefinition[] {
  return roles.map((r) => ({
    name: r.name,
    label: r.name,
    description: r.description,
    jwtRole: r.name,
    permissions: new Set(r.permissions),
  }));
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

// ---------------------------------------------------------------------------
// A step's permission
// ---------------------------------------------------------------------------

/**
 * Whether a role that has the permissions that can accepts may run a step
 * that needs permission. A step with no permission is refused, not allowed:
 * the UI cannot tell what the API will ask for. A role has only the
 * permissions that the catalog lists, so a permission that the controller
 * does not know is refused too.
 */
export function permitted(
  permission: Permission | undefined,
  can: (permission: Permission) => boolean,
): boolean {
  return permission !== undefined && can(permission);
}
