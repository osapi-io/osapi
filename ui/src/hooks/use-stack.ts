import { useState, useCallback } from "react";

export type BlockStatus =
  | "unchecked"
  | "pending"
  | "ready"
  | "applying"
  | "applied"
  | "error";

export interface BlockType {
  type: string;
  label: string;
  description: string;
  category: string;
  group: string;
  /** Blocks that need no input are ready immediately */
  readyOnAdd?: boolean;
}

export interface BlockCategory {
  name: string;
  label: string;
  types: BlockType[];
}

export interface BlockGroup {
  name: string;
  label: string;
  categories: BlockCategory[];
}

export const ALL_BLOCK_TYPES: BlockType[] = [
  // ── Services ──────────────────────────────────────────────────────────────
  {
    type: "cron-create",
    label: "Create",
    description: "Create a cron job",
    category: "cron",
    group: "services",
  },
  {
    type: "cron-list",
    label: "List",
    description: "List cron jobs",
    category: "cron",
    group: "services",
    readyOnAdd: true,
  },
  {
    type: "cron-delete",
    label: "Delete",
    description: "Delete a cron job",
    category: "cron",
    group: "services",
  },
  {
    type: "cron-get",
    label: "Get",
    description: "Get cron job details",
    category: "cron",
    group: "services",
  },
  {
    type: "cron-update",
    label: "Update",
    description: "Update a cron job",
    category: "cron",
    group: "services",
  },
  {
    type: "service-list",
    label: "List",
    description: "List systemd services",
    category: "service",
    group: "services",
    readyOnAdd: true,
  },
  {
    type: "service-get",
    label: "Get",
    description: "Get service details",
    category: "service",
    group: "services",
  },
  {
    type: "service-create",
    label: "Create",
    description: "Create a systemd unit",
    category: "service",
    group: "services",
  },
  {
    type: "service-update",
    label: "Update",
    description: "Update a systemd unit",
    category: "service",
    group: "services",
  },
  {
    type: "service-delete",
    label: "Delete",
    description: "Delete a systemd unit",
    category: "service",
    group: "services",
  },
  {
    type: "service-start",
    label: "Start",
    description: "Start a service",
    category: "service",
    group: "services",
  },
  {
    type: "service-stop",
    label: "Stop",
    description: "Stop a service",
    category: "service",
    group: "services",
  },
  {
    type: "service-restart",
    label: "Restart",
    description: "Restart a service",
    category: "service",
    group: "services",
  },
  {
    type: "service-enable",
    label: "Enable",
    description: "Enable a service at boot",
    category: "service",
    group: "services",
  },
  {
    type: "service-disable",
    label: "Disable",
    description: "Disable a service at boot",
    category: "service",
    group: "services",
  },

  // ── Software ───────────────────────────────────────────────────────────────
  {
    type: "package-list",
    label: "List",
    description: "List installed packages",
    category: "package",
    group: "software",
    readyOnAdd: true,
  },
  {
    type: "package-get",
    label: "Get",
    description: "Get package details",
    category: "package",
    group: "software",
  },
  {
    type: "package-install",
    label: "Install",
    description: "Install a package",
    category: "package",
    group: "software",
  },
  {
    type: "package-remove",
    label: "Remove",
    description: "Remove a package",
    category: "package",
    group: "software",
  },
  {
    type: "package-update",
    label: "Update",
    description: "Update a package",
    category: "package",
    group: "software",
  },
  {
    type: "package-check-updates",
    label: "Check Updates",
    description: "Check for available updates",
    category: "package",
    group: "software",
    readyOnAdd: true,
  },

  // ── Config ─────────────────────────────────────────────────────────────────
  {
    type: "sysctl-list",
    label: "List",
    description: "List sysctl parameters",
    category: "sysctl",
    group: "config",
    readyOnAdd: true,
  },
  {
    type: "sysctl-get",
    label: "Get",
    description: "Get a sysctl parameter",
    category: "sysctl",
    group: "config",
  },
  {
    type: "sysctl-set",
    label: "Set",
    description: "Create a sysctl parameter",
    category: "sysctl",
    group: "config",
  },
  {
    type: "sysctl-update",
    label: "Update",
    description: "Update a sysctl parameter",
    category: "sysctl",
    group: "config",
  },
  {
    type: "sysctl-delete",
    label: "Delete",
    description: "Delete a sysctl parameter",
    category: "sysctl",
    group: "config",
  },
  {
    type: "ntp-get",
    label: "Get",
    description: "Get NTP configuration",
    category: "ntp",
    group: "config",
    readyOnAdd: true,
  },
  {
    type: "ntp-set",
    label: "Set",
    description: "Configure NTP servers",
    category: "ntp",
    group: "config",
  },
  {
    type: "ntp-update",
    label: "Update",
    description: "Update NTP configuration",
    category: "ntp",
    group: "config",
  },
  {
    type: "ntp-delete",
    label: "Delete",
    description: "Remove NTP configuration",
    category: "ntp",
    group: "config",
  },
  {
    type: "timezone-get",
    label: "Get",
    description: "Get system timezone",
    category: "timezone",
    group: "config",
    readyOnAdd: true,
  },
  {
    type: "timezone-set",
    label: "Set",
    description: "Set system timezone",
    category: "timezone",
    group: "config",
  },
  {
    type: "hostname-get",
    label: "Get",
    description: "Get system hostname",
    category: "hostname",
    group: "config",
    readyOnAdd: true,
  },
  {
    type: "hostname-set",
    label: "Set",
    description: "Set system hostname",
    category: "hostname",
    group: "config",
  },

  // ── Node ──────────────────────────────────────────────────────────────────
  {
    type: "node-status",
    label: "Status",
    description: "Get node status and conditions",
    category: "node-info",
    group: "node",
    readyOnAdd: true,
  },
  {
    type: "node-load",
    label: "Load",
    description: "Get load averages",
    category: "node-info",
    group: "node",
    readyOnAdd: true,
  },
  {
    type: "node-uptime",
    label: "Uptime",
    description: "Get system uptime",
    category: "node-info",
    group: "node",
    readyOnAdd: true,
  },
  {
    type: "node-os",
    label: "OS",
    description: "Get OS information",
    category: "node-info",
    group: "node",
    readyOnAdd: true,
  },
  {
    type: "power-reboot",
    label: "Reboot",
    description: "Reboot the host",
    category: "power",
    group: "node",
  },
  {
    type: "power-shutdown",
    label: "Shutdown",
    description: "Shutdown the host",
    category: "power",
    group: "node",
  },
  {
    type: "process-list",
    label: "List",
    description: "List running processes",
    category: "process",
    group: "node",
    readyOnAdd: true,
  },
  {
    type: "process-get",
    label: "Get",
    description: "Get process details",
    category: "process",
    group: "node",
  },
  {
    type: "process-signal",
    label: "Signal",
    description: "Send signal to a process",
    category: "process",
    group: "node",
  },
  {
    type: "log-query",
    label: "Query",
    description: "Query system logs",
    category: "log",
    group: "node",
  },
  {
    type: "log-sources",
    label: "Sources",
    description: "List log sources",
    category: "log",
    group: "node",
    readyOnAdd: true,
  },
  {
    type: "log-query-unit",
    label: "Query Unit",
    description: "Query logs for a systemd unit",
    category: "log",
    group: "node",
  },

  // ── Networking ─────────────────────────────────────────────────────────────
  {
    type: "dns-list",
    label: "List",
    description: "List DNS servers for an interface",
    category: "dns",
    group: "networking",
  },
  {
    type: "dns-update",
    label: "Update",
    description: "Update DNS servers for an interface",
    category: "dns",
    group: "networking",
  },
  {
    type: "dns-delete",
    label: "Delete",
    description: "Delete DNS configuration",
    category: "dns",
    group: "networking",
  },
  {
    type: "ping",
    label: "Ping",
    description: "Ping a remote host",
    category: "network",
    group: "networking",
  },
  {
    type: "interface-list",
    label: "List",
    description: "List network interfaces",
    category: "interface",
    group: "networking",
    readyOnAdd: true,
  },
  {
    type: "interface-get",
    label: "Get",
    description: "Get interface details",
    category: "interface",
    group: "networking",
  },
  {
    type: "interface-create",
    label: "Create",
    description: "Create a network interface",
    category: "interface",
    group: "networking",
  },
  {
    type: "interface-update",
    label: "Update",
    description: "Update a network interface",
    category: "interface",
    group: "networking",
  },
  {
    type: "interface-delete",
    label: "Delete",
    description: "Delete a network interface",
    category: "interface",
    group: "networking",
  },
  {
    type: "route-list",
    label: "List",
    description: "List network routes",
    category: "route",
    group: "networking",
    readyOnAdd: true,
  },
  {
    type: "route-get",
    label: "Get",
    description: "Get route details",
    category: "route",
    group: "networking",
  },
  {
    type: "route-create",
    label: "Create",
    description: "Create a network route",
    category: "route",
    group: "networking",
  },
  {
    type: "route-update",
    label: "Update",
    description: "Update a network route",
    category: "route",
    group: "networking",
  },
  {
    type: "route-delete",
    label: "Delete",
    description: "Delete a network route",
    category: "route",
    group: "networking",
  },

  // ── Security ───────────────────────────────────────────────────────────────
  {
    type: "user-list",
    label: "List",
    description: "List user accounts",
    category: "user",
    group: "security",
    readyOnAdd: true,
  },
  {
    type: "user-get",
    label: "Get",
    description: "Get user details",
    category: "user",
    group: "security",
  },
  {
    type: "user-create",
    label: "Create",
    description: "Create a user account",
    category: "user",
    group: "security",
  },
  {
    type: "user-update",
    label: "Update",
    description: "Update a user account",
    category: "user",
    group: "security",
  },
  {
    type: "user-delete",
    label: "Delete",
    description: "Delete a user account",
    category: "user",
    group: "security",
  },
  {
    type: "user-list-keys",
    label: "List Keys",
    description: "List SSH keys for a user",
    category: "user",
    group: "security",
  },
  {
    type: "user-add-key",
    label: "Add Key",
    description: "Add an SSH key to a user",
    category: "user",
    group: "security",
  },
  {
    type: "user-remove-key",
    label: "Remove Key",
    description: "Remove an SSH key from a user",
    category: "user",
    group: "security",
  },
  {
    type: "user-change-password",
    label: "Password",
    description: "Change a user password",
    category: "user",
    group: "security",
  },
  {
    type: "group-list",
    label: "List",
    description: "List groups",
    category: "group",
    group: "security",
    readyOnAdd: true,
  },
  {
    type: "group-get",
    label: "Get",
    description: "Get group details",
    category: "group",
    group: "security",
  },
  {
    type: "group-create",
    label: "Create",
    description: "Create a group",
    category: "group",
    group: "security",
  },
  {
    type: "group-update",
    label: "Update",
    description: "Update a group",
    category: "group",
    group: "security",
  },
  {
    type: "group-delete",
    label: "Delete",
    description: "Delete a group",
    category: "group",
    group: "security",
  },
  {
    type: "certificate-list",
    label: "List",
    description: "List CA certificates",
    category: "certificate",
    group: "security",
    readyOnAdd: true,
  },
  {
    type: "certificate-create",
    label: "Create",
    description: "Add a CA certificate",
    category: "certificate",
    group: "security",
  },
  {
    type: "certificate-update",
    label: "Update",
    description: "Update a CA certificate",
    category: "certificate",
    group: "security",
  },
  {
    type: "certificate-delete",
    label: "Delete",
    description: "Remove a CA certificate",
    category: "certificate",
    group: "security",
  },

  // ── Containers ─────────────────────────────────────────────────────────────
  {
    type: "docker-create",
    label: "Create",
    description: "Create a container",
    category: "docker",
    group: "containers",
  },
  {
    type: "docker-list",
    label: "List",
    description: "List containers",
    category: "docker",
    group: "containers",
    readyOnAdd: true,
  },
  {
    type: "docker-start",
    label: "Start",
    description: "Start a container",
    category: "docker",
    group: "containers",
  },
  {
    type: "docker-stop",
    label: "Stop",
    description: "Stop a container",
    category: "docker",
    group: "containers",
  },
  {
    type: "docker-delete",
    label: "Delete",
    description: "Delete a container",
    category: "docker",
    group: "containers",
  },
  {
    type: "docker-exec",
    label: "Exec",
    description: "Execute a command in a container",
    category: "docker",
    group: "containers",
  },
  {
    type: "docker-pull",
    label: "Pull",
    description: "Pull a container image",
    category: "docker",
    group: "containers",
  },
  {
    type: "docker-rm-image",
    label: "Remove Image",
    description: "Remove a container image",
    category: "docker",
    group: "containers",
  },
  {
    type: "docker-inspect",
    label: "Inspect",
    description: "Inspect a container",
    category: "docker",
    group: "containers",
  },

  // ── Files ──────────────────────────────────────────────────────────────────
  {
    type: "file-list",
    label: "List",
    description: "List files in object store",
    category: "file",
    group: "files",
    readyOnAdd: true,
  },
  {
    type: "file-upload",
    label: "Upload",
    description: "Upload to object store",
    category: "file",
    group: "files",
  },
  {
    type: "file-deploy",
    label: "Deploy",
    description: "Deploy file to hosts",
    category: "file-deploy",
    group: "files",
  },
  {
    type: "file-undeploy",
    label: "Undeploy",
    description: "Remove deployed file",
    category: "file-deploy",
    group: "files",
  },
  {
    type: "file-status",
    label: "Status",
    description: "Check deploy status",
    category: "file-deploy",
    group: "files",
  },
  {
    type: "file-delete",
    label: "Delete",
    description: "Delete file from object store",
    category: "file",
    group: "files",
  },
  {
    type: "file-stale",
    label: "Stale",
    description: "Check for stale deployments",
    category: "file-deploy",
    group: "files",
    readyOnAdd: true,
  },

  // ── Command ────────────────────────────────────────────────────────────────
  {
    type: "command",
    label: "Execute",
    description: "Run a command on hosts",
    category: "command",
    group: "command",
  },
  {
    type: "command-shell",
    label: "Shell",
    description: "Run a shell command on hosts",
    category: "command",
    group: "command",
  },

  // ── Hardware ───────────────────────────────────────────────────────────────
  {
    type: "disk-info",
    label: "Disk",
    description: "Get disk information",
    category: "disk",
    group: "hardware",
    readyOnAdd: true,
  },
  {
    type: "memory-info",
    label: "Memory",
    description: "Get memory information",
    category: "memory",
    group: "hardware",
    readyOnAdd: true,
  },

  // ── Audit ──────────────────────────────────────────────────────────────────
  {
    type: "audit-list",
    label: "List",
    description: "List audit log entries",
    category: "audit",
    group: "audit",
    readyOnAdd: true,
  },
  {
    type: "audit-get",
    label: "Get",
    description: "Get audit entry details",
    category: "audit",
    group: "audit",
  },
  {
    type: "audit-export",
    label: "Export",
    description: "Export audit log",
    category: "audit",
    group: "audit",
    readyOnAdd: true,
  },
];

