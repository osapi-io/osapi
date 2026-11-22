import { useMemo, type ReactNode } from "react";
import { useVimScroll } from "@/hooks/use-vim-scroll";
import { ContentArea } from "@/components/layout/content-area";
import { AgentCard } from "@/components/domain/agent-card";
import { ComponentRow } from "@/components/domain/component-row";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/ui/page-header";
import { ErrorBanner } from "@/components/ui/error-banner";
import { StatCard } from "@/components/ui/stat-card";
import { SectionLabel } from "@/components/ui/section-label";
import { Text } from "@/components/ui/text";
import { useHealth } from "@/hooks/use-health";
import { useAgents } from "@/hooks/use-agents";
import { Activity } from "lucide-react";
import type { ComponentEntry } from "@/sdk/gen/schemas";
import { statusVariant } from "@/lib/status";
import { Spinner } from "@/components/ui/spinner";
import { Count } from "@/components/ui/count";
import { KVList } from "@/components/ui/kv-list";
import { formatBytes } from "@/lib/format";

// SubRow is one row under a component card: a sub-component of the
// component, or a check that the controller made, such as its connection
// to NATS.
type SubRow = {
  name: string;
  status: string;
  address?: string;
  url?: string;
};

// subRows are the sub-components of entry, as the API sorts them.
function subRows(entry?: ComponentEntry): SubRow[] {
  return (entry?.components ?? []).map((c) => ({
    name: c.type ?? "",
    status: c.status ?? "",
    address: c.address,
    url: c.url,
  }));
}

// RegistryHeader is one row of osapi client health status's component
// table: hostname, status, conditions, age, CPU, memory and version, in that
// order.
function RegistryHeader({
  entry,
  reference,
}: {
  entry?: ComponentEntry;
  // reference is the controller's version, which the page header shows.
  reference?: string;
}) {
  if (!entry) return null;
  return (
    <div className="border-b border-border/30 px-3 py-2 text-xs">
      <div className="flex items-center gap-3 whitespace-nowrap">
        {entry.hostname && (
          <span
            className="min-w-0 truncate font-medium text-text"
            title={entry.hostname}
          >
            {entry.hostname}
          </span>
        )}
        {entry.status && (
          <Badge variant={statusVariant(entry.status)}>{entry.status}</Badge>
        )}
        {entry.conditions && entry.conditions.length > 0 && (
          <span className="text-status-pending">
            {entry.conditions.join(", ")}
          </span>
        )}
        <div className="ml-auto flex shrink-0 items-center gap-3 text-text-muted">
          {entry.age && <span>up {entry.age}</span>}
          {entry.cpu_percent != null && (
            <span>{entry.cpu_percent.toFixed(1)}%</span>
          )}
          {entry.mem_bytes != null && (
            <span>{formatBytes(entry.mem_bytes)}</span>
          )}
        </div>
      </div>
      {/* The version only when it is not the one in the page header: a
          mismatch is what a reader compares for. */}
      {entry.version && entry.version !== reference && (
        <p
          className="mt-0.5 truncate font-mono text-[10px] text-status-pending"
          title={entry.version}
        >
          {entry.version}
        </p>
      )}
    </div>
  );
}

