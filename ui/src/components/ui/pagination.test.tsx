import { describe, expect, mock, test } from "bun:test";
import { fireEvent, render, screen } from "@testing-library/react";
import { Pagination } from "@/components/ui/pagination";

describe("Pagination", () => {
  test("one page shows nothing", () => {
    const { container } = render(
      <Pagination offset={0} limit={20} total={5} onChange={() => {}} />,
    );
    expect(container.textContent).toBe("");
  });

  test("Next moves one page on, and Previous is off on the first page", () => {
    const onChange = mock(() => {});
    render(<Pagination offset={0} limit={20} total={45} onChange={onChange} />);
    expect(screen.getByText("Page 1 of 3")).toBeTruthy();
    expect(
      (screen.getByText("Previous").closest("button") as HTMLButtonElement)
        .disabled,
    ).toBe(true);
    fireEvent.click(screen.getByText("Next"));
    expect(onChange).toHaveBeenCalledWith(20);
  });

  test("Next is off on the last page", () => {
    render(
      <Pagination offset={40} limit={20} total={45} onChange={() => {}} />,
    );
    expect(
      (screen.getByText("Next").closest("button") as HTMLButtonElement)
        .disabled,
    ).toBe(true);
  });
});
