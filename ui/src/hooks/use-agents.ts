import { getAgents } from "@/sdk/gen/agent-management-api-agent-operations/agent-management-api-agent-operations";
import type { AgentInfo, ListAgentsResponse } from "@/sdk/gen/schemas";
import { useAuth } from "@/hooks/use-auth";
import { unwrap, usePolling } from "@/hooks/use-polling";

export type { AgentInfo };

/** The agents of the fleet, by hostname, each intervalMs. */
export function useAgents(intervalMs = 10000) {
  const { canCall } = useAuth();
  const { data, error, loading, refresh } = usePolling(
    async () => {
      const body = unwrap<ListAgentsResponse>(
        await getAgents(),
        "Failed to fetch agents",
      );
      // The registry gives no stable order, so the cards would swap places
      // on each poll.
      return [...body.items].sort((a, b) =>
        a.hostname.localeCompare(b.hostname),
      );
    },
    [],
    { enabled: canCall("GetAgents"), intervalMs },
  );
  return { agents: data ?? [], error, loading, refresh };
}
