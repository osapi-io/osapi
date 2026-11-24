import { getJobByID } from "@/sdk/gen/job-management-api-job-operations/job-management-api-job-operations";
import type { JobDetailResponse } from "@/sdk/gen/schemas";
import { useAuth } from "@/hooks/use-auth";
import { unwrap, usePolling } from "@/hooks/use-polling";

/**
 * One job by its ID, read one time, or each intervalMs while it is not
 * done. A null id reads nothing.
 */
export function useJob(id: string | null, intervalMs = 0) {
  const { canCall } = useAuth();
  return usePolling(
    async () => {
      const res = await getJobByID(id as string);
      if (res.status === 404) throw new Error(`Job ${id} not found`);
      return unwrap<JobDetailResponse>(res, "Failed to load the job");
    },
    [id],
    { enabled: id !== null && canCall("getJobByID"), intervalMs },
  );
}
