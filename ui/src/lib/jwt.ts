// JWT decoding, with no verification: the controller verifies each token.

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

// jwtClaimStrings is the list claim of the token, such as its roles or its
// targets, or an empty list when the token does not have it.
export function jwtClaimStrings(token: string, claim: string): string[] {
  const value = decodeJwtPayload(token)?.[claim];
  if (Array.isArray(value)) return value.map(String);
  return [];
}
