// ---------------------------------------------------------------------------
// The catalog of the behaviors, which the controller serves at
// GET /api/catalog (useCatalog)
//
// Each behavior that the API serves is in it, with each of its methods: the
// form that the method's input makes, the route that runs it, its permission,
// the platforms that it runs on, the behaviors that it calls, its examples,
// and the columns of the rows that it answers. A behavior that changes
// changes its page and its steps; nobody edits the UI. The types come from
// the spec, through the generated client.
// ---------------------------------------------------------------------------

import type {
  CatalogBehavior,
  CatalogCall,
  CatalogColumn,
  CatalogExample,
  CatalogField,
  CatalogMethod,
  CatalogMethodKind,
} from "@/sdk/gen/schemas";

export type {
  CatalogBehavior,
  CatalogCall,
  CatalogColumn,
  CatalogExample,
  CatalogField,
  CatalogMethod,
};

/**
 * What a method does to a host: a read changes nothing, a write changes what
 * the host is configured to be, and an action makes something happen, such
 * as a restart or a reboot. It is the verb of the method's permission.
 */
export type MethodKind = CatalogMethodKind;

export const methodKindLabels: Record<MethodKind, string> = {
  read: "Read",
  write: "Write",
  action: "Action",
};

/** The name of a method in words: update is Update. */
export function methodLabel(name: string): string {
  return name.charAt(0).toUpperCase() + name.slice(1);
}

/** The value of a form field as the API takes it. */
export function fieldValue(field: CatalogField, raw: unknown): unknown {
  switch (field.kind) {
    case "number":
      return Number(raw);
    case "bool":
      return raw === true || raw === "true";
    case "list":
      return (Array.isArray(raw) ? raw.map(String) : String(raw).split(","))
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
  if (Array.isArray(raw)) return raw.length > 0;
  return raw !== undefined && raw !== null && String(raw).trim() !== "";
}

/** Whether each required field of the method has a value. */
export function methodReady(
  method: CatalogMethod,
  data: Record<string, unknown>,
): boolean {
  return method.fields.every((f) => !f.required || hasValue(data[f.name]));
}

/**
 * A value as one segment of a path. It keeps "@" and ":", which a path
 * segment allows: the controller reads a path key as it is sent, so an
 * escaped @fact. reference would not resolve.
 */
export function pathSegment(value: string): string {
  return encodeURIComponent(value).replace(/%40/g, "@").replace(/%3A/g, ":");
}

/** The input of a method as the API takes it: each field that has a value. */
export function methodInput(
  method: CatalogMethod,
  data: Record<string, unknown>,
): Record<string, unknown> {
  const out: Record<string, unknown> = {};
  for (const f of method.fields) {
    const raw = data[f.name];
    // A flag goes only once it is set, so an untouched one keeps the API's default.
    if (raw === undefined || (!hasValue(raw) && f.kind !== "bool")) continue;
    out[f.name] = fieldValue(f, raw);
  }
  return out;
}

/** The request of a method: its path, query and body. */
export function methodRequest(
  method: CatalogMethod,
  target: string,
  data: Record<string, unknown>,
): { url: string; init: RequestInit } {
  let path = (method.path ?? "").replace("{hostname}", pathSegment(target));
  const query = new URLSearchParams();
  const body: Record<string, unknown> = {};
  const input = methodInput(method, data);
  for (const f of method.fields) {
    if (!(f.name in input)) continue;
    const value = input[f.name];
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
  const init: RequestInit = { method: method.http_method };
  if (method.http_method !== "GET" && Object.keys(body).length > 0) {
    init.body = JSON.stringify(body);
  }
  return { url: qs ? `${path}?${qs}` : path, init };
}
