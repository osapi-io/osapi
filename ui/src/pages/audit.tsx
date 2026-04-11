import { useState, useEffect, useCallback } from "react";
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
import { Dropdown } from "@/components/ui/dropdown";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
  Loader2,
  Download,
  FileText,
  ChevronLeft,
  ChevronRight,
} from "lucide-react";
import {
  getAuditLogs,
  getAuditExport,
} from "@/sdk/gen/audit-log-api-audit/audit-log-api-audit";
import type { AuditEntry, ListAuditResponse } from "@/sdk/gen/schemas";

const PAGE_SIZE = 20;

const METHOD_OPTIONS = [
  { value: "", label: "All Methods" },
  { value: "GET", label: "GET" },
  { value: "POST", label: "POST" },
  { value: "PUT", label: "PUT" },
  { value: "DELETE", label: "DELETE" },
];

function statusBadgeVariant(code: number) {
  if (code >= 200 && code < 300) return "ready" as const;
  if (code >= 400 && code < 500) return "error" as const;
  if (code >= 500) return "error" as const;
  return "muted" as const;
}

function methodBadgeVariant(method: string) {
  switch (method) {
    case "GET":
      return "ready" as const;
    case "POST":
      return "running" as const;
    case "PUT":
      return "applied" as const;
    case "DELETE":
      return "error" as const;
    default:
      return "muted" as const;
  }
}

function formatTimestamp(ts: string) {
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

export function Audit() {
  useVimScroll();
  const { can } = useAuth();

  const [data, setData] = useState<ListAuditResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [exporting, setExporting] = useState(false);
  const [offset, setOffset] = useState(0);
  const [userFilter, setUserFilter] = useState("");
  const [methodFilter, setMethodFilter] = useState("");

  const fetchLogs = useCallback(async () => {
    try {
      const result = await getAuditLogs({
        limit: PAGE_SIZE,
        offset,
      });
      if (result.status === 200) {
        setData(result.data as ListAuditResponse);
        setError(null);
      } else {
        setError("Failed to fetch audit logs");
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to fetch audit logs");
    } finally {
      setLoading(false);
    }
  }, [offset]);

  useEffect(() => {
    let mounted = true;

    const poll = async () => {
      if (!mounted) return;
      await fetchLogs();
    };

    poll();
    const id = setInterval(poll, 10000);
    return () => {
      mounted = false;
      clearInterval(id);
    };
  }, [fetchLogs]);

  const handleExport = async () => {
    setExporting(true);
    try {
      const result = await getAuditExport();
      if (result.status === 200) {
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

  if (!can("audit:read")) {
    return (
      <ContentArea>
        <PageHeader title="Audit Log" subtitle="Insufficient permissions" />
        <ErrorBanner message="You do not have permission to view audit logs." />
      </ContentArea>
    );
  }

  // Client-side filtering on fetched page
  const filteredItems = (data?.items ?? []).filter((entry) => {
    if (
      userFilter &&
      !entry.user.toLowerCase().includes(userFilter.toLowerCase())
    ) {
      return false;
    }
    if (methodFilter && entry.method !== methodFilter) {
      return false;
    }
    return true;
  });

  const totalItems = data?.total_items ?? 0;
  const currentPage = Math.floor(offset / PAGE_SIZE) + 1;
  const totalPages = Math.ceil(totalItems / PAGE_SIZE);

  return (
    <ContentArea>
      <PageHeader
        title="Audit Log"
        subtitle="API request history and access trail"
        actions={
          <>
            {data && (
              <Badge variant="muted">
                {totalItems.toLocaleString()} entries
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

      {loading && (
        <div className="flex items-center justify-center py-20">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      )}

      {error && <ErrorBanner message={error} />}

      {!loading && !error && (
        <>
          {/* Filters */}
          <div className="mb-4">
            <SectionLabel icon={FileText}>Filters</SectionLabel>
            <Card>
              <CardContent>
                <div className="flex items-end gap-4">
                  <div className="w-48">
                    <Text variant="label" as="label" className="mb-1 block">
                      User
                    </Text>
                    <Input
                      value={userFilter}
                      onChange={(e) => setUserFilter(e.target.value)}
                      placeholder="Filter by user..."
                    />
                  </div>
                  <Dropdown
                    label="Method"
                    value={methodFilter}
                    onChange={setMethodFilter}
                    options={METHOD_OPTIONS}
                    className="w-40"
                  />
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Table */}
          <div className="mb-4">
            <SectionLabel>Entries ({filteredItems.length} shown)</SectionLabel>
            <DataTable
              columns={[
                {
                  header: "Timestamp",
                  cell: (e: AuditEntry) => (
                    <Text variant="mono-muted">
                      {formatTimestamp(e.timestamp)}
                    </Text>
                  ),
                },
                {
                  header: "User",
                  cell: (e: AuditEntry) => <Text variant="mono">{e.user}</Text>,
                },
                {
                  header: "Method",
                  cell: (e: AuditEntry) => (
                    <Badge variant={methodBadgeVariant(e.method)}>
                      {e.method}
                    </Badge>
                  ),
                },
                {
                  header: "Path",
                  cell: (e: AuditEntry) => (
                    <Text variant="mono-muted" truncate className="max-w-xs">
                      {e.path}
                    </Text>
                  ),
                },
                {
                  header: "Status",
                  align: "center",
                  cell: (e: AuditEntry) => (
                    <Badge variant={statusBadgeVariant(e.response_code)}>
                      {e.response_code}
                    </Badge>
                  ),
                },
                {
                  header: "Duration",
                  align: "right",
                  cell: (e: AuditEntry) => (
                    <Text variant="muted">{e.duration_ms}ms</Text>
                  ),
                },
              ]}
              rows={filteredItems}
              getRowKey={(e) => e.id}
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
    </ContentArea>
  );
}
