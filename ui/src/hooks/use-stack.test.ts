import { describe, expect, test } from "bun:test";
import {
  ALL_BLOCK_TYPES,
  BLOCK_AREAS,
  blockRunnable,
  newBlock,
  type Block,
} from "@/hooks/use-stack";
import { AREAS } from "@/lib/blocks.gen";
import { blockKind } from "@/lib/permissions";

const blockType = (type: string) => {
  const bt = ALL_BLOCK_TYPES.find((b) => b.type === type);
  if (!bt) throw new Error(`no block type ${type}`);
  return bt;
};

// Apply is enabled when a block can run, as useStack's canApply is.
const canApply = (blocks: Block[]) => blocks.some(blockRunnable);

describe("a new block", () => {
  test("a write starts with no target, and Apply is disabled", () => {
    // power reboot takes no input, so it is ready when it is added.
    const b = newBlock(blockType("power-reboot"), "b-1");
    expect(b.status).toBe("ready");
    expect(b.target).toBe("");
    expect(canApply([b])).toBe(false);
  });

  test("a write with a target picked can apply", () => {
    const b = newBlock(blockType("power-reboot"), "b-1");
    expect(canApply([{ ...b, target: "web-01" }])).toBe(true);
  });

  test("a delete and an action start with no target", () => {
    expect(newBlock(blockType("user-delete"), "b-1").target).toBe("");
    expect(newBlock(blockType("service-restart"), "b-2").target).toBe("");
  });

  test("a read starts with _all and can apply", () => {
    const b = newBlock(blockType("status-get"), "b-1");
    expect(b.target).toBe("_all");
    expect(canApply([b])).toBe(true);
  });

  test("a block that runs on the controller keeps _all", () => {
    expect(newBlock(blockType("object-upload"), "b-1").target).toBe("_all");
  });
});

describe("blockKind", () => {
  test("is the verb of the block's permission", () => {
    expect(blockKind("dns-get")).toBe("read");
    expect(blockKind("dns-update")).toBe("write");
    expect(blockKind("power-reboot")).toBe("action");
    expect(blockKind("command-shell")).toBe("action");
    expect(blockKind("docker-exec")).toBe("action");
    expect(blockKind("object-upload")).toBe("write");
    expect(blockKind("no-such-block")).toBeUndefined();
  });

  test("each block in the panel has one", () => {
    const missing = ALL_BLOCK_TYPES.filter(
      (bt) => blockKind(bt.type) === undefined,
    ).map((bt) => bt.type);
    expect(missing).toEqual([]);
  });
});

describe("the areas of the panel", () => {
  test("are the generated AREAS that have a block, in their order", () => {
    const withBlocks = AREAS.filter((a) =>
      ALL_BLOCK_TYPES.some((b) => b.area === a.name),
    );
    expect(BLOCK_AREAS.map((a) => [a.name, a.label])).toEqual(
      withBlocks.map((a) => [a.name, a.label]),
    );
  });

  test("leave out an area with no block, such as fleet and audit", () => {
    expect(BLOCK_AREAS.some((a) => a.name === "fleet")).toBe(false);
    expect(BLOCK_AREAS.some((a) => a.name === "audit")).toBe(false);
    expect(BLOCK_AREAS.find((a) => a.name === "objectstore")?.label).toBe(
      "Object Store",
    );
    expect(BLOCK_AREAS.find((a) => a.name === "logs")?.label).toBe("Logs");
  });
});

describe("the categories of an area", () => {
  const labels = (area: string) =>
    BLOCK_AREAS.find((a) => a.name === area)?.categories.map((c) => c.label);

  test("are labeled by the behavior's title", () => {
    expect(labels("software")).toContain("Package updates");
    expect(labels("containers")).toEqual(["Container", "Container image"]);
  });

  test("put a nested behavior after its parent, as the CLI does", () => {
    const security = labels("security") ?? [];
    expect(security.indexOf("User SSH key")).toBe(security.indexOf("User") + 1);
    const networking = labels("networking") ?? [];
    expect(networking.indexOf("Interface route")).toBe(
      networking.indexOf("Interface") + 1,
    );
  });

  test("of a hand-written block are labeled by the block", () => {
    expect(labels("objectstore")).toEqual(["File", "File deploy"]);
  });
});
