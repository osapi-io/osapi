// A catalog for the tests, in the shape that GET /api/catalog answers. It
// has the behaviors, the roles and the operations that the tests name; the
// controller serves the whole catalog.

import type {
  CatalogBehavior,
  CatalogField,
  CatalogMethod,
  CatalogResponse,
  Permission,
} from "@/sdk/gen/schemas";
import { buildIndex } from "@/lib/catalog-index";

const field = (
  name: string,
  inPart: CatalogField["in"],
  extra: Partial<CatalogField> = {},
): CatalogField => ({
  name,
  in: inPart,
  kind: "string",
  required: inPart === "path",
  doc: `The ${name}`,
  ...extra,
});

const method = (
  behavior: string,
  name: string,
  kind: CatalogMethod["kind"],
  extra: Partial<CatalogMethod> = {},
): CatalogMethod => {
  const verb = kind === "action" ? "execute" : kind;
  return {
    name,
    kind,
    doc: `${name} the ${behavior}`,
    http_method: kind === "read" ? "GET" : "POST",
    path: `/api/node/{hostname}/${behavior}`,
    permission: `${behavior}:${verb}` as Permission,
    platforms: ["debian", "darwin"],
    runs_on: ["debian", "darwin", "fedora", "suse", "*"],
    input: "Input {}",
    output: "Output {}",
    fields: [],
    columns: [],
    calls: [],
    examples: [],
    ...extra,
  };
};

const behavior = (
  name: string,
  title: string,
  area: string,
  methods: CatalogMethod[],
  extra: Partial<CatalogBehavior> = {},
): CatalogBehavior => ({
  name,
  title,
  area,
  doc: `The ${title}`,
  internal: false,
  platforms: ["debian", "darwin"],
  uses: [],
  used_by: [],
  chain: name,
  methods,
  ...extra,
});

const behaviors: CatalogBehavior[] = [
  behavior("service", "Service", "services", [
    method("service", "restart", "action", {
      fields: [field("name", "path")],
    }),
  ]),
  behavior("user", "User", "config", [
    method("user", "delete", "write", { fields: [field("name", "path")] }),
    method("user", "update", "write", {
      fields: [
        field("name", "path"),
        field("password", "body", { secret: true }),
      ],
    }),
  ]),
  behavior("sysctl", "Sysctl", "config", [
    method("sysctl", "update", "write", {
      fields: [field("key", "path"), field("value", "body")],
    }),
  ]),
  behavior("status", "Status", "node", [method("status", "get", "read")]),
  behavior("power", "Power", "node", [
    method("power", "reboot", "action", { disruptive: true }),
  ]),
  behavior(
    "dns",
    "DNS",
    "networking",
    [
      method("dns", "get", "read", {
        fields: [field("interface_name", "path")],
        runs_on: ["debian", "darwin", "fedora", "*"],
      }),
      method("dns", "update", "write", {
        fields: [
          field("interface_name", "path"),
          field("servers", "body", { kind: "list" }),
        ],
        runs_on: ["debian", "fedora", "*"],
        calls: [{ behavior: "netplan", method: "apply" }],
      }),
    ],
    { doc: "The DNS resolvers of each interface", uses: ["netplan"] },
  ),
  behavior("connectivity", "Connectivity", "networking", [
    method("connectivity", "ping", "action"),
  ]),
  behavior("command", "Command", "command", [
    method("command", "exec", "action"),
  ]),
  behavior(
    "netplan",
    "Netplan",
    "",
    [
      {
        ...method("netplan", "apply", "write"),
        http_method: undefined,
        path: undefined,
        permission: undefined,
      },
    ],
    { internal: true, used_by: ["dns"] },
  ),
];

const permissions = [
  ...new Set(
    behaviors.flatMap((b) =>
      b.methods.flatMap((m) => (m.permission ? [m.permission] : [])),
    ),
  ),
  "file:read",
  "file:write",
  "agent:read",
  "catalog:read",
] as Permission[];

/** The catalog of the tests, as the controller answers it. */
export const testCatalog: CatalogResponse = {
  behaviors,
  areas: [
    { name: "services", label: "Services" },
    { name: "config", label: "Config" },
    { name: "node", label: "Node" },
    { name: "networking", label: "Networking" },
    { name: "objectstore", label: "Object Store" },
    { name: "command", label: "Command" },
    { name: "fleet", label: "Fleet" },
  ],
  permissions,
  roles: [
    {
      name: "admin",
      description: "The owner of the fleet.",
      permissions,
    },
    {
      name: "read",
      description: "A person who looks.",
      permissions: permissions.filter((p) => p.endsWith(":read")),
    },
    {
      name: "observer",
      description: "A dashboard or a monitor.",
      permissions: ["agent:read", "catalog:read"],
    },
  ],
  operations: [
    { id: "DeleteFileByName", permission: "file:write" },
    { id: "GetAgents", permission: "agent:read" },
    { id: "GetCatalog", permission: "catalog:read" },
    { id: "GetFileStale", permission: "file:read" },
    { id: "GetFiles", permission: "file:read" },
    { id: "PostFile", permission: "file:write" },
  ],
  status_tones: { ok: "ok", ready: "ok", skipped: "warn", failed: "err" },
  count_tones: { completed: "ok", skipped: "warn", failed: "err" },
  host_families: ["debian", "darwin", "fedora", "suse"],
  targets: [
    {
      value: "_all",
      label: "All agents",
      doc: "Each agent answers, one row for each host.",
    },
    {
      value: "_any",
      label: "Any agent",
      doc: "One agent, any that is free, answers.",
    },
  ],
  controller: [
    {
      group: "file",
      method: "list",
      doc: "List stored files",
      operation: "GetFiles",
      permission: "file:read",
    },
    {
      group: "file",
      method: "upload",
      doc: "Upload a file",
      operation: "PostFile",
      permission: "file:write",
    },
    {
      group: "file",
      method: "get",
      doc: "Get a file",
      operation: "GetFileByName",
      permission: "file:read",
    },
    {
      group: "file",
      method: "delete",
      doc: "Delete a file",
      operation: "DeleteFileByName",
      permission: "file:write",
    },
    {
      group: "file",
      method: "stale",
      doc: "List the stale deployments",
      operation: "GetFileStale",
      permission: "file:read",
    },
  ],
  render_tags: [
    "-",
    "bytes",
    "age",
    "time",
    "duration",
    "duration=ms",
    "status",
    "body",
    "short=",
  ],
  outcomes: ["ok", "changed", "skipped", "timeout", "failed"],
};

/** The test catalog, indexed as useCatalog gives it. */
export const testIndex = buildIndex(testCatalog);
