import { Check, Lock, TriangleAlert } from "lucide-react";
import { Card } from "@/components/ui/card";
import { Text } from "@/components/ui/text";
import type { Preflight } from "@/lib/plan";

/**
 * PlanPreflight is what a plan needs before it runs, as the controller will
 * check it: each permission of its steps against the viewer's role, the
 * hosts where a method does not run, and the steps that wait for input or a
 * target. It refuses nothing; Run runs each step that can run.
 */
export function PlanPreflight({
  check,
  role,
}: {
  check: Preflight;
  role: string;
}) {
  const missing = check.permissions.filter((p) => !p.allowed);
  return (
    <Card className="mb-4 space-y-2 py-3">
      <div className="flex flex-wrap items-baseline gap-x-2 gap-y-1 text-xs">
        <span className="w-20 shrink-0 text-text-muted">Needs</span>
        {check.permissions.length === 0 && (
          <span className="text-text-muted">no permission</span>
        )}
        {check.permissions.map((p) => (
          <span
            key={p.permission}
            title={`For ${p.runs.join(", ")}`}
            className="inline-flex items-center gap-1 font-mono text-text"
          >
            {p.allowed ? (
              <Check className="h-3 w-3 text-text-muted" />
            ) : (
              <Lock className="h-3 w-3 text-status-error" />
            )}
            {p.permission}
          </span>
        ))}
        <span className="ml-auto text-text-muted">
          {missing.length === 0
            ? `your role, ${role}, has each`
            : `your role, ${role}, does not have ${missing.length}: the controller refuses ${missing.length === 1 ? "that step" : "those steps"}`}
        </span>
      </div>
      {check.skips.length > 0 && (
        <div className="flex gap-x-2 text-xs">
          <span className="w-20 shrink-0 text-text-muted">Skips</span>
          <ul className="space-y-0.5">
            {check.skips.map((s) => (
              <li
                key={`${s.stepId}-${s.hostname}`}
                className="flex items-center gap-1.5"
              >
                <TriangleAlert className="h-3 w-3 text-status-pending" />
                <span className="font-mono text-text">{s.run}</span>
                <span className="text-text-muted">
                  does not run on {s.hostname} ({s.platform}); the host reports
                  skipped
                </span>
              </li>
            ))}
          </ul>
        </div>
      )}
      {check.pending > 0 && (
        <Text variant="warn" as="p" className="pl-22">
          {check.pending} {check.pending === 1 ? "step waits" : "steps wait"}{" "}
          for input or a target, and Run leaves{" "}
          {check.pending === 1 ? "it" : "them"} out.
        </Text>
      )}
    </Card>
  );
}