export function Dashboard() {
  useVimScroll();
  const {
    data: health,
    error: healthErr,
    loading: healthLoading,
  } = useHealth();
  const {
    agents,
    error: agentErr,
    loading: agentLoading,
    refresh: refreshAgents,
  } = useAgents();

  // Filter out pending agents — they're managed in Admin > Admission.
  const activeAgents = useMemo(
    () => agents.filter((a) => a.state !== "Pending"),
    [agents],
  );

  const loading = healthLoading || agentLoading;
  const error = healthErr || agentErr;

  const controllerEntry = health?.registry?.find(
    (r) => r.type === "controller",
  );
  const natsEntry = health?.registry?.find((r) => r.type === "nats");
  const grouped = useMemo(
    () =>
      health
        ? {
            controller: [
              ...subRows(controllerEntry),
              ...(health.checks ?? []).map((c): SubRow => ({
                name: `${c.name} (connectivity)`,
                status: c.status,
              })),
            ],
            nats: subRows(natsEntry),
          }
        : null,
    [health, controllerEntry, natsEntry],
  );

  return (
    <ContentArea>
      <PageHeader
        title="Dashboard"
        subtitle="Fleet health overview"
        actions={
          health && (
            <>
              <Badge variant={statusVariant(health.status)}>
                {health.status}
              </Badge>
            </>
          )
        }
      />

      {loading && <Spinner />}

      {error && <ErrorBanner message={error} />}

      {!loading && !error && health && (
        <>
          {/* Row 1: what needs attention. The NATS version is in the NATS
              card and the consumer count is on the stream's line. */}
          <div className="mb-4 grid grid-cols-2 gap-3 lg:grid-cols-4">
            {health.agents && (
              <StatCard
                label="Agents"
                value={`${health.agents.ready}/${health.agents.total}`}
                detail={
                  health.agents.ready === health.agents.total ? (
                    "ready"
                  ) : (
                    <span className="text-status-pending">
                      {health.agents.total - health.agents.ready} not ready
                    </span>
                  )
                }
              />
            )}
            {health.jobs && (
              <StatCard
                label="Failed jobs"
                value={<Count n={health.jobs.failed} kind="failed" bare />}
                detail={
                  <>
                    of {health.jobs.total}:{" "}
                    <Count n={health.jobs.completed} kind="completed" />,{" "}
                    <Count n={health.jobs.unprocessed} kind="unprocessed" />
                  </>
                }
              />
            )}
            {health.jobs && (
              <StatCard
                label="Dead letters"
                value={<Count n={health.jobs.dlq} kind="dlq" bare />}
                detail="jobs that used up their retries"
              />
            )}
            <StatCard
              label="Controller"
              value={`up ${health.uptime}`}
              detail={health.version}
              truncateDetail
            />
          </div>

          {/* Row 2: the controller on the left; NATS and its storage on
              the right, so the short NATS card leaves no gap. */}
          <div className="mb-6 grid grid-cols-12 gap-4">
            {grouped && grouped.controller.length > 0 && (
              <div className="col-span-12 lg:col-span-7">
                <SectionLabel>Controller</SectionLabel>
                <Card>
                  <CardContent className="p-0">
                    <RegistryHeader
                      entry={controllerEntry}
                      reference={health.version}
                    />
                    {grouped.controller.map((row, i) => (
                      <ComponentRow
                        key={row.name}
                        name={row.name}
                        status={row.status}
                        address={row.address}
                        url={row.url}
                        isLast={i === grouped.controller.length - 1}
                      />
                    ))}
                  </CardContent>
                </Card>
              </div>
            )}
            <div className="col-span-12 space-y-4 lg:col-span-5">
              {grouped && grouped.nats.length > 0 && (
                <div>
                  <SectionLabel>NATS Server</SectionLabel>
                  <Card>
                    <CardContent className="p-0">
                      <RegistryHeader
                        entry={natsEntry}
                        reference={health.version}
                      />
                      {health.nats && (
                        <p className="truncate border-b border-border/20 px-3 py-1.5 text-xs text-text-muted">
                          NATS {health.nats.version}, {health.nats.url}
                        </p>
                      )}
                      {grouped.nats.map((row, i) => (
                        <ComponentRow
                          key={row.name}
                          name={row.name}
                          status={row.status}
                          address={row.address}
                          url={row.url}
                          isLast={i === grouped.nats.length - 1}
                        />
                      ))}
                    </CardContent>
                  </Card>
                </div>
              )}

              {/* Storage, as osapi client health status prints it: a line
                  for each stream, bucket and object store. */}
              {Boolean(
                health.streams?.length ||
                health.kv_buckets?.length ||
                health.object_stores?.length,
              ) && (
                <div>
                  <SectionLabel>Storage</SectionLabel>
                  <Card>
                    <CardContent>
                      <KVList
                        pairs={[
                          ...(health.streams ?? []).map(
                            (st): [string, ReactNode] => [
                              "Stream",
                              <>
                                {st.name}{" "}
                                <span className="text-text-muted">
                                  ({st.messages} msgs, {formatBytes(st.bytes)},{" "}
                                  {st.consumers} consumers)
                                </span>
                              </>,
                            ],
                          ),
                          ...(health.kv_buckets ?? []).map(
                            (kv): [string, ReactNode] => [
                              "Bucket",
                              <>
                                {kv.name}{" "}
                                <span className="text-text-muted">
                                  ({kv.keys} keys, {formatBytes(kv.bytes)})
                                </span>
                              </>,
                            ],
                          ),
                          ...(health.object_stores ?? []).map(
                            (os): [string, ReactNode] => [
                              "Object store",
                              <>
                                {os.name}{" "}
                                <span className="text-text-muted">
                                  ({formatBytes(os.size)})
                                </span>
                              </>,
                            ],
                          ),
                        ]}
                      />
                    </CardContent>
                  </Card>
                </div>
              )}
            </div>
          </div>

          {/* Agents — pending agents are managed in Admin > Admission */}
          <section>
            <SectionLabel icon={Activity}>
              Agents ({activeAgents.length})
            </SectionLabel>
            {activeAgents.length === 0 && (
              <Text variant="muted" as="p">
                None.
              </Text>
            )}
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
              {activeAgents.map((agent) => (
                <AgentCard
                  key={agent.hostname}
                  agent={agent}
                  controllerVersion={health.version}
                  components={
                    health?.registry?.find(
                      (r) =>
                        r.type === "agent" && r.hostname === agent.hostname,
                    )?.components
                  }
                  onRefresh={refreshAgents}
                />
              ))}
            </div>
          </section>
        </>
      )}
    </ContentArea>
  );
}
