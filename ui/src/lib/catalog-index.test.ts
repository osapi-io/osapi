import { describe, expect, test } from "bun:test";
import {
  behaviorMatches,
  buildIndex,
  nested,
  permissionKind,
  stepMatches,
} from "@/lib/catalog-index";
import type { CatalogBehavior } from "@/lib/catalog";
import { testCatalog, testIndex as catalogIndex } from "@/test/catalog";

const behavior = (
  name: string,
  extra: Partial<CatalogBehavior> = {},
): CatalogBehavior => ({
  name,
  title: name.toUpperCase(),
  area: "networking",
  doc: `The ${name}`,
  internal: false,
  platforms: [],
  uses: [],
  used_by: [],
  chain: name,
  methods: [],
  ...extra,
});

describe("the catalog", () => {
  test("has a section for each area with a behavior, in the order of the areas", () => {
    const names = catalogIndex.sections.map((s) => s.area);
    const order = testCatalog.areas
      .map((a) => a.name)
      .filter((a) => names.includes(a));
    expect(names).toEqual(order);
    expect(names).not.toContain("fleet");
  });

  test("has the internal behaviors apart, so a plan cannot run them", () => {
    const internal = catalogIndex.internal.map((b) => b.name);
    expect(internal).toContain("netplan");
    expect(catalogIndex.steps.some((s) => s.run.startsWith("netplan."))).toBe(
      false,
    );
  });

  test("a step is a method of a behavior, by its run name", () => {
    const step = catalogIndex.step("dns.update");
    expect(step?.title).toBe("DNS");
    expect(step?.label).toBe("Update");
    expect(step?.kind).toBe("write");
    expect(step?.controller).toBe(false);
    expect(step?.method?.calls.length).toBeGreaterThan(0);
  });

  test("the object store's steps run on the controller", () => {
    const upload = catalogIndex.step("objectstore.upload");
    expect(upload?.controller).toBe(true);
    expect(upload?.area).toBe("objectstore");
    expect(upload?.permission).toBe("file:write");
    expect(upload?.kind).toBe("write");
    expect(catalogIndex.step("objectstore.list")?.kind).toBe("read");
  });

  test("a behavior is found by its name", () => {
    expect(catalogIndex.behavior("dns")?.title).toBe("DNS");
    expect(catalogIndex.behavior("no-such")).toBeUndefined();
  });

  test("each method's kind is the verb of its permission", () => {
    for (const s of catalogIndex.steps) {
      if (s.permission) expect(s.kind).toBe(permissionKind(s.permission));
    }
  });
});

describe("permissionKind", () => {
  test("is read, write, or an action for another verb", () => {
    expect(permissionKind("dns:read")).toBe("read");
    expect(permissionKind("dns:write")).toBe("write");
    expect(permissionKind("power:execute")).toBe("action");
    expect(permissionKind("docker:exec")).toBe("action");
  });
});

describe("nested", () => {
  test("puts a behavior right after its parent", () => {
    const got = nested([
      behavior("sshkey", { parent: "user" }),
      behavior("group"),
      behavior("user"),
    ]).map((b) => b.name);
    expect(got).toEqual(["group", "user", "sshkey"]);
  });

  test("keeps a behavior whose parent is elsewhere", () => {
    expect(
      nested([behavior("route", { parent: "nowhere" })]).map((b) => b.name),
    ).toEqual(["route"]);
  });
});

describe("buildIndex", () => {
  test("leaves out an area with no behavior", () => {
    const index = buildIndex({ ...testCatalog, behaviors: [behavior("dns")] });
    expect(index.sections.map((s) => s.area)).toEqual(["networking"]);
  });

  test("indexes the roles, the operations and the tones", () => {
    expect(catalogIndex.roles.map((r) => r.name)).toEqual([
      "admin",
      "read",
      "observer",
    ]);
    expect(catalogIndex.roles[2].permissions.has("agent:read")).toBe(true);
    expect(catalogIndex.operations.get("GetCatalog")).toBe("catalog:read");
    expect(catalogIndex.statusTones.failed).toBe("err");
    expect(catalogIndex.countTones.skipped).toBe("warn");
    expect(catalogIndex.hostFamilies).toContain("darwin");
    expect(catalogIndex.permissions).toContain("catalog:read");
  });

  test("takes each step of the controller from the catalog, with its permission", () => {
    expect(catalogIndex.step("objectstore.list")?.permission).toBe("file:read");
    expect(catalogIndex.step("objectstore.upload")?.kind).toBe("write");
    const index = buildIndex({ ...testCatalog, controller: [] });
    expect(index.step("objectstore.list")).toBeUndefined();
  });

  test("offers only the methods of the controller that a plan can run", () => {
    expect(catalogIndex.step("objectstore.get")).toBeUndefined();
  });

  test("has the targets, the render tags and the outcomes of the catalog", () => {
    expect(catalogIndex.targets.map((t) => t.value)).toEqual(["_all", "_any"]);
    expect(catalogIndex.renderTags).toContain("bytes");
    expect(catalogIndex.outcomes[0]).toBe("ok");
  });
});

describe("the filters", () => {
  const dns = catalogIndex.behavior("dns");
  test("match each word, in a title, a name, a doc or a method", () => {
    if (!dns) throw new Error("no dns");
    expect(behaviorMatches(dns, "")).toBe(true);
    expect(behaviorMatches(dns, "dns update")).toBe(true);
    expect(behaviorMatches(dns, "DNS.update")).toBe(true);
    expect(behaviorMatches(dns, "resolvers")).toBe(true);
    expect(behaviorMatches(dns, "dns reboot")).toBe(false);
  });

  test("match a step by its behavior, its method or its doc", () => {
    const step = catalogIndex.step("dns.update");
    if (!step) throw new Error("no dns.update");
    expect(stepMatches(step, "dns update")).toBe(true);
    expect(stepMatches(step, "dns.update")).toBe(true);
    expect(stepMatches(step, "dns get")).toBe(false);
  });
});
