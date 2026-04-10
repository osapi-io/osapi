import { useCallback, useRef, useSyncExternalStore } from "react";
import { getJobByID } from "@/sdk/gen/job-management-api-job-operations/job-management-api-job-operations";
import type {
  JobDetailResponse,
  JobDetailResponseTimelineItem,
} from "@/sdk/gen/schemas";
import { Badge } from "@/components/ui/badge";
import { CodeBlock } from "@/components/ui/code-block";
import { IdBadge } from "@/components/ui/id-badge";
import { KeyValue } from "@/components/ui/key-value";
import { Text } from "@/components/ui/text";
import { ChevronDown, ChevronRight, Loader2 } from "lucide-react";

// Shared state between inline pill + panel for the same job ID
const jobStates = new Map<
  string,
  {
    listeners: Set<() => void>;
    expanded: boolean;
    job: JobDetailResponse | null;
    loading: boolean;
    error: string | null;
  }
>();

function getJobState(jobId: string) {
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

function notify(jobId: string) {
  const s = jobStates.get(jobId);
  if (s) for (const fn of s.listeners) fn();
}

function update(
  jobId: string,
  patch: Partial<Omit<ReturnType<typeof getJobState>, "listeners">>,
) {
  const s = getJobState(jobId);
  Object.assign(s, patch);
  notify(jobId);
}

type JobSnapshot = {
  expanded: boolean;
  job: JobDetailResponse | null;
  loading: boolean;
  error: string | null;
};

function useJobDetail(jobId: string) {
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

function statusVariant(status?: string) {
  switch (status) {
    case "completed":
      return "ready" as const;
    case "processing":
      return "running" as const;
    case "failed":
    case "partial_failure":
      return "error" as const;
    default:
      return "muted" as const;
  }
}

function TimelineRow({ event }: { event: JobDetailResponseTimelineItem }) {
  return (
    <div className="flex items-start gap-3 py-1.5">
      <div className="mt-1 h-1.5 w-1.5 shrink-0 rounded-full bg-accent/60" />
      <div className="flex-1">
        <div className="flex items-center gap-2">
          <Text className="font-medium">{event.event}</Text>
          {event.hostname && <Text variant="muted">{event.hostname}</Text>}
          {event.timestamp && (
            <span className="ml-auto text-xs text-text-muted/50">
              {new Date(event.timestamp).toLocaleTimeString()}
            </span>
          )}
        </div>
        {event.message && (
          <Text variant="muted" as="p">
            {event.message}
          </Text>
        )}
        {event.error && (
          <Text variant="error" as="p">
            {event.error}
          </Text>
        )}
      </div>
    </div>
  );
}

interface JobDetailProps {
  jobId: string;
  /** Render as the clickable pill in the header */
  inline?: boolean;
  /** Render as the full-width expandable panel */
  panel?: boolean;
}

export function JobDetail({ jobId, inline, panel }: JobDetailProps) {
  const { expanded, toggle, job, loading, error } = useJobDetail(jobId);

  // Inline mode: clickable pill
  if (inline) {
    return (
      <IdBadge interactive onClick={toggle}>
        {loading ? (
          <Loader2 className="h-3 w-3 animate-spin" />
        ) : expanded ? (
          <ChevronDown className="h-3 w-3" />
        ) : (
          <ChevronRight className="h-3 w-3" />
        )}
        {jobId.slice(0, 8)}
      </IdBadge>
    );
  }

  // Panel mode: full-width expandable detail
  if (panel) {
    if (!expanded) return null;

    if (error) {
      return (
        <div className="border-b border-border/40 px-4 py-2">
          <Text variant="error" as="p">
            {error}
          </Text>
        </div>
      );
    }

    if (!job) return null;

    return (
      <div className="border-b border-border/40 bg-[#050505] px-4 py-3">
        {/* Header */}
        <div className="mb-2 flex items-center gap-2">
          <Text variant="mono-muted">{job.id}</Text>
          <Badge variant={statusVariant(job.status)}>{job.status}</Badge>
          {job.changed != null && (
            <Badge variant={job.changed ? "applied" : "muted"}>
              {job.changed ? "changed" : "unchanged"}
            </Badge>
          )}
        </div>

        {/* Details grid */}
        <div className="flex flex-col gap-1 text-xs">
          {job.hostname && <KeyValue label="Host" value={job.hostname} />}
          {job.created && (
            <KeyValue
              label="Created"
              value={new Date(job.created).toLocaleString()}
            />
          )}
          {job.updated_at && job.updated_at.length > 0 && (
            <KeyValue
              label="Updated"
              value={new Date(job.updated_at).toLocaleString()}
            />
          )}
        </div>

        {/* Error */}
        {job.error && (
          <div className="mt-2 rounded bg-status-error/5 px-3 py-2">
            <p className="text-xs font-semibold uppercase tracking-wider text-status-error">
              Error
            </p>
            <Text variant="error" as="p" className="mt-0.5">
              {job.error}
            </Text>
          </div>
        )}

        {/* Timeline */}
        {job.timeline && job.timeline.length > 0 && (
          <div className="mt-3 border-t border-border/30 pt-2">
            <p className="mb-1 text-xs font-semibold uppercase tracking-wider text-text-muted">
              Timeline
            </p>
            {job.timeline.map((event, i) => (
              <TimelineRow key={i} event={event} />
            ))}
          </div>
        )}

        {/* Operation */}
        {job.operation && Object.keys(job.operation).length > 0 && (
          <div className="mt-3 border-t border-border/30 pt-2">
            <p className="mb-1 text-xs font-semibold uppercase tracking-wider text-text-muted">
              Operation
            </p>
            <CodeBlock variant="muted">
              {JSON.stringify(job.operation, null, 2)}
            </CodeBlock>
          </div>
        )}
      </div>
    );
  }

  return null;
}
