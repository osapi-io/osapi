import { describe, expect, test } from "bun:test";
import { AREAS } from "@/lib/catalog.gen";
import {
  behaviorMatches,
  buildIndex,
  catalogIndex,
  nested,
  permissionKind,
  stepMatches,
} from "@/lib/catalog-index";
import type { CatalogBehavior } from "@/lib/catalog";

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
  usedBy: [],
  chain: name,
  methods: [],
  ...extra,
});

describe("the catalog", () => {
  test("has a section for each area with a behavior, in the order of AREAS", () => {
    const names = catalogIndex.sections.map((s) => s.area);
    const order = AREAS.map((a) => a.name).filter((a) => names.includes(a));
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
    const index = buildIndex([behavior("dns")]);
    expect(index.sections.map((s) => s.area)).toEqual(["networking"]);
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
