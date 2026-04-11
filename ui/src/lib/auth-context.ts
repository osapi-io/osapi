import { createContext } from "react";
import type { RoleDefinition, Permission } from "@/lib/permissions";

export interface AuthState {
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

export const AuthContext = createContext<AuthState | null>(null);
