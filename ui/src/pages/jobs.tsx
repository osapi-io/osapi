import { useState, useEffect, useCallback } from "react";
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
  ChevronLeft,
  ChevronRight,
  Trash2,
  RefreshCw,
  ChevronDown,
  ChevronUp,
} from "lucide-react";
import {
  getJobs,
  getJobByID,
  deleteJobByID,
  retryJobByID,
} from "@/sdk/gen/job-management-api-job-operations/job-management-api-job-operations";
import type { JobDetailResponse, ListJobsResponse } from "@/sdk/gen/schemas";
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
  { value: "", label: "All Statuses" },
  { value: "submitted", label: "Submitted" },
  { value: "processing", label: "Processing" },
  { value: "completed", label: "Completed" },
  { value: "failed", label: "Failed" },
  { value: "partial_failure", label: "Partial Failure" },
  { value: "skipped", label: "Skipped" },
];

export function Jobs() {
  useVimScroll();
  const { can } = useAuth();

  const [data, setData] = useState<ListJobsResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [offset, setOffset] = useState(0);
  const [statusFilter, setStatusFilter] = useState("");
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [expandedJob, setExpandedJob] = useState<JobDetailResponse | null>(
    null,
  );
  const [expandLoading, setExpandLoading] = useState(false);
  const [confirmDelete, setConfirmDelete] = useState<string | null>(null);
  const [actionLoading, setActionLoading] = useState(false);

  const fetchJobs = useCallback(async () => {
    try {
      const params: Record<string, unknown> = {
        limit: PAGE_SIZE,
        offset,
      };
      if (statusFilter) {
        params.status = statusFilter;
      }
      const result = await getJobs(params as Parameters<typeof getJobs>[0]);
      if (result.status === 200) {
        setData(result.data as ListJobsResponse);
        setError(null);
      } else {
        setError("Failed to fetch jobs");
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to fetch jobs");
    } finally {
      setLoading(false);
    }
  }, [offset, statusFilter]);

  useEffect(() => {
    let mounted = true;

    const poll = async () => {
      if (!mounted) return;
      await fetchJobs();
    };

    poll();
    const id = setInterval(poll, 10000);
    return () => {
      mounted = false;
      clearInterval(id);
    };
  }, [fetchJobs]);

  const handleExpand = async (jobId: string) => {
    if (expandedId === jobId) {
      setExpandedId(null);
      setExpandedJob(null);
      return;
    }
    setExpandedId(jobId);
    setExpandLoading(true);
    try {
      const result = await getJobByID(jobId);
      if (result.status === 200) {
        setExpandedJob(result.data as JobDetailResponse);
      }
    } catch {
      // Silently fail — row just won't expand with detail
    } finally {
      setExpandLoading(false);
    }
  };

  const handleDelete = async (jobId: string) => {
    setActionLoading(true);
    try {
      const result = await deleteJobByID(jobId);
      if (result.status === 204) {
        setConfirmDelete(null);
        await fetchJobs();
      } else {
        setError("Failed to delete job");
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to delete job");
    } finally {
      setActionLoading(false);
    }
  };

  const handleRetry = async (jobId: string) => {
    setActionLoading(true);
    try {
      const result = await retryJobByID(jobId, {});
      if (result.status === 201) {
        await fetchJobs();
      } else {
        setError("Failed to retry job");
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to retry job");
    } finally {
      setActionLoading(false);
    }
  };

  if (!can("job:read")) {
    return <PermissionDenied title="Jobs" what="jobs" />;
  }

  const items = data?.items ?? [];
  const totalItems = data?.total_items ?? 0;
  const currentPage = Math.floor(offset / PAGE_SIZE) + 1;
  const totalPages = Math.max(1, Math.ceil(totalItems / PAGE_SIZE));

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
              getRowKey={(job, i) => job.id ?? i}
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
                ...(can("job:write")
                  ? [
                      {
                        header: "Actions",
                        align: "right" as const,
                        cell: (job: Job) =>
                          (job.status === "failed" ||
                            job.status === "partial_failure") && (
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
          {totalPages > 1 && (
            <div className="flex items-center justify-between">
              <Text variant="muted">
                Page {currentPage} of {totalPages}
              </Text>
              <div className="flex items-center gap-2">
                <Button
                  variant="ghost"
                  size="sm"
                  disabled={offset === 0}
                  onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}
                >
                  <ChevronLeft className="h-4 w-4" />
                  Previous
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  disabled={offset + PAGE_SIZE >= totalItems}
                  onClick={() => setOffset(offset + PAGE_SIZE)}
                >
                  Next
                  <ChevronRight className="h-4 w-4" />
                </Button>
              </div>
            </div>
          )}
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
