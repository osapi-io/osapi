import { useMemo } from "react";
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
import { DataTable } from "@/components/ui/data-table";
import { Text } from "@/components/ui/text";
import { useHealth } from "@/hooks/use-health";
import { useAgents } from "@/hooks/use-agents";
import { Activity, Loader2 } from "lucide-react";
import type { ComponentHealth, ComponentEntry } from "@/sdk/gen/schemas";

function formatBytes(bytes: number) {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  if (bytes < 1024 * 1024 * 1024)
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`;
}

type GroupedComponents = {
  controller: [string, ComponentHealth][];
  nats: [string, ComponentHealth][];
  agent: [string, ComponentHealth][];
};

function groupComponents(
  components: Record<string, ComponentHealth>,
): GroupedComponents {
  const groups: GroupedComponents = { controller: [], nats: [], agent: [] };
  for (const [key, comp] of Object.entries(components)) {
    const lower = key.toLowerCase();
    // Strip the prefix for display: "Controller.Api" → "Api"
    const label = key.includes(".") ? key.split(".").slice(1).join(".") : key;
    if (lower.startsWith("controller.")) {
      groups.controller.push([label, comp]);
    } else if (lower.startsWith("nats.")) {
      groups.nats.push([label, comp]);
    } else if (lower.startsWith("agent.")) {
      groups.agent.push([label, comp]);
    } else {
      groups.controller.push([label, comp]);
    }
  }
  return groups;
}

function RegistryHeader({ entry }: { entry?: ComponentEntry }) {
  if (!entry) return null;
  return (
    <div className="flex items-center gap-3 border-b border-border/30 px-3 py-2 text-xs">
      {entry.hostname && (
        <span className="font-medium text-text">{entry.hostname}</span>
      )}
      {entry.status && (
        <Badge
          variant={
            entry.status === "Ready" || entry.status === "ok"
              ? "ready"
              : "error"
          }
        >
          {entry.status}
        </Badge>
      )}
      <div className="ml-auto flex items-center gap-3 text-text-muted">
        {entry.cpu_percent != null && (
          <span>{entry.cpu_percent.toFixed(1)}% cpu</span>
        )}
        {entry.mem_bytes != null && <span>{formatBytes(entry.mem_bytes)}</span>}
        {entry.age && <span>{entry.age}</span>}
      </div>
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

  const loading = healthLoading || agentLoading;
  const error = healthErr || agentErr;

  const components = health?.components;
  const grouped = useMemo(
    () => (components ? groupComponents(components) : null),
    [components],
  );

  const controllerEntry = health?.registry?.find(
    (r) => r.type === "controller",
  );
  const natsEntry = health?.registry?.find((r) => r.type === "nats");

  return (
    <ContentArea>
      <PageHeader
        title="Dashboard"
        subtitle="Fleet health overview"
        actions={
          health && (
            <>
              <Badge variant={health.status === "ok" ? "ready" : "error"}>
                {health.status}
              </Badge>
              {health.version && (
                <Badge variant="muted">v{health.version}</Badge>
              )}
              {health.uptime && <Badge variant="muted">{health.uptime}</Badge>}
            </>
          )
        }
      />

      {loading && (
        <div className="flex items-center justify-center py-20">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      )}

      {error && <ErrorBanner message={error} />}

      {!loading && !error && health && (
        <>
          {/* Row 1: Summary stats */}
          <div className="mb-4 grid grid-cols-4 gap-3">
            {health.nats && (
              <StatCard
                label="NATS"
                value={health.nats.version}
                detail={health.nats.url}
                truncateDetail
              />
            )}
            {health.jobs && (
              <StatCard
                label="Jobs"
                value={health.jobs.total}
                detail={`${health.jobs.completed} done / ${health.jobs.failed} failed`}
              />
            )}
            {health.agents && (
              <StatCard
                label="Agents"
                value={`${health.agents.ready}/${health.agents.total}`}
                detail="ready"
              />
            )}
            {health.consumers && (
              <StatCard
                label="Consumers"
                value={health.consumers.total}
                detail="JetStream"
              />
            )}
          </div>

          {/* Row 2: Controller + NATS components */}
          {grouped && (
            <div className="mb-4 grid grid-cols-12 gap-4">
              {grouped.controller.length > 0 && (
                <div className="col-span-7">
                  <SectionLabel>Controller</SectionLabel>
                  <Card>
                    <CardContent className="p-0">
                      <RegistryHeader entry={controllerEntry} />
                      {grouped.controller.map(([name, comp], i) => (
                        <ComponentRow
                          key={name}
                          name={name}
                          status={comp.status}
                          address={comp.address}
                          isLast={i === grouped.controller.length - 1}
                        />
                      ))}
                    </CardContent>
                  </Card>
                </div>
              )}
              {grouped.nats.length > 0 && (
                <div className="col-span-5">
                  <SectionLabel>NATS Server</SectionLabel>
                  <Card>
                    <CardContent className="p-0">
                      <RegistryHeader entry={natsEntry} />
                      {grouped.nats.map(([name, comp], i) => (
                        <ComponentRow
                          key={name}
                          name={name}
                          status={comp.status}
                          address={comp.address}
                          isLast={i === grouped.nats.length - 1}
                        />
                      ))}
                    </CardContent>
                  </Card>
                </div>
              )}
            </div>
          )}

          {/* Row 3: Streams + Object Store */}
          {health.streams && health.streams.length > 0 && (
            <div className="mb-4 grid grid-cols-12 gap-4">
              <div className="col-span-7">
                <SectionLabel>Streams ({health.streams.length})</SectionLabel>
                <DataTable
                  columns={[
                    {
                      header: "Name",
                      cell: (s) => <Text variant="mono">{s.name}</Text>,
                    },
                    {
                      header: "Messages",
                      align: "right",
                      cell: (s) => (
                        <Text variant="muted">
                          {s.messages.toLocaleString()}
                        </Text>
                      ),
                    },
                    {
                      header: "Size",
                      align: "right",
                      cell: (s) => (
                        <Text variant="muted">{formatBytes(s.bytes)}</Text>
                      ),
                    },
                    {
                      header: "Consumers",
                      align: "right",
                      cell: (s) => <Text>{s.consumers}</Text>,
                    },
                  ]}
                  rows={health.streams}
                  getRowKey={(s) => s.name}
                />
              </div>

              {health.object_stores && health.object_stores.length > 0 && (
                <div className="col-span-5">
                  <SectionLabel>
                    Object Store ({health.object_stores.length})
                  </SectionLabel>
                  <DataTable
                    columns={[
                      {
                        header: "Bucket",
                        cell: (os) => <Text variant="mono">{os.name}</Text>,
                      },
                      {
                        header: "Size",
                        align: "right",
                        cell: (os) => (
                          <Text variant="muted">{formatBytes(os.size)}</Text>
                        ),
                      },
                    ]}
                    rows={health.object_stores}
                    getRowKey={(os) => os.name}
                  />
                </div>
              )}
            </div>
          )}

          {/* Row 4: KV Stores */}
          {health.kv_buckets && health.kv_buckets.length > 0 && (
            <div className="mb-6">
              <SectionLabel>
                KV Stores ({health.kv_buckets.length})
              </SectionLabel>
              <DataTable
                columns={[
                  {
                    header: "Bucket",
                    cell: (kv) => <Text variant="mono">{kv.name}</Text>,
                  },
                  {
                    header: "Keys",
                    align: "right",
                    cell: (kv) => (
                      <Text variant="muted">{kv.keys.toLocaleString()}</Text>
                    ),
                  },
                  {
                    header: "Size",
                    align: "right",
                    cell: (kv) => (
                      <Text variant="muted">{formatBytes(kv.bytes)}</Text>
                    ),
                  },
                ]}
                rows={health.kv_buckets}
                getRowKey={(kv) => kv.name}
              />
            </div>
          )}

          {/* Agents */}
          <section>
            <SectionLabel icon={Activity}>
              Agents ({agents.length})
            </SectionLabel>
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
              {agents.map((agent) => (
                <AgentCard
                  key={agent.hostname}
                  agent={agent}
                  components={grouped?.agent}
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
