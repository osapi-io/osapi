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

const VIEWER_PERMISSIONS: Permission[] = ROLE_PERMISSIONS.read;
const OPERATOR_PERMISSIONS: Permission[] = ROLE_PERMISSIONS.write;
const ADMIN_PERMISSIONS: Permission[] = ROLE_PERMISSIONS.admin;

export const ROLES: RoleDefinition[] = [
  {
    name: "admin",
    label: "admin",
    description: "Full access including audit logs",
    jwtRole: "admin",
    permissions: new Set(ADMIN_PERMISSIONS),
  },
  {
    name: "operator",
    label: "write",
    description: "Deploy, execute, and manage resources",
    jwtRole: "write",
    permissions: new Set(OPERATOR_PERMISSIONS),
  },
  {
    name: "viewer",
    label: "read",
    description: "Read-only access to all resources",
    jwtRole: "read",
    permissions: new Set(VIEWER_PERMISSIONS),
  },
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
  "cron-create": "postNodeSchedule",
  "cron-list": "getNodeSchedule",
  "cron-delete": "deleteNodeSchedule",
  "cron-get": "getNodeScheduleByName",
  "cron-update": "putNodeSchedule",
  "file-list": "getFiles",
  "file-upload": "postFile",
  "file-deploy": "postNodeFileDeploy",
  "file-undeploy": "postNodeFileUndeploy",
  "file-status": "getNodeFileStatus",
  "file-delete": "deleteFileByName",
  "file-stale": "getFileStale",
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
  "ntp-get": "getNodeNtp",
  "ntp-set": "postNodeNtp",
  "ntp-update": "putNodeNtp",
  "ntp-delete": "deleteNodeNtp",
  "power-reboot": "postNodePowerReboot",
  "power-shutdown": "postNodePowerShutdown",
  "process-list": "getNodeProcess",
  "process-get": "getNodeProcessByPid",
  "process-signal": "postNodeProcessSignal",
  "log-query": "getNodeLog",
  "log-sources": "getNodeLogSource",
  "log-query-unit": "getNodeLogUnit",
  "audit-list": "getAuditLogs",
  "audit-get": "getAuditLogByID",
  "audit-export": "getAuditExport",
  "user-list": "getNodeUser",
  "user-get": "getNodeUserByName",
  "user-create": "postNodeUser",
  "user-update": "putNodeUser",
  "user-delete": "deleteNodeUser",
  "user-list-keys": "getNodeUserSSHKey",
  "user-add-key": "postNodeUserSSHKey",
  "user-remove-key": "deleteNodeUserSSHKey",
  "user-change-password": "postNodeUserPassword",
  "group-list": "getNodeGroup",
  "group-get": "getNodeGroupByName",
  "group-create": "postNodeGroup",
  "group-update": "putNodeGroup",
  "group-delete": "deleteNodeGroup",
  "certificate-list": "getNodeCertificateCa",
  "certificate-create": "postNodeCertificateCa",
  "certificate-update": "putNodeCertificateCa",
  "certificate-delete": "deleteNodeCertificateCa",
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
