import { Count } from "@/components/ui/count";
import type { HostOutcome } from "@/lib/outcome";

interface ResultSummaryProps {
  outcomes: HostOutcome[];
  jobId?: string;
}

const order: HostOutcome[] = ["ok", "changed", "skipped", "timeout", "failed"];

/**
 * ResultSummary is the line under a result, as the CLI's printSummary
 * prints it: "2 hosts: 1 ok, 1 skipped", each count in its tone, then the
 * whole job ID in gray.
 */
export function ResultSummary({ outcomes, jobId }: ResultSummaryProps) {
  const n = outcomes.length;
  const parts = order
    .map((o) => [o, outcomes.filter((x) => x === o).length] as const)
    .filter(([, c]) => c > 0);

  return (
    <span className="text-xs">
      {n} {n === 1 ? "host" : "hosts"}:{" "}
      {parts.map(([o, c], i) => (
        <span key={o}>
          {i > 0 && ", "}
          <Count n={c} kind={o} />
        </span>
      ))}
      {jobId && (
        <span className="ml-3 font-mono text-text-muted">job {jobId}</span>
      )}
    </span>
  );
}
