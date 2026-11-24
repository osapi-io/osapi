import { describe, expect, test } from "bun:test";
import {
  hasAllPermissions,
  hasPermission,
  permitted,
  rolesOf,
  type Permission,
} from "@/lib/permissions";
import { testCatalog, testIndex as catalogIndex } from "@/test/catalog";

const canAs = (name: string) => {
  const role = catalogIndex.roles.find((r) => r.name === name);
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
    expect(permitted("dns:read", () => false)).toBe(false);
  });

  test("each step of a plan has a permission", () => {
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

describe("rolesOf", () => {
  test("is each role of the catalog, in order, with its permissions", () => {
    const roles = rolesOf(testCatalog.roles);
    expect(roles.map((r) => r.name)).toEqual(["admin", "read", "observer"]);
    expect(roles[2]).toMatchObject({
      label: "observer",
      jwtRole: "observer",
      description: "A dashboard or a monitor.",
    });
    expect(hasAllPermissions(roles[2], ["agent:read", "catalog:read"])).toBe(
      true,
    );
    expect(hasAllPermissions(roles[2], ["agent:read", "dns:read"])).toBe(false);
  });
});
