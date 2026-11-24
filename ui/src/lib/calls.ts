import { runName } from "@/lib/catalog-index";
import type { CatalogCall } from "@/lib/catalog";

/** A call and its depth: 0 for a call that the method makes itself. */
export interface CallRow {
  depth: number;
  call: CatalogCall;
  /** Its place in the tree, such as 0.1: a stable key for a row. */
  path: string;
}

/**
 * The calls of a method as rows of a tree, each after the call that made
 * it, as osapi dsl chain get prints them.
 */
export function callRows(
  calls: readonly CatalogCall[],
  depth = 0,
  parent = "",
): CallRow[] {
  return calls.flatMap((call, i) => {
    const path = parent ? `${parent}.${i}` : `${i}`;
    return [
      { depth, call, path },
      ...callRows(call.calls ?? [], depth + 1, path),
    ];
  });
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
