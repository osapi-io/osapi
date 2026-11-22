import { cn } from "@/lib/cn";
import type { ReactNode } from "react";

interface KVListProps {
  /** Each pair's label and value; a pair whose value is null is left out. */
  pairs: [string, ReactNode][];
  className?: string;
}

/**
 * KVList is a block of key-value pairs with their values lined up, as the
 * CLI's PrintKV prints one.
 */
export function KVList({ pairs, className }: KVListProps) {
  return (
    <dl
      className={cn(
        "grid grid-cols-[max-content_1fr] gap-x-3 gap-y-0.5 text-xs",
        className,
      )}
    >
      {pairs
        .filter(([, v]) => v != null && v !== false)
        .map(([k, v], i) => (
          <div key={`${k}-${i}`} className="contents">
            <dt className="text-text-muted">{k}:</dt>
            <dd className="min-w-0 text-text">{v}</dd>
          </div>
        ))}
    </dl>
  );
}
