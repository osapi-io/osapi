/** What one host's result counts as in a summary line, as the CLI counts it. */
export type HostOutcome = "ok" | "changed" | "skipped" | "failed";

/** hostOutcome is the outcome of one host's result. */
export function hostOutcome(r: {
  status?: string;
  error?: string;
  changed?: boolean;
  exit_code?: number;
}): HostOutcome {
  if (r.status === "skipped" || r.status === "skip") return "skipped";
  if (r.status === "failed" || r.error) return "failed";
  if (r.exit_code !== undefined && r.exit_code !== 0) return "failed";
  return r.changed ? "changed" : "ok";
}

const worst: HostOutcome[] = ["failed", "skipped", "changed", "ok"];

/**
 * hostOutcomes is one outcome for each host of a result's entries, the
 * worst of the host's entries, as the CLI's summary counts hosts: a result
 * with several entries for one host, such as a cron list, counts it once.
 */
export function hostOutcomes(
  results: {
    hostname?: string;
    status?: string;
    error?: string;
    changed?: boolean;
    exit_code?: number;
  }[],
): HostOutcome[] {
  const byHost = new Map<string, HostOutcome>();
  results.forEach((r, i) => {
    const host = r.hostname ?? `#${i}`;
    const o = hostOutcome(r);
    const cur = byHost.get(host);
    if (!cur || worst.indexOf(o) < worst.indexOf(cur)) byHost.set(host, o);
  });
  return [...byHost.values()];
}
