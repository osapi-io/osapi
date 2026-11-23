import { useState, useCallback, useRef } from "react";
import { useVimNav } from "@/hooks/use-vim-nav";
import { useBlockFocus } from "@/hooks/use-block-focus";
import { ContentArea } from "@/components/layout/content-area";
import { BlockCard } from "@/components/domain/block-card";
import { BlockStack } from "@/components/domain/block-stack";
import { ResultCard } from "@/components/domain/result-card";
import { RunLinks } from "@/components/domain/run-links";
import { traceIdOf } from "@/lib/trace";
import { FileUploadBlock } from "@/components/domain/file-upload-block";
import { SingleInputBlock } from "@/components/domain/single-input-block";
import { FileDeleteBlock } from "@/components/domain/file-delete-block";
import { ApplyButton } from "@/components/domain/apply-button";
import {
  useStack,
  ALL_BLOCK_TYPES,
  BLOCK_AREAS,
  blockRunnable,
  type BlockStatus,
} from "@/hooks/use-stack";
import { useCommands } from "@/hooks/use-commands";
import {
  postFile,
  getFiles,
  deleteFileByName,
  getFileStale,
} from "@/sdk/gen/file-management-api-file-operations/file-management-api-file-operations";
import {
  getAuditLogs,
  getAuditLogByID,
  getAuditExport,
} from "@/sdk/gen/audit-log-api-audit/audit-log-api-audit";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/ui/page-header";
import { SectionLabel } from "@/components/ui/section-label";
import { EmptyState } from "@/components/ui/empty-state";
import { Text } from "@/components/ui/text";
import { Chip } from "@/components/ui/chip";
import { useAuth } from "@/hooks/use-auth";
import {
  BLOCK_PERMISSIONS,
  blockAllowed,
  blockKind,
  blockKindLabels,
} from "@/lib/permissions";
import { BlockKindIcon } from "@/components/domain/block-kind";
import { BEHAVIOR_BLOCKS, type BehaviorBlockType } from "@/lib/blocks.gen";
import { runBehavior } from "@/lib/behavior-block";
import { BehaviorBlockForm } from "@/components/domain/behavior-block-form";
import {
  Calendar,
  Container,
  Terminal,
  Lock,
  Server,
  Cog,
  Package,
  SlidersHorizontal,
  Clock,
  Globe,
  Tag,
  Power,
  Activity,
  Network,
  GitBranch,
  Users,
  UsersRound,
  FolderOpen,
  Monitor,
  HardDrive,
  MemoryStick,
  ClipboardList,
  FileDown,
  Gauge,
  Timer,
  Info,
  RefreshCw,
} from "lucide-react";

const blockIcons: Record<
  string,
  React.ComponentType<{ className?: string }>
> = {
  // Services
  "service-list": Cog,
  // Software
  "package-list": Package,
  "updatepackages-list": RefreshCw,
  // Config
  "schedule-list": Calendar,
  "sysctl-list": SlidersHorizontal,
  "ntp-get": Clock,
  "timezone-get": Globe,
  "hostname-get": Tag,
  "file-status": FileDown,
  // Node
  "status-get": Monitor,
  "load-get": Gauge,
  "uptime-get": Timer,
  "os-get": Info,
  "power-reboot": Power,
  "process-list": Activity,
  // Networking
  "dns-get": Server,
  "interface-list": Network,
  "route-list": GitBranch,
  // Security
  "user-list": Users,
  "group-list": UsersRound,
  // Containers
  "docker-list": Container,
  // Command
  "command-exec": Terminal,
  // Hardware
  "disk-get": HardDrive,
  "memory-get": MemoryStick,
  // Files
  "object-list": FolderOpen,
  // Audit
  "audit-list": ClipboardList,
};

