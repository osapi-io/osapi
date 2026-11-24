import { Badge } from "@/components/ui/badge";
import type { ComponentEntry } from "@/sdk/gen/schemas";
import { statusVariant } from "@/lib/status";
import { formatBytes } from "@/lib/format";

// RegistryHeader is one row of osapi client health status's component
// table: hostname, status, conditions, age, CPU, memory and version, in that
// order.
export function RegistryHeader({
  entry,
  reference,
}: {
  entry?: ComponentEntry;
  // reference is the controller's version, which the page header shows.
  reference?: string;
}) {
  if (!entry) return null;
  return (
    <div className="border-b border-border/30 px-3 py-2 text-xs">
      <div className="flex items-center gap-3 whitespace-nowrap">
        {entry.hostname && (
          <span
            className="min-w-0 truncate font-medium text-text"
            title={entry.hostname}
          >
            {entry.hostname}
          </span>
        )}
        {entry.status && (
          <Badge variant={statusVariant(entry.status)}>{entry.status}</Badge>
        )}
        {entry.conditions && entry.conditions.length > 0 && (
          <span className="text-status-pending">
            {entry.conditions.join(", ")}
          </span>
        )}
        <div className="ml-auto flex shrink-0 items-center gap-3 text-text-muted">
          {entry.age && <span>up {entry.age}</span>}
          {entry.cpu_percent != null && (
            <span>{entry.cpu_percent.toFixed(1)}%</span>
          )}
          {entry.mem_bytes != null && (
            <span>{formatBytes(entry.mem_bytes)}</span>
          )}
        </div>
      </div>
      {/* The version only when it is not the one in the page header: a
          mismatch is what a reader compares for. */}
      {entry.version && entry.version !== reference && (
        <p
          className="mt-0.5 truncate font-mono text-[10px] text-status-pending"
          title={entry.version}
        >
          {entry.version}
        </p>
      )}
    </div>
  );
}
