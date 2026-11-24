import { useState, useEffect } from "react";
import { TraceSpanStatus } from "@/sdk/gen/schemas";
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
import type { TraceNode, TraceResponse } from "@/sdk/gen/schemas";
import { statusVariant } from "@/lib/status";
import { PermissionDenied } from "@/components/layout/permission-denied";
import { Spinner } from "@/components/ui/spinner";
import { KVList } from "@/components/ui/kv-list";
import { formatDurationMs, formatTime } from "@/lib/format";

// TraceRow is one span of the tree the API answers, with the tree lines
// before its name and the place of its bar in the timing column.
interface TraceRow {
  node: TraceNode;
  // lines are the tree lines before the name, such as "│  └─ ".
  lines: string;
  // offset and width place the span's bar, as fractions of the trace.
  offset: number;
  width: number;
}

const startOf = (n: TraceNode) => new Date(n.span.start_time).getTime();

// flatten is the rows of the span tree, each child under its parent, in
// the order the API gives them.
function flatten(
  nodes: TraceNode[],
  lead = "",
  root = true,
): [TraceNode, string][] {
  return nodes.flatMap((n, i): [TraceNode, string][] => {
    const last = i === nodes.length - 1;
    const mark = root ? "" : last ? "└─ " : "├─ ";
    const under = root ? "" : last ? "   " : "│  ";
    return [
      [n, lead + mark],
      ...flatten(n.children ?? [], lead + under, false),
    ];
  });
}

export function Traces() {
  const { canCall } = useAuth();
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
    getTrace(id, { all: showNATS })
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
  }, [id, showNATS]);

  const current = id && result?.id === id ? result : null;
  const loading = !!id && !current;
  const trace = current?.trace ?? null;
  const error = current?.error ?? null;

  if (!canCall("getTrace")) {
    return <PermissionDenied title="Traces" what="traces" />;
  }

  // The API answers the tree and the summary; the page only lays out the
  // rows and the bar of each span against the whole trace.
  const flat = trace ? flatten(trace.spans) : [];
  const start = trace?.started ? new Date(trace.started).getTime() : 0;
  const end = Math.max(
    start,
    ...flat.map(([n]) => startOf(n) + n.span.duration_ms),
  );
  const total = Math.max(end - start, 1);
  const rows: TraceRow[] = flat.map(([n, lines]) => ({
    node: n,
    lines,
    offset: (startOf(n) - start) / total,
    width: Math.max(n.span.duration_ms / total, 0.005),
  }));
  const natsCount = trace?.nats_calls ?? 0;

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
                  ["Services", (trace.services ?? []).join(", ")],
                  ["Failed spans", trace.failed_spans],
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
            {(natsCount > 0 || showNATS) && (
              <Checkbox
                id="show-nats"
                checked={showNATS}
                onChange={(e) => setShowNATS(e.target.checked)}
                label={
                  showNATS
                    ? "Show the NATS calls, the bookkeeping of the job queue"
                    : `Show ${natsCount} NATS calls (${trace.nats_time}), the bookkeeping of the job queue`
                }
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
                    <span className="text-text">{r.node.name}</span>
                    {r.node.clock_skew && (
                      <span
                        className="ml-2 text-status-pending"
                        title={`This span started ${r.node.clock_skew} before its parent, so the clock of its host is behind. It and the spans under it are moved to the parent's start.`}
                      >
                        clock −{r.node.clock_skew}
                      </span>
                    )}
                  </span>
                ),
              },
              {
                header: "On",
                cell: (r: TraceRow) => <Text variant="mono">{r.node.on}</Text>,
              },
              {
                header: "Duration",
                align: "right",
                cell: (r: TraceRow) => (
                  <Text variant="muted">
                    {formatDurationMs(r.node.span.duration_ms)}
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
                        r.node.span.status === TraceSpanStatus.error
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
                  <Badge variant={statusVariant(r.node.status)}>
                    {r.node.status}
                  </Badge>
                ),
              },
              {
                header: "Detail",
                cell: (r: TraceRow) => (
                  <Text variant="mono-muted" truncate className="max-w-md">
                    {r.node.detail}
                  </Text>
                ),
              },
            ]}
            rows={rows}
            getRowKey={(r) => r.node.span.span_id}
          />
        </div>
      )}
    </ContentArea>
  );
}
