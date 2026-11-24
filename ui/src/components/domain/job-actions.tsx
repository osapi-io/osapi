import { RefreshCw, Trash2 } from "lucide-react";
import { IconButton } from "@/components/ui/icon-button";
import { GetJobsStatus, type ListJobsResponse } from "@/sdk/gen/schemas";

type Job = NonNullable<ListJobsResponse["items"]>[number];

interface JobActionsProps {
  job: Job;
  busy: boolean;
  onRetry: (id: string) => void;
  onDelete: (id: string) => void;
}

/** Retry and delete on the row of a job that failed. */
export function JobActions({ job, busy, onRetry, onDelete }: JobActionsProps) {
  const failed =
    job.status === GetJobsStatus.failed ||
    job.status === GetJobsStatus.partial_failure;
  if (!failed || !job.id) return null;
  const id = job.id;
  return (
    <div className="flex items-center justify-end gap-1">
      <IconButton
        icon={RefreshCw}
        onClick={(e) => {
          e.stopPropagation();
          onRetry(id);
        }}
        disabled={busy}
        title="Retry job"
      />
      <IconButton
        icon={Trash2}
        variant="danger"
        onClick={(e) => {
          e.stopPropagation();
          onDelete(id);
        }}
        disabled={busy}
        title="Delete job"
      />
    </div>
  );
}
