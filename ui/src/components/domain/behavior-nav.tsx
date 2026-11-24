import { Link } from "react-router-dom";
import { ChevronLeft } from "lucide-react";
import { usePlan } from "@/hooks/use-plan";

export function BackLink({ area }: { area?: string }) {
  return (
    <Link
      to="/behaviors"
      className="mb-4 inline-flex items-center gap-1 text-sm text-text-muted transition-colors hover:text-text"
    >
      <ChevronLeft className="h-4 w-4" />
      Behaviors
      {area && <span className="text-text-muted/60">/ {area}</span>}
    </Link>
  );
}

/** The way to the plan, with its count of steps, once it has one. */
export function PlanLink() {
  const { steps } = usePlan();
  if (steps.length === 0) return null;
  return (
    <Link
      to="/plans/new"
      className="inline-flex h-8 items-center gap-2 rounded-md border border-border px-3 text-xs font-semibold text-text transition-colors hover:border-white/20"
    >
      Open the plan: {steps.length} {steps.length === 1 ? "step" : "steps"}
    </Link>
  );
}
