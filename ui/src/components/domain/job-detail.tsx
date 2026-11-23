import { useCallback, useRef, useSyncExternalStore } from "react";
import { getJobByID } from "@/sdk/gen/job-management-api-job-operations/job-management-api-job-operations";
import type { JobDetailResponse } from "@/sdk/gen/schemas";
import { Badge } from "@/components/ui/badge";
import { IdBadge } from "@/components/ui/id-badge";
import { KVList } from "@/components/ui/kv-list";
import { DataTable } from "@/components/ui/data-table";
import { SectionLabel } from "@/components/ui/section-label";
import { Text } from "@/components/ui/text";
import { formatDuration, formatTime, summarize } from "@/lib/format";
import { hostSteps } from "@/lib/steps";
import { ChainSection } from "@/components/domain/step-tree";
import { ChevronDown, ChevronRight, Loader2 } from "lucide-react";
import { statusVariant } from "@/lib/status";

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
        job {jobId}
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
      <div className="border-b border-border/40 bg-surface px-4 py-3">
        <JobView job={job} />
      </div>
    );
  }

  return null;
}

/** The fields of an object as key-value pairs, as the CLI's printObject prints them. */
function ObjectKV({ v }: { v: unknown }) {
  const pairs: [string, string][] =
    v && typeof v === "object" && !Array.isArray(v)
      ? Object.keys(v)
          .sort()
          .map((k) => [k, summarize((v as Record<string, unknown>)[k])])
      : [["Value", summarize(v)]];
  return <KVList pairs={pairs} />;
}

/**
 * JobView is one job, as osapi client job get prints it: the job, its
 * request, the agents' responses and states, the chain of calls on each
 * host, the timeline and the result.
 * The Jobs page and the job panel of a step's result on the Plans page both show it.
 */
export function JobView({ job }: { job: JobDetailResponse }) {
  // The API sorts both by hostname.
  const states = job.agent_states ?? [];
  const responses = job.responses ?? [];
  const count = (status: string) =>
    states.filter((s) => s.status === status).length;
  const skipped = count("skipped");

  return (
    <div className="space-y-3">
      <div>
        <SectionLabel>Job</SectionLabel>
        <KVList
          pairs={[
            ["Job ID", <span className="font-mono">{job.id}</span>],
            [
              "Status",
              <Badge variant={statusVariant(job.status)}>{job.status}</Badge>,
            ],
            ["Hostname", job.hostname],
            ["Target", job.target],
            ["Created", job.created && formatTime(job.created)],
            ["Updated", job.updated_at && formatTime(job.updated_at)],
            [
              "Error",
              job.error && (
                <span className="text-status-error">{job.error}</span>
              ),
            ],
            [
              "Agents",
              states.length > 1 &&
                `${states.length} total (${count("completed")} completed, ${count("failed")} failed, ${count("started")} processing${skipped > 0 ? `, ${skipped} skipped` : ""})`,
            ],
          ]}
        />
      </div>

      {job.operation && (
        <div>
          <SectionLabel>Request</SectionLabel>
          <ObjectKV v={job.operation} />
        </div>
      )}

      {responses.length > 0 && (
        <div>
          <SectionLabel>Agent Responses</SectionLabel>
          <DataTable
            rows={responses}
            getRowKey={(r) => r.hostname}
            columns={[
              { header: "Hostname", cell: (r) => r.hostname },
              {
                header: "Status",
                cell: (r) => (
                  <Badge variant={statusVariant(r.status)}>{r.status}</Badge>
                ),
              },
              { header: "Data", cell: (r) => summarize(r.data) },
              {
                header: "Error",
                cell: (r) => (
                  <span className="text-status-error">{r.error}</span>
                ),
              },
            ]}
          />
        </div>
      )}

      <ChainSection hosts={hostSteps(job)} lines={job.chain} />

      {states.length > 0 && (
        <div>
          <SectionLabel>Agent States</SectionLabel>
          <DataTable
            rows={states}
            getRowKey={(st) => st.hostname}
            columns={[
              { header: "Hostname", cell: (st) => st.hostname },
              {
                header: "Status",
                cell: (st) => (
                  <Badge variant={statusVariant(st.status)}>{st.status}</Badge>
                ),
              },
              {
                header: "Duration",
                cell: (st) => (
                  <span className="text-text-muted">
                    {formatDuration(st.duration)}
                  </span>
                ),
              },
              {
                header: "Error",
                cell: (st) => (
                  <span className="text-status-error">{st.error}</span>
                ),
              },
            ]}
          />
        </div>
      )}

      <div>
        <SectionLabel>Timeline</SectionLabel>
        <DataTable
          rows={job.timeline ?? []}
          getRowKey={(_, i) => i}
          columns={[
            {
              header: "Timestamp",
              cell: (e) => (
                <span className="text-text-muted">
                  {formatTime(e.timestamp)}
                </span>
              ),
            },
            {
              header: "Event",
              cell: (e) => (
                <Badge variant={statusVariant(e.event)}>{e.event}</Badge>
              ),
            },
            { header: "Hostname", cell: (e) => e.hostname },
            { header: "Message", cell: (e) => e.message },
            {
              header: "Error",
              cell: (e) => <span className="text-status-error">{e.error}</span>,
            },
          ]}
        />
      </div>

      {job.result !== undefined && job.result !== null && (
        <div>
          <SectionLabel>Result</SectionLabel>
          <ObjectKV v={job.result} />
        </div>
      )}
    </div>
  );
}
