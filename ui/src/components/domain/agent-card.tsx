import { useState } from "react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { HealthDot } from "@/components/ui/health-dot";
import { Text } from "@/components/ui/text";
import { LabelTag } from "@/components/ui/label-tag";
import { ConditionAlert } from "@/components/ui/condition-alert";
import { MetricValue } from "@/components/ui/metric-value";
import { useAuth } from "@/hooks/use-auth";
import {
  Monitor,
  Clock,
  Cpu,
  HardDrive,
  Pause,
  Play,
  Loader2,
} from "lucide-react";
import type { AgentInfo, ComponentHealth } from "@/sdk/gen/schemas";
import {
  drainAgent,
  undrainAgent,
} from "@/sdk/gen/agent-management-api-agent-operations/agent-management-api-agent-operations";

interface AgentCardProps {
  agent: AgentInfo;
  components?: [string, ComponentHealth][];
  onRefresh?: () => void;
}

function formatBytes(bytes: number) {
  if (bytes < 1024 * 1024 * 1024)
    return `${(bytes / (1024 * 1024)).toFixed(0)} MB`;
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`;
}

function isHealthy(status: string) {
  return status === "ok" || status === "healthy" || status === "ready";
}

function stateVariant(state?: string) {
  switch (state) {
    case "Draining":
      return "running" as const;
    case "Cordoned":
      return "error" as const;
    default:
      return undefined;
  }
}

export function AgentCard({ agent, components, onRefresh }: AgentCardProps) {
  const { can } = useAuth();
  const [acting, setActing] = useState(false);

  const isReady = agent.status === "Ready";
  const isDrained = agent.state === "Draining" || agent.state === "Cordoned";
  const canWrite = can("agent:write");
  const variant = isReady ? "active" : "error";
  const badgeVariant = isReady ? "ready" : "error";
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
      variant={activeConditions.length > 0 ? "error" : variant}
      className="flex flex-col"
    >
      <CardHeader>
        <Monitor className="h-4 w-4 text-primary" />
        <CardTitle className="truncate" title={agent.hostname}>
          {agent.hostname}
        </CardTitle>
        <div className="ml-auto flex items-center gap-1.5">
          {agent.state && agent.state !== "Ready" && (
            <Badge variant={stateVariant(agent.state)}>{agent.state}</Badge>
          )}
          <Badge variant={badgeVariant}>{agent.status}</Badge>
        </div>
      </CardHeader>
      <CardContent className="flex flex-1 flex-col">
        {/* Stats row */}
        <div className="grid grid-cols-3 gap-2">
          {agent.load_average && (
            <MetricValue icon={Cpu}>
              {agent.load_average["1min"].toFixed(2)}
            </MetricValue>
          )}
          {agent.memory && (
            <MetricValue icon={HardDrive}>
              {formatBytes(agent.memory.used)}/{formatBytes(agent.memory.total)}
            </MetricValue>
          )}
          {agent.uptime && (
            <MetricValue icon={Clock}>{agent.uptime}</MetricValue>
          )}
        </div>

        {/* OS + arch */}
        {agent.os_info && (
          <Text variant="muted" as="p" className="mt-1.5">
            {agent.os_info.distribution} {agent.os_info.version}
            {agent.architecture ? ` / ${agent.architecture}` : ""}
            {agent.cpu_count ? ` / ${agent.cpu_count} cpu` : ""}
          </Text>
        )}

        {/* Agent components */}
        {components && components.length > 0 && (
          <div className="mt-2 flex items-center gap-3">
            {components.map(([name, comp]) => (
              <div key={name} className="flex items-center gap-1 text-xs">
                <HealthDot status={isHealthy(comp.status)} />
                <span className="text-text-muted">{name}</span>
              </div>
            ))}
          </div>
        )}

        {/* Bottom section — pinned to bottom of card */}
        <div className="mt-auto pt-2">
          {/* Conditions */}
          {activeConditions.length > 0 && (
            <div className="space-y-1">
              {activeConditions.map((c) => (
                <ConditionAlert key={c.type} type={c.type} />
              ))}
            </div>
          )}

          {/* Labels */}
          {agent.labels && Object.keys(agent.labels).length > 0 && (
            <div className="mt-2 flex flex-wrap gap-1">
              {Object.entries(agent.labels).map(([k, v]) => (
                <LabelTag key={k}>
                  {k}:{v}
                </LabelTag>
              ))}
            </div>
          )}

          {/* Drain / Undrain */}
          {canWrite && (
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
