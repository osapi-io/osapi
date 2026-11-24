// Format checks for a pasted bearer token, before the sign-in page keeps it.
// The server is the only verifier; this only refuses a token that the server
// is certain to refuse, with a reason the person can act on.

const BASE64URL = /^[A-Za-z0-9_-]+$/;

export type TokenPayload = { roles?: unknown };

export type TokenCheck =
  { ok: true; payload: TokenPayload } | { ok: false; error: string };

// Decodes one base64url segment. atob ignores whitespace, and the server does
// not, so the caller must refuse whitespace before it calls this.
function decodeSegment(segment: string): string {
  const b64 = segment.replace(/-/g, "+").replace(/_/g, "/");
  return atob(b64.padEnd(b64.length + ((4 - (b64.length % 4)) % 4), "="));
}

export function checkTokenFormat(raw: string): TokenCheck {
  const token = raw.trim();

  if (/\s/.test(token)) {
    return {
      ok: false,
      error:
        "The token has a space or a line break in it, usually from a wrapped line in the terminal. Copy it again, e.g. osapi controller token generate ... | awk '/Token:/{print $2}' | pbcopy",
    };
  }

  const parts = token.split(".");
  if (parts.length !== 3) {
    return {
      ok: false,
      error: `A JWT has three parts separated by dots. This token has ${parts.length}.`,
    };
  }

  const bad = parts.findIndex((p) => !BASE64URL.test(p));
  if (bad !== -1) {
    const name = ["header", "payload", "signature"][bad];
    return {
      ok: false,
      error: `The token's ${name} has a character that is not base64url (A-Z, a-z, 0-9, - and _). Copy it again.`,
    };
  }

  let payload: TokenPayload;
  try {
    payload = JSON.parse(decodeSegment(parts[1])) as TokenPayload;
  } catch {
    return { ok: false, error: "The token's payload is not valid JSON." };
  }

  if (!Array.isArray(payload.roles)) {
    return {
      ok: false,
      error:
        "Token is missing roles claim. Generate one with: osapi controller token generate",
    };
  }

  return { ok: true, payload };
}
