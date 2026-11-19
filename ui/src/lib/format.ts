// The CLI's formatters, in internal/cli/ui.go, so a value reads the same in
// the UI as in the terminal. Each one returns its input unchanged when the
// input does not parse, as the Go one does.

const pad = (n: number) => String(n).padStart(2, "0");

/** An RFC 3339 time as local "2006-01-02 15:04:05", as FormatTime prints it. */
export function formatTime(s?: string): string {
  if (!s) return "";
  const d = new Date(s);
  if (Number.isNaN(d.getTime())) return s;
  return (
    `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ` +
    `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
  );
}

/** A span in milliseconds as "3d 4h", "12h 30m", "45m" or "30s", as FormatAge prints it. */
export function formatAge(ms: number): string {
  if (!(ms > 0)) return "";
  const totalMinutes = Math.floor(ms / 60_000);
  const totalHours = Math.floor(ms / 3_600_000);
  const days = Math.floor(totalHours / 24);
  const hours = totalHours % 24;
  const minutes = totalMinutes % 60;
  if (days > 0) return `${days}d ${hours}h`;
  if (hours > 0) return `${hours}h ${minutes}m`;
  if (minutes > 0) return `${minutes}m`;
  return `${Math.floor(ms / 1000)}s`;
}

/** The age of an RFC 3339 time, from now. */
export function formatAgeSince(s?: string): string {
  if (!s) return "";
  const t = new Date(s).getTime();
  if (Number.isNaN(t)) return s;
  return formatAge(Date.now() - t);
}

/**
 * The API's uptime, "2 days, 23 hours, 50 minutes", as FormatAge prints it:
 * "2d 23h". A string in a different form comes back unchanged.
 */
export function formatUptime(uptime: string): string {
  const unit = (name: string) => {
    const m = uptime.match(new RegExp(`(\\d+)\\s+${name}s?\\b`));
    return m ? Number(m[1]) : undefined;
  };
  const days = unit("day");
  const hours = unit("hour");
  const minutes = unit("minute");
  if (days === undefined && hours === undefined && minutes === undefined) {
    return uptime;
  }
  return formatAge(
    ((days ?? 0) * 1440 + (hours ?? 0) * 60 + (minutes ?? 0)) * 60_000,
  );
}

// Trims a fixed-point number's trailing zeros, as Go's Duration.String does.
const trim = (n: number, digits: number) =>
  n.toFixed(digits).replace(/\.?0+$/, "");

/**
 * A duration in milliseconds, rounded and printed as FormatDuration does:
 * 6.168625ms is 6ms, 1234ms is 1.2s, 184000ms is 3m4s.
 */
export function formatDurationMs(ms: number): string {
  if (!Number.isFinite(ms)) return String(ms);
  if (ms < 1) return `${trim(Math.round(ms * 1000), 0)}µs`;
  if (ms < 1000) return `${Math.round(ms)}ms`;
  if (ms < 60_000) return `${trim(Math.round(ms / 100) / 10, 1)}s`;

  const total = Math.round(ms / 1000);
  const h = Math.floor(total / 3600);
  const m = Math.floor((total % 3600) / 60);
  const s = total % 60;
  return h > 0 ? `${h}h${m}m${s}s` : `${m}m${s}s`;
}

const goUnits: Record<string, number> = {
  ns: 1e-6,
  us: 1e-3,
  µs: 1e-3,
  ms: 1,
  s: 1000,
  m: 60_000,
  h: 3_600_000,
};

/** A Go duration string, such as "6.168625ms" or "1m2.5s", as FormatDuration prints it. */
export function formatDuration(s?: string): string {
  if (!s) return "";
  const parts = [...s.matchAll(/(\d+(?:\.\d+)?)(ns|us|µs|ms|s|m|h)/g)];
  if (parts.length === 0 || parts.map((p) => p[0]).join("") !== s) return s;
  return formatDurationMs(
    parts.reduce((ms, [, n, unit]) => ms + Number(n) * goUnits[unit], 0),
  );
}

/** A byte count as "5.2 KB" or "1.0 MB", as FormatBytes prints it. */
export function formatBytes(b: number): string {
  const kb = 1024;
  const mb = kb * 1024;
  const gb = mb * 1024;
  if (b >= gb) return `${(b / gb).toFixed(1)} GB`;
  if (b >= mb) return `${(b / mb).toFixed(1)} MB`;
  if (b >= kb) return `${(b / kb).toFixed(1)} KB`;
  return `${b} B`;
}

/** A fingerprint cut at 20 characters, as the CLI's pending agent list prints it. */
export function shortFingerprint(fp: string): string {
  return fp.length > 20 ? `${fp.slice(0, 20)}...` : fp;
}

/**
 * A value for one cell or key-value pair, as the CLI's Summarize prints it:
 * a scalar as itself, a list by its length, an object by its field count,
 * and nothing as "none".
 */
export function summarize(v: unknown): string {
  if (v === null || v === undefined) return "none";
  if (Array.isArray(v)) return v.length === 1 ? "1 item" : `${v.length} items`;
  if (typeof v === "object") {
    const n = Object.keys(v).length;
    return n === 1 ? "1 field" : `${n} fields`;
  }
  return String(v);
}

/** The type of a job's operation, as osapi client job list prints it. */
export function operationType(operation?: { type: string }): string {
  return operation?.type || "Unknown";
}
