import { useEffect, useState } from "react";
import { getHealthStatus } from "@/sdk/gen/health-check-api-health/health-check-api-health";
import type { StatusResponse } from "@/sdk/gen/schemas";

export function useHealth(intervalMs = 10000) {
  const [data, setData] = useState<StatusResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let mounted = true;

    const poll = async () => {
      try {
        const result = await getHealthStatus();
        if (!mounted) return;
        if (result.status === 200 || result.status === 503) {
          setData(result.data as StatusResponse);
          setError(null);
        } else {
          setError("Failed to fetch health");
        }
        setLoading(false);
      } catch (e) {
        if (mounted) {
          setError(e instanceof Error ? e.message : "Failed to fetch health");
          setLoading(false);
        }
      }
    };

    poll();
    const id = setInterval(poll, intervalMs);
    return () => {
      mounted = false;
      clearInterval(id);
    };
  }, [intervalMs]);

  return { data, error, loading };
}
