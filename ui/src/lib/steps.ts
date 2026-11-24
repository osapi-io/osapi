import type { JobDetailResponse, Step } from "@/sdk/gen/schemas";

/** One step of a tree, with how deep it is: 0 for a call that the job made. */
export interface StepRow {
  depth: number;
  step: Step;
  /** Its place in the tree, such as 0.1: a stable key for a row. */
  path: string;
}

/**
 * The steps of a host as rows of a tree, each after the step that called
 * it, as osapi client job get prints them: dns update, then netplan update
 * under it, then managedfile update under that.
 */
export function stepRows(
  steps: readonly Step[],
  depth = 0,
  parent = "",
): StepRow[] {
  return steps.flatMap((step, i) => {
    const path = parent ? `${parent}.${i}` : `${i}`;
    return [
      { depth, step, path },
      ...stepRows(step.steps ?? [], depth + 1, path),
    ];
  });
}

/** The steps of one host. */
export interface HostSteps {
  hostname: string;
  steps: Step[];
}

/** The responses of a job that have steps, each with its host. */
export function hostSteps(job: JobDetailResponse): HostSteps[] {
  return (job.responses ?? [])
    .filter((r) => (r.steps ?? []).length > 0)
    .map((r) => ({ hostname: r.hostname, steps: r.steps ?? [] }));
}
