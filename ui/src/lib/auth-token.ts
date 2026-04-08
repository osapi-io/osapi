// Module-level token store for the fetch mutator (non-React code).
// Kept separate from auth.tsx to avoid react-refresh warnings.

let _currentToken: string | null = null;

export function getAuthToken(): string | null {
  return _currentToken;
}

export function setAuthToken(token: string | null) {
  _currentToken = token;
}
