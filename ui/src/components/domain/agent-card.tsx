import type { ReactNode } from "react";
import { useDrain } from "@/hooks/use-drain";
import { AgentState } from "@/sdk/gen/schemas";
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
import { AgentTimeline } from "@/components/domain/agent-timeline";
import { useAuth } from "@/hooks/use-auth";
import { formatAgeSince, formatBytes } from "@/lib/format";
import { Monitor, Clock, Cpu, Pause, Play, Loader2 } from "lucide-react";
import type { AgentInfo, ComponentEntry } from "@/sdk/gen/schemas";
import { useStatus } from "@/hooks/use-status";

interface AgentCardProps {
  agent: AgentInfo;
  // components are the agent's sub-components, from its entry in the
  // health registry.
  components?: ComponentEntry[];
  // controllerVersion is the controller's build. The card shows the agent's
  // version only when it differs, since a mismatch is what a reader needs.
  controllerVersion?: string;
  onRefresh?: () => void;
}

function NotReported() {
  return <span className="text-text-muted">-</span>;
}

export function AgentCard({
  agent,
  components,
  controllerVersion,
  onRefresh,
}: AgentCardProps) {
  const { canCall } = useAuth();
  const statusColor = useStatus();
  const isPending = agent.state === AgentState.Pending;
  const isDrained =
    agent.state === AgentState.Draining || agent.state === AgentState.Cordoned;
  const canWrite = canCall("DrainAgent");
  const badgeVariant = statusColor.variant(agent.state);
  // The API lists only the conditions that hold.
  const activeConditions = agent.conditions ?? [];

  const { pending, acting, note, act } = useDrain(agent, onRefresh);

  // What the badge shows: the request while the agent has not acted on it,
  // else the state the agent reports.
  const shownState =
    pending === "drain"
      ? AgentState.Draining
      : pending === "undrain"
        ? "Undraining"
        : agent.state;

  return (
    <Card
      variant={activeConditions.length > 0 ? "warn" : "active"}
      className="flex flex-col"
    >
      <CardHeader>
        <Monitor className="h-4 w-4 text-text-muted" />
        <CardTitle className="truncate" title={agent.hostname}>
          {agent.hostname}
        </CardTitle>
        <div className="ml-auto flex items-center gap-1.5">
          {shownState && shownState !== AgentState.Ready ? (
            <Badge
              variant={pending ? "pending" : statusColor.variant(shownState)}
            >
              {shownState}
            </Badge>
          ) : (
            <Badge variant={badgeVariant}>{agent.state}</Badge>
          )}
        </div>
      </CardHeader>
      <CardContent className="flex flex-1 flex-col">
        {isPending && (
          <Text variant="muted" as="p" className="mt-1 italic">
            Awaiting admission
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
              <span title={agent.uptime && `system up ${agent.uptime}`}>
                up {formatAgeSince(agent.started_at)}
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
            // The config that the controller holds for the agent, and
            // whether the agent applied it yet.
            ...(agent.config_wanted
              ? [
                  [
                    "Config",
                    <span
                      key="c"
                      className={
                        agent.config_behind ? "text-status-pending" : undefined
                      }
                    >
                      v{agent.config_version ?? 0}
                      {agent.config_behind &&
                        `, applying v${agent.config_wanted}`}
                    </span>,
                  ] as [string, ReactNode],
                ]
              : []),
            // The version only when it is not the controller's: a
            // mismatch is the case that matters.
            ...(agent.version &&
            controllerVersion &&
            agent.version !== controllerVersion
              ? [
                  [
                    "Version",
                    <span key="v" className="text-status-pending">
                      {agent.version}
                    </span>,
                  ] as [string, ReactNode],
                ]
              : []),
          ]}
        />

        {components && components.length > 0 && (
          <div className="mt-2 flex items-center gap-3">
            {components.map((comp) => (
              <div key={comp.type} className="flex items-center gap-1 text-xs">
                <HealthDot status={statusColor.dot(comp.status)} />
                <span className="text-text-muted">{comp.type}</span>
              </div>
            ))}
          </div>
        )}

        {/* The identity, which nobody scans, behind a disclosure. */}
        {(agent.machine_id || agent.fingerprint) && (
          <details className="mt-2 text-xs">
            <summary className="cursor-pointer text-text-muted hover:text-text">
              Identity
            </summary>
            <div className="mt-1 space-y-0.5">
              {agent.machine_id && (
                <CopyField label="Machine ID" value={agent.machine_id} />
              )}
              {agent.fingerprint && (
                <CopyField label="Fingerprint" value={agent.fingerprint} />
              )}
            </div>
          </details>
        )}

        {!isPending && <AgentTimeline hostname={agent.hostname} />}

        <div className="mt-auto pt-2">
          {/* Drain / Undrain — hidden when pending */}
          {canWrite && !isPending && (
            <div className="mt-2 border-t border-border/30 pt-2">
              {(pending ? pending === "drain" : isDrained) ? (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => act("undrain")}
                  disabled={acting || pending !== null}
                >
                  {acting || pending ? (
                    <Loader2 className="h-3 w-3 animate-spin" />
                  ) : (
                    <Play className="h-3 w-3" />
                  )}
                  {pending === "drain" ? "Draining" : "Undrain"}
                </Button>
              ) : (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => act("drain")}
                  disabled={acting || pending !== null}
                >
                  {acting || pending ? (
                    <Loader2 className="h-3 w-3 animate-spin" />
                  ) : (
                    <Pause className="h-3 w-3" />
                  )}
                  {pending === "undrain" ? "Undraining" : "Drain"}
                </Button>
              )}
              {note && (
                <Text
                  variant={note.error ? "warn" : "muted"}
                  as="p"
                  className="mt-1 text-xs"
                >
                  {note.text}
                </Text>
              )}
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
