import { useState } from "react";
import { Code, Trash2, Copy, Check } from "lucide-react";
import { SearchBox } from "@/components/ui/search-box";
import type { FileInfo, FileUploadResponse } from "@/sdk/gen/schemas";
import { JobDetail } from "@/components/domain/job-detail";
import { CodeBlock } from "@/components/ui/code-block";
import { KeyValue } from "@/components/ui/key-value";
import { StatusIcon } from "@/components/ui/status-icon";
import { Text } from "@/components/ui/text";
import { HealthDot } from "@/components/ui/health-dot";
import { CollapsibleSection } from "@/components/ui/collapsible-section";
import { IconButton } from "@/components/ui/icon-button";
import { HostGroupHeader } from "@/components/domain/host-group-header";
import { ResultSummary } from "@/components/domain/result-summary";
import { hostOutcomes } from "@/lib/outcome";
import { formatBytes, formatDurationMs, formatTime } from "@/lib/format";
import {
  hasValue,
  type CatalogColumn,
  type CatalogMethod,
} from "@/lib/catalog";

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

// ---------------------------------------------------------------------------
// Shared
// ---------------------------------------------------------------------------

/** HostError is the error of one host, under its row. */
function HostError({ children }: { children: React.ReactNode }) {
  return (
    <Text variant="error" as="p" className="px-4 pb-2">
      {children}
    </Text>
  );
}

function ChangedLabel({ changed }: { changed?: boolean }) {
  if (changed === undefined || changed === null) return null;
  return (
    <Text variant={changed ? "primary" : "label"}>
      {changed ? "changed" : "unchanged"}
    </Text>
  );
}

function RawJsonToggle({ data }: { data: unknown }) {
  const [open, setOpen] = useState(false);
  const [copied, setCopied] = useState(false);
  const json = JSON.stringify(data, null, 2);

  const handleCopy = async (e: React.MouseEvent) => {
    e.stopPropagation();
    await navigator.clipboard.writeText(json);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };

  return (
    <CollapsibleSection
      open={open}
      onToggle={() => setOpen(!open)}
      icon={Code}
      label="JSON"
      rightContent={
        open ? (
          <span
            onClick={handleCopy}
            className="flex items-center gap-1 rounded px-1.5 py-0.5 hover:bg-white/5"
          >
            {copied ? (
              <Check className="h-3 w-3 text-primary" />
            ) : (
              <Copy className="h-3 w-3" />
            )}
            {copied ? "Copied" : "Copy"}
          </span>
        ) : undefined
      }
    >
      <CodeBlock variant="json">{json}</CodeBlock>
    </CollapsibleSection>
  );
}

function ResultShell({
  jobId,
  label,
  summary,
  children,
  data,
}: {
  jobId?: string;
  label: string;
  /** The summary line; for a result with per-host results, it is counted from them. */
  summary?: React.ReactNode;
  children?: React.ReactNode;
  data: unknown;
}) {
  return (
    <div className="rounded-lg border border-border/60 bg-surface shadow-[inset_0_1px_0_rgba(255,255,255,0.02)]">
      <div className="flex items-center gap-2 border-b border-border/40 px-4 py-2.5">
        {summary ??
          (resultsOf(data).length > 0 && (
            <ResultSummary outcomes={hostOutcomes(resultsOf(data))} />
          ))}
        <Text variant="muted" className="font-semibold">
          {label}
        </Text>
        {jobId && (
          <div className="ml-auto">
            <JobDetail jobId={jobId} inline />
          </div>
        )}
      </div>
      {jobId && <JobDetail jobId={jobId} panel />}
      {children}
      <RawJsonToggle data={data} />
    </div>
  );
}

// ---------------------------------------------------------------------------
// File Upload (single result, no hosts)
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// List renderers (expanded, with search)
// ---------------------------------------------------------------------------

function useListSearch() {
  const [search, setSearch] = useState("");
  return { search, setSearch };
}

