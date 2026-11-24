import { useLayoutEffect, useRef } from "react";
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
  // The callbacks are read when a command runs, not when the commands are
  // registered. A caller's callback is often a new function on each render,
  // and registering on each new one re-renders the page, which registers
  // again: a loop that never lets a navigation away from the page finish.
  const calls = useRef({ run, clear, showYAML, add });
  useLayoutEffect(() => {
    calls.current = { run, clear, showYAML, add };
  });
  useCommands(
    [
      {
        id: "cmd:run",
        name: "run",
        description: "Run the plan",
        category: "actions",
        action: () => {
          if (canRun && !running) calls.current.run();
        },
      },
      {
        id: "cmd:clear",
        name: "clear",
        description: "Remove each step",
        category: "actions",
        action: () => calls.current.clear(),
      },
      {
        id: "cmd:yaml",
        name: "yaml",
        description: "Show the plan as YAML",
        category: "actions",
        action: () => calls.current.showYAML(),
      },
      ...defs
        .filter((d) => allowed(d.permission))
        .map((d) => ({
          id: `step:${d.run}`,
          name: d.run.replace(".", " "),
          description: d.doc,
          category: "steps",
          action: () => calls.current.add(d),
        })),
    ],
    [canRun, running, defs, allowed],
  );
}
