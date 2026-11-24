import { KeyValue } from "@/components/ui/key-value";
import { HostGroupHeader } from "@/components/domain/host-group-header";
import {
  hasValue,
  type CatalogColumn,
  type CatalogMethod,
} from "@/lib/catalog";
import { ResultShell, HostError } from "@/components/domain/result-shell";
import { ListSearchBar } from "@/components/domain/list-search";
import { useListSearch } from "@/hooks/use-list-search";
import { matches } from "@/lib/result-format";
import { formatCell, countOf, rowText } from "@/lib/result-format";

/** The fields of an object or a list item, as label and value pairs. */
export function Pairs({
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
export function GenericRowsResult({
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
                    : def.http_method !== "GET"
                      ? changed
                        ? "changed"
                        : "unchanged"
                      : countOf(def, host.rows)
              }
            />
            {host.error && host.status !== "skipped" && (
              <HostError>{host.error}</HostError>
            )}
            {!(def.http_method !== "GET") &&
              visible.map((r) => (
                <div
                  key={JSON.stringify(r)}
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
