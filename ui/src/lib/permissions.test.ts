import { describe, expect, test } from "bun:test";
import { catalogIndex } from "@/lib/catalog-index";
import {
  getRoleByName,
  hasPermission,
  permitted,
  type Permission,
} from "@/lib/permissions";

const canAs = (name: string) => {
  const role = getRoleByName(name);
  if (!role) throw new Error(`no role ${name}`);
  return (p: Permission) => hasPermission(role, p);
};

const permissionOf = (run: string) => catalogIndex.step(run)?.permission;

describe("permitted", () => {
  test("a read-only role cannot run command exec or ping", () => {
    const can = canAs("read");
    expect(permitted(permissionOf("command.exec"), can)).toBe(false);
    expect(permitted(permissionOf("connectivity.ping"), can)).toBe(false);
    expect(permitted(permissionOf("dns.get"), can)).toBe(true);
  });

  test("the admin role may run command exec and ping", () => {
    const can = canAs("admin");
    expect(permitted(permissionOf("command.exec"), can)).toBe(true);
    expect(permitted(permissionOf("connectivity.ping"), can)).toBe(true);
  });

  test("a step with no known permission is refused, not allowed", () => {
    expect(permitted(undefined, () => true)).toBe(false);
    expect(permitted("nothing:read", () => true)).toBe(false);
  });

  test("each step of a plan has a permission that the UI knows", () => {
    const missing = catalogIndex.steps
      .filter((d) => !permitted(d.permission, () => true))
      .map((d) => d.run);
    expect(missing).toEqual([]);
  });

  test("each step is in the list once", () => {
    const runs = catalogIndex.steps.map((d) => d.run);
    expect(runs.length).toBe(new Set(runs).size);
  });
});
