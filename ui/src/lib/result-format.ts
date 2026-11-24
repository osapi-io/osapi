import { hostOutcomes } from "@/lib/outcome";
import { formatBytes, formatDurationMs, formatTime } from "@/lib/format";
import {
  hasValue,
  type CatalogColumn,
  type CatalogMethod,
} from "@/lib/catalog";

/**
 * A value as the CLI's output shows it, by its column's render tag: bytes
 * as a size, a time in the one time format, milliseconds as a duration, and
 * short=N cut to N characters. A list is joined.
 */
export function formatCell(c: CatalogColumn, v: unknown): string {
  if (Array.isArray(v)) return v.map((x) => formatCell(c, x)).join(", ");
  if (v !== null && typeof v === "object") return objectText(c, v);
  const r = c.render ?? "";
  if (r === "bytes" && typeof v === "number") return formatBytes(v);
  if (r === "time" && typeof v === "string") return formatTime(v);
  if (r === "duration=ms" && typeof v === "number") {
    return formatDurationMs(v);
  }
  if (r.startsWith("short=")) {
    const n = Number(r.slice("short=".length));
    const t = String(v);
    return t.length > n ? `${t.slice(0, n)}...` : t;
  }
  if (typeof v === "boolean") return v ? "yes" : "no";
  return String(v);
}

/** An object as one line, each field by its label: Total 1.9 GB, Used 227 MB. */
export function objectText(c: CatalogColumn, v: object): string {
  const cols = c.columns ?? [];
  const o = v as Record<string, unknown>;
  if (cols.length === 0) {
    return Object.entries(o)
      .map(([k, x]) => `${k} ${String(x)}`)
      .join(", ");
  }
  return cols
    .filter((sub) => hasValue(o[sub.name]))
    .map((sub) => `${sub.label.toLowerCase()} ${formatCell(sub, o[sub.name])}`)
    .join(", ");
}

/**
 * What a host's rows count: its rows, or, for one row that is only a list,
 * such as the disks of a host, the items of the list: 4 disks.
 */
export function countOf(
  def: CatalogMethod,
  rows: readonly Record<string, unknown>[],
): string {
  const shown = def.columns.filter((c) => c.name !== "changed");
  const list = shown.length === 1 && shown[0].kind === "list" ? shown[0] : null;
  if (rows.length === 1 && list && Array.isArray(rows[0][list.name])) {
    const n = (rows[0][list.name] as unknown[]).length;
    const noun = list.label.toLowerCase();
    return `${n} ${n === 1 ? noun.replace(/s$/, "") : noun}`;
  }
  return `${rows.length} ${rows.length === 1 ? "row" : "rows"}`;
}

/** The text of every value of a row, for the search. */
export function rowText(cols: readonly CatalogColumn[], r: object): string[] {
  const o = r as Record<string, unknown>;
  return cols.map((c) => (hasValue(o[c.name]) ? formatCell(c, o[c.name]) : ""));
}

export function formatSize(bytes: number) {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

// resultsOf is a result's per-host results, or none.
export function resultsOf(data: unknown): Parameters<typeof hostOutcomes>[0] {
  const r = (data as { results?: unknown })?.results;
  return Array.isArray(r) ? r : [];
}

export function matches(
  term: string,
  ...fields: (string | number | boolean | undefined | null)[]
) {
  if (!term) return true;
  const t = term.toLowerCase();
  return fields.some((f) => f != null && String(f).toLowerCase().includes(t));
}
