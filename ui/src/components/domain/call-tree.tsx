import { Link } from "react-router-dom";
import { CornerDownRight } from "lucide-react";
import { runName } from "@/lib/catalog-index";
import { callRows } from "@/lib/calls";
import type { CatalogCall } from "@/lib/catalog";

/**
 * CallTree is the chain of a method: the method, then each call that its
 * examples make to another behavior, under the call that made it. Each call
 * opens the page of its behavior. The job view draws the same tree from a
 * run, with what each call changed.
 */
export function CallTree({
  behavior,
  method,
  calls,
}: {
  behavior: string;
  method: string;
  calls: readonly CatalogCall[];
}) {
  return (
    <ul className="space-y-0.5 font-mono text-xs">
      <li className="text-text">{runName(behavior, method)}</li>
      {callRows(calls).map(({ depth, call }, i) => (
        <li
          key={i}
          className="flex items-center gap-2"
          style={{ paddingLeft: `${depth * 1.25}rem` }}
        >
          <CornerDownRight className="ml-1 h-3 w-3 shrink-0 text-text-muted/60" />
          <Link
            to={`/behaviors/${call.behavior}#${call.method}`}
            className="text-text-muted transition-colors hover:text-heading-1"
          >
            {runName(call.behavior, call.method)}
          </Link>
        </li>
      ))}
    </ul>
  );
}
