import { useVimScroll } from "@/hooks/use-vim-scroll";
import { ContentArea } from "@/components/layout/content-area";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/ui/page-header";
import { SectionLabel } from "@/components/ui/section-label";
import { DataTable } from "@/components/ui/data-table";
import { Text } from "@/components/ui/text";
import { useAuth } from "@/hooks/use-auth";
import { ROLES, ALL_PERMISSIONS, BLOCK_PERMISSIONS } from "@/lib/permissions";
import { Shield, Check, X } from "lucide-react";

// Group permissions by resource
function groupPermissions(perms: string[]) {
  const groups: Record<string, string[]> = {};
  for (const p of perms) {
    const [resource] = p.split(":");
    (groups[resource] ??= []).push(p);
  }
  return groups;
}

export function Roles() {
  useVimScroll();
  const { role: currentRole, tokenRoles } = useAuth();
  const grouped = groupPermissions([...ALL_PERMISSIONS]);

  return (
    <ContentArea>
      <PageHeader
        title="Roles & Permissions"
        subtitle="RBAC role definitions and permission mappings"
      />

      {/* Current session */}
      <div className="mb-6">
        <SectionLabel>Current Session</SectionLabel>
        <Card>
          <CardContent>
            <div className="flex items-center gap-3">
              <Badge variant="muted">
                <Shield className="h-3 w-3" />
                {currentRole.label}
              </Badge>
              <Text variant="muted">{currentRole.description}</Text>
              <Text variant="muted" className="ml-auto">
                {currentRole.permissions.size} permissions
              </Text>
            </div>
            {tokenRoles.length > 0 && (
              <Text variant="muted" as="p" className="mt-2">
                Token roles: {tokenRoles.join(", ")}
              </Text>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Role definitions */}
      <div className="mb-6">
        <SectionLabel>Role Definitions</SectionLabel>
        <DataTable
          compact={false}
          rows={ROLES}
          getRowKey={(r) => r.name}
          columns={[
            {
              header: "Role",
              cell: (r) => <Badge variant="muted">{r.label}</Badge>,
            },
            {
              header: "Description",
              cell: (r) => <span className="text-text">{r.description}</span>,
            },
            {
              header: "Permissions",
              align: "right",
              cell: (r) => (
                <span className="text-text-muted">{r.permissions.size}</span>
              ),
            },
          ]}
        />
      </div>

      {/* Permission matrix */}
      <div className="mb-6">
        <SectionLabel>Permission Matrix</SectionLabel>
        <DataTable
          compact={false}
          rows={Object.entries(grouped).flatMap(([resource, perms]) =>
            perms.map((perm) => ({ resource, perm })),
          )}
          getRowKey={(row) => row.perm}
          group={(row) => row.resource}
          columns={[
            {
              header: "Permission",
              cell: (row) => (
                <span className="font-mono text-text-muted">{row.perm}</span>
              ),
            },
            // A role has a permission or not: yes or a blank, as osapi
            // token roles prints it. A yes is not a status, so it is not
            // green.
            ...ROLES.map((r) => ({
              header: r.label,
              align: "center" as const,
              cell: (row: { resource: string; perm: string }) =>
                r.permissions.has(
                  row.perm as (typeof ALL_PERMISSIONS)[number],
                ) ? (
                  <span className="text-text">yes</span>
                ) : null,
            })),
          ]}
        />
      </div>

      {/* Block permissions */}
      <div className="mb-6">
        <SectionLabel>Block Permissions</SectionLabel>
        <Text variant="muted" as="p" className="mb-2">
          Required permission for each Configure block type
        </Text>
        <DataTable
          rows={Object.entries(BLOCK_PERMISSIONS)}
          getRowKey={([block]) => block}
          columns={[
            {
              header: "Block",
              cell: ([block]) => (
                <span className="font-mono text-text">{block}</span>
              ),
            },
            {
              header: "Permission",
              cell: ([, perm]) => (
                <span className="font-mono text-text-muted">{perm}</span>
              ),
            },
            ...ROLES.map((r) => ({
              header: r.label,
              align: "center" as const,
              cell: ([, perm]: [
                string,
                (typeof BLOCK_PERMISSIONS)[keyof typeof BLOCK_PERMISSIONS],
              ]) => {
                const has = r.permissions.has(perm);
                return has ? (
                  <Check className="mx-auto h-3.5 w-3.5 text-primary" />
                ) : (
                  <X className="mx-auto h-3.5 w-3.5 text-text-muted/30" />
                );
              },
            })),
          ]}
        />
      </div>
    </ContentArea>
  );
}
