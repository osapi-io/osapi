// ---------------------------------------------------------------------------
// A plan: steps that run in order, each a method of a behavior with its
// target and its input. Its shape is the step of a workflow file,
// {run, target, with}, so a plan exports as the YAML that a workflow reads.
// ---------------------------------------------------------------------------

import type { AgentInfo } from "@/sdk/gen/schemas";
import { HOST_FAMILIES } from "@/lib/catalog.gen";
import { methodInput, methodReady } from "@/lib/catalog";
import type { StepDef } from "@/lib/catalog-index";

/** What the last run of a step did. */
export type StepOutcome = "running" | "done" | "error";

/** Where a step is: it waits for input or a target, it can run, or it ran. */
export type StepStatus = "pending" | "ready" | StepOutcome;

export interface PlanStep {
  id: string;
  /** The method that the step runs, behavior.method: dns.update. */
  run: string;
  /** The hosts that it runs on: _all, _any, a hostname or a label. */
  target: string;
  /** The input of the method, by field. */
  with: Record<string, unknown>;
  outcome?: StepOutcome;
  result?: unknown;
  error?: string;
  /** The trace of the step's last call, from the traceresponse header. */
  traceId?: string;
}

/**
 * The target that a new step starts with. A read may run on each agent, so
 * it starts with _all. A write or an action starts with none: the person
 * picks the hosts that it changes, and the step does not run until then. A
 * call of the controller takes no target.
 */
export function defaultTarget(def: StepDef): string {
  return !def.controller && def.kind === "read" ? "_all" : "";
}

/** Whether the step has the input that it needs. */
export function inputReady(
  def: StepDef,
  data: Record<string, unknown>,
): boolean {
  if (def.method) return methodReady(def.method, data);
  const name = String(data.name ?? "").trim();
  switch (def.run) {
    case "objectstore.upload":
      return name !== "" && data._file != null;
    case "objectstore.delete":
      return name !== "";
    default:
      return true;
  }
}

/** Whether Run runs the step: it has its input and, on hosts, a target. */
export function stepRunnable(def: StepDef, step: PlanStep): boolean {
  return inputReady(def, step.with) && (def.controller || step.target !== "");
}

/** The status of a step: its outcome after a run, else whether it can run. */
export function stepStatus(def: StepDef, step: PlanStep): StepStatus {
  if (step.outcome) return step.outcome;
  return stepRunnable(def, step) ? "ready" : "pending";
}

// ---------------------------------------------------------------------------
// The steps, as a reducer
// ---------------------------------------------------------------------------

export type PlanAction =
  | { type: "add"; step: PlanStep }
  | { type: "remove"; id: string }
  | { type: "move"; id: string; by: -1 | 1 }
  | { type: "with"; id: string; with: Record<string, unknown> }
  | { type: "target"; id: string; target: string }
  | {
      type: "outcome";
      id: string;
      outcome: StepOutcome;
      result?: unknown;
      error?: string;
      traceId?: string;
    }
  | { type: "reset" }
  | { type: "clear" };

/** A new step of def, as the list adds it. */
export function newStep(def: StepDef, id: string): PlanStep {
  return { id, run: def.run, target: defaultTarget(def), with: {} };
}

const done = (s: PlanStep): PlanStep => ({
  ...s,
  outcome: undefined,
  result: undefined,
  error: undefined,
  traceId: undefined,
});

export function planReducer(steps: PlanStep[], action: PlanAction): PlanStep[] {
  switch (action.type) {
    case "add":
      return [...steps, action.step];
    case "remove":
      return steps.filter((s) => s.id !== action.id);
    case "move": {
      const i = steps.findIndex((s) => s.id === action.id);
      const j = i + action.by;
      if (i < 0 || j < 0 || j >= steps.length) return steps;
      const out = [...steps];
      [out[i], out[j]] = [out[j], out[i]];
      return out;
    }
    case "with":
      return steps.map((s) =>
        s.id === action.id ? { ...s, with: action.with } : s,
      );
    case "target":
      return steps.map((s) =>
        s.id === action.id ? { ...s, target: action.target } : s,
      );
    case "outcome":
      return steps.map((s) =>
        s.id === action.id
          ? {
              ...s,
              outcome: action.outcome,
              result: action.result,
              error: action.error,
              traceId: action.traceId,
            }
          : s,
      );
    case "reset":
      return steps.map(done);
    case "clear":
      return [];
  }
}

// ---------------------------------------------------------------------------
// Preflight: what the plan needs before it runs
// ---------------------------------------------------------------------------

/**
 * The platform of a host, as the behaviors name platforms: the family that
 * the agent picks providers by, which gohai detects on the host.
 */
export function hostPlatform(agent: AgentInfo): string | undefined {
  return agent.os_info?.family || undefined;
}

/**
 * Whether a method runs on a host of platform: one that its runsOn names,
 * which the catalog has from the agent's own rule, or any other Linux when
 * runsOn has "*".
 */
export function runsOnHost(
  runsOn: readonly string[],
  platform: string,
): boolean {
  if (runsOn.includes(platform)) return true;
  return !HOST_FAMILIES.includes(platform) && runsOn.includes("*");
}

