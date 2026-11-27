import { describe, expect, test } from "bun:test";
import { render, screen } from "@testing-library/react";
import { FIPSLine } from "@/components/domain/registry-header";

describe("FIPSLine", () => {
  test("an approved process is muted, with no title", () => {
    render(<FIPSLine fips="on · approved" />);
    const line = screen.getByText("FIPS on · approved");
    expect(line.className).toContain("text-text-muted");
    expect(line.getAttribute("title")).toBeNull();
  });

  test("a process with what is not approved is pending, with each gap in its title", () => {
    render(
      <FIPSLine
        fips="off · 1 algorithm not approved"
        gaps={["mode off", "md5"]}
      />,
    );
    const line = screen.getByText("FIPS off · 1 algorithm not approved");
    expect(line.className).toContain("text-status-pending");
    expect(line.getAttribute("title")).toBe("mode off\nmd5");
  });
});
