import { useState } from "react";
import { getAgentDetails } from "@/sdk/gen/agent-management-api-agent-operations/agent-management-api-agent-operations";
import type { AgentInfo, TimelineEvent } from "@/sdk/gen/schemas";
import { unwrap } from "@/hooks/use-polling";
import { formatTime } from "@/lib/format";

// loadTimeline reads the timeline of an agent. Only agent get carries it;
// the list that the dashboard polls does not.
async function loadTimeline(hostname: string): Promise<TimelineEvent[]> {
  const body = unwrap<AgentInfo>(
    await getAgentDetails(hostname),
    "Failed to fetch the agent",
  );
  return body.timeline ?? [];
}

interface AgentTimelineProps {
  hostname: string;
  // load is the call that reads the timeline. Tests replace it.
  load?: (hostname: string) => Promise<TimelineEvent[]>;
}

/** The events of an agent, newest first, as agent get prints them. It reads them when the reader opens it. */
export function AgentTimeline({
  hostname,
  load = loadTimeline,
}: AgentTimelineProps) {
  const [events, setEvents] = useState<TimelineEvent[] | null>(null);
  const [error, setError] = useState<string | null>(null);

  const onToggle = (e: React.SyntheticEvent<HTMLDetailsElement>) => {
    if (!e.currentTarget.open) return;
    setError(null);
    load(hostname)
      .then((ev) =>
        setEvents(
          [...ev].sort((a, b) => b.timestamp.localeCompare(a.timestamp)),
        ),
      )
      .catch((err: Error) => setError(err.message));
  };

  return (
    <details className="mt-2 text-xs" onToggle={onToggle}>
      <summary className="cursor-pointer text-text-muted hover:text-text">
        Timeline
      </summary>
      <div className="mt-1 space-y-0.5">
        {error && <p className="text-status-error">{error}</p>}
        {!error && events === null && (
          <p className="text-text-muted">Loading</p>
        )}
        {events?.length === 0 && <p className="text-text-muted">No events</p>}
        {events?.map((ev) => (
          <div
            key={`${ev.timestamp}-${ev.event}`}
            className="grid grid-cols-[auto_auto_1fr] gap-x-2"
          >
            <span className="text-text-muted tabular-nums">
              {formatTime(ev.timestamp)}
            </span>
            <span className={ev.error ? "text-status-error" : "text-text"}>
              {ev.event}
            </span>
            <span
              className="min-w-0 truncate text-text-muted"
              title={ev.error || ev.message}
            >
              {ev.error || ev.message}
            </span>
          </div>
        ))}
      </div>
    </details>
  );
}
