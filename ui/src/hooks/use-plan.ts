import { useContext } from "react";
import { PlanContext, type PlanState } from "@/lib/plan-context";

export type { PlanState };

/** usePlan is the plan that is being composed, from PlanProvider. */
export function usePlan(): PlanState {
  const ctx = useContext(PlanContext);
  if (!ctx) throw new Error("usePlan must be used within PlanProvider");
  return ctx;
}
