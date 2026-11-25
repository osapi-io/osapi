import { useMemo } from "react";
import { useAgents } from "@/hooks/use-agents";
import { useCatalog } from "@/hooks/use-catalog";

export interface TargetOption {
  value: string;
  label: string;
  group: "builtin" | "hostname" | "label";
}

export function useTargets() {
  const { agents } = useAgents(30000);
  const { targets } = useCatalog();

  const options = useMemo<TargetOption[]>(() => {
    const opts: TargetOption[] = targets.map((t) => ({
      value: t.value,
      label: t.label,
      group: "builtin",
    }));

    // Add hostnames
    for (const agent of agents) {
      if (agent.hostname) {
        opts.push({
          value: agent.hostname,
          label: agent.hostname,
          group: "hostname",
        });
      }
    }

    // Collect unique label key:value pairs across all agents
    const seen = new Set<string>();
    for (const agent of agents) {
      if (agent.labels) {
        for (const [key, val] of Object.entries(agent.labels)) {
          const target = `${key}:${val}`;
          if (!seen.has(target)) {
            seen.add(target);
            opts.push({ value: target, label: target, group: "label" });
          }
        }
      }
    }

    return opts;
  }, [agents, targets]);

  return { options };
}
