import { describe, expect, test } from "bun:test";
import { callRows, flatCalls } from "@/lib/calls";

const calls = [
  { behavior: "netplan", method: "get" },
  {
    behavior: "netplan",
    method: "update",
    calls: [{ behavior: "managedfile", method: "update" }],
  },
];

describe("callRows", () => {
  test("puts each call after the call that made it, one level deeper", () => {
    expect(
      callRows(calls).map((r) => [r.depth, r.call.behavior, r.call.method]),
    ).toEqual([
      [0, "netplan", "get"],
      [0, "netplan", "update"],
      [1, "managedfile", "update"],
    ]);
  });
});

describe("flatCalls", () => {
  test("is each call once, in the order that they run", () => {
    expect(
      flatCalls([...calls, { behavior: "netplan", method: "get" }]),
    ).toEqual(["netplan.get", "netplan.update", "managedfile.update"]);
  });
});
