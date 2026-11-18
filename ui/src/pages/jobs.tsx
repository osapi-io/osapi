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
import { CodeBlock } from "@/components/ui/code-block";
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

function formatTimestamp(ts?: string) {
  if (!ts) return "-";
  try {
    const d = new Date(ts);
    return d.toLocaleString(undefined, {
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
    });
  } catch {
    return ts;
  }
}

function truncateId(id?: string) {
  if (!id) return "-";
  return id.length > 12 ? `${id.slice(0, 12)}...` : id;
}

function getOperationType(job: JobDetailResponse): string {
  if (!job.operation) return "-";
  const keys = Object.keys(job.operation);
  // Common keys: category, operation, hostname
  const op = job.operation.operation as string | undefined;
  const cat = job.operation.category as string | undefined;
  if (cat && op) return `${cat}.${op}`;
  if (op) return op;
  if (keys.length > 0) return keys.join(", ");
  return "-";
}

function JobDetailPanel({ job }: { job: JobDetailResponse }) {
  return (
    <div className="space-y-3 py-2">
      <div className="grid grid-cols-2 gap-x-6 gap-y-2">
        <div>
          <Text variant="label">Job ID</Text>
          <Text variant="mono" as="p">
            {job.id}
          </Text>
        </div>
        <div>
          <Text variant="label">Status</Text>
          <div className="mt-0.5">
            <Badge variant={statusVariant(job.status)}>{job.status}</Badge>
          </div>
        </div>
        <div>
          <Text variant="label">Hostname</Text>
          <Text variant="mono" as="p">
            {job.hostname ?? "-"}
          </Text>
        </div>
        <div>
          <Text variant="label">Created</Text>
          <Text variant="muted" as="p">
            {formatTimestamp(job.created)}
          </Text>
        </div>
        {job.updated_at && (
          <div>
            <Text variant="label">Updated</Text>
            <Text variant="muted" as="p">
              {formatTimestamp(job.updated_at)}
            </Text>
          </div>
        )}
        {job.changed != null && (
          <div>
            <Text variant="label">Changed</Text>
            <Text as="p">{String(job.changed)}</Text>
          </div>
        )}
      </div>

      {job.error && (
        <div>
          <Text variant="label">Error</Text>
          <Text variant="error" as="p" className="mt-0.5">
            {job.error}
          </Text>
        </div>
      )}

      {job.operation && (
        <div>
          <Text variant="label">Operation</Text>
          <CodeBlock variant="json" maxHeight="max-h-32" className="mt-1">
            {JSON.stringify(job.operation, null, 2)}
          </CodeBlock>
        </div>
      )}

      {job.result != null && (
        <div>
          <Text variant="label">Result</Text>
          <CodeBlock variant="json" maxHeight="max-h-32" className="mt-1">
            {JSON.stringify(job.result, null, 2)}
          </CodeBlock>
        </div>
      )}

      {job.timeline && job.timeline.length > 0 && (
        <div>
          <Text variant="label">Timeline</Text>
          <div className="mt-1 space-y-1">
            {job.timeline.map((evt, i) => (
              <div key={i} className="flex items-center gap-2 text-xs">
                <Text variant="mono-muted">
                  {formatTimestamp(evt.timestamp)}
                </Text>
                <Badge variant={statusVariant(evt.event)}>{evt.event}</Badge>
                {evt.hostname && <Text variant="muted">{evt.hostname}</Text>}
                {evt.message && <Text variant="muted">{evt.message}</Text>}
                {evt.error && <Text variant="error">{evt.error}</Text>}
              </div>
            ))}
          </div>
        </div>
      )}

      {job.agent_states && Object.keys(job.agent_states).length > 0 && (
        <div>
          <Text variant="label">Agent States</Text>
          <div className="mt-1 space-y-1">
            {Object.entries(job.agent_states).map(([host, state]) => (
              <div key={host} className="flex items-center gap-2 text-xs">
                <Text variant="mono">{host}</Text>
                <Badge variant={statusVariant(state.status)}>
                  {state.status}
                </Badge>
                {state.duration && (
                  <Text variant="muted">{state.duration}</Text>
                )}
                {state.error && <Text variant="error">{state.error}</Text>}
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

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
            <>
              <Badge variant="muted">{totalItems.toLocaleString()} jobs</Badge>
              {data.status_counts && (
                <>
                  {Object.entries(data.status_counts).map(([status, count]) => (
                    <Badge
                      key={status}
                      variant={count === 0 ? "muted" : statusVariant(status)}
                    >
                      {status}: {count}
                    </Badge>
                  ))}
                </>
              )}
            </>
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
            <SectionLabel>Jobs ({items.length} shown)</SectionLabel>
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
                  <JobDetailPanel job={expandedJob} />
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
                  header: "ID",
                  cell: (job: Job) => (
                    <Text variant="mono" title={job.id}>
                      {truncateId(job.id)}
                    </Text>
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
                  header: "Operation",
                  cell: (job: Job) => (
                    <Text variant="mono-muted">{getOperationType(job)}</Text>
                  ),
                },
                {
                  header: "Hostname",
                  cell: (job: Job) => (
                    <Text variant="mono-muted">{job.hostname ?? "-"}</Text>
                  ),
                },
                {
                  header: "Created",
                  cell: (job: Job) => (
                    <Text variant="muted">{formatTimestamp(job.created)}</Text>
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
            <Text variant="mono">{truncateId(confirmDelete ?? undefined)}</Text>
            ? This action cannot be undone.
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
