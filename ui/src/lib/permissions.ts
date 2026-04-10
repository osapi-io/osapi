// ---------------------------------------------------------------------------
// Permissions — mirrors osapi's pkg/sdk/client/permissions.go
// ---------------------------------------------------------------------------

export type Permission =
  | "agent:read"
  | "agent:write"
  | "node:read"
  | "network:read"
  | "network:write"
  | "job:read"
  | "job:write"
  | "health:read"
  | "audit:read"
  | "command:execute"
  | "file:read"
  | "file:write"
  | "docker:read"
  | "docker:write"
  | "docker:execute"
  | "cron:read"
  | "cron:write"
  | "service:read"
  | "service:write"
  | "package:read"
  | "package:write"
  | "sysctl:read"
  | "sysctl:write"
  | "ntp:read"
  | "ntp:write"
  | "timezone:read"
  | "timezone:write"
  | "hostname:read"
  | "hostname:write"
  | "power:execute"
  | "process:read"
  | "process:write"
  | "log:read"
  | "user:read"
  | "user:write"
  | "group:read"
  | "group:write"
  | "certificate:read"
  | "certificate:write";

export const ALL_PERMISSIONS: Permission[] = [
  "agent:read",
  "agent:write",
  "node:read",
  "network:read",
  "network:write",
  "job:read",
  "job:write",
  "health:read",
  "audit:read",
  "command:execute",
  "file:read",
  "file:write",
  "docker:read",
  "docker:write",
  "docker:execute",
  "cron:read",
  "cron:write",
  "service:read",
  "service:write",
  "package:read",
  "package:write",
  "sysctl:read",
  "sysctl:write",
  "ntp:read",
  "ntp:write",
  "timezone:read",
  "timezone:write",
  "hostname:read",
  "hostname:write",
  "power:execute",
  "process:read",
  "process:write",
  "log:read",
  "user:read",
  "user:write",
  "group:read",
  "group:write",
  "certificate:read",
  "certificate:write",
];

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

const VIEWER_PERMISSIONS: Permission[] = [
  "agent:read",
  "node:read",
  "network:read",
  "job:read",
  "health:read",
  "file:read",
  "docker:read",
  "cron:read",
  "service:read",
  "package:read",
  "sysctl:read",
  "ntp:read",
  "timezone:read",
  "hostname:read",
  "process:read",
  "log:read",
  "user:read",
  "group:read",
  "certificate:read",
];

const OPERATOR_PERMISSIONS: Permission[] = [
  ...VIEWER_PERMISSIONS,
  "agent:write",
  "network:write",
  "job:write",
  "file:write",
  "docker:write",
  "docker:execute",
  "command:execute",
  "cron:write",
  "service:write",
  "package:write",
  "sysctl:write",
  "ntp:write",
  "timezone:write",
  "hostname:write",
  "power:execute",
  "process:write",
  "user:write",
  "group:write",
  "certificate:write",
];

const ADMIN_PERMISSIONS: Permission[] = [...OPERATOR_PERMISSIONS, "audit:read"];

export const ROLES: RoleDefinition[] = [
  {
    name: "admin",
    label: "Admin",
    description: "Full access including audit logs",
    jwtRole: "admin",
    permissions: new Set(ADMIN_PERMISSIONS),
  },
  {
    name: "operator",
    label: "Operator",
    description: "Deploy, execute, and manage resources",
    jwtRole: "write",
    permissions: new Set(OPERATOR_PERMISSIONS),
  },
  {
    name: "viewer",
    label: "Viewer",
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
// Block → Permission mapping (derived from OpenAPI spec security scopes)
// ---------------------------------------------------------------------------

export const BLOCK_PERMISSIONS: Record<string, Permission> = {
  // Cron group
  "cron-create": "cron:write",
  "cron-list": "cron:read",
  "cron-delete": "cron:write",
  "cron-get": "cron:read",
  "cron-update": "cron:write",
  // File group
  "file-list": "file:read",
  "file-upload": "file:write",
  "file-deploy": "file:write",
  "file-undeploy": "file:write",
  "file-status": "file:read",
  "file-delete": "file:write",
  "file-stale": "file:read",
  // Docker/Containers group
  "docker-create": "docker:write",
  "docker-list": "docker:read",
  "docker-start": "docker:write",
  "docker-stop": "docker:write",
  "docker-delete": "docker:write",
  "docker-exec": "docker:execute",
  "docker-pull": "docker:write",
  "docker-rm-image": "docker:write",
  "docker-inspect": "docker:read",
  // Command group
  command: "command:execute",
  "command-shell": "command:execute",
  // Networking group
  "dns-list": "network:read",
  "dns-update": "network:write",
  "dns-delete": "network:write",
  ping: "network:write",
  "interface-list": "network:read",
  "interface-get": "network:read",
  "interface-create": "network:write",
  "interface-update": "network:write",
  "interface-delete": "network:write",
  "route-list": "network:read",
  "route-get": "network:read",
  "route-create": "network:write",
  "route-update": "network:write",
  "route-delete": "network:write",
  // Services group
  "service-list": "service:read",
  "service-get": "service:read",
  "service-create": "service:write",
  "service-update": "service:write",
  "service-delete": "service:write",
  "service-start": "service:write",
  "service-stop": "service:write",
  "service-restart": "service:write",
  "service-enable": "service:write",
  "service-disable": "service:write",
  // Software/Package group
  "package-list": "package:read",
  "package-get": "package:read",
  "package-install": "package:write",
  "package-remove": "package:write",
  "package-update": "package:write",
  "package-check-updates": "package:read",
  // Config group
  "sysctl-list": "sysctl:read",
  "sysctl-get": "sysctl:read",
  "sysctl-set": "sysctl:write",
  "sysctl-update": "sysctl:write",
  "sysctl-delete": "sysctl:write",
  "ntp-get": "ntp:read",
  "ntp-set": "ntp:write",
  "ntp-update": "ntp:write",
  "ntp-delete": "ntp:write",
  "timezone-get": "timezone:read",
  "timezone-set": "timezone:write",
  "hostname-get": "hostname:read",
  "hostname-set": "hostname:write",
  // Node group
  "node-status": "node:read",
  "node-load": "node:read",
  "node-uptime": "node:read",
  "node-os": "node:read",
  "power-reboot": "power:execute",
  "power-shutdown": "power:execute",
  "process-list": "process:read",
  "process-get": "process:read",
  "process-signal": "process:write",
  "log-query": "log:read",
  "log-sources": "log:read",
  "log-query-unit": "log:read",
  // Hardware group
  "disk-info": "node:read",
  "memory-info": "node:read",
  // Audit group
  "audit-list": "audit:read",
  "audit-get": "audit:read",
  "audit-export": "audit:read",
  // Security group
  "user-list": "user:read",
  "user-get": "user:read",
  "user-create": "user:write",
  "user-update": "user:write",
  "user-delete": "user:write",
  "user-list-keys": "user:read",
  "user-add-key": "user:write",
  "user-remove-key": "user:write",
  "user-change-password": "user:write",
  "group-list": "group:read",
  "group-get": "group:read",
  "group-create": "group:write",
  "group-update": "group:write",
  "group-delete": "group:write",
  "certificate-list": "certificate:read",
  "certificate-create": "certificate:write",
  "certificate-update": "certificate:write",
  "certificate-delete": "certificate:write",
};
