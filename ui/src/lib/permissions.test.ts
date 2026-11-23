import { describe, expect, test } from "bun:test";
import { ALL_BLOCK_TYPES } from "@/hooks/use-stack";
import {
  BLOCK_PERMISSIONS,
  blockAllowed,
  getRoleByName,
  hasPermission,
  type Permission,
} from "@/lib/permissions";

const canAs = (name: string) => {
  const role = getRoleByName(name);
  if (!role) throw new Error(`no role ${name}`);
  return (p: Permission) => hasPermission(role, p);
};

describe("blockAllowed", () => {
  test("a read-only role sees command exec and ping disabled", () => {
    const can = canAs("read");
    expect(blockAllowed("command-exec", can)).toBe(false);
    expect(blockAllowed("connectivity-ping", can)).toBe(false);
    expect(blockAllowed("dns-get", can)).toBe(true);
  });

  test("the admin role may use command exec and ping", () => {
    const can = canAs("admin");
    expect(blockAllowed("command-exec", can)).toBe(true);
    expect(blockAllowed("connectivity-ping", can)).toBe(true);
  });

  test("a block with no known permission is disabled, not allowed", () => {
    expect(blockAllowed("no-such-block", () => true)).toBe(false);
  });

  test("each block in the panel has a permission", () => {
    const missing = ALL_BLOCK_TYPES.filter(
      (bt) => BLOCK_PERMISSIONS[bt.type] === undefined,
    ).map((bt) => bt.type);
    expect(missing).toEqual([]);
  });

  test("each block is in the panel once", () => {
    const types = ALL_BLOCK_TYPES.map((bt) => bt.type);
    expect(types.length).toBe(new Set(types).size);
  });
});
