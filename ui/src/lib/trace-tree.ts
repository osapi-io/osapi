import type { TraceSpan } from "@/sdk/gen/schemas";

// The attributes a span's detail shows, the first one it has: what the span
// acted on. Mirrors internal/cli/trace.go, so the CLI and the UI agree.
const DETAIL = [
  "file.path",
  "nats.kv.key",
  "nats.object.name",
  "messaging.destination.name",
  "db.namespace",
  "http.response.status_code",
];

export interface TraceRow {
  span: TraceSpan;
  // lines are the tree lines before the name, such as "│  └─ ".
  lines: string;
  host: string;
  status: string;
  detail: string;
  // offset and width place the span's bar, as fractions of the trace.
  offset: number;
  width: number;
}

const startOf = (s: TraceSpan) => new Date(s.start_time).getTime();

// traceRows returns the spans as a tree: each child under its parent, in
// start order. A span whose parent is not in the trace is a root. The host is
// the agent of the nearest job.process at or above the span.
export function traceRows(spans: TraceSpan[]): TraceRow[] {
  const ids = new Set(spans.map((s) => s.span_id));
  const children = new Map<string, TraceSpan[]>();
  const roots: TraceSpan[] = [];
  for (const s of spans) {
    const parent = s.parent_span_id ?? "";
    if (!parent || !ids.has(parent)) {
      roots.push(s);
      continue;
    }
    children.set(parent, [...(children.get(parent) ?? []), s]);
  }
  const byStart = (list: TraceSpan[]) =>
    [...list].sort((a, b) => startOf(a) - startOf(b));

  const begin = Math.min(...spans.map(startOf));
  const end = Math.max(...spans.map((s) => startOf(s) + s.duration_ms));
  const total = Math.max(end - begin, 1);

  const rows: TraceRow[] = [];
  const walk = (s: TraceSpan, lead: string, branch: string, host: string) => {
    const agent = s.attributes["osapi.agent.hostname"];
    if (s.name === "job.process" && agent) host = agent;

    let status: string = s.status;
    if (s.name === "job.process" && s.attributes["job.status"]) {
      status = s.attributes["job.status"];
    }
    const detail =
      s.status === "error" && s.status_message
        ? s.status_message
        : (DETAIL.map((k) => s.attributes[k]).find((v) => v) ?? "");

    rows.push({
      span: s,
      lines: lead + branch,
      host,
      status,
      detail,
      offset: (startOf(s) - begin) / total,
      width: Math.max(s.duration_ms / total, 0.002),
    });

    const next =
      lead + (branch === "├─ " ? "│  " : branch === "└─ " ? "   " : "");
    const kids = byStart(children.get(s.span_id) ?? []);
    kids.forEach((k, i) =>
      walk(k, next, i === kids.length - 1 ? "└─ " : "├─ ", host),
    );
  };
  for (const r of byStart(roots)) walk(r, "", "", "");

  return rows;
}

// formatMs is a duration in the units a reader can use: µs, ms or s.
export function formatMs(ms: number): string {
  if (ms < 1) return `${Math.round(ms * 1000)}µs`;
  if (ms < 1000) return `${Math.round(ms)}ms`;
  return `${(ms / 1000).toFixed(1)}s`;
}
