import { useState, useEffect } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Loader2, Search, GitBranch } from "lucide-react";
import { spinnerColor } from "@/lib/spinner-color";
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
import { Input } from "@/components/ui/input";
import { getTrace } from "@/sdk/gen/telemetry-api-telemetry/telemetry-api-telemetry";
import type { TraceResponse } from "@/sdk/gen/schemas";
import { traceRows, formatMs, type TraceRow } from "@/lib/trace-tree";

function statusVariant(status: string) {
  switch (status) {
    case "ok":
    case "completed":
      return "ready" as const;
    case "skipped":
      return "applied" as const;
    case "error":
    case "failed":
      return "error" as const;
    default:
      return "muted" as const;
  }
}

export function Traces() {
  const { can } = useAuth();
  const navigate = useNavigate();
  const { id } = useParams();
  const [input, setInput] = useState(id ?? "");
  // result is the answer for one trace ID. While the ID in the URL has no
  // answer yet, the page is loading.
  const [result, setResult] = useState<{
    id: string;
    trace?: TraceResponse;
    error?: string;
  } | null>(null);

  useEffect(() => {
    if (!id) return;
    let mounted = true;
    getTrace(id)
      .then((r) => {
        if (!mounted) return;
        if (r.status === 200) {
          setResult({ id, trace: r.data as TraceResponse });
        } else {
          setResult({ id, error: r.data.error ?? "Failed to read the trace" });
        }
      })
      .catch((e) => {
        if (mounted) {
          setResult({
            id,
            error: e instanceof Error ? e.message : "Failed to read the trace",
          });
        }
      });
    return () => {
      mounted = false;
    };
  }, [id]);

  const current = id && result?.id === id ? result : null;
  const loading = !!id && !current;
  const trace = current?.trace ?? null;
  const error = current?.error ?? null;

  if (!can("trace:read")) {
    return (
      <ContentArea>
        <PageHeader title="Traces" subtitle="Insufficient permissions" />
        <ErrorBanner message="You do not have permission to read traces." />
      </ContentArea>
    );
  }

  const rows = trace ? traceRows(trace.spans) : [];
  const failed = trace?.spans.filter((s) => s.status === "error").length ?? 0;
  const services = [...new Set(trace?.spans.map((s) => s.service) ?? [])];

  return (
    <ContentArea>
      <PageHeader
        title="Traces"
        subtitle="The span tree of a request, read through the controller"
        actions={
          trace && (
            <>
              <Badge variant="muted">{trace.spans.length} spans</Badge>
              {failed > 0 && <Badge variant="error">{failed} failed</Badge>}
            </>
          )
        }
      />

      <div className="mb-4">
        <SectionLabel icon={Search}>Trace ID</SectionLabel>
        <Card>
          <CardContent>
            <form
              className="flex items-end gap-4"
              onSubmit={(e) => {
                e.preventDefault();
                const next = input.trim().toLowerCase();
                if (next) navigate(`/admin/traces/${next}`);
              }}
            >
              <div className="w-96">
                <Input
                  value={input}
                  onChange={(e) => setInput(e.target.value)}
                  placeholder="The trace ID a command prints after its job ID"
                />
              </div>
              <Button type="submit" variant="secondary" size="sm">
                Show trace
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>

      {loading && (
        <div className="flex items-center justify-center py-20">
          <Loader2 className={`h-8 w-8 animate-spin ${spinnerColor}`} />
        </div>
      )}

      {error && <ErrorBanner message={error} />}

      {!loading && trace && (
        <div className="mb-4">
          <SectionLabel icon={GitBranch}>
            Spans ({trace.spans.length}) · {services.join(", ")}
          </SectionLabel>
          <DataTable
            columns={[
              {
                header: "Span",
                cell: (r: TraceRow) => (
                  <span className="whitespace-pre font-mono">
                    <span className="text-text-muted">{r.lines}</span>
                    <span className="text-text">{r.span.name}</span>
                  </span>
                ),
              },
              {
                header: "Host",
                cell: (r: TraceRow) => <Text variant="mono">{r.host}</Text>,
              },
              {
                header: "Duration",
                align: "right",
                cell: (r: TraceRow) => (
                  <Text variant="muted">{formatMs(r.span.duration_ms)}</Text>
                ),
              },
              {
                header: "Timing",
                className: "w-48",
                cell: (r: TraceRow) => (
                  <div className="relative h-2 w-48 rounded bg-white/5">
                    <div
                      className={
                        r.span.status === "error"
                          ? "absolute h-2 rounded bg-status-error"
                          : "absolute h-2 rounded bg-primary"
                      }
                      style={{
                        left: `${r.offset * 100}%`,
                        width: `${r.width * 100}%`,
                      }}
                    />
                  </div>
                ),
              },
              {
                header: "Status",
                align: "center",
                cell: (r: TraceRow) => (
                  <Badge variant={statusVariant(r.status)}>{r.status}</Badge>
                ),
              },
              {
                header: "Detail",
                cell: (r: TraceRow) => (
                  <Text variant="mono-muted" truncate className="max-w-md">
                    {r.detail}
                  </Text>
                ),
              },
            ]}
            rows={rows}
            getRowKey={(r) => r.span.span_id}
          />
        </div>
      )}
    </ContentArea>
  );
}
