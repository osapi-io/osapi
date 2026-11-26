import { RotateCcw } from "lucide-react";
import { DataTable } from "@/components/ui/data-table";
import { Text } from "@/components/ui/text";
import { formatTime, operationType } from "@/lib/format";
import type { DeadLetter } from "@/sdk/gen/schemas";

/**
 * DeadLetters is the jobs whose delivery to an agent used up its retries,
 * the newest first, as osapi client job deadletter list prints them. A
 * dead letter whose job is still there can go back to the queue: onRetry
 * makes a new job with the same operation. Without onRetry, the table has
 * no actions.
 */
export function DeadLetters({
  items,
  total,
  busy,
  onRetry,
}: {
  items: DeadLetter[];
  total: number;
  busy: boolean;
  onRetry?: (jobId: string) => void;
}) {
  if (items.length === 0) {
    return (
      <Text variant="muted">No dead letters: each job reached an agent.</Text>
    );
  }

  return (
    <>
      <Text variant="muted">
        {items.length} of {total} {total === 1 ? "dead letter" : "dead letters"}
        , the newest first
      </Text>
      <DataTable
        rows={items}
        getRowKey={(dl) => String(dl.seq)}
        columns={[
          {
            header: "Time",
            cell: (dl: DeadLetter) => (
              <Text variant="muted">{formatTime(dl.time)}</Text>
            ),
          },
          {
            header: "Job ID",
            cell: (dl: DeadLetter) => (
              <Text variant="mono-muted">{dl.job_id || "gone"}</Text>
            ),
          },
          {
            header: "Operation",
            cell: (dl: DeadLetter) => (
              <Text variant="mono-muted">{operationType(dl.operation)}</Text>
            ),
          },
          {
            header: "Target",
            cell: (dl: DeadLetter) => (
              <Text variant="mono-muted">{dl.target || "-"}</Text>
            ),
          },
          {
            header: "Deliveries",
            cell: (dl: DeadLetter) => <Text>{dl.deliveries}</Text>,
          },
          {
            header: "Error",
            cell: (dl: DeadLetter) => (
              <Text variant="muted">{dl.error || "-"}</Text>
            ),
          },
          ...(onRetry
            ? [
                {
                  header: "Actions",
                  align: "right" as const,
                  cell: (dl: DeadLetter) =>
                    dl.job_id ? (
                      <button
                        type="button"
                        title="Retry job"
                        disabled={busy}
                        onClick={() => dl.job_id && onRetry(dl.job_id)}
                        className="rounded p-1 text-text-muted hover:text-primary disabled:opacity-50"
                      >
                        <RotateCcw className="h-3.5 w-3.5" />
                      </button>
                    ) : null,
                },
              ]
            : []),
        ]}
      />
    </>
  );
}
