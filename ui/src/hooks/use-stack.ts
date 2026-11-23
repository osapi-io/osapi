import { useState, useCallback } from "react";
import { BEHAVIOR_BLOCKS } from "@/lib/blocks.gen";
import { blockKind, type HandBlockType } from "@/lib/permissions";

export type BlockStatus =
  "unchecked" | "pending" | "ready" | "applying" | "applied" | "error";

export interface BlockType {
  type: string;
  label: string;
  description: string;
  category: string;
  /** The area of the Add Block panel that the block goes in. */
  area: string;
  /** Blocks that need no input are ready immediately */
  readyOnAdd?: boolean;
}

export interface BlockCategory {
  name: string;
  label: string;
  types: BlockType[];
}

export interface BlockArea {
  name: string;
  label: string;
  categories: BlockCategory[];
}

// The blocks that are not a behavior's. Each has an operation in
// BLOCK_OPERATIONS, which gives its permission: a type with none fails the
// type check here.
const HAND_BLOCK_TYPES: (BlockType & { type: HandBlockType })[] = [
  // ── Files ──────────────────────────────────────────────────────────────────
  {
    type: "object-list",
    label: "List",
    description: "List files in object store",
    category: "file",
    area: "files",
    readyOnAdd: true,
  },
  {
    type: "object-upload",
    label: "Upload",
    description: "Upload to object store",
    category: "file",
    area: "files",
  },
  {
    type: "object-delete",
    label: "Delete",
    description: "Delete file from object store",
    category: "file",
    area: "files",
  },
  {
    type: "object-stale",
    label: "Stale",
    description: "Check for stale deployments",
    category: "file-deploy",
    area: "files",
    readyOnAdd: true,
  },

  // ── Audit ──────────────────────────────────────────────────────────────────
  {
    type: "audit-list",
    label: "List",
    description: "List audit log entries",
    category: "audit",
    area: "audit",
    readyOnAdd: true,
  },
  {
    type: "audit-get",
    label: "Get",
    description: "Get audit entry details",
    category: "audit",
    area: "audit",
  },
  {
    type: "audit-export",
    label: "Export",
    description: "Export audit log",
    category: "audit",
    area: "audit",
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
    area: def.area,
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
  file: "File",
  "file-deploy": "File Deploy",
  audit: "Audit",
};

const behaviorTitles: Record<string, string> = Object.fromEntries(
  Object.values(BEHAVIOR_BLOCKS).map((def) => [def.behavior, def.title]),
);

function buildCategories(area: string): BlockCategory[] {
  const types = ALL_BLOCK_TYPES.filter((b) => b.area === area);
  const cats = [...new Set(types.map((b) => b.category))];
  return cats.map((cat) => ({
    name: cat,
    label: categoryNames[cat] || behaviorTitles[cat] || cat,
    types: types.filter((b) => b.category === cat),
  }));
}

/**
 * The areas of the panel: the AreaName of each behavior, from behavior.Area
 * in its definition, and the areas of the hand-written blocks. A new
 * AreaName fails the type check here until the panel has it.
 */
type BehaviorArea =
  (typeof BEHAVIOR_BLOCKS)[keyof typeof BEHAVIOR_BLOCKS]["area"];
type PanelArea = BehaviorArea | "files" | "audit";
type PanelAreaName =
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
const areaInPanel = (a: PanelArea): PanelAreaName => a;
void areaInPanel;

export const BLOCK_AREAS: (BlockArea & { name: PanelAreaName })[] = [
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
export const BLOCK_CATEGORIES: BlockCategory[] = BLOCK_AREAS.flatMap(
  (a) => a.categories,
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

/**
 * The target that a new block starts with. A read of a behavior may run on
 * each agent, so it starts with _all. A write or an action starts with none:
 * the user picks the hosts that it changes, and the block does not apply
 * until then. A block that is not a behavior's runs on the controller, and
 * its target is not used.
 */
export function defaultTarget(type: string): string {
  return type in BEHAVIOR_BLOCKS && blockKind(type) !== "read" ? "" : "_all";
}

/** Whether Apply runs the block: it is ready and it has a target. */
export function blockRunnable(block: Block): boolean {
  return block.status === "ready" && block.target !== "";
}

/** A new block of a type, as the panel adds it. */
export function newBlock(blockType: BlockType, id: string): Block {
  return {
    id,
    type: blockType.type,
    label: blockType.label,
    description: blockType.description,
    status: blockType.readyOnAdd ? "ready" : "pending",
    target: defaultTarget(blockType.type),
    data: {},
  };
}

export function useStack() {
  const [blocks, setBlocks] = useState<Block[]>([]);

  const addBlock = useCallback((blockType: BlockType): string => {
    const id = `${blockType.type}-${nextId++}`;
    setBlocks((prev) => [...prev, newBlock(blockType, id)]);
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

  const clearBlocks = useCallback(() => {
    setBlocks([]);
  }, []);

  const readyBlocks = blocks.filter(blockRunnable);
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
    clearBlocks,
  };
}
