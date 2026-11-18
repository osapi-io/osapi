import { cn } from "@/lib/cn";
import { COUNT_TONES } from "@/lib/status.gen";

const toneClass = {
  ok: "text-status-ready",
  warn: "text-status-pending",
  err: "text-status-error",
};

interface CountProps {
  n: number;
  /** What it counts, such as "failed"; it follows the number unless bare. */
  kind: string;
  /** Show the number alone, where a label already says what it counts. */
  bare?: boolean;
}

/**
 * Count is a number of something, as the CLI's cli.Count prints it: a zero
 * is dim, and a count above zero takes its kind's tone from COUNT_TONES,
 * which is generated from internal/cli/theme.go.
 */
export function Count({ n, kind, bare }: CountProps) {
  const tone = COUNT_TONES[kind.toLowerCase()];
  return (
    <span>
      <span
        className={cn(
          n === 0 ? "text-text-muted" : tone ? toneClass[tone] : "text-text",
        )}
      >
        {n}
      </span>
      {!bare && ` ${kind}`}
    </span>
  );
}
