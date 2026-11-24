import { useState, useCallback, type ReactNode } from "react";
import { OPERATION_PERMISSIONS, type OperationId } from "@/lib/operations.gen";
import {
  ROLES,
  getRoleByName,
  getRoleByJwtRole,
  hasPermission,
  hasAllPermissions,
  type Permission,
} from "@/lib/permissions";
import { setAuthToken as setAuthTokenValue } from "@/lib/auth-token";
import { AuthContext } from "@/lib/auth-context";
import { jwtClaimStrings } from "@/lib/jwt";

// ---------------------------------------------------------------------------
// Auth provider
// ---------------------------------------------------------------------------

const STORAGE_KEY_ROLE = "osapi-role-override";
const STORAGE_KEY_TOKEN = "osapi-token";

// Token from env var (set in .env.local) — skips sign-in when present
const ENV_TOKEN = import.meta.env.OSAPI_BEARER_TOKEN || "";

export function AuthProvider({ children }: { children: ReactNode }) {
  const [roleOverride, setRoleOverrideState] = useState<string | null>(() => {
    return localStorage.getItem(STORAGE_KEY_ROLE);
  });

  const [token, setTokenState] = useState<string | null>(() => {
    return ENV_TOKEN || localStorage.getItem(STORAGE_KEY_TOKEN) || null;
  });

  const tokenRoles = token ? jwtClaimStrings(token, "roles") : [];
  const tokenTargets = token ? jwtClaimStrings(token, "targets") : [];
  const isAuthenticated = token !== null && tokenRoles.length > 0;

  let role;
  if (roleOverride) {
    role = getRoleByName(roleOverride) ?? ROLES[ROLES.length - 1];
  } else if (tokenRoles.length > 0) {
    role =
      getRoleByJwtRole(tokenRoles[0]) ??
      getRoleByName(tokenRoles[0]) ??
      ROLES[ROLES.length - 1];
  } else {
    role = ROLES[ROLES.length - 1];
  }

  const can = useCallback(
    (permission: Permission) => hasPermission(role, permission),
    [role],
  );

  const canCall = useCallback(
    (operation: OperationId) =>
      hasPermission(role, OPERATION_PERMISSIONS[operation]),
    [role],
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

  const setToken = useCallback((t: string) => {
    setAuthTokenValue(t);
    setTokenState(t);
    localStorage.setItem(STORAGE_KEY_TOKEN, t);
  }, []);

  const clearToken = useCallback(() => {
    setAuthTokenValue(null);
    setTokenState(null);
    localStorage.removeItem(STORAGE_KEY_TOKEN);
  }, []);

  return (
    <AuthContext.Provider
      value={{
        isAuthenticated,
        role,
        roleOverride,
        token,
        tokenRoles,
        tokenTargets,
        can,
        canCall,
        canAll,
        setRoleOverride,
        setToken,
        clearToken,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}
