import { describe, expect, test } from "bun:test";
import type { JobDetailResponse } from "@/sdk/gen/schemas";
import { hostSteps, stepRows } from "@/lib/steps";

// A dns update on two hosts: web-01 changed netplan and its file, and
// web-02 failed in managedfile. db-01 answered with no steps.
const job: JobDetailResponse = {
  id: "j-1",
  status: "partial_failure",
  chain: [
    "web-01: netplan update (changed) → managedfile update (changed)",
    "web-02: netplan update → managedfile update (failed)",
  ],
  responses: [
    { hostname: "db-01", status: "completed" },
    {
      hostname: "web-01",
      status: "completed",
      steps: [
        {
          behavior: "netplan",
          method: "update",
          changed: true,
          duration_ms: 40,
          steps: [
            {
              behavior: "managedfile",
              method: "update",
              changed: true,
              duration_ms: 12,
            },
          ],
        },
      ],
    },
    {
      hostname: "web-02",
      status: "failed",
      steps: [
        {
          behavior: "netplan",
          method: "update",
          duration_ms: 30,
          steps: [
            {
              behavior: "managedfile",
              method: "update",
              error: "permission denied",
              duration_ms: 5,
            },
          ],
        },
      ],
    },
  ],
};

describe("hostSteps", () => {
  test("has each host that answered with steps, in order", () => {
    expect(hostSteps(job).map((h) => h.hostname)).toEqual(["web-01", "web-02"]);
  });

  test("is empty for a job with no responses", () => {
    expect(hostSteps({ id: "j-2" })).toEqual([]);
  });
});

describe("stepRows", () => {
  test("puts each step under the step that called it", () => {
    const rows = stepRows(hostSteps(job)[0].steps);
    expect(
      rows.map((r) => [r.depth, `${r.step.behavior}.${r.step.method}`]),
    ).toEqual([
      [0, "netplan.update"],
      [1, "managedfile.update"],
    ]);
  });

  test("keeps the error of a nested step that failed", () => {
    const rows = stepRows(hostSteps(job)[1].steps);
    expect(rows[1].step.error).toBe("permission denied");
    expect(rows[1].depth).toBe(1);
  });

  test("goes as deep as the calls go", () => {
    const rows = stepRows([
      {
        behavior: "a",
        method: "m",
        duration_ms: 1,
        steps: [
          {
            behavior: "b",
            method: "m",
            duration_ms: 1,
            steps: [{ behavior: "c", method: "m", duration_ms: 1 }],
          },
          { behavior: "d", method: "m", duration_ms: 1 },
        ],
      },
    ]);
    expect(rows.map((r) => [r.depth, r.step.behavior])).toEqual([
      [0, "a"],
      [1, "b"],
      [2, "c"],
      [1, "d"],
    ]);
  });
});
