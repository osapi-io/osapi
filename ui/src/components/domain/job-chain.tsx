import { useJob } from "@/hooks/use-job";
import { ChainSection } from "@/components/domain/step-tree";
import { hostSteps } from "@/lib/steps";

/**
 * JobChain is the chain of calls that a step's job made on each host, under
 * the step's result: dns.update, then netplan.update under it, with what
 * each call changed. It shows nothing for a job whose method called no other
 * behavior.
 */
export function JobChain({ jobId }: { jobId: string }) {
  // The result links the job; a chain that does not load is left out.
  const { data: job } = useJob(jobId);

  if (!job) return null;
  const hosts = hostSteps(job);
  if (hosts.length === 0) return null;
  return (
    <div className="rounded-lg border border-border/60 bg-surface px-4 py-3">
      <ChainSection hosts={hosts} />
    </div>
  );
}