/** The agents that a target names: each, one host, or the hosts of a label. */
export function targetHosts(target: string, agents: AgentInfo[]): AgentInfo[] {
  if (target === "_all") return agents;
  if (target === "" || target === "_any") return [];
  const byName = agents.filter((a) => a.hostname === target);
  if (byName.length > 0) return byName;
  const i = target.indexOf(":");
  if (i < 0) return [];
  const key = target.slice(0, i);
  const value = target.slice(i + 1);
  return agents.filter((a) => a.labels?.[key] === value);
}

export interface PermissionCheck {
  permission: string;
  allowed: boolean;
  /** The steps that need it, by run name. */
  runs: string[];
}

export interface PlatformSkip {
  stepId: string;
  run: string;
  hostname: string;
  platform: string;
}

export interface Preflight {
  permissions: PermissionCheck[];
  /** The hosts that a step skips, because its method does not run there. */
  skips: PlatformSkip[];
  /** The steps that wait for input or a target. */
  pending: number;
}

/**
 * The checks of a plan before it runs: each permission that its steps need
 * against the viewer's role, the hosts of a target whose platform a method
 * does not run on, and the steps that cannot run yet. The controller makes
 * the same permission check on each call; this says it first.
 */
export function preflight(
  steps: PlanStep[],
  stepDef: (run: string) => StepDef | undefined,
  can: (permission: string) => boolean,
  agents: AgentInfo[],
): Preflight {
  const perms = new Map<string, PermissionCheck>();
  const skips: PlatformSkip[] = [];
  let pending = 0;
  for (const step of steps) {
    const def = stepDef(step.run);
    if (!def) continue;
    if (!stepRunnable(def, step)) pending++;
    if (def.permission) {
      const p = perms.get(def.permission) ?? {
        permission: def.permission,
        allowed: can(def.permission),
        runs: [],
      };
      if (!p.runs.includes(step.run)) p.runs.push(step.run);
      perms.set(def.permission, p);
    }
    if (!def.method) continue;
    for (const host of targetHosts(step.target, agents)) {
      const platform = hostPlatform(host);
      if (platform && !runsOnHost(def.method.runsOn, platform)) {
        skips.push({
          stepId: step.id,
          run: step.run,
          hostname: host.hostname,
          platform,
        });
      }
    }
  }
  return { permissions: [...perms.values()], skips, pending };
}

// ---------------------------------------------------------------------------
// YAML export
// ---------------------------------------------------------------------------

/** A value as YAML: a plain word as it is, else quoted as JSON quotes it. */
function yamlScalar(v: unknown): string {
  if (typeof v === "number" || typeof v === "boolean") return String(v);
  const s = String(v);
  return /^[A-Za-z0-9_./@-][A-Za-z0-9_./@ -]*$/.test(s) &&
    !/^(true|false|null|yes|no|on|off|~)$/i.test(s) &&
    !/^[-+]?(\d+\.?\d*|\.\d+)([eE][-+]?\d+)?$/.test(s) &&
    !s.endsWith(" ")
    ? s
    : JSON.stringify(s);
}

function yamlValue(v: unknown, indent: string): string {
  if (Array.isArray(v)) return `[${v.map(yamlScalar).join(", ")}]`;
  if (v !== null && typeof v === "object") {
    const entries = Object.entries(v as Record<string, unknown>);
    if (entries.length === 0) return "{}";
    return entries
      .map(([k, x]) => `\n${indent}  ${k}: ${yamlValue(x, `${indent}  `)}`)
      .join("");
  }
  return yamlScalar(v);
}

/**
 * The plan as a workflow file: each step's run, its target and its input,
 * as the API takes the input. A step with no target says so, and a call of
 * the controller, which a workflow cannot run yet, is a comment. A secret is
 * left out.
 */
export function planYAML(
  steps: PlanStep[],
  stepDef: (run: string) => StepDef | undefined,
): string {
  const lines = ["name: new-plan", "steps:"];
  if (steps.length === 0) lines[1] = "steps: []";
  for (const step of steps) {
    const def = stepDef(step.run);
    if (!def) continue;
    if (def.controller) {
      lines.push(
        `  # ${step.run} runs on the controller; a workflow cannot run it yet`,
      );
      continue;
    }
    lines.push(`  - run: ${step.run}`);
    lines.push(
      step.target === ""
        ? "    # target: none yet; the step does not run until it has one"
        : `    target: ${yamlScalar(step.target)}`,
    );
    const input: Record<string, unknown> = {};
    const raw = def.method ? safeInput(def, step) : {};
    for (const f of def.method?.fields ?? []) {
      if (!(f.name in raw)) continue;
      input[f.name] = f.secret ? "<secret>" : raw[f.name];
    }
    if (Object.keys(input).length > 0) {
      lines.push(`    with:${yamlValue(input, "    ")}`);
    }
  }
  return `${lines.join("\n")}\n`;
}

/** The input of a step, or none when a field does not parse, such as JSON. */
function safeInput(def: StepDef, step: PlanStep): Record<string, unknown> {
  try {
    return def.method ? methodInput(def.method, step.with) : {};
  } catch {
    return {};
  }
}
