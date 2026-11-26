import type { PKICert } from "@/sdk/gen/schemas";

/** The days left below which a certificate needs attention. */
export const CERT_WARN_DAYS = 30;

/** CertSummary is what the dashboard says about the certificates. */
export interface CertSummary {
  /** The certificate that expires first, of those that osapi read. */
  soonest?: PKICert;
  /** How many osapi read, and how many the fleet has. */
  ok: number;
  total: number;
}

/**
 * certSummary is the certificate that expires first, of the ones that
 * osapi read, and the count of those. A certificate that a host has not
 * made is not one that expires.
 */
export function certSummary(items: PKICert[]): CertSummary {
  const read = items.filter(
    (c) => c.state === "ok" && typeof c.days_left === "number",
  );
  const soonest = read.reduce<PKICert | undefined>(
    (min, c) =>
      min === undefined || (c.days_left ?? 0) < (min.days_left ?? 0) ? c : min,
    undefined,
  );
  return { soonest, ok: read.length, total: items.length };
}
