import { useState } from "react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { HealthDot } from "@/components/ui/health-dot";
import { Text } from "@/components/ui/text";
import { LabelTag } from "@/components/ui/label-tag";
import { ConditionAlert } from "@/components/ui/condition-alert";
import { CopyField } from "@/components/ui/copy-field";
import { MetricValue } from "@/components/ui/metric-value";
import { KVList } from "@/components/ui/kv-list";
import { useAuth } from "@/hooks/use-auth";
import { formatAgeSince, formatBytes } from "@/lib/format";
import { Monitor, Clock, Cpu, Pause, Play, Loader2 } from "lucide-react";
import type { AgentInfo, ComponentHealth } from "@/sdk/gen/schemas";
import {
  drainAgent,
  undrainAgent,
} from "@/sdk/gen/agent-management-api-agent-operations/agent-management-api-agent-operations";
import { statusDot, statusVariant } from "@/lib/status";

interface AgentCardProps {
  agent: AgentInfo;
  components?: [string, ComponentHealth][];
  onRefresh?: () => void;
}

function NotReported() {
  return <span className="text-text-muted">-</span>;
}

export function AgentCard({ agent, components, onRefresh }: AgentCardProps) {
  const { can } = useAuth();
  const [acting, setActing] = useState(false);

  const isReady = agent.status === "Ready";
  const isPending = agent.state === "Pending";
  const isDrained = agent.state === "Draining" || agent.state === "Cordoned";
  const canWrite = can("agent:write");
  const variant = isPending ? "active" : isReady ? "active" : "error";
  const badgeVariant = statusVariant(isPending ? agent.state : agent.status);
  const activeConditions = agent.conditions?.filter((c) => c.status) ?? [];

  const handleDrain = async () => {
    setActing(true);
    try {
      await drainAgent(agent.hostname);
      onRefresh?.();
    } catch {
      // silent — next poll will show state
    }
    setActing(false);
  };

  const handleUndrain = async () => {
    setActing(true);
    try {
      await undrainAgent(agent.hostname);
      onRefresh?.();
    } catch {
      // silent
    }
    setActing(false);
  };

  return (
    <Card
      variant={activeConditions.length > 0 ? "warn" : variant}
      className="flex flex-col"
    >
      <CardHeader>
        <Monitor className="h-4 w-4 text-tint" />
        <CardTitle className="truncate" title={agent.hostname}>
          {agent.hostname}
        </CardTitle>
        <div className="ml-auto flex items-center gap-1.5">
          {agent.state && agent.state !== "Ready" ? (
            <Badge variant={statusVariant(agent.state)}>{agent.state}</Badge>
          ) : (
            <Badge variant={badgeVariant}>{agent.status}</Badge>
          )}
        </div>
      </CardHeader>
      <CardContent className="flex flex-1 flex-col">
        {/* In the order of osapi client agent list: machine ID,
            conditions, labels, age, load, OS, version. */}
        {(agent.machine_id || agent.fingerprint) && (
          <div className="mt-1 space-y-0.5">
            {agent.machine_id && (
              <CopyField label="Machine ID" value={agent.machine_id} />
            )}
            {agent.fingerprint && (
              <CopyField label="Fingerprint" value={agent.fingerprint} />
            )}
          </div>
        )}

        {isPending && (
          <Text variant="muted" as="p" className="mt-1 italic">
            Awaiting PKI enrollment
          </Text>
        )}

        {activeConditions.length > 0 && (
          <div className="mt-2 space-y-1">
            {activeConditions.map((c) => (
              <ConditionAlert key={c.type} type={c.type} />
            ))}
          </div>
        )}

        {agent.labels && Object.keys(agent.labels).length > 0 && (
          <div className="mt-2 flex flex-wrap gap-1">
            {Object.entries(agent.labels).map(([k, v]) => (
              <LabelTag key={k}>
                {k}:{v}
              </LabelTag>
            ))}
          </div>
        )}

        {/* A value that the agent does not report is a dash. */}
        <div className="mt-2 grid grid-cols-2 gap-2">
          <MetricValue icon={Clock}>
            {agent.started_at ? (
              <span title={agent.uptime}>
                {formatAgeSince(agent.started_at)}
              </span>
            ) : (
              <NotReported />
            )}
          </MetricValue>
          <MetricValue icon={Cpu}>
            {agent.load_average ? (
              agent.load_average["1min"].toFixed(2)
            ) : (
              <NotReported />
            )}
          </MetricValue>
        </div>

        <KVList
          className="mt-1.5"
          pairs={[
            [
              "Memory",
              agent.memory &&
                [
                  `${formatBytes(agent.memory.total)} total`,
                  agent.memory.used > 0 &&
                    `${formatBytes(agent.memory.used)} used`,
                  `${formatBytes(agent.memory.free)} free`,
                ]
                  .filter(Boolean)
                  .join(", "),
            ],
            [
              "OS",
              agent.os_info && (
                <>
                  {agent.os_info.distribution}{" "}
                  <span className="text-text-muted">
                    {agent.os_info.version}
                  </span>
                </>
              ),
            ],
            ["Version", agent.version],
          ]}
        />

        {components && components.length > 0 && (
          <div className="mt-2 flex items-center gap-3">
            {components.map(([name, comp]) => (
              <div key={name} className="flex items-center gap-1 text-xs">
                <HealthDot status={statusDot(comp.status)} />
                <span className="text-text-muted">{name}</span>
              </div>
            ))}
          </div>
        )}

        <div className="mt-auto pt-2">
          {/* Drain / Undrain — hidden when pending */}
          {canWrite && !isPending && (
            <div className="mt-2 border-t border-border/30 pt-2">
              {isDrained ? (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={handleUndrain}
                  disabled={acting}
                >
                  {acting ? (
                    <Loader2 className="h-3 w-3 animate-spin" />
                  ) : (
                    <Play className="h-3 w-3" />
                  )}
                  Undrain
                </Button>
              ) : (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={handleDrain}
                  disabled={acting}
                >
                  {acting ? (
                    <Loader2 className="h-3 w-3 animate-spin" />
                  ) : (
                    <Pause className="h-3 w-3" />
                  )}
                  Drain
                </Button>
              )}
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
