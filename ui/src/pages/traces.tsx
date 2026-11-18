import { useState, useEffect } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Search, GitBranch } from "lucide-react";
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
import { Checkbox } from "@/components/ui/checkbox";
import { getTrace } from "@/sdk/gen/telemetry-api-telemetry/telemetry-api-telemetry";
import type { TraceResponse } from "@/sdk/gen/schemas";
import { traceRows, natsCalls, type TraceRow } from "@/lib/trace-tree";
import { statusVariant } from "@/lib/status";
import { PermissionDenied } from "@/components/layout/permission-denied";
import { Spinner } from "@/components/ui/spinner";
import { KVList } from "@/components/ui/kv-list";
import { formatDurationMs, formatTime } from "@/lib/format";

export function Traces() {
  const { can } = useAuth();
  const navigate = useNavigate();
  const { id } = useParams();
  const [input, setInput] = useState(id ?? "");
  const [showNATS, setShowNATS] = useState(false);
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
    return <PermissionDenied title="Traces" what="traces" />;
  }

  const rows = trace ? traceRows(trace.spans, showNATS) : [];
  const nats = trace ? natsCalls(trace.spans) : { count: 0, ms: 0 };
  const failed = trace?.spans.filter((s) => s.status === "error").length ?? 0;
  const services = [...new Set(trace?.spans.map((s) => s.service) ?? [])];
  // The trace's start and its whole duration, as osapi client telemetry
  // trace get prints them: from the first span's start to the last end.
  const starts =
    trace?.spans.map((s) => new Date(s.start_time).getTime()) ?? [];
  const start = starts.length ? Math.min(...starts) : 0;
  const end = trace?.spans.length
    ? Math.max(
        ...trace.spans.map(
          (s) => new Date(s.start_time).getTime() + s.duration_ms,
        ),
      )
    : 0;

  return (
    <ContentArea>
      <PageHeader
        title="Traces"
        subtitle="The span tree of a request, read through the controller"
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

      {loading && <Spinner />}

      {error && <ErrorBanner message={error} />}

      {!loading && trace && (
        <div className="mb-4">
          <SectionLabel>Trace</SectionLabel>
          <Card>
            <CardContent>
              <KVList
                pairs={[
                  ["ID", <span className="font-mono">{id}</span>],
                  ["Started", formatTime(new Date(start).toISOString())],
                  ["Duration", formatDurationMs(end - start)],
                  ["Services", services.join(", ")],
                  ["Failed spans", failed > 0 && failed],
                ]}
              />
            </CardContent>
          </Card>
        </div>
      )}

      {!loading && trace && (
        <div className="mb-4">
          <div className="flex items-center justify-between">
            <SectionLabel icon={GitBranch}>Spans ({rows.length})</SectionLabel>
            {nats.count > 0 && (
              <Checkbox
                id="show-nats"
                checked={showNATS}
                onChange={(e) => setShowNATS(e.target.checked)}
                label={`Show ${nats.count} NATS calls (${formatDurationMs(nats.ms)}), the bookkeeping of the job queue`}
              />
            )}
          </div>
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
                header: "Hostname",
                cell: (r: TraceRow) => <Text variant="mono">{r.host}</Text>,
              },
              {
                header: "Duration",
                align: "right",
                cell: (r: TraceRow) => (
                  <Text variant="muted">
                    {formatDurationMs(r.span.duration_ms)}
                  </Text>
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
