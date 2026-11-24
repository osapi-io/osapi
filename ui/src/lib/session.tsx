import { useState, useCallback, type ReactNode } from "react";
import { setAuthToken as setAuthTokenValue } from "@/lib/auth-token";
import { SessionContext } from "@/lib/session-context";
import { jwtClaimStrings } from "@/lib/jwt";

const STORAGE_KEY_TOKEN = "osapi-token";

// Token from env var (set in .env.local) — skips sign-in when present
const ENV_TOKEN = import.meta.env.OSAPI_BEARER_TOKEN || "";

/**
 * SessionProvider keeps the token. It needs no catalog, so the sign-in page
 * works before the catalog loads.
 */
export function SessionProvider({ children }: { children: ReactNode }) {
  const [token, setTokenState] = useState<string | null>(() => {
    return ENV_TOKEN || localStorage.getItem(STORAGE_KEY_TOKEN) || null;
  });

  const tokenRoles = token ? jwtClaimStrings(token, "roles") : [];
  const tokenTargets = token ? jwtClaimStrings(token, "targets") : [];
  const isAuthenticated = token !== null && tokenRoles.length > 0;

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
    <SessionContext.Provider
      value={{
        isAuthenticated,
        token,
        tokenRoles,
        tokenTargets,
        setToken,
        clearToken,
      }}
    >
      {children}
    </SessionContext.Provider>
  );
}
