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
  "validate.target": ["osapi.target"],
  "target.resolve": ["osapi.target"],
};

// The attributes that count what an operation read, and the key of the
// count in the detail: job.list's job.count is jobs=2.
const OP_COUNTS: [string, string][] = [
  ["job.count", "jobs"],
  ["osapi.agent.count", "agents"],
  ["osapi.audit.count", "entries"],
  ["osapi.trace.spans", "spans"],
];

// The key that each attribute has in the detail: the word a reader uses for
// it, not the OpenTelemetry name.
const ATTR_KEYS: Record<string, string> = {
  "file.path": "path",
  "nats.object.name": "object",
  "messaging.destination.name": "subject",
  "db.namespace": "bucket",
  "http.response.status_code": "status",
  "job.id": "job",
  "osapi.target": "target",
  "osapi.agent.hostname": "host",
  "osapi.agent.fingerprint": "fingerprint",
  "osapi.audit.id": "entry",
  "osapi.trace.id": "trace",
};

// logfmtValue quotes a value that has a space, an equals sign or a quote,
// as logfmt does, so a reader can tell where it ends.
const logfmtValue = (v: string) =>
  v === "" || /[ ="]/.test(v) ? JSON.stringify(v) : v;

export interface TraceRow {
  span: TraceSpan;
  // lines are the tree lines before the name, such as "│  └─ ".
  lines: string;
  // on is what ran the span, the role of its service: controller, agent or
  // cli. An agent span's host is host= in its job.process's detail.
  on: string;
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

// spanDetail is what a span acted on, as key=value pairs, led by error= on a
// failed span.
function spanDetail(s: TraceSpan): string {
  const pairs: [string, string][] = [];
  if (s.status === "error" && s.status_message)
    pairs.push(["error", s.status_message]);
  pairs.push(...actedOn(s));
  return pairs.map(([k, v]) => `${k}=${logfmtValue(v)}`).join(" ");
}

// actedOn is what a span acted on, as key and value, by the first rule that
// fits it.
function actedOn(s: TraceSpan): [string, string][] {
  const a = s.attributes;
  const inputs = Object.keys(a)
    .filter((k) => k.startsWith(INPUT))
    .sort()
    .map((k): [string, string] => [k.slice(INPUT.length), a[k]]);
  if (inputs.length) return inputs;
  if (s.name === "job.process" && a["osapi.agent.hostname"])
    return [["host", a["osapi.agent.hostname"]]];
  if (a["nats.kv.key"])
    return [["key", keyMeaning(a["nats.kv.key"], a["db.namespace"] ?? "")]];
  for (const [attr, key] of OP_COUNTS) {
    if (a[attr] !== undefined && a[attr] !== "") return [[key, a[attr]]];
  }
  const op = (OP_DETAIL[s.name] ?? []).find((k) => a[k]);
  if (op) return [[ATTR_KEYS[op], a[op]]];
  if (s.name === "kv.keys") return [["bucket", a["db.namespace"] ?? ""]];
  // The request span has the size too, from otelecho; it shows its status.
  const size = a["http.response.body.size"];
  if (size && s.name === "response.write")
    return [["size", formatBytes(Number(size))]];
  // A job's subject names a host by its machine ID; the operation and the
  // target that the caller gave say the same thing in words.
  if (a["job.type"] && a["osapi.target"])
    return [
      ["op", a["job.type"]],
      ["target", a["osapi.target"]],
    ];
  const k = DETAIL.find((d) => a[d]);
  return k ? [[ATTR_KEYS[k], a[k]]] : [];
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
  const walk = (s: TraceSpan, lead: string, branch: string) => {
    rows.push({
      span: s,
      lines: lead + branch,
      on: s.service.replace(/^osapi-/, ""),
      status: spanStatus(s),
      detail: spanDetail(s),
      offset: (startOf(s) - begin) / total,
      width: Math.max(s.duration_ms / total, 0.002),
    });
    const next =
      lead + (branch === "├─ " ? "│  " : branch === "└─ " ? "   " : "");
    const kids = byStart(children.get(s.span_id) ?? []);
    kids.forEach((k, i) =>
      walk(k, next, i === kids.length - 1 ? "└─ " : "├─ "),
    );
  };
  for (const r of byStart(roots)) walk(r, "", "");

  return rows;
}
