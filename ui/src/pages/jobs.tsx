import { useState } from "react";
import { Pagination } from "@/components/ui/pagination";
import { unwrap, usePolling } from "@/hooks/use-polling";
import { useJob } from "@/hooks/use-job";
import { enumLabel } from "@/lib/format";
import { GetJobsStatus } from "@/sdk/gen/schemas";
import { Link, useParams } from "react-router-dom";
import { useVimScroll } from "@/hooks/use-vim-scroll";
import { useAuth } from "@/hooks/use-auth";
import { ContentArea } from "@/components/layout/content-area";
import { PageHeader } from "@/components/ui/page-header";
import { SectionLabel } from "@/components/ui/section-label";
import { ErrorBanner } from "@/components/ui/error-banner";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Text } from "@/components/ui/text";
import { Dropdown } from "@/components/ui/dropdown";
import { Card, CardContent } from "@/components/ui/card";
import { Modal } from "@/components/ui/modal";
import {
  Loader2,
  Briefcase,
  Trash2,
  RefreshCw,
  ChevronDown,
  ChevronUp,
} from "lucide-react";
import {
  getJobs,
  deleteJobByID,
  retryJobByID,
} from "@/sdk/gen/job-management-api-job-operations/job-management-api-job-operations";
import type { ListJobsResponse } from "@/sdk/gen/schemas";
import { statusVariant } from "@/lib/status";
import { PermissionDenied } from "@/components/layout/permission-denied";
import { Spinner } from "@/components/ui/spinner";
import { DataTable } from "@/components/ui/data-table";
import { Count } from "@/components/ui/count";
import { JobView } from "@/components/domain/job-detail";
import { formatTime, operationType } from "@/lib/format";
import { IconButton } from "@/components/ui/icon-button";

/** A job as the list gives it. */
type Job = NonNullable<ListJobsResponse["items"]>[number];

const PAGE_SIZE = 20;

const STATUS_OPTIONS = [
  { value: "", label: "All statuses" },
  ...Object.values(GetJobsStatus).map((v) => ({
    value: v,
    label: enumLabel(v),
  })),
];

// LinkedJob is the job that /admin/jobs/:id names, such as the job of a
// step's result on the Plans page: its detail, above the list, whether
// or not the job is on the page of the list that shows.
function LinkedJob({ id }: { id: string }) {
  const { data: job, error: failed } = useJob(id);

  return (
    <div className="mb-4">
      <SectionLabel>
        Job <span className="font-mono">{id}</span>
        <Link
          to="/admin/jobs"
          className="ml-3 text-xs text-text-muted transition-colors hover:text-text"
        >
          close
        </Link>
      </SectionLabel>
      <Card>
        <CardContent>
          {failed ? (
            <Text variant="muted">{failed}</Text>
          ) : job ? (
            <JobView job={job} />
          ) : (
            <Spinner size="inline" />
          )}
        </CardContent>
      </Card>
    </div>
  );
}

export function Jobs() {
  const { id: linkedId } = useParams();
  useVimScroll();
  const { canCall } = useAuth();

  const [offset, setOffset] = useState(0);
  const [statusFilter, setStatusFilter] = useState("");
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
          ...(statusFilter ? { status: statusFilter as GetJobsStatus } : {}),
        }),
        "Failed to fetch jobs",
      ),
    [offset, statusFilter],
    { enabled: canCall("getJobs") },
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
      } else {
        setActionError("Failed to retry job");
      }
    } catch (e) {
      setActionError(e instanceof Error ? e.message : "Failed to retry job");
    } finally {
      setActionLoading(false);
    }
  };

  if (!canCall("getJobs")) {
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

      {!loading && !error && (
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
                    <Badge variant={statusVariant(job.status)}>
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
                ...(canCall("retryJobByID")
                  ? [
                      {
                        header: "Actions",
                        align: "right" as const,
                        cell: (job: Job) =>
                          (job.status === GetJobsStatus.failed ||
                            job.status === GetJobsStatus.partial_failure) && (
                            <div className="flex items-center justify-end gap-1">
                              <IconButton
                                icon={RefreshCw}
                                onClick={(e) => {
                                  e.stopPropagation();
                                  if (job.id) handleRetry(job.id);
                                }}
                                disabled={actionLoading}
                                title="Retry job"
                              />
                              <IconButton
                                icon={Trash2}
                                variant="danger"
                                onClick={(e) => {
                                  e.stopPropagation();
                                  setConfirmDelete(job.id ?? null);
                                }}
                                disabled={actionLoading}
                                title="Delete job"
                              />
                            </div>
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

      {/* Delete confirmation modal */}
      <Modal
        open={confirmDelete !== null}
        onClose={() => setConfirmDelete(null)}
      >
        <div className="space-y-4">
          <Text as="h3" className="text-lg font-semibold">
            Delete Job
          </Text>
          <Text variant="muted" as="p">
            Are you sure you want to delete job{" "}
            <Text variant="mono">{confirmDelete}</Text>? This action cannot be
            undone.
          </Text>
          <div className="flex justify-end gap-2">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setConfirmDelete(null)}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              size="sm"
              onClick={() => confirmDelete && handleDelete(confirmDelete)}
              disabled={actionLoading}
            >
              {actionLoading ? (
                <Loader2 className="h-3.5 w-3.5 animate-spin" />
              ) : (
                <Trash2 className="h-3.5 w-3.5" />
              )}
              Delete
            </Button>
          </div>
        </div>
      </Modal>
    </ContentArea>
  );
}
