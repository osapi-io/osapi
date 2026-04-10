// Module-level token store for the fetch mutator (non-React code).
// Kept separate from auth.tsx to avoid react-refresh warnings.
//
// Initialized synchronously at import time from env var or localStorage so
// that fetches fired during the first render (before AuthProvider's useEffect
// runs) include the Bearer header.

const STORAGE_KEY_TOKEN = "osapi-token";
const ENV_TOKEN = import.meta.env.OSAPI_BEARER_TOKEN || "";

let _currentToken: string | null =
  ENV_TOKEN || localStorage.getItem(STORAGE_KEY_TOKEN) || null;

export function getAuthToken(): string | null {
  return _currentToken;
}

export function setAuthToken(token: string | null) {
  _currentToken = token;
}
