import { useEffect, useState, type ReactNode } from "react";
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
import type { AgentInfo, ComponentEntry } from "@/sdk/gen/schemas";
import {
  drainAgent,
  undrainAgent,
} from "@/sdk/gen/agent-management-api-agent-operations/agent-management-api-agent-operations";
import { statusDot, statusVariant } from "@/lib/status";

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
  // A drain or undrain the controller accepted, until the agent's state
  // shows it. The agent acts on it at its next heartbeat, which can be 20s
  // away, so the card shows the request at once and polls faster till then.
  const [pending, setPending] = useState<"drain" | "undrain" | null>(null);
  const [acting, setActing] = useState(false);
  const [note, setNote] = useState<{ text: string; error?: boolean } | null>(
    null,
  );

  const isPending = agent.state === "Pending";
  const isDrained = agent.state === "Draining" || agent.state === "Cordoned";
  const canWrite = canCall("drainAgent");
  const badgeVariant = statusVariant(agent.state);
  // The API lists only the conditions that hold.
  const activeConditions = agent.conditions ?? [];

  // The request is done when the state the agent reports matches it: a
  // drain when the agent is Cordoned, its jobs stopped, not when it is
  // only Draining.
  const caughtUp =
    (pending === "drain" && agent.state === "Cordoned") ||
    (pending === "undrain" && !isDrained);

  useEffect(() => {
    if (!pending || caughtUp) return;
    const poll = setInterval(() => onRefresh?.(), 2000);
    const giveUp = setTimeout(() => {
      setNote({
        text:
          pending === "drain" && agent.state === "Draining"
            ? `${agent.hostname} is still draining: its jobs have not stopped.`
            : `${agent.hostname} has not picked up the ${pending} yet; it applies at the agent's next heartbeat.`,
        error: true,
      });
      setPending(null);
    }, 60000);
    return () => {
      clearInterval(poll);
      clearTimeout(giveUp);
    };
  }, [pending, caughtUp, onRefresh, agent.hostname, agent.state]);

  // Once the state matches, the request is done.
  if (pending && caughtUp) {
    setPending(null);
    setNote(null);
  }

  const act = async (kind: "drain" | "undrain") => {
    setActing(true);
    setNote(null);
    try {
      const res =
        kind === "drain"
          ? await drainAgent(agent.hostname)
          : await undrainAgent(agent.hostname);
      if (res.status >= 300) {
        const body = res.data as { error?: string } | undefined;
        throw new Error(body?.error ?? `${kind} failed: ${res.status}`);
      }
      setPending(kind);
      setNote({
        text:
          kind === "drain"
            ? `Drain accepted. ${agent.hostname} takes no new jobs from its next heartbeat.`
            : `Undrain accepted. ${agent.hostname} takes jobs again from its next heartbeat.`,
      });
      onRefresh?.();
    } catch (e) {
      setNote({
        text: e instanceof Error ? e.message : `${kind} failed`,
        error: true,
      });
    }
    setActing(false);
  };

  // What the badge shows: the request while the agent has not acted on it,
  // else the state the agent reports.
  const shownState =
    pending === "drain"
      ? "Draining"
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
          {shownState && shownState !== "Ready" ? (
            <Badge variant={pending ? "pending" : statusVariant(shownState)}>
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
                <HealthDot status={statusDot(comp.status)} />
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
