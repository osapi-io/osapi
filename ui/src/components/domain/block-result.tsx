import { useState } from "react";
import { ChevronDown, ChevronRight } from "lucide-react";
import { Text } from "@/components/ui/text";
import type { CommandResultItem } from "@/sdk/gen/schemas";

interface BlockResultProps {
  type: string;
  result: unknown;
}

function CommandResultView({
  result,
}: {
  result: { job_id?: string; results: CommandResultItem[] };
}) {
  const [expanded, setExpanded] = useState(true);
  const allOk = result.results.every(
    (r) => !r.error && (r.exit_code === 0 || r.exit_code === undefined),
  );
  const hostCount = result.results.length;

  return (
    <div className="mt-3 rounded-md border border-border bg-[#050505]">
      {/* Collapsible header */}
      <button
        onClick={() => setExpanded(!expanded)}
        className="flex w-full items-center gap-2 px-3 py-2 text-left"
      >
        {expanded ? (
          <ChevronDown className="h-3.5 w-3.5 text-text-muted" />
        ) : (
          <ChevronRight className="h-3.5 w-3.5 text-text-muted" />
        )}
        <Text variant="label">
          {hostCount} host{hostCount !== 1 ? "s" : ""}
        </Text>
        <span
          className={`ml-auto text-xs font-medium ${allOk ? "text-primary" : "text-status-error"}`}
        >
          {allOk ? "ok" : "failed"}
        </span>
        {result.job_id && (
          <span className="text-[10px] text-text-muted/50 font-mono">
            {result.job_id.slice(0, 8)}
          </span>
        )}
      </button>

      {/* Expandable body */}
      {expanded && (
        <div className="space-y-px border-t border-border">
          {result.results.map((r, i) => {
            const ok =
              !r.error && (r.exit_code === 0 || r.exit_code === undefined);
            return (
              <div key={i} className="px-3 py-2">
                <div className="flex items-center justify-between">
                  <Text className="font-medium">{r.hostname}</Text>
                  <div className="flex items-center gap-2">
                    {r.duration_ms !== undefined && (
                      <span className="text-[10px] text-text-muted">
                        {r.duration_ms}ms
                      </span>
                    )}
                    <span
                      className={`text-[10px] font-medium ${ok ? "text-primary" : "text-status-error"}`}
                    >
                      {r.error ? "error" : `exit ${r.exit_code ?? 0}`}
                    </span>
                  </div>
                </div>
                {r.error && (
                  <pre className="mt-1 whitespace-pre-wrap font-mono text-[11px] text-status-error">
                    {r.error}
                  </pre>
                )}
                {r.stdout && (
                  <pre className="mt-1 max-h-32 overflow-auto whitespace-pre-wrap rounded bg-background/50 px-2 py-1.5 font-mono text-[11px] text-text/80">
                    {r.stdout}
                  </pre>
                )}
                {r.stderr && (
                  <pre className="mt-1 max-h-32 overflow-auto whitespace-pre-wrap rounded bg-status-running/5 px-2 py-1.5 font-mono text-[11px] text-status-running">
                    {r.stderr}
                  </pre>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}

export function BlockResult({ type, result }: BlockResultProps) {
  if (!result) return null;

  const data = result as Record<string, unknown>;

  if (type === "command" && Array.isArray(data.results)) {
    return (
      <CommandResultView
        result={
          data as unknown as { job_id?: string; results: CommandResultItem[] }
        }
      />
    );
  }

  // Generic fallback: collapsible JSON
  return <GenericResult data={data} />;
}

function GenericResult({ data }: { data: Record<string, unknown> }) {
  const [expanded, setExpanded] = useState(false);

  return (
    <div className="mt-3 rounded-md border border-border bg-[#050505]">
      <button
        onClick={() => setExpanded(!expanded)}
        className="flex w-full items-center gap-2 px-3 py-2 text-left"
      >
        {expanded ? (
          <ChevronDown className="h-3.5 w-3.5 text-text-muted" />
        ) : (
          <ChevronRight className="h-3.5 w-3.5 text-text-muted" />
        )}
        <Text variant="label">Response</Text>
        {"job_id" in data ? (
          <span className="ml-auto text-[10px] text-text-muted/50 font-mono">
            {String(data.job_id).slice(0, 8)}
          </span>
        ) : null}
      </button>
      {expanded && (
        <pre className="max-h-40 overflow-auto border-t border-border px-3 py-2 font-mono text-[11px] text-text/80">
          {JSON.stringify(data, null, 2)}
        </pre>
      )}
    </div>
  );
}
