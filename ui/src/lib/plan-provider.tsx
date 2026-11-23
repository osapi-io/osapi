import {
  useCallback,
  useMemo,
  useReducer,
  useRef,
  type ReactNode,
} from "react";
import { PlanContext, type PlanState } from "@/lib/plan-context";
import { newStep, planReducer, type StepOutcome } from "@/lib/plan";
import type { StepDef } from "@/lib/catalog-index";

/**
 * PlanProvider keeps the plan that is being composed while the person moves
 * between pages, so a method added from a behavior's page is a step of the
 * plan on the Plans page. A plan is not saved: it exports as YAML.
 */
export function PlanProvider({ children }: { children: ReactNode }) {
  const [steps, dispatch] = useReducer(planReducer, []);
  const nextId = useRef(1);

  const addStep = useCallback((def: StepDef): string => {
    const id = `${def.run}-${nextId.current++}`;
    dispatch({ type: "add", step: newStep(def, id) });
    return id;
  }, []);
  const removeStep = useCallback(
    (id: string) => dispatch({ type: "remove", id }),
    [],
  );
  const moveStep = useCallback(
    (id: string, by: -1 | 1) => dispatch({ type: "move", id, by }),
    [],
  );
  const setWith = useCallback(
    (id: string, data: Record<string, unknown>) =>
      dispatch({ type: "with", id, with: data }),
    [],
  );
  const setTarget = useCallback(
    (id: string, target: string) => dispatch({ type: "target", id, target }),
    [],
  );
  const setOutcome = useCallback(
    (
      id: string,
      outcome: StepOutcome,
      result?: unknown,
      error?: string,
      traceId?: string,
    ) => dispatch({ type: "outcome", id, outcome, result, error, traceId }),
    [],
  );
  const resetSteps = useCallback(() => dispatch({ type: "reset" }), []);
  const clearSteps = useCallback(() => dispatch({ type: "clear" }), []);

  const value = useMemo<PlanState>(
    () => ({
      steps,
      addStep,
      removeStep,
      moveStep,
      setWith,
      setTarget,
      setOutcome,
      resetSteps,
      clearSteps,
    }),
    [
      steps,
      addStep,
      removeStep,
      moveStep,
      setWith,
      setTarget,
      setOutcome,
      resetSteps,
      clearSteps,
    ],
  );

  return <PlanContext.Provider value={value}>{children}</PlanContext.Provider>;
}
