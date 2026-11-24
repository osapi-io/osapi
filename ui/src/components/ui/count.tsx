import { cn } from "@/lib/cn";
import { useCatalog } from "@/hooks/use-catalog";

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
 * is dim, and a count above zero takes its kind's tone from the count tones
 * of the catalog: the CLI's CountTones.
 */
export function Count({ n, kind, bare }: CountProps) {
  const { countTones } = useCatalog();
  const tone = countTones[kind.toLowerCase()];
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
