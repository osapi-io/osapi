import { useCallback, useRef, useSyncExternalStore } from "react";
import { getJobByID } from "@/sdk/gen/job-management-api-job-operations/job-management-api-job-operations";
import type { JobDetailResponse } from "@/sdk/gen/schemas";

export function getJobState(jobId: string) {
  if (!jobStates.has(jobId)) {
    jobStates.set(jobId, {
      listeners: new Set(),
      expanded: false,
      job: null,
      loading: false,
      error: null,
    });
  }
  return jobStates.get(jobId)!;
}

export function notify(jobId: string) {
  const s = jobStates.get(jobId);
  if (s) for (const fn of s.listeners) fn();
}

export function update(
  jobId: string,
  patch: Partial<Omit<ReturnType<typeof getJobState>, "listeners">>,
) {
  const s = getJobState(jobId);
  Object.assign(s, patch);
  notify(jobId);
}

export type JobSnapshot = {
  expanded: boolean;
  job: JobDetailResponse | null;
  loading: boolean;
  error: string | null;
};

export function useJobDetail(jobId: string) {
  const state = getJobState(jobId);
  const cachedSnapshot = useRef<JobSnapshot | null>(null);

  const subscribe = useCallback(
    (cb: () => void) => {
      state.listeners.add(cb);
      return () => {
        state.listeners.delete(cb);
      };
    },
    [state],
  );

  const getSnapshot = useCallback((): JobSnapshot => {
    const s = getJobState(jobId);
    const prev = cachedSnapshot.current;
    if (
      prev &&
      prev.expanded === s.expanded &&
      prev.job === s.job &&
      prev.loading === s.loading &&
      prev.error === s.error
    ) {
      return prev;
    }
    const next: JobSnapshot = {
      expanded: s.expanded,
      job: s.job,
      loading: s.loading,
      error: s.error,
    };
    cachedSnapshot.current = next;
    return next;
  }, [jobId]);

  const snapshot = useSyncExternalStore(subscribe, getSnapshot);

  const toggle = useCallback(async () => {
    if (snapshot.expanded) {
      update(jobId, { expanded: false });
      return;
    }

    if (!snapshot.job && !snapshot.loading) {
      update(jobId, { loading: true, error: null });
      try {
        const result = await getJobByID(jobId);
        if (result.status === 200) {
          update(jobId, { job: result.data as JobDetailResponse });
        } else {
          update(jobId, { error: "Failed to fetch job" });
        }
      } catch (e) {
        update(jobId, {
          error: e instanceof Error ? e.message : "Failed to fetch job",
        });
      }
      update(jobId, { loading: false });
    }

    update(jobId, { expanded: true });
  }, [jobId, snapshot.expanded, snapshot.job, snapshot.loading]);

  return { ...snapshot, toggle };
}

// Shared state between inline pill + panel for the same job ID
export const jobStates = new Map<
  string,
  {
    listeners: Set<() => void>;
    expanded: boolean;
    job: JobDetailResponse | null;
    loading: boolean;
    error: string | null;
  }
>();
