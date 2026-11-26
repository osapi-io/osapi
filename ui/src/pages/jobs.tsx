import { useState } from "react";
import { JobActions } from "@/components/domain/job-actions";
import { ConfirmModal } from "@/components/ui/confirm-modal";
import { Pagination } from "@/components/ui/pagination";
import { unwrap, usePolling } from "@/hooks/use-polling";
import { useJob } from "@/hooks/use-job";
import { enumLabel } from "@/lib/format";
import { JobStatus } from "@/sdk/gen/schemas";
import { useParams, useSearchParams } from "react-router-dom";
import { useVimScroll } from "@/hooks/use-vim-scroll";
import { useAuth } from "@/hooks/use-auth";
import { ContentArea } from "@/components/layout/content-area";
import { PageHeader } from "@/components/ui/page-header";
import { SectionLabel } from "@/components/ui/section-label";
import { ErrorBanner } from "@/components/ui/error-banner";
import { Badge } from "@/components/ui/badge";
import { Text } from "@/components/ui/text";
import { Dropdown } from "@/components/ui/dropdown";
import { Card, CardContent } from "@/components/ui/card";
import { Briefcase, ChevronDown, ChevronUp } from "lucide-react";
import {
  getJobs,
  deleteJobByID,
  retryJobByID,
} from "@/sdk/gen/job-management-api-job-operations/job-management-api-job-operations";
import type { ListJobsResponse } from "@/sdk/gen/schemas";
import { useStatus } from "@/hooks/use-status";
import { PermissionDenied } from "@/components/layout/permission-denied";
import { Spinner } from "@/components/ui/spinner";
import { DataTable } from "@/components/ui/data-table";
import { Count } from "@/components/ui/count";
import { JobView } from "@/components/domain/job-detail";
import { formatTime, operationType } from "@/lib/format";
import { LinkedJob } from "@/components/domain/linked-job";
import { DeadLetters } from "@/components/domain/dead-letters";
import { listDeadLetters } from "@/sdk/gen/job-management-api-job-operations/job-management-api-job-operations";
import type { DeadLetterList } from "@/sdk/gen/schemas";

/** A job as the list gives it. */
type Job = NonNullable<ListJobsResponse["items"]>[number];

const PAGE_SIZE = 20;

const STATUS_OPTIONS = [
  { value: "", label: "All statuses" },
  ...Object.values(JobStatus).map((v) => ({
    value: v,
    label: enumLabel(v),
  })),
];

