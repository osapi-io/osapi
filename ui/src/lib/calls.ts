import { runName } from "@/lib/catalog-index";
import type { CatalogCall } from "@/lib/catalog";

/** A call and its depth: 0 for a call that the method makes itself. */
export interface CallRow {
  depth: number;
  call: CatalogCall;
}

/**
 * The calls of a method as rows of a tree, each after the call that made
 * it, as osapi behavior chain get prints them.
 */
export function callRows(calls: readonly CatalogCall[], depth = 0): CallRow[] {
  return calls.flatMap((call) => [
    { depth, call },
    ...callRows(call.calls ?? [], depth + 1),
  ]);
}

/** The calls of a method, each once, in the order that they run. */
export function flatCalls(calls: readonly CatalogCall[]): string[] {
  const out: string[] = [];
  for (const { call } of callRows(calls)) {
    const name = runName(call.behavior, call.method);
    if (!out.includes(name)) out.push(name);
  }
  return out;
}
