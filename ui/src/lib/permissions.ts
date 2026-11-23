// ---------------------------------------------------------------------------
// Permissions — generated from Go, in permissions.gen.ts
// ---------------------------------------------------------------------------

import {
  ALL_PERMISSIONS,
  ROLE_PERMISSIONS,
  type Permission,
} from "./permissions.gen";
import { OPERATION_PERMISSIONS, type OperationId } from "./operations.gen";
import { BEHAVIOR_BLOCKS } from "./blocks.gen";

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
// Block → operation → permission
//
// Each block names the operation that it calls. Its permission is that
// operation's, from the scopes of the spec in operations.gen.ts. Thus a
// block cannot check another permission than the API asks for, and an
// operation that is renamed or removed fails the type check here.
// ---------------------------------------------------------------------------

export const BLOCK_OPERATIONS: Record<string, OperationId> = {
  "object-list": "getFiles",
  "object-upload": "postFile",
  "object-delete": "deleteFileByName",
  "object-stale": "getFileStale",
  "docker-create": "postNodeContainerDocker",
  "docker-list": "getNodeContainerDocker",
  "docker-start": "postNodeContainerDockerStart",
  "docker-stop": "postNodeContainerDockerStop",
  "docker-delete": "deleteNodeContainerDockerByID",
  "docker-exec": "postNodeContainerDockerExec",
  "docker-pull": "postNodeContainerDockerPull",
  "docker-rm-image": "deleteNodeContainerDockerImage",
  "docker-inspect": "getNodeContainerDockerByID",
  "command-shell": "postNodeCommandShell",
  "audit-list": "getAuditLogs",
  "audit-get": "getAuditLogByID",
  "audit-export": "getAuditExport",
};

export const BLOCK_PERMISSIONS: Record<string, Permission> = Object.fromEntries(
  [
    ...Object.entries(BLOCK_OPERATIONS).map(([block, op]) => [
      block,
      OPERATION_PERMISSIONS[op],
    ]),
    // A behavior block's permission is its method's, from blocks.gen.ts.
    ...Object.values(BEHAVIOR_BLOCKS).map((def) => [def.type, def.permission]),
  ],
);
