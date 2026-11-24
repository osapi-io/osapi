import { useState, useEffect } from "react";
import { Pagination } from "@/components/ui/pagination";
import { unwrap, usePolling } from "@/hooks/use-polling";
import { useVimScroll } from "@/hooks/use-vim-scroll";
import { useAuth } from "@/hooks/use-auth";
import { ContentArea } from "@/components/layout/content-area";
import { PageHeader } from "@/components/ui/page-header";
import { SectionLabel } from "@/components/ui/section-label";
import { DataTable } from "@/components/ui/data-table";
import { ErrorBanner } from "@/components/ui/error-banner";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Text } from "@/components/ui/text";
import { Card, CardContent } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Loader2,
  Download,
  FileText,
  ChevronDown,
  ChevronUp,
} from "lucide-react";
import {
  getAuditLogs,
  getAuditExport,
} from "@/sdk/gen/audit-log-api-audit/audit-log-api-audit";
import type { AuditEntry, ListAuditResponse } from "@/sdk/gen/schemas";
import { PermissionDenied } from "@/components/layout/permission-denied";
import { Spinner } from "@/components/ui/spinner";
import { formatTime } from "@/lib/format";
import { AuditDetail } from "@/components/domain/audit-detail";
import { FilterInput } from "@/components/ui/filter-input";

const PAGE_SIZE = 20;

// The filters of the list, as osapi client audit list has them. The
// controller filters the whole log, so a page of the list is a page of the
// entries that match.
interface Filters {
  reads: boolean;
  user: string;
  operation: string;
  status: string;
}

