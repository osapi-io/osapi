import { useState, useCallback, type ReactNode } from "react";
import {
  hasPermission,
  hasAllPermissions,
  type Permission,
} from "@/lib/permissions";
import { AuthContext } from "@/lib/auth-context";
import { useCatalog } from "@/hooks/use-catalog";
import { useSession } from "@/hooks/use-session";
import type { OperationID } from "@/sdk/gen/schemas";

// ---------------------------------------------------------------------------
// Auth provider
// ---------------------------------------------------------------------------

const STORAGE_KEY_ROLE = "osapi-role-override";

/**
 * AuthProvider gives the role of the token, or of the role switcher, and
 * what it may call. The roles and the permission of each operation come from
 * the catalog, so it is inside CatalogProvider.
 */
export function AuthProvider({ children }: { children: ReactNode }) {
  const session = useSession();
  const { roles, operations } = useCatalog();
  const [roleOverride, setRoleOverrideState] = useState<string | null>(() => {
    return localStorage.getItem(STORAGE_KEY_ROLE);
  });

  // The last role has the least: the UI falls back to it.
  const least = roles[roles.length - 1];
  const byName = (name: string | null) =>
    name ? roles.find((r) => r.name === name) : undefined;
  const role =
    (roleOverride ? byName(roleOverride) : byName(session.tokenRoles[0])) ??
    least;

  const can = useCallback(
    (permission: Permission) => hasPermission(role, permission),
    [role],
  );

  const canCall = useCallback(
    (operation: OperationID) => {
      const permission = operations.get(operation);
      return permission !== undefined && hasPermission(role, permission);
    },
    [role, operations],
  );

  const canAll = useCallback(
    (permissions: Permission[]) => hasAllPermissions(role, permissions),
    [role],
  );

  const setRoleOverride = useCallback((r: string | null) => {
    setRoleOverrideState(r);
    if (r) {
      localStorage.setItem(STORAGE_KEY_ROLE, r);
    } else {
      localStorage.removeItem(STORAGE_KEY_ROLE);
    }
  }, []);

  return (
    <AuthContext.Provider
      value={{
        ...session,
        role,
        roleOverride,
        can,
        canCall,
        canAll,
        setRoleOverride,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}
