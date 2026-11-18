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
