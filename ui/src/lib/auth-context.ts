import { createContext } from "react";
import type { RoleDefinition, Permission } from "@/lib/permissions";
import type { OperationId } from "@/lib/operations.gen";

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
  /** Hosts the token may act on, from JWT; empty is each host */
  tokenTargets: string[];
  /** Check if current role has a permission */
  can: (permission: Permission) => boolean;
  /**
   * Whether the current role may call an API operation: the permission that
   * the API checks, from the generated OPERATION_PERMISSIONS. A page and its
   * link ask this of the operation that they call, so they cannot disagree.
   */
  canCall: (operation: OperationId) => boolean;
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
