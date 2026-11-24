import { useJob } from "@/hooks/use-job";
import { Card } from "@/components/ui/card";
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
    <Card variant="inset" className="px-4 py-3">
      <ChainSection hosts={hosts} />
    </Card>
  );
}
