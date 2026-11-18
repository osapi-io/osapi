import type { TraceSpan } from "@/sdk/gen/schemas";
import { formatBytes } from "@/lib/format";

// Mirrors internal/cli/trace.go, so the CLI and the UI show a trace the same
// way.

// The attributes a span's detail shows when it has nothing more specific:
// what the span acted on.
const DETAIL = [
  "file.path",
  "nats.object.name",
  "messaging.destination.name",
  "db.namespace",
  "http.response.status_code",
];

// The attribute that a store or client operation's detail shows, by span
// name: what it acted on. Its first attribute that is set wins.
const OP_DETAIL: Record<string, string[]> = {
  "job.get": ["job.id"],
  "job.retry": ["job.id"],
  "job.delete": ["job.id"],
  "agent.get": ["osapi.target"],
  "agent.timeline": ["osapi.agent.hostname"],
  "enrollment.accept": ["osapi.agent.hostname", "osapi.agent.fingerprint"],
  "enrollment.reject": ["osapi.agent.hostname"],
  "enrollment.remove": ["osapi.agent.hostname"],
  "audit.get": ["osapi.audit.id"],
  "traces.read": ["osapi.trace.id"],
};

// The attributes that count what an operation read, and the nouns of the
// count: job.list's job.count is "2 jobs".
const OP_COUNTS: [string, string, string][] = [
  ["job.count", "job", "jobs"],
  ["osapi.agent.count", "agent", "agents"],
  ["osapi.audit.count", "entry", "entries"],
  ["osapi.trace.spans", "span", "spans"],
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

// isNATSSpan reports whether a span is a NATS call that nats-client made: the
// bookkeeping of the job queue.
export function isNATSSpan(name: string): boolean {
  return ["kv.", "objectstore.", "stream.", "nats."].some((p) =>
    name.startsWith(p),
  );
}

// natsCalls counts the NATS spans and the time they took.
export function natsCalls(spans: TraceSpan[]): { count: number; ms: number } {
  const nats = spans.filter((s) => isNATSSpan(s.name));
  return {
    count: nats.length,
    ms: nats.reduce((t, s) => t + s.duration_ms, 0),
  };
}

// keyMeaning says what an osapi key is for, from its shape.
function keyMeaning(key: string, bucket: string): string {
  const parts = key.split(".");
  if (parts[0] === "jobs" && parts.length === 2) return "job";
  if (parts[0] === "status" && parts.length >= 3) return `status: ${parts[2]}`;
  if (parts[0] === "responses" && key.endsWith(".>")) return "answers";
  if (parts[0] === "responses") return "response";
  if (parts[0] === "agents") return "registration";
  if (parts[0] === "facts") return "facts";
  if (parts[0] === "drain") return "drain flag";
  if (parts[0] === "timeline" && parts.length >= 3)
    return `timeline: ${parts[2]}`;
  if (bucket.endsWith("file-state")) return "file record";
  return key;
}

const INPUT = "osapi.behavior.input.";

function spanStatus(s: TraceSpan): string {
  if (s.name === "job.process" && s.attributes["job.status"]) {
    return s.attributes["job.status"];
  }
  if (s.status === "error") return "error";
  if (s.attributes["osapi.behavior.changed"] === "true") return "changed";
  if (s.attributes["osapi.behavior.changed"] === "false") return "unchanged";
  return "ok";
}

function spanDetail(s: TraceSpan): string {
  const a = s.attributes;
  if (s.status === "error" && s.status_message) return s.status_message;
  const inputs = Object.keys(a)
    .filter((k) => k.startsWith(INPUT))
    .sort()
    .map((k) => `${k.slice(INPUT.length)}=${a[k]}`)
    .join(" ");
  if (inputs) return inputs;
  if (a["nats.kv.key"])
    return keyMeaning(a["nats.kv.key"], a["db.namespace"] ?? "");
  for (const [key, one, many] of OP_COUNTS) {
    const n = a[key];
    if (n !== undefined && n !== "") return `${n} ${n === "1" ? one : many}`;
  }
  const op = (OP_DETAIL[s.name] ?? []).map((k) => a[k]).find((v) => v);
  if (op) return op;
  if (s.name === "kv.keys") return `list keys in ${a["db.namespace"] ?? ""}`;
  // The request span has the size too, from otelecho; it shows its status.
  const size = a["http.response.body.size"];
  if (size && s.name === "response.write") return formatBytes(Number(size));
  // A job's subject names a host by its machine ID; the operation and the
  // target that the caller gave say the same thing in words.
  if (a["job.type"] && a["osapi.target"])
    return `${a["job.type"]} to ${a["osapi.target"]}`;
  return DETAIL.map((k) => a[k]).find((v) => v) ?? "";
}

// traceRows returns the spans as a tree: each child under its parent, in
// start order. Without all, the NATS spans are left out, and a span under one
// moves up to the span above it.
export function traceRows(spans: TraceSpan[], all: boolean): TraceRow[] {
  const byId = new Map(spans.map((s) => [s.span_id, s]));
  const parentOf = (s: TraceSpan): string => {
    let p = s.parent_span_id ?? "";
    for (;;) {
      const ps = byId.get(p);
      if (!ps) return "";
      if (all || !isNATSSpan(ps.name)) return p;
      p = ps.parent_span_id ?? "";
    }
  };
  const shown = spans.filter((s) => all || !isNATSSpan(s.name));
  const children = new Map<string, TraceSpan[]>();
  const roots: TraceSpan[] = [];
  for (const s of shown) {
    const p = parentOf(s);
    if (!p) {
      roots.push(s);
      continue;
    }
    children.set(p, [...(children.get(p) ?? []), s]);
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
    rows.push({
      span: s,
      lines: lead + branch,
      host,
      status: spanStatus(s),
      detail: spanDetail(s),
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
