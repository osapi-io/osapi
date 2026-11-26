import { useState } from "react";
import { getAgentDetails } from "@/sdk/gen/agent-management-api-agent-operations/agent-management-api-agent-operations";
import type { AgentInfo, TimelineEvent } from "@/sdk/gen/schemas";
import { unwrap } from "@/hooks/use-polling";
import { formatAgeSince, formatTime } from "@/lib/format";
import { cn } from "@/lib/cn";

// loadTimeline reads the timeline of an agent. Only agent get carries it;
// the list that the dashboard polls does not.
async function loadTimeline(hostname: string): Promise<TimelineEvent[]> {
  const body = unwrap<AgentInfo>(
    await getAgentDetails(hostname),
    "Failed to fetch the agent",
  );
  return body.timeline ?? [];
}

// shown is how many events the card shows before "Show all".
const shown = 8;

// tone is the color of an event's dot: what it means for the agent.
function tone(ev: TimelineEvent): string {
  if (ev.error) return "bg-status-error";
  switch (ev.event) {
    case "started":
    case "admitted":
    case "undrain":
      return "bg-primary";
    case "stopped":
    case "drain":
    case "cordoned":
      return "bg-status-pending";
    case "revoked":
    case "rejected":
    case "deleted":
      return "bg-status-error";
    default:
      return "bg-status-running";
  }
}

// detail is the message without the event's own name at its start, as in
// "started, version v1": the event already says it.
function detail(ev: TimelineEvent): string {
  const text = ev.error || ev.message || "";
  const rest = text.startsWith(ev.event)
    ? text.slice(ev.event.length).replace(/^[,:\s]+/, "")
    : text;
  return rest;
}

// newestFirst orders by time. The timestamps have fractions of different
// lengths, so their text order is not their time order.
function newestFirst(events: TimelineEvent[]): TimelineEvent[] {
  return [...events].sort(
    (a, b) => Date.parse(b.timestamp) - Date.parse(a.timestamp),
  );
}

interface AgentTimelineProps {
  hostname: string;
  // load is the call that reads the timeline. Tests replace it.
  load?: (hostname: string) => Promise<TimelineEvent[]>;
}

/** The events of an agent, newest first, on a rail. It reads them when the reader opens it. */
export function AgentTimeline({
  hostname,
  load = loadTimeline,
}: AgentTimelineProps) {
  const [events, setEvents] = useState<TimelineEvent[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [all, setAll] = useState(false);

  const onToggle = (e: React.SyntheticEvent<HTMLDetailsElement>) => {
    if (!e.currentTarget.open) return;
    setError(null);
    load(hostname)
      .then((ev) => setEvents(newestFirst(ev)))
      .catch((err: Error) => setError(err.message));
  };

  const visible = events && (all ? events : events.slice(0, shown));

  return (
    <details className="mt-2 min-w-0 text-xs" onToggle={onToggle}>
      <summary className="cursor-pointer text-text-muted hover:text-text">
        Timeline
      </summary>
      <div className="mt-1.5">
        {error && <p className="text-status-error">{error}</p>}
        {!error && events === null && (
          <p className="text-text-muted">Loading</p>
        )}
        {events?.length === 0 && <p className="text-text-muted">No events</p>}
        {visible && visible.length > 0 && (
          <ol className="relative ml-1 border-l border-border/60">
            {visible.map((ev, i) => (
              <li
                key={`${ev.timestamp}-${ev.event}-${i}`}
                className="relative min-w-0 pb-1.5 pl-3 last:pb-0"
              >
                <span
                  className={cn(
                    "absolute -left-[4px] top-[5px] h-[7px] w-[7px] rounded-full",
                    tone(ev),
                  )}
                />
                <div className="flex min-w-0 items-baseline gap-2">
                  <span
                    className={ev.error ? "text-status-error" : "text-text"}
                  >
                    {ev.event}
                  </span>
                  <span
                    className="ml-auto shrink-0 text-text-muted tabular-nums"
                    title={formatTime(ev.timestamp)}
                  >
                    {formatAgeSince(ev.timestamp)} ago
                  </span>
                </div>
                {detail(ev) && (
                  <div
                    className="min-w-0 truncate text-text-muted"
                    title={detail(ev)}
                  >
                    {detail(ev)}
                  </div>
                )}
              </li>
            ))}
          </ol>
        )}
        {events && events.length > shown && (
          <button
            type="button"
            className="mt-1 text-text-muted underline-offset-2 hover:text-text hover:underline"
            onClick={() => setAll(!all)}
          >
            {all ? "Show fewer" : `Show all ${events.length}`}
          </button>
        )}
      </div>
    </details>
  );
}
