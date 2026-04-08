import {
  createContext,
  useContext,
  useState,
  useCallback,
  useEffect,
  type ReactNode,
} from "react";
import {
  ROLES,
  getRoleByName,
  getRoleByJwtRole,
  hasPermission,
  hasAllPermissions,
  type RoleDefinition,
  type Permission,
} from "@/lib/permissions";
import { setAuthToken as setAuthTokenValue } from "@/lib/auth-token";

// ---------------------------------------------------------------------------
// JWT decoding (no verification — that's the server's job)
// ---------------------------------------------------------------------------

function decodeJwtPayload(token: string): Record<string, unknown> | null {
  try {
    const parts = token.split(".");
    if (parts.length !== 3) return null;
    const payload = atob(parts[1].replace(/-/g, "+").replace(/_/g, "/"));
    return JSON.parse(payload);
  } catch {
    return null;
  }
}

function extractRolesFromJwt(token: string): string[] {
  const payload = decodeJwtPayload(token);
  if (!payload) return [];
  // osapi puts roles in the "roles" claim
  const roles = payload.roles;
  if (Array.isArray(roles)) return roles.map(String);
  return [];
}

// ---------------------------------------------------------------------------
// Auth context
// ---------------------------------------------------------------------------

interface AuthState {
  /** Whether the user has a valid token */
  isAuthenticated: boolean;
  /** Current effective role */
  role: RoleDefinition;
  /** Role override from dropdown (null = use token role) */
  roleOverride: string | null;
  /** JWT token if provided */
  token: string | null;
  /** Roles extracted from JWT */
  tokenRoles: string[];
  /** Check if current role has a permission */
  can: (permission: Permission) => boolean;
  /** Check if current role has all permissions */
  canAll: (permissions: Permission[]) => boolean;
  /** Set the role override (dropdown) */
  setRoleOverride: (role: string | null) => void;
  /** Set the JWT token */
  setToken: (token: string) => void;
  /** Clear the token */
  clearToken: () => void;
}

const AuthContext = createContext<AuthState | null>(null);

const STORAGE_KEY_ROLE = "osapi-role-override";
const STORAGE_KEY_TOKEN = "osapi-token";

// Token from env var (set in .env.local) — skips sign-in when present
const ENV_TOKEN = import.meta.env.OSAPI_BEARER_TOKEN || "";

export function AuthProvider({ children }: { children: ReactNode }) {
  const [roleOverride, setRoleOverrideState] = useState<string | null>(() => {
    return localStorage.getItem(STORAGE_KEY_ROLE);
  });

  const [token, setTokenState] = useState<string | null>(() => {
    // Env var takes priority, then localStorage, then null (show sign-in)
    return ENV_TOKEN || localStorage.getItem(STORAGE_KEY_TOKEN) || null;
  });

  // Keep module-level token in sync for the fetch mutator
  useEffect(() => {
    setAuthTokenValue(token);
  }, [token]);

  const tokenRoles = token ? extractRolesFromJwt(token) : [];
  const isAuthenticated = token !== null && tokenRoles.length > 0;

  // Resolve effective role: override > token > default to viewer
  let role: RoleDefinition;
  if (roleOverride) {
    role = getRoleByName(roleOverride) ?? ROLES[ROLES.length - 1];
  } else if (tokenRoles.length > 0) {
    // Use the highest-privilege role from the token
    role =
      getRoleByJwtRole(tokenRoles[0]) ??
      getRoleByName(tokenRoles[0]) ??
      ROLES[ROLES.length - 1];
  } else {
    role = ROLES[ROLES.length - 1]; // viewer
  }

  const can = useCallback(
    (permission: Permission) => hasPermission(role, permission),
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
    setTokenState(t);
    localStorage.setItem(STORAGE_KEY_TOKEN, t);
  }, []);

  const clearToken = useCallback(() => {
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
        can,
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

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
