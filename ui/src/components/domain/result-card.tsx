import { Trash2 } from "lucide-react";
import type { FileInfo, FileUploadResponse } from "@/sdk/gen/schemas";
import { KeyValue } from "@/components/ui/key-value";
import { StatusIcon } from "@/components/ui/status-icon";
import { Text } from "@/components/ui/text";
import { HealthDot } from "@/components/ui/health-dot";
import { IconButton } from "@/components/ui/icon-button";
import { type CatalogMethod } from "@/lib/catalog";
import { ResultShell, ChangedLabel } from "@/components/domain/result-shell";
import { formatSize } from "@/lib/result-format";
import { GenericRowsResult } from "@/components/domain/rows-result";

interface ResultCardProps {
  /** The step that ran, behavior.method: dns.update. */
  run: string;
  /** The method of a behavior that ran, from the catalog. */
  method?: CatalogMethod;
  /** The title of the method's behavior, such as DNS. */
  title?: string;
  result: unknown;
  onAction?: (action: string, name: string) => void;
}

function FileUploadResult({ data }: { data: FileUploadResponse }) {
  return (
    <ResultShell
      label="uploaded"
      summary={<HealthDot status={data.changed ? "ok" : "muted"} />}
      data={data}
    >
      <div className="flex items-center gap-4 px-4 py-2.5">
        <StatusIcon status="ok" />
        <Text className="font-medium">{data.name}</Text>
        <Text variant="muted">{formatSize(data.size)}</Text>
        <Text variant="muted">{data.content_type}</Text>
        <ChangedLabel changed={data.changed} />
        <Text variant="mono-muted" className="ml-auto">
          sha:{data.sha256.slice(0, 12)}
        </Text>
      </div>
    </ResultShell>
  );
}

export function ResultCard({
  run,
  method,
  title,
  result,
  onAction,
}: ResultCardProps) {
  if (!result) return null;
  const data = result as Record<string, unknown>;

  // A method of a behavior shows its rows as the CLI does, from the columns
  // of its output: no behavior has a card of its own.
  if (method) {
    return (
      <GenericRowsResult
        data={data}
        def={method}
        noun={`${(title ?? run).toLowerCase()} ${method.name}`}
      />
    );
  }

  // File list
  if (run === "objectstore.list" && "files" in data) {
    const files = (data as { files: FileInfo[] }).files;
    return (
      <ResultShell
        label={`${files.length} file${files.length !== 1 ? "s" : ""}`}
        data={data}
      >
        {files.map((f) => (
          <div
            key={f.name}
            className="flex items-center gap-4 border-b border-border/50 px-4 py-2 last:border-0"
          >
            <StatusIcon status="ok" />
            <KeyValue label="name" value={f.name} variant="strong" />
            <KeyValue label="size" value={formatSize(f.size)} />
            <KeyValue label="type" value={f.content_type} />
            {f.source && (
              <KeyValue label="source" value={f.source} variant="accent" />
            )}
            <Text variant="mono-muted">sha:{f.sha256.slice(0, 12)}</Text>
            {onAction && (
              <IconButton
                icon={Trash2}
                variant="danger"
                onClick={() => onAction("objectstore.delete", f.name)}
                title={`Delete ${f.name}`}
                className="ml-auto"
              />
            )}
          </div>
        ))}
      </ResultShell>
    );
  }

  if (run === "objectstore.upload" && "name" in data && "sha256" in data) {
    return <FileUploadResult data={data as unknown as FileUploadResponse} />;
  }

  // File stale — list of stale deployments
  if (run === "objectstore.stale" && Array.isArray(data.stale)) {
    type StaleEntry = {
      object_name: string;
      hostname: string;
      provider: string;
      path: string;
      deployed_sha: string;
      current_sha: string;
      deployed_at: string;
    };
    const entries = data.stale as StaleEntry[];
    return (
      <ResultShell
        label={`${entries.length} stale deployment${entries.length !== 1 ? "s" : ""}`}
        summary={<HealthDot status={entries.length > 0 ? "error" : "ok"} />}
        data={data}
      >
        {entries.map((e) => (
          <div
            key={`${e.hostname}:${e.path}`}
            className="flex items-center gap-4 border-b border-border/50 px-4 py-2 last:border-0"
          >
            <StatusIcon status="failed" />
            <KeyValue label="host" value={e.hostname} variant="strong" />
            <KeyValue label="object" value={e.object_name} variant="accent" />
            <KeyValue label="path" value={e.path} />
            <Text variant="mono-muted" className="ml-auto">
              sha:{e.current_sha.slice(0, 12)}
            </Text>
          </div>
        ))}
        {entries.length === 0 && (
          <div className="px-4 py-3">
            <Text variant="muted">None.</Text>
          </div>
        )}
      </ResultShell>
    );
  }

  // Fallback — unknown type, show in shell with JSON toggle
  return <ResultShell label="result" summary={null} data={data} />;
}
