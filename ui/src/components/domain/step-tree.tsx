import { CornerDownRight } from "lucide-react";
import { SectionLabel } from "@/components/ui/section-label";
import { Text } from "@/components/ui/text";
import { formatDurationMs } from "@/lib/format";
import { stepRows, type HostSteps } from "@/lib/steps";
import type { Step } from "@/sdk/gen/schemas";

/**
 * StepTree is the calls that a method made on one host: each step as
 * behavior.method, under the step that called it, with whether it changed
 * the host, its error and how long it ran.
 */
function StepTree({ steps }: { steps: Step[] }) {
  return (
    <ul className="space-y-0.5 text-xs">
      {stepRows(steps).map(({ depth, step }, i) => (
        <li
          key={i}
          className="flex items-center gap-2"
          style={{ paddingLeft: `${depth * 1.25}rem` }}
        >
          {depth > 0 && (
            <CornerDownRight className="h-3 w-3 shrink-0 text-text-muted/60" />
          )}
          <span className="font-mono text-text">
            {step.behavior}.{step.method}
          </span>
          {step.error ? (
            <span className="text-status-error">failed: {step.error}</span>
          ) : (
            step.changed && <span className="text-status-applied">changed</span>
          )}
          <span className="text-text-muted">
            {formatDurationMs(step.duration_ms)}
          </span>
        </li>
      ))}
    </ul>
  );
}

/**
 * ChainSection is the chain of calls on each host, as osapi client job get
 * and audit get print it: the tree of each host, or the chain lines when no
 * host has a tree. The job view and the audit view both show it.
 */
export function ChainSection({
  hosts,
  lines,
}: {
  hosts: HostSteps[];
  lines?: string[];
}) {
  if (hosts.length > 0) {
    return (
      <div>
        <SectionLabel>Chain</SectionLabel>
        <div className="space-y-2">
          {hosts.map((c) => (
            <div key={c.hostname}>
              <Text variant="mono-muted" size="xs" as="p" className="mb-0.5">
                {c.hostname}
              </Text>
              <StepTree steps={c.steps} />
            </div>
          ))}
        </div>
      </div>
    );
  }
  if ((lines ?? []).length === 0) {
    return null;
  }

  return (
    <div>
      <SectionLabel>Chain</SectionLabel>
      {lines?.map((line) => (
        <Text key={line} variant="mono" size="xs" as="p">
          {line}
        </Text>
      ))}
    </div>
  );
}
