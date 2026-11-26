import { describe, expect, test } from "bun:test";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { AgentTimeline } from "@/components/domain/agent-timeline";
import type { TimelineEvent } from "@/sdk/gen/schemas";

const events: TimelineEvent[] = [
  { timestamp: "2026-11-25T10:00:00Z", event: "started", message: "v1.2.0" },
  {
    timestamp: "2026-11-25T11:00:00Z",
    event: "config_applied",
    message: "config version 2, labels group=web",
  },
  {
    timestamp: "2026-11-25T09:00:00Z",
    event: "drain",
    error: "the agent did not stop",
  },
];

function open(container: HTMLElement) {
  const details = container.querySelector("details")!;
  details.open = true;
  fireEvent(details, new Event("toggle"));
}

describe("AgentTimeline", () => {
  test("reads nothing until the reader opens it", () => {
    let calls = 0;
    render(
      <AgentTimeline
        hostname="web-01"
        load={async () => {
          calls++;
          return events;
        }}
      />,
    );
    expect(screen.getByText("Timeline")).toBeTruthy();
    expect(calls).toBe(0);
  });

  test("shows the events newest first, with an error in place of a message", async () => {
    let asked = "";
    const { container } = render(
      <AgentTimeline
        hostname="web-01"
        load={async (h) => {
          asked = h;
          return events;
        }}
      />,
    );
    open(container);
    await waitFor(() =>
      expect(screen.getByText("config_applied")).toBeTruthy(),
    );
    expect(asked).toBe("web-01");
    const names = [...container.querySelectorAll("li")].map(
      (row) => row.querySelector("div > span")!.textContent,
    );
    expect(names).toEqual(["config_applied", "started", "drain"]);
    expect(screen.getByText("the agent did not stop")).toBeTruthy();
  });

  test("orders by time, also when the fractions of the timestamps differ in length", async () => {
    const { container } = render(
      <AgentTimeline
        hostname="web-01"
        load={async () => [
          { timestamp: "2026-11-25T17:04:56.1-08:00", event: "undrain" },
          { timestamp: "2026-11-25T12:26:40.123456-08:00", event: "started" },
        ]}
      />,
    );
    open(container);
    await waitFor(() => expect(screen.getByText("undrain")).toBeTruthy());
    const names = [...container.querySelectorAll("li")].map(
      (row) => row.querySelector("div > span")!.textContent,
    );
    expect(names).toEqual(["undrain", "started"]);
  });

  test("shows the newest eight, and all on request", async () => {
    const many = Array.from({ length: 12 }, (_, i) => ({
      timestamp: `2026-11-25T10:${String(i).padStart(2, "0")}:00Z`,
      event: "started",
      message: `started, version v${i}`,
    }));
    const { container } = render(
      <AgentTimeline hostname="web-01" load={async () => many} />,
    );
    open(container);
    await waitFor(() => expect(screen.getByText("Show all 12")).toBeTruthy());
    expect(container.querySelectorAll("li").length).toBe(8);
    expect(screen.getByText("version v11")).toBeTruthy();
    fireEvent.click(screen.getByText("Show all 12"));
    expect(container.querySelectorAll("li").length).toBe(12);
  });

  test("an agent with no events says so", async () => {
    const { container } = render(
      <AgentTimeline hostname="web-01" load={async () => []} />,
    );
    open(container);
    await waitFor(() => expect(screen.getByText("No events")).toBeTruthy());
  });

  test("a failed read shows its error", async () => {
    const { container } = render(
      <AgentTimeline
        hostname="web-01"
        load={async () => {
          throw new Error("agent not found: web-01");
        }}
      />,
    );
    open(container);
    await waitFor(() =>
      expect(screen.getByText("agent not found: web-01")).toBeTruthy(),
    );
  });
});