export function Configure() {
  const { can } = useAuth();
  const {
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
  } = useStack();
  const [applying, setApplying] = useState(false);

  // After Apply, what the run did, counted from the blocks' states.
  const applied = blocks.filter((b) => b.status === "applied").length;
  const failed = blocks.filter((b) => b.status === "error").length;
  const applySummary = hasApplied
    ? `${applied + failed} applied: ${applied} ok${failed ? `, ${failed} failed` : ""}`
    : `${blocks.length} block${blocks.length === 1 ? "" : "s"}`;
  const [activeArea, setActiveArea] = useState("services");
  const [focusedId, setFocusedId] = useState<string | null>(null);
  const blockRefs = useRef<Map<string, HTMLDivElement>>(new Map());
  const focusBlock = useBlockFocus(blockRefs);

  const resultRefs = useRef<Map<string, HTMLDivElement>>(new Map());

  const scrollToBlock = useCallback(async (id: string) => {
    await new Promise((r) => setTimeout(r, 50));
    const el = blockRefs.current.get(id);
    if (!el) return;
    el.scrollIntoView({ behavior: "smooth", block: "center" });
  }, []);

  const addBlockAndFocus = useCallback(
    (bt: Parameters<typeof addBlock>[0]) => {
      const newId = addBlock(bt);
      // Focus the new block, which brings it into view.
      setFocusedId(newId);
      focusBlock(newId);
    },
    [addBlock, focusBlock],
  );

  const scrollToResult = useCallback(async (id: string) => {
    await new Promise((r) => setTimeout(r, 150));
    const el = resultRefs.current.get(id);
    if (el) {
      // Pin result top ~30% from viewport top — enough to see the
      // block above while giving max room for host rows below
      const y =
        el.getBoundingClientRect().top +
        window.scrollY -
        window.innerHeight * 0.3;
      window.scrollTo({ top: y, behavior: "smooth" });
    } else {
      // No result card (error case) — just re-center the block
      blockRefs.current.get(id)?.scrollIntoView({
        behavior: "smooth",
        block: "center",
      });
    }
  }, []);

  const handleApply = useCallback(async () => {
    setApplying(true);

    for (const block of blocks) {
      if (!blockRunnable(block)) continue;

      // Focus and scroll to the executing block
      setFocusedId(block.id);
      await scrollToBlock(block.id);

      setBlockStatus(block.id, "applying");

      try {
        let result: { data: unknown };

        const t = block.target;

        const behaviorDef = BEHAVIOR_BLOCKS[block.type as BehaviorBlockType];
        if (behaviorDef) {
          result = await runBehavior(behaviorDef, t, block.data);
          setBlockStatus(
            block.id,
            "applied",
            undefined,
            result.data,
            traceIdOf(result),
          );
          await scrollToResult(block.id);
          continue;
        }

        switch (block.type) {
          case "object-list":
            result = await getFiles();
            break;
          case "object-upload": {
            const file = block.data._file as globalThis.File;
            result = await postFile({
              name: block.data.name as string,
              content_type: ((block.data.content_type as string) || "raw") as
                "raw" | "template",
              file,
            });
            break;
          }
          case "object-delete":
            result = await deleteFileByName(block.data.name as string);
            break;
          case "object-stale":
            result = await getFileStale();
            break;
          case "audit-list":
            result = await getAuditLogs();
            break;
          case "audit-get":
            result = await getAuditLogByID(block.data.id as string);
            break;
          case "audit-export":
            result = await getAuditExport();
            break;
          default:
            continue;
        }

        setBlockStatus(
          block.id,
          "applied",
          undefined,
          result.data,
          traceIdOf(result),
        );
      } catch (err) {
        setBlockStatus(
          block.id,
          "error",
          err instanceof Error ? err.message : "Unknown error",
          undefined,
          (err as { traceId?: string } | undefined)?.traceId,
        );
      }

      // Wait for result card to render, then scroll result into view
      await scrollToResult(block.id);
    }

    setFocusedId(null);
    setApplying(false);
  }, [blocks, setBlockStatus, scrollToBlock, scrollToResult]);

  // Register commands for the / palette
  useCommands(
    [
      {
        id: "cmd:run",
        name: "run",
        description: "Apply all blocks",
        category: "actions",
        action: () => {
          if (canApply && !applying) handleApply();
        },
      },
      {
        id: "cmd:clear",
        name: "clear",
        description: "Remove all blocks",
        category: "actions",
        action: () => clearBlocks(),
      },
      ...ALL_BLOCK_TYPES.map((bt) => ({
        id: `block:${bt.type}`,
        name: `${bt.category} ${bt.label.toLowerCase()}`,
        description: bt.description,
        category: "blocks",
        action: () => addBlockAndFocus(bt),
      })),
    ],
    [
      canApply,
      applying,
      handleApply,
      resetBlocks,
      clearBlocks,
      addBlockAndFocus,
    ],
  );

  // Vim-style keyboard navigation
  useVimNav({
    items: blocks,
    focusedId,
    setFocusedId,
    onDelete: removeBlock,
    onExecute: handleApply,
    canExecute: canApply,
    executing: applying,
    itemRefs: blockRefs,
  });

  const renderBlockForm = (block: (typeof blocks)[0]) => {
    const behaviorDef = BEHAVIOR_BLOCKS[block.type as BehaviorBlockType];
    if (behaviorDef) {
      return (
        <BehaviorBlockForm
          blockId={block.id}
          def={behaviorDef}
          data={block.data}
          onChange={(data) => updateBlockData(block.id, data)}
          onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
        />
      );
    }

    switch (block.type) {
      case "object-list":
        return null;
      case "object-upload":
        return (
          <FileUploadBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
          />
        );
      case "object-delete":
        return (
          <FileDeleteBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
          />
        );
      case "object-stale":
        return null;
      case "audit-list":
      case "audit-export":
        return null;
      case "audit-get":
        return (
          <SingleInputBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
            field="id"
            label="Audit Entry ID"
            placeholder="abc123"
          />
        );
      default:
        return (
          <Text variant="muted" size="sm" className="block py-4 text-center">
            Coming soon
          </Text>
        );
    }
  };

  const handleResultAction = useCallback(
    async (action: string, name: string) => {
      try {
        if (action === "object-delete") {
          await deleteFileByName(name);
        }
      } catch {
        // silent — user will see it didn't disappear
      }
    },
    [],
  );

  // Build interleaved list: [block, result?, block, result?, ...]
  const stackItems: React.ReactNode[] = [];
  for (let i = 0; i < blocks.length; i++) {
    const block = blocks[i];
    stackItems.push(
      <BlockCard
        key={block.id}
        ref={(el) => {
          if (el) blockRefs.current.set(block.id, el);
          else blockRefs.current.delete(block.id);
        }}
        label={block.label}
        description={block.description}
        kind={blockKind(block.type)}
        // A block that waits for its target is not ready to apply.
        status={
          block.status === "ready" && !blockRunnable(block)
            ? "pending"
            : block.status
        }
        focused={focusedId === block.id}
        error={block.error}
        target={block.target}
        onTargetChange={(t) => setBlockTarget(block.id, t)}
        onRemove={() => removeBlock(block.id)}
        onFocusCard={() => setFocusedId(block.id)}
      >
        {renderBlockForm(block)}
      </BlockCard>,
    );
    // The job and the trace of the call, also for a call that failed.
    const jobId = (block.result as { job_id?: unknown } | undefined)?.job_id;
    if (block.result != null || block.traceId) {
      stackItems.push(
        <RunLinks
          key={`${block.id}-links`}
          jobId={typeof jobId === "string" ? jobId : undefined}
          traceId={block.traceId}
          className="-mt-1 mb-1 px-1"
        />,
      );
    }
    if (block.result != null) {
      stackItems.push(
        <div
          key={`${block.id}-result`}
          ref={(el) => {
            if (el) resultRefs.current.set(block.id, el);
            else resultRefs.current.delete(block.id);
          }}
        >
          <ResultCard
            type={block.type}
            result={block.result}
            onAction={handleResultAction}
          />
        </div>,
      );
    }
  }

  return (
    <ContentArea>
      <PageHeader
        title="Configure"
        subtitle="Add blocks to a stack, then apply them to your hosts"
      />

      <div className="grid grid-cols-12 gap-6">
        {/* Sidebar */}
        {/* The panel stays in view as the stack grows, and scrolls on its
            own when it is taller than the window, so each chip is reachable. */}
        <aside className="col-span-3 space-y-3 self-start sticky top-20 max-h-[calc(100vh-6rem)] overflow-y-auto pr-1">
          <div className="flex items-center justify-between">
            <SectionLabel>Add Block</SectionLabel>
            {blocks.length > 0 && (
              <Badge
                key={blocks.length}
                variant={canApply ? "ready" : "pending"}
                className="animate-pulse-once"
              >
                {blocks.length} block{blocks.length !== 1 ? "s" : ""}
              </Badge>
            )}
          </div>

          {/* Area pills */}
          <div className="flex flex-wrap gap-1.5">
            {BLOCK_AREAS.map((area) => (
              <Chip
                key={area.name}
                size="sm"
                active={activeArea === area.name}
                onClick={() => setActiveArea(area.name)}
              >
                {area.label}
              </Chip>
            ))}
          </div>

          {/* Categories for the active area */}
          {BLOCK_AREAS.find((a) => a.name === activeArea)?.categories.map(
            (cat) => {
              const CatIcon = blockIcons[cat.types[0]?.type] || Terminal;
              return (
                <div key={cat.name}>
                  <div className="mb-2 flex items-center gap-2 px-1">
                    <CatIcon className="h-4 w-4 text-text-muted" />
                    <Text size="sm" className="font-semibold">
                      {cat.label}
                    </Text>
                  </div>
                  <div className="flex flex-wrap gap-1.5">
                    {cat.types.map((bt) => {
                      const perm = BLOCK_PERMISSIONS[bt.type];
                      const allowed = blockAllowed(bt.type, can);
                      const kind = blockKind(bt.type);
                      return (
                        <Chip
                          key={bt.type}
                          onClick={(e) => {
                            (e.target as HTMLElement).blur();
                            addBlockAndFocus(bt);
                          }}
                          disabled={!allowed}
                          title={
                            allowed
                              ? `${kind ? `${blockKindLabels[kind]}: ` : ""}${bt.description}`
                              : `Requires ${perm ?? "a permission that is not known"}`
                          }
                        >
                          {kind && <BlockKindIcon kind={kind} />}
                          {bt.label}
                          {!allowed && (
                            <Lock className="h-3 w-3 text-text-muted/40" />
                          )}
                        </Chip>
                      );
                    })}
                  </div>
                </div>
              );
            },
          )}
        </aside>

        {/* Main */}
        <main className="col-span-9">
          {blocks.length === 0 ? (
            <EmptyState message="Add blocks from the panel to build a stack" />
          ) : (
            <>
              <BlockStack>{stackItems}</BlockStack>
              <ApplyButton
                summary={applySummary}
                onClear={clearBlocks}
                ready={canApply}
                applying={applying}
                hasApplied={hasApplied}
                onApply={handleApply}
                onReset={resetBlocks}
              />
              <div className="h-24" />
            </>
          )}
        </main>
      </div>
    </ContentArea>
  );
}
