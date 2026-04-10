import { useVimScroll } from "@/hooks/use-vim-scroll";
import { ContentArea } from "@/components/layout/content-area";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/ui/page-header";
import { SectionLabel } from "@/components/ui/section-label";
import { DataTable } from "@/components/ui/data-table";
import { Text } from "@/components/ui/text";
import { useAuth } from "@/lib/auth";
import { ROLES, ALL_PERMISSIONS, BLOCK_PERMISSIONS } from "@/lib/permissions";
import { Shield, Check, X } from "lucide-react";
import { cn } from "@/lib/cn";

function roleBadgeVariant(name: string) {
  switch (name) {
    case "admin":
      return "error" as const;
    case "operator":
      return "running" as const;
    default:
      return "muted" as const;
  }
}

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
              <Badge variant={roleBadgeVariant(currentRole.name)}>
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
              cell: (r) => (
                <Badge variant={roleBadgeVariant(r.name)}>{r.label}</Badge>
              ),
            },
            {
              header: "JWT Value",
              cell: (r) => (
                <span
                  className={cn(
                    "font-mono text-text-muted",
                    r.name === currentRole.name && "bg-primary/5",
                  )}
                >
                  {r.jwtRole}
                </span>
              ),
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
        <Card>
          <CardContent className="p-0">
            <table className="w-full text-xs">
              <thead>
                <tr className="border-b border-border/40 text-left text-text-muted">
                  <th className="px-4 py-2.5 font-medium">Permission</th>
                  {ROLES.map((r) => (
                    <th
                      key={r.name}
                      className="px-4 py-2.5 text-center font-medium"
                    >
                      {r.label}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {Object.entries(grouped).map(([resource, perms]) => (
                  <>
                    <tr key={`${resource}-header`}>
                      <td
                        colSpan={ROLES.length + 1}
                        className="border-t border-border/30 bg-card px-4 py-1.5 font-semibold uppercase tracking-wider text-text-muted"
                      >
                        {resource}
                      </td>
                    </tr>
                    {perms.map((perm) => (
                      <tr
                        key={perm}
                        className="border-b border-border/10 last:border-0"
                      >
                        <td className="px-4 py-1.5 font-mono text-text">
                          {perm}
                        </td>
                        {ROLES.map((r) => {
                          const has = r.permissions.has(
                            perm as (typeof ALL_PERMISSIONS)[number],
                          );
                          return (
                            <td
                              key={r.name}
                              className="px-4 py-1.5 text-center"
                            >
                              {has ? (
                                <Check className="mx-auto h-3.5 w-3.5 text-primary" />
                              ) : (
                                <X className="mx-auto h-3.5 w-3.5 text-text-muted/30" />
                              )}
                            </td>
                          );
                        })}
                      </tr>
                    ))}
                  </>
                ))}
              </tbody>
            </table>
          </CardContent>
        </Card>
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