export function Jobs() {
  const { id: linkedId } = useParams();
  useVimScroll();
  const { canCall } = useAuth();
  const statusColor = useStatus();

  const [offset, setOffset] = useState(0);
  // ?status=failed, from the dashboard's Failed jobs card, is the first
  // filter.
  const [searchParams] = useSearchParams();
  const [statusFilter, setStatusFilter] = useState(
    searchParams.get("status") ?? "",
  );
  // ?view=deadletters, from the dashboard's Dead letters card, shows the
  // jobs whose delivery used up its retries instead of the job list.
  const [view, setView] = useState(searchParams.get("view") ?? "");
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [confirmDelete, setConfirmDelete] = useState<string | null>(null);
  const [actionLoading, setActionLoading] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);

  const jobs = usePolling(
    async () =>
      unwrap<ListJobsResponse>(
        await getJobs({
          limit: PAGE_SIZE,
          offset,
          ...(statusFilter ? { status: statusFilter as JobStatus } : {}),
        }),
        "Failed to fetch jobs",
      ),
    [offset, statusFilter],
    { enabled: canCall("GetJobs") },
  );
  const deadLetters = usePolling(
    async () =>
      unwrap<DeadLetterList>(
        await listDeadLetters({ limit: PAGE_SIZE }),
        "Failed to fetch the dead letters",
      ),
    [],
    { enabled: view === "deadletters" && canCall("ListDeadLetters") },
  );
  const { data, loading } = jobs;
  const error = actionError ?? jobs.error;
  const fetchJobs = jobs.refresh;
  const expanded = useJob(expandedId);
  const expandedJob = expanded.data;
  const expandLoading = expanded.loading;

  const handleExpand = (jobId: string) => {
    setExpandedId(expandedId === jobId ? null : jobId);
  };

  const handleDelete = async (jobId: string) => {
    setActionLoading(true);
    try {
      const result = await deleteJobByID(jobId);
      if (result.status === 204) {
        setConfirmDelete(null);
        setActionError(null);
        await fetchJobs();
      } else {
        setActionError("Failed to delete job");
      }
    } catch (e) {
      setActionError(e instanceof Error ? e.message : "Failed to delete job");
    } finally {
      setActionLoading(false);
    }
  };

  const handleRetry = async (jobId: string) => {
    setActionLoading(true);
    try {
      const result = await retryJobByID(jobId, {});
      if (result.status === 201) {
        setActionError(null);
        await fetchJobs();
        await deadLetters.refresh();
      } else {
        setActionError("Failed to retry job");
      }
    } catch (e) {
      setActionError(e instanceof Error ? e.message : "Failed to retry job");
    } finally {
      setActionLoading(false);
    }
  };

  if (!canCall("GetJobs")) {
    return <PermissionDenied title="Jobs" what="jobs" />;
  }

  const items = data?.items ?? [];
  const totalItems = data?.total_items ?? 0;

  return (
    <ContentArea>
      <PageHeader
        title="Jobs"
        subtitle="Async job queue and history"
        actions={
          data && (
            // As osapi client job list prints them: N jobs, then each
            // status's count in its tone.
            <Text variant="muted">
              {totalItems} {totalItems === 1 ? "job" : "jobs"}:{" "}
              {Object.entries(data.status_counts ?? {}).map(([st, n], i) => (
                <span key={st}>
                  {i > 0 && ", "}
                  <Count n={n} kind={st} />
                </span>
              ))}
            </Text>
          )
        }
      />

      {linkedId && <LinkedJob id={linkedId} />}

      {loading && <Spinner />}

      {error && <ErrorBanner message={error} />}

      {canCall("ListDeadLetters") && (
        <div className="mb-4 flex gap-2">
          <button
            type="button"
            onClick={() => setView("")}
            className={
              view === "" ? "font-semibold text-primary" : "text-text-muted"
            }
          >
            Jobs
          </button>
          <button
            type="button"
            onClick={() => setView("deadletters")}
            className={
              view === "deadletters"
                ? "font-semibold text-primary"
                : "text-text-muted"
            }
          >
            Dead letters
          </button>
        </div>
      )}

      {view === "deadletters" && canCall("ListDeadLetters") && (
        <div className="mb-4">
          <SectionLabel>Dead letters</SectionLabel>
          {deadLetters.error && <ErrorBanner message={deadLetters.error} />}
          {deadLetters.loading && <Spinner />}
          {deadLetters.data && (
            <DeadLetters
              items={deadLetters.data.items}
              total={deadLetters.data.total_items}
              busy={actionLoading}
              onRetry={canCall("RetryJobByID") ? handleRetry : undefined}
            />
          )}
        </div>
      )}

      {view !== "deadletters" && !loading && !error && (
        <>
          {/* Filters */}
          <div className="mb-4">
            <SectionLabel icon={Briefcase}>Filters</SectionLabel>
            <Card>
              <CardContent>
                <div className="flex items-end gap-4">
                  <Dropdown
                    label="Status"
                    value={statusFilter}
                    onChange={(v) => {
                      setStatusFilter(v);
                      setOffset(0);
                    }}
                    options={STATUS_OPTIONS}
                    className="w-48"
                  />
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Table */}
          <div className="mb-4">
            <SectionLabel>
              Jobs{statusFilter && ` ${statusFilter}`} ({items.length} of{" "}
              {totalItems})
            </SectionLabel>
            <DataTable
              rows={items}
              getRowKey={(job) => job.id ?? ""}
              onRowClick={(job) => job.id && handleExpand(job.id)}
              expanded={(job) =>
                expandedId !== job.id ? undefined : expandLoading ? (
                  <div className="flex items-center justify-center py-4">
                    <Spinner size="inline" />
                  </div>
                ) : expandedJob ? (
                  <JobView job={expandedJob} />
                ) : (
                  <Text variant="muted">Failed to load details</Text>
                )
              }
              columns={[
                {
                  header: "",
                  cell: (job: Job) =>
                    expandedId === job.id ? (
                      <ChevronUp className="h-3.5 w-3.5 text-text-muted" />
                    ) : (
                      <ChevronDown className="h-3.5 w-3.5 text-text-muted" />
                    ),
                },
                {
                  header: "Job ID",
                  cell: (job: Job) => (
                    <Text variant="mono-muted">{job.id}</Text>
                  ),
                },
                {
                  header: "Status",
                  cell: (job: Job) => (
                    <Badge variant={statusColor.variant(job.status)}>
                      {job.status}
                    </Badge>
                  ),
                },
                {
                  header: "Created",
                  cell: (job: Job) => (
                    <Text variant="muted">{formatTime(job.created)}</Text>
                  ),
                },
                {
                  // The target the request gave; a job stored before targets
                  // were kept has only the host that answered.
                  header: "Target",
                  cell: (job: Job) => (
                    <Text variant="mono-muted">
                      {job.target ?? job.hostname ?? "-"}
                    </Text>
                  ),
                },
                {
                  header: "Operation",
                  cell: (job: Job) => (
                    <Text variant="mono-muted">
                      {operationType(job.operation)}
                    </Text>
                  ),
                },
                ...(canCall("RetryJobByID")
                  ? [
                      {
                        header: "Actions",
                        align: "right" as const,
                        cell: (job: Job) => (
                          <JobActions
                            job={job}
                            busy={actionLoading}
                            onRetry={handleRetry}
                            onDelete={setConfirmDelete}
                          />
                        ),
                      },
                    ]
                  : []),
              ]}
            />
          </div>

          {/* Pagination */}
          <Pagination
            offset={offset}
            limit={PAGE_SIZE}
            total={totalItems}
            onChange={setOffset}
          />
        </>
      )}

      <ConfirmModal
        open={confirmDelete !== null}
        title="Delete job"
        action="Delete"
        busy={actionLoading}
        onConfirm={() => confirmDelete && handleDelete(confirmDelete)}
        onClose={() => setConfirmDelete(null)}
      >
        Delete job <Text variant="mono">{confirmDelete}</Text>? This cannot be
        undone.
      </ConfirmModal>
    </ContentArea>
  );
}
