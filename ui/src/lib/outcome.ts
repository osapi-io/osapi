import { RowStatus } from "@/sdk/gen/schemas";

/**
 * What one host's result counts as in a summary line: its status, or
 * changed for a host whose write changed it.
 */
export type HostOutcome = RowStatus | "changed";

/**
 * hostOutcome is the outcome of one host's result. A skipped or timed-out
 * host carries an error, so its status is read first: neither is a failure.
 */
export function hostOutcome(r: {
  status?: RowStatus;
  error?: string;
  changed?: boolean;
  exit_code?: number;
}): HostOutcome {
  if (r.status === RowStatus.skipped) return RowStatus.skipped;
  if (r.status === RowStatus.timeout) return RowStatus.timeout;
  if (r.status === RowStatus.failed || r.error) return RowStatus.failed;
  if (r.exit_code !== undefined && r.exit_code !== 0) return RowStatus.failed;
  return r.changed ? "changed" : RowStatus.ok;
}

const worst: HostOutcome[] = [
  RowStatus.failed,
  RowStatus.timeout,
  RowStatus.skipped,
  "changed",
  RowStatus.ok,
];

/**
 * hostOutcomes is one outcome for each host of a result's entries, the
 * worst of the host's entries, as the CLI's summary counts hosts: a result
 * with several entries for one host, such as a cron list, counts it once.
 */
export function hostOutcomes(
  results: {
    hostname?: string;
    status?: RowStatus;
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
