import { createContext } from "react";

/**
 * The token of the person who signed in. The sign-in page sets it, before
 * the catalog loads; the roles and the permissions come with the catalog,
 * in AuthState.
 */
export interface SessionState {
  /** Whether the user has a valid token */
  isAuthenticated: boolean;
  /** JWT token if provided */
  token: string | null;
  /** Roles extracted from JWT */
  tokenRoles: string[];
  /** Hosts the token may act on, from JWT; empty is each host */
  tokenTargets: string[];
  /** Set the JWT token */
  setToken: (token: string) => void;
  /** Clear the token */
  clearToken: () => void;
}

export const SessionContext = createContext<SessionState | null>(null);
