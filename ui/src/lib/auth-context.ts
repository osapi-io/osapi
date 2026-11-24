import { createContext } from "react";
import type { RoleDefinition, Permission } from "@/lib/permissions";
import type { SessionState } from "@/lib/session-context";
import type { OperationID } from "@/sdk/gen/schemas";

export interface AuthState extends SessionState {
  /** Current effective role */
  role: RoleDefinition;
  /** Role override from dropdown (null = use token role) */
  roleOverride: string | null;
  /** Check if current role has a permission */
  can: (permission: Permission) => boolean;
  /**
   * Whether the current role may call an API operation: the permission that
   * the API checks, from the operations of the catalog. A page and its link
   * ask this of the operation that they call, so they cannot disagree.
   */
  canCall: (operation: OperationID) => boolean;
  /** Check if current role has all permissions */
  canAll: (permissions: Permission[]) => boolean;
  /** Set the role override (dropdown) */
  setRoleOverride: (role: string | null) => void;
}

export const AuthContext = createContext<AuthState | null>(null);
