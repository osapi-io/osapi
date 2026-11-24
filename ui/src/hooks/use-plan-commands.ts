import { useCommands } from "@/hooks/use-commands";
import type { StepDef } from "@/lib/catalog-index";

interface PlanCommands {
  canRun: boolean;
  running: boolean;
  run: () => void;
  clear: () => void;
  showYAML: () => void;
  defs: readonly StepDef[];
  allowed: (permission?: string) => boolean;
  add: (def: StepDef) => void;
}

/** The commands of the command bar on the Plans page: run, clear, yaml, and a step for each method. */
export function usePlanCommands({
  canRun,
  running,
  run,
  clear,
  showYAML,
  defs,
  allowed,
  add,
}: PlanCommands) {
  useCommands(
    [
      {
        id: "cmd:run",
        name: "run",
        description: "Run the plan",
        category: "actions",
        action: () => {
          if (canRun && !running) run();
        },
      },
      {
        id: "cmd:clear",
        name: "clear",
        description: "Remove each step",
        category: "actions",
        action: () => clear(),
      },
      {
        id: "cmd:yaml",
        name: "yaml",
        description: "Show the plan as YAML",
        category: "actions",
        action: () => showYAML(),
      },
      ...defs
        .filter((d) => allowed(d.permission))
        .map((d) => ({
          id: `step:${d.run}`,
          name: d.run.replace(".", " "),
          description: d.doc,
          category: "steps",
          action: () => add(d),
        })),
    ],
    [canRun, running, run, clear, showYAML, defs, allowed, add],
  );
}
