import { useEffect, useState, useCallback } from "react";
import { listAgents } from "@/sdk/gen/agent-management-api-agent-operations/agent-management-api-agent-operations";
import type { AgentInfo, ListAgentsResponse } from "@/sdk/gen/schemas";

export type { AgentInfo };

export function useAgents(intervalMs = 10000) {
  const [agents, setAgents] = useState<AgentInfo[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshKey, setRefreshKey] = useState(0);

  useEffect(() => {
    let cancelled = false;
    const doFetch = async () => {
      try {
        const result = await listAgents();
        if (cancelled) return;
        if (result.status === 200) {
          setAgents((result.data as ListAgentsResponse).agents);
          setError(null);
        } else {
          setError("Failed to fetch agents");
        }
      } catch (e) {
        if (cancelled) return;
        setError(e instanceof Error ? e.message : "Failed to fetch agents");
      }
      if (!cancelled) setLoading(false);
    };
    doFetch();
    const id = setInterval(doFetch, intervalMs);
    return () => {
      cancelled = true;
      clearInterval(id);
    };
  }, [intervalMs, refreshKey]);

  const refresh = useCallback(() => setRefreshKey((k) => k + 1), []);

  return { agents, error, loading, refresh };
}
