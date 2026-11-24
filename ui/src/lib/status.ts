/** The tone of a word or a count, as the CLI colors it. */
export type Tone = "ok" | "warn" | "err";

/** A tone for each word, or each kind of count, from the catalog. */
export type Tones = Readonly<Record<string, Tone>>;

/** A badge variant that a status word can take. */
export type StatusVariant = "ready" | "pending" | "error" | "muted";

const variants: Record<Tone, StatusVariant> = {
  ok: "ready",
  warn: "pending",
  err: "error",
};

/**
 * statusVariant is the badge variant of a status word, in the color the CLI
 * prints it in: green for ok, amber for attention, red for wrong. As the CLI
 * does, it reads the first word of the text, lowercased, so "Ready" and
 * "ok nats://..." both match. A word the CLI does not color is muted. The
 * tones are the CLI's StatusTones, which the catalog serves; useStatus gives
 * them.
 */
export function statusVariant(
  tones: Tones,
  status?: string | null,
): StatusVariant {
  const word = (status ?? "").trim().split(/\s+/)[0].toLowerCase();
  const tone = tones[word];

  return tone ? variants[tone] : "muted";
}

/** statusDot is the HealthDot color of a status word. */
export function statusDot(
  tones: Tones,
  status?: string | null,
): "ok" | "warn" | "error" | "muted" {
  const v = statusVariant(tones, status);

  return v === "ready" ? "ok" : v === "pending" ? "warn" : v;
}

/** statusText is the Text variant of a status word. */
export function statusText(
  tones: Tones,
  status?: string | null,
): "primary" | "warn" | "error" | "default" {
  const v = statusVariant(tones, status);

  return v === "ready"
    ? "primary"
    : v === "pending"
      ? "warn"
      : v === "error"
        ? "error"
        : "default";
}
