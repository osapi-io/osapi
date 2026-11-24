import { getHealthStatus } from "@/sdk/gen/health-check-api-health/health-check-api-health";
import type { StatusResponse } from "@/sdk/gen/schemas";
import { useAuth } from "@/hooks/use-auth";
import { unwrap, usePolling } from "@/hooks/use-polling";

/** The health of each component, each intervalMs. */
export function useHealth(intervalMs = 10000) {
  const { canCall } = useAuth();
  const { data, error, loading } = usePolling(
    async () =>
      unwrap<StatusResponse>(await getHealthStatus(), "Failed to fetch health"),
    [],
    { enabled: canCall("getHealthStatus"), intervalMs },
  );
  return { data, error, loading };
}
