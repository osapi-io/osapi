import { createContext } from "react";
import type { StepDef } from "@/lib/catalog-index";
import type { PlanStep, StepOutcome } from "@/lib/plan";

export interface PlanState {
  steps: PlanStep[];
  /** Adds a step that runs def, and gives its id. */
  addStep: (def: StepDef) => string;
  removeStep: (id: string) => void;
  moveStep: (id: string, by: -1 | 1) => void;
  setWith: (id: string, data: Record<string, unknown>) => void;
  setTarget: (id: string, target: string) => void;
  setOutcome: (
    id: string,
    outcome: StepOutcome,
    result?: unknown,
    error?: string,
    traceId?: string,
  ) => void;
  /** Clears the outcome of each step, so the plan can run again. */
  resetSteps: () => void;
  clearSteps: () => void;
}

export const PlanContext = createContext<PlanState | null>(null);
