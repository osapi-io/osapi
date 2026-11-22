// ---------------------------------------------------------------------------
// Behavior blocks — generated from the behavior registry, in blocks.gen.ts
//
// Each method of a behavior that the API serves is a block: its form comes
// from the method's input, its request from the method's route, its
// permission from the definition, and its result card from the rows that it
// answers. A behavior that changes changes its block; nobody edits the UI.
// ---------------------------------------------------------------------------

import { apiFetch } from "@/sdk/fetch";

export interface BehaviorBlockField {
  readonly name: string;
  readonly in: "path" | "query" | "body";
  readonly kind: "string" | "number" | "bool" | "list" | "object";
  readonly required: boolean;
  readonly doc: string;
  readonly options?: readonly string[];
  readonly secret?: boolean;
}

export interface BehaviorBlockColumn {
  readonly name: string;
  readonly doc: string;
}

export interface BehaviorBlockDef {
  readonly type: string;
  readonly behavior: string;
  readonly title: string;
  readonly method: string;
  readonly label: string;
  readonly description: string;
  readonly httpMethod: string;
  readonly path: string;
  readonly permission: string;
  readonly write: boolean;
  readonly fields: readonly BehaviorBlockField[];
  readonly columns: readonly BehaviorBlockColumn[];
}

/** The value of a form field as the API takes it. */
export function fieldValue(field: BehaviorBlockField, raw: unknown): unknown {
  switch (field.kind) {
    case "number":
      return Number(raw);
    case "bool":
      return raw === true || raw === "true";
    case "list":
      return String(raw)
        .split(",")
        .map((v) => v.trim())
        .filter((v) => v !== "");
    case "object":
      return JSON.parse(String(raw));
    default:
      return String(raw);
  }
}

/** Whether a field has a value: a form leaves an empty one out. */
export function hasValue(raw: unknown): boolean {
  return raw !== undefined && raw !== null && String(raw).trim() !== "";
}

/** Whether each required field of the block has a value. */
export function blockReady(
  def: BehaviorBlockDef,
  data: Record<string, unknown>,
): boolean {
  return def.fields.every((f) => !f.required || hasValue(data[f.name]));
}

/**
 * A value as one segment of a path. It keeps "@" and ":", which a path
 * segment allows: the controller reads a path key as it is sent, so an
 * escaped @fact. reference would not resolve.
 */
export function pathSegment(value: string): string {
  return encodeURIComponent(value).replace(/%40/g, "@").replace(/%3A/g, ":");
}

/** The request of a block: its path, query and body. */
export function behaviorRequest(
  def: BehaviorBlockDef,
  target: string,
  data: Record<string, unknown>,
): { url: string; init: RequestInit } {
  let path = def.path.replace("{hostname}", pathSegment(target));
  const query = new URLSearchParams();
  const body: Record<string, unknown> = {};
  for (const f of def.fields) {
    const raw = data[f.name];
    // A flag goes only once it is set, so an untouched one keeps the API's default.
    if (raw === undefined || (!hasValue(raw) && f.kind !== "bool")) continue;
    const value = fieldValue(f, raw);
    if (f.in === "path") {
      path = path.replace(`{${f.name}}`, pathSegment(String(value)));
    } else if (f.in === "query") {
      for (const v of Array.isArray(value) ? value : [value]) {
        query.append(f.name, String(v));
      }
    } else {
      body[f.name] = value;
    }
  }
  const qs = query.toString();
  const init: RequestInit = { method: def.httpMethod };
  if (def.write && Object.keys(body).length > 0) {
    init.body = JSON.stringify(body);
  }
  return { url: qs ? `${path}?${qs}` : path, init };
}

/**
 * Runs a block against the target, as the generated client would. A refusal
 * (a status of 400 or more) throws the API's error, so the block shows it.
 */
export async function runBehavior(
  def: BehaviorBlockDef,
  target: string,
  data: Record<string, unknown>,
): Promise<{ data: unknown }> {
  const { url, init } = behaviorRequest(def, target, data);
  const res = await apiFetch<{ data: unknown; status: number }>(url, init);
  if (res.status >= 400) {
    const body = res.data as { error?: string; reason?: string } | undefined;
    const reason = body?.reason ? ` (${body.reason})` : "";
    throw new Error(`${body?.error ?? `HTTP ${res.status}`}${reason}`);
  }
  return res;
}