export function Audit() {
  useVimScroll();
  const { canCall } = useAuth();

  const [exportError, setError] = useState<string | null>(null);
  const [exporting, setExporting] = useState(false);
  const [offset, setOffset] = useState(0);
  // Reads are hidden at first: the UI's own polling reads fill the log, and
  // what an operator looks for is a change.
  const [filters, setFilters] = useState<Filters>({
    reads: false,
    user: "",
    operation: "",
    status: "",
  });
  const [applied, setApplied] = useState<Filters>(filters);
  const [expandedId, setExpandedId] = useState<string | null>(null);

  // Apply the text filters when the typing stops, not at each key.
  useEffect(() => {
    const t = setTimeout(() => {
      setApplied(filters);
      setOffset(0);
    }, 300);
    return () => clearTimeout(t);
  }, [filters]);

  const logs = usePolling(
    async () => {
      const status = Number(applied.status);
      return unwrap<ListAuditResponse>(
        await getAuditLogs({
          limit: PAGE_SIZE,
          offset,
          writes: applied.reads ? undefined : true,
          user: applied.user.trim() || undefined,
          operation: applied.operation.trim() || undefined,
          status: Number.isInteger(status) && status > 0 ? status : undefined,
        }),
        "Failed to fetch audit logs",
      );
    },
    [offset, applied],
    { enabled: canCall("getAuditLogs") },
  );
  const { data, loading } = logs;
  const error = exportError ?? logs.error;

  const handleExport = async () => {
    setExporting(true);
    try {
      const result = await getAuditExport();
      if (result.status === 200) {
        setError(null);
        const exportData = result.data as ListAuditResponse;
        const blob = new Blob([JSON.stringify(exportData, null, 2)], {
          type: "application/json",
        });
        const url = URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = url;
        a.download = `audit-export-${new Date().toISOString().slice(0, 10)}.json`;
        a.click();
        URL.revokeObjectURL(url);
      } else {
        setError("Failed to export audit logs");
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to export audit logs");
    } finally {
      setExporting(false);
    }
  };

  if (!canCall("getAuditLogs")) {
    return <PermissionDenied title="Audit Log" what="audit entries" />;
  }

  const items = data?.items ?? [];

  const totalItems = data?.total_items ?? 0;

  return (
    <ContentArea>
      <PageHeader
        title="Audit Log"
        subtitle="API request history and access trail"
        actions={
          <>
            {data && (
              <Badge variant="muted">
                {totalItems} {totalItems === 1 ? "entry" : "entries"}
              </Badge>
            )}
            <Button
              variant="secondary"
              size="sm"
              onClick={handleExport}
              disabled={exporting}
            >
              {exporting ? (
                <Loader2 className="h-3.5 w-3.5 animate-spin" />
              ) : (
                <Download className="h-3.5 w-3.5" />
              )}
              Export JSON
            </Button>
          </>
        }
      />

      {loading && <Spinner />}

      {error && <ErrorBanner message={error} />}

      {!loading && !error && (
        <>
          {/* Filters */}
          <div className="mb-4">
            <SectionLabel icon={FileText}>Filters</SectionLabel>
            <Card>
              <CardContent>
                <div className="flex flex-wrap items-end gap-4">
                  <FilterInput
                    label="User"
                    value={filters.user}
                    placeholder="ops@example.com"
                    onChange={(user) => setFilters({ ...filters, user })}
                  />
                  <FilterInput
                    label="Operation or path"
                    value={filters.operation}
                    placeholder="hostname"
                    onChange={(operation) =>
                      setFilters({ ...filters, operation })
                    }
                  />
                  <FilterInput
                    label="Status"
                    value={filters.status}
                    placeholder="403"
                    className="w-24"
                    onChange={(status) => setFilters({ ...filters, status })}
                  />
                  <div className="pb-2">
                    <Checkbox
                      id="audit-reads"
                      checked={filters.reads}
                      onChange={(e) =>
                        setFilters({ ...filters, reads: e.target.checked })
                      }
                      label="Show reads"
                    />
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Table */}
          <div className="mb-4">
            <SectionLabel>
              {applied.reads ? "Audit entries" : "Changes"} ({items.length} of{" "}
              {totalItems})
            </SectionLabel>
            <DataTable
              // Who did what, as osapi client audit list prints it; a click
              // opens the entry in full.
              onRowClick={(e: AuditEntry) =>
                setExpandedId(expandedId === e.id ? null : e.id)
              }
              expanded={(e: AuditEntry) =>
                expandedId === e.id ? <AuditDetail entry={e} /> : undefined
              }
              columns={[
                {
                  header: "",
                  cell: (e: AuditEntry) =>
                    expandedId === e.id ? (
                      <ChevronUp className="h-3.5 w-3.5 text-text-muted" />
                    ) : (
                      <ChevronDown className="h-3.5 w-3.5 text-text-muted" />
                    ),
                },
                {
                  header: "Time",
                  className: "whitespace-nowrap",
                  cell: (e: AuditEntry) => (
                    <Text variant="mono-muted">{formatTime(e.timestamp)}</Text>
                  ),
                },
                {
                  header: "User",
                  cell: (e: AuditEntry) => <Text variant="mono">{e.user}</Text>,
                },
                {
                  header: "Method",
                  cell: (e: AuditEntry) => (
                    <Text className="font-semibold">{e.method}</Text>
                  ),
                },
                {
                  header: "Operation",
                  cell: (e: AuditEntry) => (
                    <Text
                      variant="mono-muted"
                      truncate
                      className="max-w-xs"
                      title={e.path}
                    >
                      {e.operation_id || e.path}
                    </Text>
                  ),
                },
                {
                  header: "Status",
                  align: "center",
                  cell: (e: AuditEntry) => <Text>{e.response_code}</Text>,
                },
                {
                  header: "Duration",
                  align: "right",
                  cell: (e: AuditEntry) => (
                    <Text variant="muted">{e.duration_ms}ms</Text>
                  ),
                },
              ]}
              rows={items}
              getRowKey={(e) => e.id}
            />
          </div>

          {/* Pagination */}
          <Pagination
            offset={offset}
            limit={PAGE_SIZE}
            total={totalItems}
            onChange={setOffset}
          />
        </>
      )}
    </ContentArea>
  );
}