function matches(
  term: string,
  ...fields: (string | number | boolean | undefined | null)[]
) {
  if (!term) return true;
  const t = term.toLowerCase();
  return fields.some((f) => f != null && String(f).toLowerCase().includes(t));
}

function ListSearchBar({
  search,
  setSearch,
  hasItems,
}: {
  search: string;
  setSearch: (v: string) => void;
  hasItems: boolean;
}) {
  if (!hasItems) return null;
  return (
    <div className="border-b border-border/40 px-4 py-2">
      <SearchBox
        value={search}
        onChange={setSearch}
        onClose={() => setSearch("")}
        placeholder="Filter..."
        autoFocus={false}
      />
    </div>
  );
}

// ---------------------------------------------------------------------------
// Router
// ---------------------------------------------------------------------------

/**
 * A value as the CLI's output shows it, by its column's render tag: bytes
 * as a size, a time in the one time format, milliseconds as a duration, and
 * short=N cut to N characters. A list is joined.
 */
function formatCell(c: CatalogColumn, v: unknown): string {
  if (Array.isArray(v)) return v.map((x) => formatCell(c, x)).join(", ");
  if (v !== null && typeof v === "object") return objectText(c, v);
  const r = c.render ?? "";
  if (r === "bytes" && typeof v === "number") return formatBytes(v);
  if (r === "time" && typeof v === "string") return formatTime(v);
  if (r === "duration=ms" && typeof v === "number") {
    return formatDurationMs(v);
  }
  if (r.startsWith("short=")) {
    const n = Number(r.slice("short=".length));
    const t = String(v);
    return t.length > n ? `${t.slice(0, n)}...` : t;
  }
  if (typeof v === "boolean") return v ? "yes" : "no";
  return String(v);
}

/** An object as one line, each field by its label: Total 1.9 GB, Used 227 MB. */
function objectText(c: CatalogColumn, v: object): string {
  const cols = c.columns ?? [];
  const o = v as Record<string, unknown>;
  if (cols.length === 0) {
    return Object.entries(o)
      .map(([k, x]) => `${k} ${String(x)}`)
      .join(", ");
  }
  return cols
    .filter((sub) => hasValue(o[sub.name]))
    .map((sub) => `${sub.label.toLowerCase()} ${formatCell(sub, o[sub.name])}`)
    .join(", ");
}

/**
 * What a host's rows count: its rows, or, for one row that is only a list,
 * such as the disks of a host, the items of the list: 4 disks.
 */
function countOf(
  def: CatalogMethod,
  rows: readonly Record<string, unknown>[],
): string {
  const shown = def.columns.filter((c) => c.name !== "changed");
  const list = shown.length === 1 && shown[0].kind === "list" ? shown[0] : null;
  if (rows.length === 1 && list && Array.isArray(rows[0][list.name])) {
    const n = (rows[0][list.name] as unknown[]).length;
    const noun = list.label.toLowerCase();
    return `${n} ${n === 1 ? noun.replace(/s$/, "") : noun}`;
  }
  return `${rows.length} ${rows.length === 1 ? "row" : "rows"}`;
}

/** The text of every value of a row, for the search. */
function rowText(cols: readonly CatalogColumn[], r: object): string[] {
  const o = r as Record<string, unknown>;
  return cols.map((c) => (hasValue(o[c.name]) ? formatCell(c, o[c.name]) : ""));
}

/** The fields of an object or a list item, as label and value pairs. */
function Pairs({
  cols,
  item,
  strongFirst,
}: {
  cols: readonly CatalogColumn[];
  item: Record<string, unknown>;
  strongFirst?: boolean;
}) {
  return (
    <>
      {cols
        .filter((c) => c.kind !== "list" && hasValue(item[c.name]))
        .map((c, j) => (
          <KeyValue
            key={c.name}
            label={c.label.toLowerCase()}
            value={formatCell(c, item[c.name])}
            variant={strongFirst && j === 0 ? "strong" : undefined}
          />
        ))}
    </>
  );
}

