import { describe, expect, mock, test } from "bun:test";
import { fireEvent, render, screen } from "@testing-library/react";
import { DeadLetters } from "@/components/domain/dead-letters";
import type { DeadLetter } from "@/sdk/gen/schemas";

const withJob: DeadLetter = {
  seq: 654,
  time: "2026-11-26T10:00:00Z",
  job_id: "550e8400-e29b-41d4-a716-446655440000",
  operation: { type: "node.hostname.get" },
  target: "_all",
  error: "agent did not answer",
  deliveries: 5,
  consumer: "jobs-agent",
};
const gone: DeadLetter = {
  seq: 653,
  time: "2026-11-26T09:00:00Z",
  deliveries: 5,
  consumer: "jobs-agent",
};

describe("DeadLetters", () => {
  test("no dead letters says so", () => {
    render(<DeadLetters items={[]} total={0} busy={false} />);
    expect(screen.getByText(/No dead letters/)).toBeTruthy();
  });

  test("each dead letter shows its job, and one whose job is there retries", () => {
    const onRetry = mock(() => {});
    render(
      <DeadLetters
        items={[withJob, gone]}
        total={654}
        busy={false}
        onRetry={onRetry}
      />,
    );
    expect(screen.getByText(/2 of 654 dead letters/)).toBeTruthy();
    expect(screen.getByText("node.hostname.get")).toBeTruthy();
    expect(screen.getByText("agent did not answer")).toBeTruthy();
    expect(screen.getByText("gone")).toBeTruthy();
    const retries = screen.getAllByTitle("Retry job");
    expect(retries.length).toBe(1);
    fireEvent.click(retries[0]);
    expect(onRetry).toHaveBeenCalledWith(withJob.job_id);
  });

  test("without retry, there is no action", () => {
    render(<DeadLetters items={[withJob]} total={1} busy={false} />);
    expect(screen.queryByTitle("Retry job")).toBeNull();
    expect(screen.getByText(/1 of 1 dead letter,/)).toBeTruthy();
  });
});
