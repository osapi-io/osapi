import { describe, expect, test } from "bun:test";
import { catalogIndex, type StepDef } from "@/lib/catalog-index";
import type { AgentInfo } from "@/sdk/gen/schemas";
import {
  defaultTarget,
  hostPlatform,
  inputReady,
  newStep,
  planReducer,
  planYAML,
  preflight,
  runsOnHost,
  stepStatus,
  targetHosts,
  type PlanStep,
} from "@/lib/plan";

const def = (run: string): StepDef => {
  const d = catalogIndex.step(run);
  if (!d) throw new Error(`no step ${run}`);
  return d;
};

const agent = (
  hostname: string,
  family?: string,
  labels?: Record<string, string>,
): AgentInfo =>
  ({
    hostname,
    state: "Ready",
    labels,
    os_info: family ? { distribution: "any", version: "1", family } : undefined,
  }) as AgentInfo;

describe("a new step", () => {
  test("a read starts with _all and is ready", () => {
    const s = newStep(def("status.get"), "s-1");
    expect(s.target).toBe("_all");
    expect(stepStatus(def("status.get"), s)).toBe("ready");
  });

  test("a write and an action start with no target, and are pending", () => {
    const reboot = newStep(def("power.reboot"), "s-1");
    expect(reboot.target).toBe("");
    expect(stepStatus(def("power.reboot"), reboot)).toBe("pending");
    expect(defaultTarget(def("user.delete"))).toBe("");
    expect(defaultTarget(def("service.restart"))).toBe("");
  });

  test("a write with a target and its input is ready", () => {
    const s = newStep(def("power.reboot"), "s-1");
    expect(stepStatus(def("power.reboot"), { ...s, target: "web-01" })).toBe(
      "ready",
    );
  });

  test("a call of the controller takes no target", () => {
    const d = def("objectstore.list");
    const s = newStep(d, "s-1");
    expect(s.target).toBe("");
    expect(stepStatus(d, s)).toBe("ready");
  });

  test("an upload needs a file and a name, and a delete a name", () => {
    expect(inputReady(def("objectstore.upload"), { name: "a" })).toBe(false);
    expect(
      inputReady(def("objectstore.upload"), { name: "a", _file: {} }),
    ).toBe(true);
    expect(inputReady(def("objectstore.delete"), {})).toBe(false);
    expect(inputReady(def("objectstore.delete"), { name: "a" })).toBe(true);
  });

  test("its outcome is its status once it ran", () => {
    const s = newStep(def("status.get"), "s-1");
    expect(stepStatus(def("status.get"), { ...s, outcome: "error" })).toBe(
      "error",
    );
  });
});

describe("planReducer", () => {
  const a = newStep(def("status.get"), "a");
  const b = newStep(def("dns.update"), "b");

  test("adds, moves and removes steps", () => {
    let steps = planReducer([], { type: "add", step: a });
    steps = planReducer(steps, { type: "add", step: b });
    steps = planReducer(steps, { type: "move", id: "b", by: -1 });
    expect(steps.map((s) => s.id)).toEqual(["b", "a"]);
    expect(planReducer(steps, { type: "move", id: "b", by: -1 })).toBe(steps);
    steps = planReducer(steps, { type: "remove", id: "b" });
    expect(steps.map((s) => s.id)).toEqual(["a"]);
    expect(planReducer(steps, { type: "clear" })).toEqual([]);
  });

  test("sets the input, the target and the outcome, and resets", () => {
    let steps = planReducer([b], { type: "with", id: "b", with: { x: 1 } });
    steps = planReducer(steps, { type: "target", id: "b", target: "web-01" });
    steps = planReducer(steps, {
      type: "outcome",
      id: "b",
      outcome: "done",
      result: { ok: true },
      traceId: "t",
    });
    expect(steps[0]).toMatchObject({
      with: { x: 1 },
      target: "web-01",
      outcome: "done",
      result: { ok: true },
      traceId: "t",
    });
    steps = planReducer(steps, { type: "reset" });
    expect(steps[0].outcome).toBeUndefined();
    expect(steps[0].result).toBeUndefined();
    expect(steps[0].target).toBe("web-01");
  });
});

