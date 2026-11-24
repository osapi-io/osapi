// ---------------------------------------------------------------------------
// The catalog, indexed for the pages: the behaviors by name, the sections of
// the Behaviors page by area, the definition of each step that a plan can
// run, by its run name, such as dns.update, the roles, the permission of
// each operation, and the tones of the statuses and the counts.
// ---------------------------------------------------------------------------

import {
  methodLabel,
  type CatalogBehavior,
  type CatalogMethod,
  type MethodKind,
} from "@/lib/catalog";
import { rolesOf, type RoleDefinition } from "@/lib/permissions";
import type { Tones } from "@/lib/status";
import type {
  CatalogArea,
  CatalogResponse,
  OperationID,
  Permission,
} from "@/sdk/gen/schemas";

/** The behaviors of one area, in the order of the registry. */
export interface CatalogSection {
  area: string;
  label: string;
  behaviors: CatalogBehavior[];
}

/**
 * What a step of a plan runs: a method of a behavior, which runs on the
 * hosts of its target, or a call of the controller, such as an upload to
 * the object store, which takes no target.
 */
export interface StepDef {
  /** The step's name in a plan, behavior.method: dns.update. */
  run: string;
  /** The behavior's title, or the controller's subject: DNS. */
  title: string;
  /** The method in words: Update. */
  label: string;
  kind: MethodKind;
  doc: string;
  permission?: Permission;
  area: string;
  /** It runs on the controller, not on the hosts: it takes no target. */
  controller: boolean;
  behavior?: CatalogBehavior;
  method?: CatalogMethod;
}

/**
 * The calls of the controller that a plan can run, under the Object Store
 * area. Each names the operation that it calls, which gives its permission,
 * so a step cannot check another permission than the API asks for, and an
 * operation that is renamed fails the type check here.
 */
const CONTROLLER_STEPS: {
  method: string;
  doc: string;
  operation: OperationID;
}[] = [
  {
    method: "list",
    doc: "List the files in the object store",
    operation: "GetFiles",
  },
  {
    method: "upload",
    doc: "Upload a file to the object store",
    operation: "PostFile",
  },
  {
    method: "delete",
    doc: "Delete a file from the object store",
    operation: "DeleteFileByName",
  },
  {
    method: "stale",
    doc: "List the deployments whose file changed after they were deployed",
    operation: "GetFileStale",
  },
];

/** The kind of a permission: its verb read, write, or a verb of an action. */
export function permissionKind(permission: string): MethodKind {
  const verb = permission.slice(permission.indexOf(":") + 1);
  return verb === "read" ? "read" : verb === "write" ? "write" : "action";
}

/** The run name of a method of a behavior: dns.update. */
export function runName(behavior: string, method: string): string {
  return `${behavior}.${method}`;
}

/**
 * The behaviors with each one under a parent right after its parent, such
 * as the SSH keys of a user after the user, else in the order given.
 */
export function nested(behaviors: CatalogBehavior[]): CatalogBehavior[] {
  const names = new Set(behaviors.map((b) => b.name));
  const out: CatalogBehavior[] = [];
  const add = (b: CatalogBehavior) => {
    out.push(b);
    for (const child of behaviors) {
      if (child.parent === b.name) add(child);
    }
  };
  for (const b of behaviors) {
    if (!b.parent || !names.has(b.parent)) add(b);
  }
  return out;
}

export interface CatalogIndex {
  behaviors: readonly CatalogBehavior[];
  /** Each area, in order, with its label. */
  areas: readonly CatalogArea[];
  /** Each permission that osapi knows. */
  permissions: readonly Permission[];
  /** The built-in roles, in order; the last has the least. */
  roles: RoleDefinition[];
  /** The permission of each operation that needs one. */
  operations: ReadonlyMap<OperationID, Permission>;
  /** The tone of each status word, as the CLI colors it. */
  statusTones: Tones;
  /** The tone of a count above zero, by what it counts. */
  countTones: Tones;
  /** The host platforms that gohai names; runs_on has "*" for any other Linux. */
  hostFamilies: readonly string[];
  /** The sections of the Behaviors page: each area with a behavior. */
  sections: CatalogSection[];
  /** The behaviors that only other behaviors call. */
  internal: CatalogBehavior[];
  behavior: (name: string) => CatalogBehavior | undefined;
  /** Each step that a plan can run, in the order of the sections. */
  steps: StepDef[];
  step: (run: string) => StepDef | undefined;
}

/** The catalog that the controller serves, indexed for the pages. */
export function buildIndex(catalog: CatalogResponse): CatalogIndex {
  const behaviors = catalog.behaviors;
  const operations = new Map(
    catalog.operations.map((o) => [o.id, o.permission]),
  );
  const byName = new Map(behaviors.map((b) => [b.name, b]));
  const served = behaviors.filter((b) => !b.internal);
  const steps: StepDef[] = [];
  const sections: CatalogSection[] = [];
  for (const a of catalog.areas) {
    const inArea = nested(served.filter((b) => b.area === a.name));
    for (const b of inArea) {
      for (const m of b.methods) {
        steps.push({
          run: runName(b.name, m.name),
          title: b.title,
          label: methodLabel(m.name),
          kind: m.kind,
          doc: m.doc,
          permission: m.permission,
          area: a.name,
          controller: false,
          behavior: b,
          method: m,
        });
      }
    }
    if (a.name === "objectstore") {
      for (const c of CONTROLLER_STEPS) {
        const permission = operations.get(c.operation);
        if (!permission) continue;
        steps.push({
          run: runName("objectstore", c.method),
          title: "Object store",
          label: methodLabel(c.method),
          kind: permissionKind(permission),
          doc: c.doc,
          permission,
          area: a.name,
          controller: true,
        });
      }
    }
    if (inArea.length > 0) {
      sections.push({ area: a.name, label: a.label, behaviors: inArea });
    }
  }
  const stepByRun = new Map(steps.map((s) => [s.run, s]));
  return {
    behaviors,
    areas: catalog.areas,
    permissions: catalog.permissions,
    roles: rolesOf(catalog.roles),
    operations,
    statusTones: catalog.status_tones as Tones,
    countTones: catalog.count_tones as Tones,
    hostFamilies: catalog.host_families,
    sections,
    internal: behaviors.filter((b) => b.internal),
    behavior: (name) => byName.get(name),
    steps,
    step: (run) => stepByRun.get(run),
  };
}

/** The words of a query: "dns update" and "dns.update" are dns and update. */
function words(query: string): string[] {
  return query
    .toLowerCase()
    .split(/[\s.]+/)
    .filter((w) => w !== "");
}

/**
 * Whether a behavior matches a filter: each word of the query is in its
 * title, its name, its doc, or the name of one of its methods.
 */
export function behaviorMatches(b: CatalogBehavior, query: string): boolean {
  const text = [b.title, b.name, b.doc, ...b.methods.map((m) => m.name)]
    .join(" ")
    .toLowerCase();
  return words(query).every((w) => text.includes(w));
}

/**
 * Whether a step matches a filter: each word of the query is in the title
 * or the name of its behavior, or in the name or the doc of its method.
 */
export function stepMatches(def: StepDef, query: string): boolean {
  const text = [def.title, def.run, def.doc].join(" ").toLowerCase();
  return words(query).every((w) => text.includes(w));
}
