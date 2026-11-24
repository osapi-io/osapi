import { describe, expect, test } from "bun:test";
import { hostOutcome, hostOutcomes } from "@/lib/outcome";
import { RowStatus } from "@/sdk/gen/schemas";

describe("hostOutcome", () => {
  test("a host that did not answer is a timeout, not a failure", () => {
    expect(
      hostOutcome({
        status: RowStatus.timeout,
        error: "timeout: agent did not respond",
      }),
    ).toBe(RowStatus.timeout);
  });

  test("a skipped host is skipped, though it carries an error", () => {
    expect(
      hostOutcome({ status: RowStatus.skipped, error: "unsupported" }),
    ).toBe(RowStatus.skipped);
  });

  test("a host with an error is failed", () => {
    expect(hostOutcome({ status: RowStatus.failed, error: "boom" })).toBe(
      RowStatus.failed,
    );
  });
});

describe("hostOutcomes", () => {
  test("a host counts once, as its worst entry", () => {
    expect(
      hostOutcomes([
        { hostname: "a", status: RowStatus.ok },
        { hostname: "a", status: RowStatus.timeout },
        { hostname: "b", status: RowStatus.ok, changed: true },
      ]),
    ).toEqual([RowStatus.timeout, "changed"]);
  });
});