describe("platforms", () => {
  test("a host's platform is the family that its agent reports", () => {
    expect(hostPlatform(agent("m", "darwin"))).toBe("darwin");
    expect(hostPlatform(agent("u", "debian"))).toBe("debian");
    expect(hostPlatform(agent("s", "suse"))).toBe("suse");
    expect(hostPlatform(agent("x"))).toBeUndefined();
  });

  test("a method runs on the hosts that its runsOn names", () => {
    expect(runsOnHost(["debian"], "debian")).toBe(true);
    expect(runsOnHost(["debian"], "suse")).toBe(false);
    expect(runsOnHost(["debian", "fedora", "*"], "nixos")).toBe(true);
    expect(runsOnHost(["debian"], "nixos")).toBe(false);
    expect(runsOnHost(["debian", "*"], "darwin")).toBe(false);
  });
});

describe("targetHosts", () => {
  const agents = [
    agent("web-01", "debian", { group: "web" }),
    agent("mac-01", "darwin"),
  ];
  test("is each host, one host, or the hosts of a label", () => {
    expect(targetHosts("_all", agents)).toHaveLength(2);
    expect(targetHosts("_any", agents)).toEqual([]);
    expect(targetHosts("", agents)).toEqual([]);
    expect(targetHosts("mac-01", agents).map((a) => a.hostname)).toEqual([
      "mac-01",
    ]);
    expect(targetHosts("group:web", agents).map((a) => a.hostname)).toEqual([
      "web-01",
    ]);
    expect(targetHosts("nothing", agents)).toEqual([]);
  });
});

describe("preflight", () => {
  const steps: PlanStep[] = [
    { ...newStep(def("dns.update"), "a"), target: "_all" },
    newStep(def("dns.get"), "b"),
    newStep(def("power.reboot"), "c"),
    { id: "x", run: "no.such", target: "", with: {} },
  ];
  const agents = [agent("web-01", "debian"), agent("mac-01", "darwin")];

  test("checks each permission, each platform, and the pending steps", () => {
    const check = preflight(
      steps,
      catalogIndex.step,
      (p) => p.endsWith(":read"),
      agents,
    );
    const dns = check.permissions.find((p) => p.permission === "dns:write");
    expect(dns?.allowed).toBe(false);
    expect(dns?.runs).toEqual(["dns.update"]);
    expect(
      check.permissions.find((p) => p.permission === "dns:read")?.allowed,
    ).toBe(true);
    expect(check.skips).toEqual([
      {
        stepId: "a",
        run: "dns.update",
        hostname: "mac-01",
        platform: "darwin",
      },
    ]);
    // dns update and dns get need an interface, and reboot a target.
    expect(check.pending).toBe(3);
  });
});

describe("planYAML", () => {
  test("is the steps as a workflow file", () => {
    const steps: PlanStep[] = [
      {
        ...newStep(def("dns.update"), "a"),
        target: "web-01",
        with: { interface_name: "eth0", servers: ["1.1.1.1", "8.8.8.8"] },
      },
      newStep(def("power.reboot"), "b"),
      newStep(def("objectstore.list"), "c"),
    ];
    expect(planYAML(steps, catalogIndex.step)).toBe(
      [
        "name: new-plan",
        "steps:",
        "  - run: dns.update",
        "    target: web-01",
        "    with:",
        "      interface_name: eth0",
        "      servers: [1.1.1.1, 8.8.8.8]",
        "  - run: power.reboot",
        "    # target: none yet; the step does not run until it has one",
        "  # objectstore.list runs on the controller; a workflow cannot run it yet",
        "",
      ].join("\n"),
    );
  });

  test("quotes a value that YAML would read as another type, and hides a secret", () => {
    const d = def("user.update");
    const secret = d.method?.fields.find((f) => f.secret);
    const steps: PlanStep[] = [
      {
        ...newStep(def("sysctl.update"), "a"),
        target: "_all",
        with: { key: "net.ipv4.ip_forward", value: "1" },
      },
    ];
    if (secret) {
      steps.push({
        ...newStep(d, "b"),
        target: "_all",
        with: { name: "joe", [secret.name]: "hunter2" },
      });
    }
    const yaml = planYAML(steps, catalogIndex.step);
    expect(yaml).toContain('      value: "1"');
    expect(yaml).not.toContain("hunter2");
  });

  test("an empty plan has no steps", () => {
    expect(planYAML([], catalogIndex.step)).toBe("name: new-plan\nsteps: []\n");
  });
});
