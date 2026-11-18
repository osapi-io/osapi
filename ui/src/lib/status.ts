import { STATUS_TONES, type Tone } from "@/lib/status.gen";

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
 * "ok nats://..." both match. A word the CLI does not color is muted.
 * STATUS_TONES is generated from internal/cli/theme.go.
 */
export function statusVariant(status?: string | null): StatusVariant {
  const word = (status ?? "").trim().split(/\s+/)[0].toLowerCase();
  const tone = STATUS_TONES[word];

  return tone ? variants[tone] : "muted";
}

/** isOK reports whether a status word is one the CLI prints in green. */
export function isOK(status?: string | null): boolean {
  return statusVariant(status) === "ready";
}

/** statusDot is the HealthDot color of a status word. */
export function statusDot(
  status?: string | null,
): "ok" | "warn" | "error" | "muted" {
  const v = statusVariant(status);

  return v === "ready" ? "ok" : v === "pending" ? "warn" : v;
}

/** statusText is the Text variant of a status word. */
export function statusText(
  status?: string | null,
): "primary" | "warn" | "error" | "default" {
  const v = statusVariant(status);

  return v === "ready"
    ? "primary"
    : v === "pending"
      ? "warn"
      : v === "error"
        ? "error"
        : "default";
}
