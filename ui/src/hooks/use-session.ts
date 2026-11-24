import { useContext } from "react";
import { SessionContext, type SessionState } from "@/lib/session-context";

export type { SessionState };

/** useSession is the token of the person who signed in. */
export function useSession(): SessionState {
  const ctx = useContext(SessionContext);
  if (!ctx) throw new Error("useSession must be used within SessionProvider");
  return ctx;
}