/**
 * The result of a method of a behavior: the rows that its method answers, one per
 * item, grouped by host, with a column for each field of the output. A host
 * that failed or was skipped is a header without rows.
 */
function GenericRowsResult({
  data,
  def,
  noun,
}: {
  data: Record<string, unknown>;
  def: CatalogMethod;
  /** What the rows are, as the summary line names them: dns get. */
  noun: string;
}) {
  const { search, setSearch } = useListSearch();
  type Row = Record<string, unknown> & {
    hostname: string;
    status?: string;
    error?: string;
  };
  const rows = (Array.isArray(data.results) ? data.results : []) as Row[];
  const hosts: {
    hostname: string;
    status?: string;
    error?: string;
    rows: Row[];
  }[] = [];
  for (const row of rows) {
    let host = hosts.find((h) => h.hostname === row.hostname);
    if (!host) {
      host = {
        hostname: row.hostname,
        status: row.status,
        error: row.error,
        rows: [],
      };
      hosts.push(host);
    }
    if (def.columns.some((c) => hasValue(row[c.name]))) host.rows.push(row);
  }
  const shown = def.columns.filter((c) => c.name !== "changed");
  const total = hosts.reduce((n, h) => n + h.rows.length, 0);
  return (
    <ResultShell
      jobId={data.job_id as string | undefined}
      label={noun}
      data={data}
    >
      {total > 1 && (
        <ListSearchBar
          search={search}
          setSearch={setSearch}
          hasItems={total > 0}
        />
      )}
      {hosts.map((host) => {
        const visible = host.rows.filter((r) =>
          matches(search, ...rowText(shown, r)),
        );
        if (search && visible.length === 0 && !host.error) return null;
        const failed =
          host.status === "failed" || (host.error && host.status !== "skipped");
        const changed = host.rows.some((r) => r.changed === true);
        return (
          <div
            key={host.hostname}
            className="border-b border-border/50 last:border-0"
          >
            <HostGroupHeader
              hostname={host.hostname}
              status={
                host.status === "skipped" ? "skipped" : failed ? "failed" : "ok"
              }
              detail={
                host.status === "skipped"
                  ? "skipped"
                  : failed
                    ? "failed"
                    : def.httpMethod !== "GET"
                      ? changed
                        ? "changed"
                        : "unchanged"
                      : countOf(def, host.rows)
              }
            />
            {host.error && host.status !== "skipped" && (
              <HostError>{host.error}</HostError>
            )}
            {!(def.httpMethod !== "GET") &&
              visible.map((r, i) => (
                <div
                  key={i}
                  className="flex flex-wrap items-center gap-x-4 gap-y-1 border-b border-border/30 px-4 py-2 pl-9 last:border-0"
                >
                  <Pairs cols={shown} item={r} strongFirst />
                  {/* A list in the row, such as the disks of a host: one
                      line for each item, under the row. */}
                  {shown
                    .filter(
                      (c) => c.kind === "list" && Array.isArray(r[c.name]),
                    )
                    .flatMap((c) =>
                      (r[c.name] as Record<string, unknown>[]).map(
                        (item, k) => (
                          <div
                            key={`${c.name}-${k}`}
                            className="flex w-full flex-wrap items-center gap-x-4 gap-y-1 pl-4"
                          >
                            <Pairs
                              cols={c.columns ?? []}
                              item={item}
                              strongFirst
                            />
                          </div>
                        ),
                      ),
                    )}
                </div>
              ))}
          </div>
        );
      })}
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
        {files.map((f, i) => (
          <div
            key={i}
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
        {entries.map((e, i) => (
          <div
            key={i}
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

function formatSize(bytes: number) {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

// resultsOf is a result's per-host results, or none.
function resultsOf(data: unknown): Parameters<typeof hostOutcomes>[0] {
  const r = (data as { results?: unknown })?.results;
  return Array.isArray(r) ? r : [];
}
