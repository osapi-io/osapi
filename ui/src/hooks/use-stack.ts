import { useState, useCallback } from "react";
import { BEHAVIOR_BLOCKS } from "@/lib/blocks.gen";

export type BlockStatus =
  "unchecked" | "pending" | "ready" | "applying" | "applied" | "error";

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

const HAND_BLOCK_TYPES: BlockType[] = [
  // ── Services ──────────────────────────────────────────────────────────────
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

  // ── Networking ─────────────────────────────────────────────────────────────
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

/** The block of each method of each behavior, from blocks.gen.ts. */
const BEHAVIOR_BLOCK_TYPES: BlockType[] = Object.values(BEHAVIOR_BLOCKS).map(
  (def) => ({
    type: def.type,
    label: def.label,
    description: def.description,
    category: def.behavior,
    group: def.area,
    readyOnAdd: !def.fields.some((f) => f.required),
  }),
);

export const ALL_BLOCK_TYPES: BlockType[] = [
  ...HAND_BLOCK_TYPES,
  ...BEHAVIOR_BLOCK_TYPES,
];

// The names of the categories of the hand-written blocks. A behavior's
// category is named by its title, from blocks.gen.ts, or here when the
// title does not read well.
const categoryNames: Record<string, string> = {
  updatepackages: "Package Index",
  ntp: "NTP",
  power: "Power",
  process: "Process",
  log: "Log",
  network: "Network",
  user: "User",
  group: "Group",
  certificate: "Certificate",
  docker: "Docker",
  file: "File",
  "file-deploy": "File Deploy",
  command: "Command",
  audit: "Audit",
};

const behaviorTitles: Record<string, string> = Object.fromEntries(
  Object.values(BEHAVIOR_BLOCKS).map((def) => [def.behavior, def.title]),
);

function buildCategories(groupName: string): BlockCategory[] {
  const types = ALL_BLOCK_TYPES.filter((b) => b.group === groupName);
  const cats = [...new Set(types.map((b) => b.category))];
  return cats.map((cat) => ({
    name: cat,
    label: categoryNames[cat] || behaviorTitles[cat] || cat,
    types: types.filter((b) => b.category === cat),
  }));
}

/**
 * The area of each behavior, from behavior.Area in its definition, is a
 * group of the panel: a new area fails the type check here until the panel
 * has a group for it.
 */
type BehaviorArea =
  (typeof BEHAVIOR_BLOCKS)[keyof typeof BEHAVIOR_BLOCKS]["area"];
type GroupName =
  | "services"
  | "software"
  | "config"
  | "node"
  | "networking"
  | "security"
  | "containers"
  | "files"
  | "command"
  | "hardware"
  | "audit";
const areaIsGroup = (a: BehaviorArea): GroupName => a;
void areaIsGroup;

export const BLOCK_GROUPS: (BlockGroup & { name: GroupName })[] = [
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
  /** The trace of the block's last call, from the answer's traceresponse header. */
  traceId?: string;
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
    (
      id: string,
      status: BlockStatus,
      error?: string,
      result?: unknown,
      traceId?: string,
    ) => {
      setBlocks((prev) =>
        prev.map((b) =>
          b.id === id ? { ...b, status, error, result, traceId } : b,
        ),
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
        // A block that is ready when added (it takes no input) is ready
        // again; any other block is ready when it has its input.
        status:
          ALL_BLOCK_TYPES.find((t) => t.type === b.type)?.readyOnAdd ||
          (b.data && Object.keys(b.data).length > 0)
            ? "ready"
            : "pending",
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
