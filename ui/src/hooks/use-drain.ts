import { useEffect, useState } from "react";
import {
  drainAgent,
  undrainAgent,
} from "@/sdk/gen/agent-management-api-agent-operations/agent-management-api-agent-operations";
import { AgentInfoState, type AgentInfo } from "@/sdk/gen/schemas";

export type DrainRequest = "drain" | "undrain";

export interface DrainNote {
  text: string;
  error?: boolean;
}

/**
 * The drain and undrain of one agent: the request, until the state that the
 * agent reports shows it, and a note for the person. onRefresh reads the
 * agent again.
 */
export function useDrain(agent: AgentInfo, onRefresh?: () => void) {
  // A drain or undrain the controller accepted, until the agent's state
  // shows it. The agent acts on it at its next heartbeat, which can be 20s
  // away, so the card shows the request at once and polls faster till then.
  const [pending, setPending] = useState<DrainRequest | null>(null);
  const [acting, setActing] = useState(false);
  const [note, setNote] = useState<DrainNote | null>(null);

  // The request is done when the state the agent reports matches it: a
  // drain when the agent is Cordoned, its jobs stopped, not when it is
  // only Draining.
  const caughtUp =
    (pending === "drain" && agent.state === AgentInfoState.Cordoned) ||
    (pending === "undrain" &&
      agent.state !== AgentInfoState.Draining &&
      agent.state !== AgentInfoState.Cordoned);

  useEffect(() => {
    if (!pending || caughtUp) return;
    const poll = setInterval(() => onRefresh?.(), 2000);
    const giveUp = setTimeout(() => {
      setNote({
        text:
          pending === "drain" && agent.state === AgentInfoState.Draining
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

  const act = async (kind: DrainRequest) => {
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

  return { pending, acting, note, act };
}
