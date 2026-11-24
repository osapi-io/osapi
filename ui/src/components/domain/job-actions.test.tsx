import { describe, expect, mock, test } from "bun:test";
import { fireEvent, render, screen } from "@testing-library/react";
import { JobActions } from "@/components/domain/job-actions";

describe("JobActions", () => {
  test("a job that did not fail has no actions", () => {
    const { container } = render(
      <JobActions
        job={{ id: "j1", status: "completed" }}
        busy={false}
        onRetry={() => {}}
        onDelete={() => {}}
      />,
    );
    expect(container.textContent).toBe("");
    expect(container.querySelector("button")).toBeNull();
  });

  test("a failed job can be retried or deleted", () => {
    const onRetry = mock(() => {});
    const onDelete = mock(() => {});
    render(
      <JobActions
        job={{ id: "j1", status: "failed" }}
        busy={false}
        onRetry={onRetry}
        onDelete={onDelete}
      />,
    );
    fireEvent.click(screen.getByTitle("Retry job"));
    fireEvent.click(screen.getByTitle("Delete job"));
    expect(onRetry).toHaveBeenCalledWith("j1");
    expect(onDelete).toHaveBeenCalledWith("j1");
  });
});
