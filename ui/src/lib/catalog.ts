// ---------------------------------------------------------------------------
// The catalog of the behaviors, generated from the registry in catalog.gen.ts
//
// Each behavior that the API serves is in it, with each of its methods: the
// form that the method's input makes, the route that runs it, its permission,
// the platforms that it runs on, the behaviors that it calls, its examples,
// and the columns of the rows that it answers. A behavior that changes
// changes its page and its steps; nobody edits the UI.
// ---------------------------------------------------------------------------

export interface CatalogField {
  readonly name: string;
  readonly in: "path" | "query" | "body";
  readonly kind: "string" | "number" | "bool" | "list" | "object";
  readonly required: boolean;
  readonly doc: string;
  readonly options?: readonly string[];
  readonly secret?: boolean;
}

/**
 * A field of the rows that a method answers, as the CLI's output shows it:
 * its label, its render tag, such as bytes, and for an object or a list of
 * objects the columns of the object.
 */
export interface CatalogColumn {
  readonly name: string;
  readonly label: string;
  readonly doc: string;
  readonly render?: string;
  readonly kind: "value" | "values" | "object" | "list";
  readonly columns?: readonly CatalogColumn[];
}

/** A call that a method makes to another behavior, with its own calls. */
export interface CatalogCall {
  readonly behavior: string;
  readonly method: string;
  readonly calls?: readonly CatalogCall[];
}

export interface CatalogExample {
  readonly doc: string;
  readonly input: unknown;
  readonly output?: unknown;
}

/**
 * What a method does to a host: a read changes nothing, a write changes what
 * the host is configured to be, and an action makes something happen, such
 * as a restart or a reboot. It is the verb of the method's permission.
 */
export type MethodKind = "read" | "write" | "action";

export const methodKindLabels: Record<MethodKind, string> = {
  read: "Read",
  write: "Write",
  action: "Action",
};

export interface CatalogMethod {
  readonly name: string;
  readonly kind: MethodKind;
  readonly doc: string;
  readonly long?: string;
  readonly httpMethod?: string;
  readonly path?: string;
  readonly permission?: string;
  readonly platforms: readonly string[];
  /** The input type and its fields, a required one marked *. */
  readonly input: string;
  /** The output type and its fields. */
  readonly output: string;
  readonly fields: readonly CatalogField[];
  readonly columns: readonly CatalogColumn[];
  readonly calls: readonly CatalogCall[];
  readonly examples: readonly CatalogExample[];
  readonly disruptive?: boolean;
  readonly disruption?: string;
}

export interface CatalogBehavior {
  readonly name: string;
  readonly title: string;
  /** The area, or "" for an internal behavior. */
  readonly area: string;
  readonly doc: string;
  /** Only other behaviors call it: it has no route. */
  readonly internal: boolean;
  readonly parent?: string;
  readonly platforms: readonly string[];
  readonly uses: readonly string[];
  readonly usedBy: readonly string[];
  readonly chain: string;
  readonly methods: readonly CatalogMethod[];
}

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
  const init: RequestInit = { method: method.httpMethod };
  if (method.httpMethod !== "GET" && Object.keys(body).length > 0) {
    init.body = JSON.stringify(body);
  }
  return { url: qs ? `${path}?${qs}` : path, init };
}