const categoryNames: Record<string, string> = {
  service: "Service",
  cron: "Cron",
  package: "Package",
  sysctl: "Sysctl",
  ntp: "NTP",
  timezone: "Timezone",
  hostname: "Hostname",
  "node-info": "Node Info",
  power: "Power",
  process: "Process",
  log: "Log",
  dns: "DNS",
  interface: "Interface",
  route: "Route",
  network: "Network",
  user: "User",
  group: "Group",
  certificate: "Certificate",
  docker: "Docker",
  file: "File",
  "file-deploy": "File Deploy",
  command: "Command",
  disk: "Disk",
  memory: "Memory",
  audit: "Audit",
};

function buildCategories(groupName: string): BlockCategory[] {
  const types = ALL_BLOCK_TYPES.filter((b) => b.group === groupName);
  const cats = [...new Set(types.map((b) => b.category))];
  return cats.map((cat) => ({
    name: cat,
    label: categoryNames[cat] || cat,
    types: types.filter((b) => b.category === cat),
  }));
}

export const BLOCK_GROUPS: BlockGroup[] = [
  {
    name: "services",
    label: "Services",
    categories: buildCategories("services"),
  },
  {
    name: "software",
    label: "Software",
    categories: buildCategories("software"),
  },
  { name: "config", label: "Config", categories: buildCategories("config") },
  { name: "node", label: "Node", categories: buildCategories("node") },
  {
    name: "networking",
    label: "Networking",
    categories: buildCategories("networking"),
  },
  {
    name: "security",
    label: "Security",
    categories: buildCategories("security"),
  },
  {
    name: "containers",
    label: "Containers",
    categories: buildCategories("containers"),
  },
  { name: "files", label: "Files", categories: buildCategories("files") },
  { name: "command", label: "Command", categories: buildCategories("command") },
  {
    name: "hardware",
    label: "Hardware",
    categories: buildCategories("hardware"),
  },
  { name: "audit", label: "Audit", categories: buildCategories("audit") },
];

