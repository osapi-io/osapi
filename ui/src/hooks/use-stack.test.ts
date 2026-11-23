import { describe, expect, test } from "bun:test";
import {
  ALL_BLOCK_TYPES,
  blockRunnable,
  newBlock,
  type Block,
} from "@/hooks/use-stack";
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
