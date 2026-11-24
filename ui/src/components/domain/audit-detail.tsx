import { Text } from "@/components/ui/text";
import { KVList } from "@/components/ui/kv-list";
import { CodeBlock } from "@/components/ui/code-block";
import { RunLinks } from "@/components/domain/run-links";
import { ChainSection } from "@/components/domain/step-tree";
import type { AuditEntry } from "@/sdk/gen/schemas";
import { formatTime } from "@/lib/format";

// AuditDetail is one entry in full: who, what, the result, the chain of
// calls that its job made on each host, and the links to its job and its
// trace.
export function AuditDetail({ entry }: { entry: AuditEntry }) {
  return (
    <div className="space-y-3 px-2 py-2">
      <KVList
        pairs={[
          ["User", entry.user],
          ["Roles", entry.roles.join(", ") || "-"],
          ["Targets", entry.targets?.join(", ") || "each host"],
          ["Source IP", entry.source_ip],
          ["Operation", entry.operation_id || "-"],
          ["Method", entry.method],
          ["Path", entry.path],
          ["Status", String(entry.response_code)],
          ["Duration", `${entry.duration_ms}ms`],
          ["Time", formatTime(entry.timestamp)],
          ["ID", entry.id],
        ]}
      />
      {entry.request_summary && (
        <div>
          <Text variant="label" as="p" className="mb-1">
            Request, with each secret replaced
          </Text>
          <CodeBlock variant="json">{entry.request_summary}</CodeBlock>
        </div>
      )}
      <ChainSection hosts={entry.steps ?? []} lines={entry.chain} />
      <RunLinks jobId={entry.job_id} traceId={entry.trace_id} />
    </div>
  );
}