/** Flat list for backwards compat (command palette, etc.) */
export const BLOCK_CATEGORIES: BlockCategory[] = BLOCK_GROUPS.flatMap(
  (g) => g.categories,
);

export interface Block {
  id: string;
  type: string;
  label: string;
  description: string;
  status: BlockStatus;
  target: string;
  data: Record<string, unknown>;
  result?: unknown;
  error?: string;
}

let nextId = 1;

export function useStack() {
  const [blocks, setBlocks] = useState<Block[]>([]);

  const addBlock = useCallback((blockType: BlockType): string => {
    const id = `${blockType.type}-${nextId++}`;
    setBlocks((prev) => [
      ...prev,
      {
        id,
        type: blockType.type,
        label: blockType.label,
        description: blockType.description,
        status: blockType.readyOnAdd ? "ready" : "pending",
        target: "_all",
        data: {},
      },
    ]);
    return id;
  }, []);

  const removeBlock = useCallback((id: string) => {
    setBlocks((prev) => prev.filter((b) => b.id !== id));
  }, []);

  const updateBlockData = useCallback(
    (id: string, data: Record<string, unknown>) => {
      setBlocks((prev) => prev.map((b) => (b.id === id ? { ...b, data } : b)));
    },
    [],
  );

  const setBlockStatus = useCallback(
    (id: string, status: BlockStatus, error?: string, result?: unknown) => {
      setBlocks((prev) =>
        prev.map((b) => (b.id === id ? { ...b, status, error, result } : b)),
      );
    },
    [],
  );

  const setBlockTarget = useCallback((id: string, target: string) => {
    setBlocks((prev) => prev.map((b) => (b.id === id ? { ...b, target } : b)));
  }, []);

  const resetBlocks = useCallback(() => {
    setBlocks((prev) =>
      prev.map((b) => ({
        ...b,
        status: b.data && Object.keys(b.data).length > 0 ? "ready" : "pending",
        result: undefined,
        error: undefined,
      })),
    );
  }, []);

  const loadBlocks = useCallback(
    (
      items: {
        type: string;
        label: string;
        description: string;
        target?: string;
        data: Record<string, unknown>;
      }[],
    ) => {
      setBlocks(
        items.map((item) => {
          const id = `${item.type}-${nextId++}`;
          const hasData = Object.keys(item.data).length > 0;
          return {
            id,
            type: item.type,
            label: item.label,
            description: item.description,
            status: hasData ? "ready" : "pending",
            target: item.target ?? "_all",
            data: { ...item.data },
          };
        }),
      );
    },
    [],
  );

  const clearBlocks = useCallback(() => {
    setBlocks([]);
  }, []);

  const readyBlocks = blocks.filter((b) => b.status === "ready");
  const hasApplied = blocks.some(
    (b) => b.status === "applied" || b.status === "error",
  );
  const canApply = readyBlocks.length > 0;

  return {
    blocks,
    canApply,
    hasApplied,
    addBlock,
    removeBlock,
    updateBlockData,
    setBlockStatus,
    setBlockTarget,
    resetBlocks,
    loadBlocks,
    clearBlocks,
  };
}
