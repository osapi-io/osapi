import { useState, useCallback, useRef } from "react";
import { useVimNav } from "@/hooks/use-vim-nav";
import { useBlockFocus } from "@/hooks/use-block-focus";
import { ContentArea } from "@/components/layout/content-area";
import { BlockCard } from "@/components/domain/block-card";
import { BlockStack } from "@/components/domain/block-stack";
import { ResultCard } from "@/components/domain/result-card";
import { ScheduleBlock } from "@/components/domain/schedule-block";
import { CommandBlock } from "@/components/domain/command-block";
import { DockerBlock } from "@/components/domain/docker-block";
import { FileBlock } from "@/components/domain/file-block";
import { FileUploadBlock } from "@/components/domain/file-upload-block";
import { SingleInputBlock } from "@/components/domain/single-input-block";
import { FileDeleteBlock } from "@/components/domain/file-delete-block";
import { ScheduleDeleteBlock } from "@/components/domain/schedule-delete-block";
import { ContainerActionBlock } from "@/components/domain/container-action-block";
import { DockerExecBlock } from "@/components/domain/docker-exec-block";
import { NtpBlock } from "@/components/domain/ntp-block";
import { PowerBlock } from "@/components/domain/power-block";
import { ProcessSignalBlock } from "@/components/domain/process-signal-block";
import { LogQueryBlock } from "@/components/domain/log-query-block";
import { UserBlock } from "@/components/domain/user-block";
import { UserPasswordBlock } from "@/components/domain/user-password-block";
import { UserSSHKeyBlock } from "@/components/domain/user-ssh-key-block";
import { UserRemoveKeyBlock } from "@/components/domain/user-remove-key-block";
import { GroupBlock } from "@/components/domain/group-block";
import { CertificateBlock } from "@/components/domain/certificate-block";
import { ApplyButton } from "@/components/domain/apply-button";
import { StackBar } from "@/components/domain/stack-bar";
import { SaveStackDialog } from "@/components/domain/save-stack-dialog";
import {
  useStack,
  ALL_BLOCK_TYPES,
  BLOCK_GROUPS,
  type BlockStatus,
} from "@/hooks/use-stack";
import { useCommands } from "@/hooks/use-commands";
import { useStacks } from "@/hooks/use-stacks";
import { features } from "@/lib/features";
import {
  postNodeSchedule,
  getNodeSchedule,
  deleteNodeSchedule,
  getNodeScheduleByName,
  putNodeSchedule,
} from "@/sdk/gen/schedule-management-api-schedule-operations/schedule-management-api-schedule-operations";
import {
  postNodeCommandExec,
  postNodeCommandShell,
} from "@/sdk/gen/command-execution-api-command-operations/command-execution-api-command-operations";
import { postNodeContainerDockerExec } from "@/sdk/gen/docker-management-api-docker-exec/docker-management-api-docker-exec";
import {
  postNodeContainerDockerPull,
  deleteNodeContainerDockerImage,
} from "@/sdk/gen/docker-management-api-docker-image/docker-management-api-docker-image";
import { postNodeNetworkPing } from "@/sdk/gen/network-management-api-network-operations/network-management-api-network-operations";
import {
  getNodeNtp,
  postNodeNtp,
  putNodeNtp,
  deleteNodeNtp,
} from "@/sdk/gen/ntp-management-api-ntp-operations/ntp-management-api-ntp-operations";
import {
  postNodePowerReboot,
  postNodePowerShutdown,
} from "@/sdk/gen/power-management-api-power-operations/power-management-api-power-operations";
import {
  getNodeProcess,
  getNodeProcessByPid,
  postNodeProcessSignal,
} from "@/sdk/gen/process-management-api-process-operations/process-management-api-process-operations";
import {
  getNodeLog,
  getNodeLogSource,
  getNodeLogUnit,
} from "@/sdk/gen/log-management-api-log-operations/log-management-api-log-operations";
import {
  postNodeContainerDocker,
  getNodeContainerDocker,
  deleteNodeContainerDockerByID,
  postNodeContainerDockerStart,
  postNodeContainerDockerStop,
  getNodeContainerDockerByID,
} from "@/sdk/gen/docker-management-api-docker-operations/docker-management-api-docker-operations";
import {
  postFile,
  getFiles,
  deleteFileByName,
  getFileStale,
} from "@/sdk/gen/file-management-api-file-operations/file-management-api-file-operations";
import {
  getNodeUser,
  getNodeUserByName,
  postNodeUser,
  putNodeUser,
  deleteNodeUser,
  getNodeUserSSHKey,
  postNodeUserSSHKey,
  deleteNodeUserSSHKey,
  postNodeUserPassword,
} from "@/sdk/gen/user-and-group-management-api-user-operations/user-and-group-management-api-user-operations";
import {
  getNodeGroup,
  getNodeGroupByName,
  postNodeGroup,
  putNodeGroup,
  deleteNodeGroup,
} from "@/sdk/gen/user-and-group-management-api-group-operations/user-and-group-management-api-group-operations";
import {
  getNodeCertificateCa,
  postNodeCertificateCa,
  putNodeCertificateCa,
  deleteNodeCertificateCa,
} from "@/sdk/gen/certificate-management-api-certificate-operations/certificate-management-api-certificate-operations";
import {
  postNodeFileDeploy,
  postNodeFileUndeploy,
  getNodeFileStatus,
} from "@/sdk/gen/node-file-operations-api-node-file-operations/node-file-operations-api-node-file-operations";
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
import { BLOCK_PERMISSIONS } from "@/lib/permissions";
import { BEHAVIOR_BLOCKS, type BehaviorBlockType } from "@/lib/blocks.gen";
import { runBehavior } from "@/lib/behavior-block";
import { BehaviorBlockForm } from "@/components/domain/behavior-block-form";
import {
  Calendar,
  Container,
  Terminal,
  Lock,
  Server,
  Radio,
  Cog,
  Package,
  SlidersHorizontal,
  Clock,
  Globe,
  Tag,
  Power,
  Activity,
  FileText,
  Network,
  GitBranch,
  Users,
  UsersRound,
  ShieldCheck,
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
  "cron-create": Calendar,
  // Software
  "package-list": Package,
  "updatepackages-list": RefreshCw,
  // Config
  "sysctl-list": SlidersHorizontal,
  "ntp-get": Clock,
  "timezone-get": Globe,
  "hostname-get": Tag,
  // System
  "power-reboot": Power,
  "process-list": Activity,
  "log-query": FileText,
  // Networking
  "dns-get": Server,
  "interface-list": Network,
  "route-list": GitBranch,
  ping: Radio,
  // Security
  "user-list": Users,
  "group-list": UsersRound,
  "certificate-list": ShieldCheck,
  // Containers
  "docker-create": Container,
  // Files
  "file-list": FolderOpen,
  "file-deploy": FileDown,
  // Command
  command: Terminal,
  // Node
  "status-get": Monitor,
  "load-get": Gauge,
  "uptime-get": Timer,
  "os-get": Info,
  // Hardware
  "disk-get": HardDrive,
  "memory-get": MemoryStick,
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
    loadBlocks,
    clearBlocks,
  } = useStack();
  const { stacks, activeStack, activeStackId, loadStack, clearActiveStack } =
    useStacks();
  const [applying, setApplying] = useState(false);

  // After Apply, what the run did, counted from the blocks' states.
  const applied = blocks.filter((b) => b.status === "applied").length;
  const failed = blocks.filter((b) => b.status === "error").length;
  const applySummary = hasApplied
    ? `${applied + failed} applied: ${applied} ok${failed ? `, ${failed} failed` : ""}`
    : `${blocks.length} block${blocks.length === 1 ? "" : "s"}`;
  const [showSaveDialog, setShowSaveDialog] = useState(false);
  const [activeGroup, setActiveGroup] = useState("services");

  const handleLoadStack = useCallback(
    (id: string) => {
      loadStack(id);
      const stack = stacks.find((s) => s.id === id);
      if (stack) {
        loadBlocks(stack.blocks);
      }
    },
    [stacks, loadStack, loadBlocks],
  );

  const handleNewStack = useCallback(() => {
    clearActiveStack();
    clearBlocks();
  }, [clearActiveStack, clearBlocks]);
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

  const getUpstreamObjects = (blockIndex: number) => {
    const names: string[] = [];
    for (let i = 0; i < blockIndex; i++) {
      const b = blocks[i];
      if (b.type === "file-upload" && b.data.name) {
        names.push(b.data.name as string);
      }
    }
    return names;
  };

  const handleApply = useCallback(async () => {
    setApplying(true);

    for (const block of blocks) {
      if (block.status !== "ready") continue;

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
          setBlockStatus(block.id, "applied", undefined, result.data);
          await scrollToResult(block.id);
          continue;
        }

        switch (block.type) {
          case "cron-create":
            result = await postNodeSchedule(t, {
              name: block.data.name as string,
              schedule: block.data.schedule as string,
              object: block.data.object as string,
              content_type: ((block.data.content_type as string) || "raw") as
                "raw" | "template",
            });
            break;
          case "cron-list":
            result = await getNodeSchedule(t);
            break;
          case "cron-delete":
            result = await deleteNodeSchedule(t, block.data.name as string);
            break;
          case "cron-get":
            result = await getNodeScheduleByName(t, block.data.name as string);
            break;
          case "cron-update":
            result = await putNodeSchedule(t, block.data.name as string, {
              schedule: block.data.schedule as string,
              object: block.data.object as string,
              content_type:
                (block.data.content_type as "raw" | "template" | undefined) ||
                undefined,
            });
            break;
          case "command": {
            const argsStr = (block.data.args as string) || "";
            result = await postNodeCommandExec(t, {
              command: block.data.command as string,
              args: argsStr ? argsStr.split(" ") : undefined,
              cwd: (block.data.cwd as string) || undefined,
            });
            break;
          }
          case "docker-create": {
            const portsStr = (block.data.ports as string) || "";
            const volsStr = (block.data.volumes as string) || "";
            const envStr = (block.data.env as string) || "";
            const dnsStr = (block.data.dns as string) || "";
            result = await postNodeContainerDocker(t, {
              image: block.data.image as string,
              name: (block.data.name as string) || undefined,
              hostname: (block.data.hostname as string) || undefined,
              ports: portsStr
                ? portsStr.split(",").map((s) => s.trim())
                : undefined,
              volumes: volsStr
                ? volsStr.split(",").map((s) => s.trim())
                : undefined,
              env: envStr ? envStr.split(",").map((s) => s.trim()) : undefined,
              dns: dnsStr ? dnsStr.split(",").map((s) => s.trim()) : undefined,
            });
            break;
          }
          case "docker-list":
            result = await getNodeContainerDocker(t);
            break;
          case "docker-start":
            result = await postNodeContainerDockerStart(
              t,
              block.data.container_id as string,
            );
            break;
          case "docker-stop":
            result = await postNodeContainerDockerStop(
              t,
              block.data.container_id as string,
            );
            break;
          case "docker-delete":
            result = await deleteNodeContainerDockerByID(
              t,
              block.data.container_id as string,
            );
            break;
          case "file-list":
            result = await getFiles();
            break;
          case "file-upload": {
            const file = block.data._file as globalThis.File;
            result = await postFile({
              name: block.data.name as string,
              content_type: ((block.data.content_type as string) || "raw") as
                "raw" | "template",
              file,
            });
            break;
          }
          case "file-deploy": {
            const ct = (block.data.content_type as string) || "raw";
            result = await postNodeFileDeploy(t, {
              object_name: block.data.object_name as string,
              path: block.data.path as string,
              mode: (block.data.mode as string) || "0644",
              owner: (block.data.owner as string) || undefined,
              group: (block.data.group as string) || undefined,
              content_type: ct as "raw" | "template",
            });
            break;
          }
          case "file-undeploy":
            result = await postNodeFileUndeploy(t, {
              path: block.data.path as string,
            });
            break;
          case "file-status":
            result = await getNodeFileStatus(t, {
              path: block.data.path as string,
            });
            break;
          case "file-delete":
            result = await deleteFileByName(block.data.name as string);
            break;
          case "docker-exec": {
            const execCmd = (block.data.command as string).trim().split(/\s+/);
            result = await postNodeContainerDockerExec(
              t,
              block.data.container_id as string,
              { command: execCmd },
            );
            break;
          }
          case "docker-pull":
            result = await postNodeContainerDockerPull(t, {
              image: block.data.image as string,
            });
            break;
          case "docker-rm-image":
            result = await deleteNodeContainerDockerImage(
              t,
              block.data.image as string,
            );
            break;
          case "docker-inspect":
            result = await getNodeContainerDockerByID(
              t,
              block.data.container_id as string,
            );
            break;
          case "file-stale":
            result = await getFileStale();
            break;
          case "command-shell":
            result = await postNodeCommandShell(t, {
              command: block.data.command as string,
            });
            break;
          case "ping":
            result = await postNodeNetworkPing(t, {
              address: block.data.address as string,
            });
            break;
          case "ntp-get":
            result = await getNodeNtp(t);
            break;
          case "ntp-set":
            result = await postNodeNtp(t, {
              servers: (block.data.servers as string)
                .split(",")
                .map((s) => s.trim())
                .filter(Boolean),
            });
            break;
          case "ntp-update":
            result = await putNodeNtp(t, {
              servers: (block.data.servers as string)
                .split(",")
                .map((s) => s.trim())
                .filter(Boolean),
            });
            break;
          case "ntp-delete":
            result = await deleteNodeNtp(t);
            break;
          case "power-reboot":
            result = await postNodePowerReboot(t, {
              ...(block.data.delay ? { delay: Number(block.data.delay) } : {}),
            });
            break;
          case "power-shutdown":
            result = await postNodePowerShutdown(t, {
              ...(block.data.delay ? { delay: Number(block.data.delay) } : {}),
            });
            break;
          case "process-list":
            result = await getNodeProcess(t);
            break;
          case "process-get":
            result = await getNodeProcessByPid(t, Number(block.data.pid));
            break;
          case "process-signal":
            result = await postNodeProcessSignal(t, Number(block.data.pid), {
              signal: block.data
                .signal as import("@/sdk/gen/schemas").ProcessSignalRequestSignal,
            });
            break;
          case "log-query":
            result = await getNodeLog(t, {
              ...(block.data.lines ? { lines: Number(block.data.lines) } : {}),
              ...(block.data.since
                ? { since: block.data.since as string }
                : {}),
              ...(block.data.priority
                ? { priority: block.data.priority as string }
                : {}),
            });
            break;
          case "log-sources":
            result = await getNodeLogSource(t);
            break;
          case "log-query-unit":
            result = await getNodeLogUnit(t, block.data.name as string);
            break;
          case "user-list":
            result = await getNodeUser(t);
            break;
          case "user-get":
            result = await getNodeUserByName(t, block.data.name as string);
            break;
          case "user-create": {
            const groupsStr = (block.data.groups as string) || "";
            result = await postNodeUser(t, {
              name: block.data.name as string,
              shell: (block.data.shell as string) || undefined,
              home: (block.data.home as string) || undefined,
              groups: groupsStr
                ? groupsStr
                    .split(",")
                    .map((s) => s.trim())
                    .filter(Boolean)
                : undefined,
            });
            break;
          }
          case "user-update": {
            const updGroupsStr = (block.data.groups as string) || "";
            result = await putNodeUser(t, block.data.name as string, {
              shell: (block.data.shell as string) || undefined,
              home: (block.data.home as string) || undefined,
              groups: updGroupsStr
                ? updGroupsStr
                    .split(",")
                    .map((s) => s.trim())
                    .filter(Boolean)
                : undefined,
            });
            break;
          }
          case "user-delete":
            result = await deleteNodeUser(t, block.data.name as string);
            break;
          case "user-list-keys":
            result = await getNodeUserSSHKey(t, block.data.name as string);
            break;
          case "user-add-key":
            result = await postNodeUserSSHKey(t, block.data.name as string, {
              key: block.data.key as string,
            });
            break;
          case "user-remove-key":
            result = await deleteNodeUserSSHKey(
              t,
              block.data.name as string,
              block.data.fingerprint as string,
            );
            break;
          case "user-change-password":
            result = await postNodeUserPassword(t, block.data.name as string, {
              password: block.data.password as string,
            });
            break;
          case "group-list":
            result = await getNodeGroup(t);
            break;
          case "group-get":
            result = await getNodeGroupByName(t, block.data.name as string);
            break;
          case "group-create":
            result = await postNodeGroup(t, {
              name: block.data.name as string,
              gid: block.data.gid ? Number(block.data.gid) : undefined,
            });
            break;
          case "group-update": {
            const membersStr = (block.data.members as string) || "";
            result = await putNodeGroup(t, block.data.name as string, {
              members: membersStr
                ? membersStr
                    .split(",")
                    .map((s) => s.trim())
                    .filter(Boolean)
                : undefined,
            });
            break;
          }
          case "group-delete":
            result = await deleteNodeGroup(t, block.data.name as string);
            break;
          case "certificate-list":
            result = await getNodeCertificateCa(t);
            break;
          case "certificate-create":
            result = await postNodeCertificateCa(t, {
              name: block.data.name as string,
              object: block.data.object as string,
            });
            break;
          case "certificate-update":
            result = await putNodeCertificateCa(t, block.data.name as string, {
              object: block.data.object as string,
            });
            break;
          case "certificate-delete":
            result = await deleteNodeCertificateCa(
              t,
              block.data.name as string,
            );
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

        setBlockStatus(block.id, "applied", undefined, result.data);
      } catch (err) {
        setBlockStatus(
          block.id,
          "error",
          err instanceof Error ? err.message : "Unknown error",
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

  const renderBlockForm = (block: (typeof blocks)[0], index: number) => {
    const upstream = getUpstreamObjects(index);

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
      case "cron-create":
        return (
          <ScheduleBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
            upstreamObjects={upstream}
          />
        );
      case "cron-list":
        return null;
      case "cron-delete":
        return (
          <ScheduleDeleteBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
          />
        );
      case "cron-get":
        return (
          <SingleInputBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
            field="name"
            label="Cron Name"
            placeholder="backup-daily"
          />
        );
      case "cron-update":
        return (
          <ScheduleBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
            upstreamObjects={upstream}
          />
        );
      case "command":
        return (
          <CommandBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
          />
        );
      case "docker-create":
        return (
          <DockerBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
          />
        );
      case "docker-list":
        return null;
      case "docker-start":
        return (
          <ContainerActionBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
            action="start"
          />
        );
      case "docker-stop":
        return (
          <ContainerActionBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
            action="stop"
          />
        );
      case "docker-delete":
        return (
          <ContainerActionBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
            action="delete"
          />
        );
      case "file-list":
        return null;
      case "file-upload":
        return (
          <FileUploadBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
          />
        );
      case "file-deploy":
        return (
          <FileBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
            upstreamObjects={upstream}
          />
        );
      case "file-delete":
        return (
          <FileDeleteBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
          />
        );
      case "file-undeploy":
      case "file-status":
        return (
          <SingleInputBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
            field="path"
            label="File Path"
            placeholder="/etc/myapp/config.yaml"
          />
        );
      case "docker-exec":
        return (
          <DockerExecBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
          />
        );
      case "docker-pull":
        return (
          <SingleInputBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
            field="image"
            label="Image"
            placeholder="nginx:latest"
          />
        );
      case "docker-rm-image":
        return (
          <SingleInputBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
            field="image"
            label="Image Name or ID"
            placeholder="nginx:latest"
          />
        );
      case "docker-inspect":
        return (
          <ContainerActionBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
            action="inspect"
          />
        );
      case "file-stale":
        return null;
      case "command-shell":
        return (
          <SingleInputBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
            field="command"
            label="Shell Command"
            placeholder="echo hello | tee /tmp/out.txt"
          />
        );
      case "ping":
        return (
          <SingleInputBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
            field="address"
            label="Host Address"
            placeholder="1.1.1.1 or @fact.custom.gateway"
            facts
          />
        );
      case "ntp-set":
      case "ntp-update":
        return (
          <NtpBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
          />
        );
      case "ntp-get":
      case "ntp-delete":
        return null;
      case "power-reboot":
      case "power-shutdown":
        return (
          <PowerBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
          />
        );
      case "process-list":
      case "log-sources":
        return null;
      case "process-get":
        return (
          <SingleInputBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
            field="pid"
            label="PID"
            placeholder="1234"
          />
        );
      case "process-signal":
        return (
          <ProcessSignalBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
          />
        );
      case "log-query":
        return (
          <LogQueryBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
          />
        );
      case "log-query-unit":
        return (
          <SingleInputBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
            field="name"
            label="Unit Name"
            placeholder="nginx.service"
          />
        );
      case "user-list":
        return null;
      case "user-create":
        return (
          <UserBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
          />
        );
      case "user-update":
        return (
          <UserBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
          />
        );
      case "user-get":
      case "user-delete":
        return (
          <SingleInputBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
            field="name"
            label="Username"
            placeholder="deploy"
          />
        );
      case "user-list-keys":
        return (
          <SingleInputBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
            field="name"
            label="Username"
            placeholder="deploy"
          />
        );
      case "user-add-key":
        return (
          <UserSSHKeyBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
          />
        );
      case "user-remove-key":
        return (
          <UserRemoveKeyBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
          />
        );
      case "user-change-password":
        return (
          <UserPasswordBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
          />
        );
      case "group-list":
        return null;
      case "group-create":
        return (
          <GroupBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
            mode="create"
          />
        );
      case "group-update":
        return (
          <GroupBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
            mode="update"
          />
        );
      case "group-get":
      case "group-delete":
        return (
          <SingleInputBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
            field="name"
            label="Group Name"
            placeholder="docker"
          />
        );
      case "certificate-list":
        return null;
      case "certificate-create":
        return (
          <CertificateBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
            mode="create"
          />
        );
      case "certificate-update":
        return (
          <CertificateBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
            mode="update"
          />
        );
      case "certificate-delete":
        return (
          <SingleInputBlock
            data={block.data}
            onChange={(data) => updateBlockData(block.id, data)}
            onStatusChange={(s: BlockStatus) => setBlockStatus(block.id, s)}
            field="name"
            label="Certificate Name"
            placeholder="my-ca"
          />
        );
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
        if (action === "cron-delete") {
          await deleteNodeSchedule("_all", name);
        } else if (action === "file-delete") {
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
        status={block.status}
        focused={focusedId === block.id}
        error={block.error}
        target={block.target}
        onTargetChange={(t) => setBlockTarget(block.id, t)}
        onRemove={() => removeBlock(block.id)}
        onFocusCard={() => setFocusedId(block.id)}
      >
        {renderBlockForm(block, i)}
      </BlockCard>,
    );
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
        subtitle={
          activeStack
            ? `Editing stack: ${activeStack.name}`
            : "Build a stack by selecting blocks"
        }
      />

      {features.stacks && (
        <StackBar
          stacks={stacks}
          activeStackId={activeStackId}
          onLoad={handleLoadStack}
          onNew={handleNewStack}
        />
      )}

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

          {/* Group pills */}
          <div className="flex flex-wrap gap-1.5">
            {BLOCK_GROUPS.map((group) => (
              <Chip
                key={group.name}
                size="sm"
                active={activeGroup === group.name}
                onClick={() => setActiveGroup(group.name)}
              >
                {group.label}
              </Chip>
            ))}
          </div>

          {/* Categories for active group */}
          {BLOCK_GROUPS.find((g) => g.name === activeGroup)?.categories.map(
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
                      const allowed = perm ? can(perm) : true;
                      return (
                        <Chip
                          key={bt.type}
                          onClick={(e) => {
                            (e.target as HTMLElement).blur();
                            addBlockAndFocus(bt);
                          }}
                          disabled={!allowed}
                          title={allowed ? bt.description : `Requires ${perm}`}
                        >
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
            <EmptyState message="Select a saved stack above or add blocks to build a new one" />
          ) : (
            <>
              <BlockStack>{stackItems}</BlockStack>
              <ApplyButton
                summary={applySummary}
                onClear={clearBlocks}
                ready={canApply}
                applying={applying}
                hasApplied={hasApplied}
                hasBlocks={blocks.length > 0}
                showSave={features.stacks}
                onApply={handleApply}
                onReset={resetBlocks}
                onSave={() => setShowSaveDialog(true)}
              />
              <div className="h-24" />
            </>
          )}
        </main>
      </div>
      <SaveStackDialog
        open={showSaveDialog}
        blockCount={blocks.length}
        onSave={(name, description) => {
          // TODO: POST /stacks — mock no-op for now
          console.log("Saving stack:", { name, description, blocks });
          setShowSaveDialog(false);
        }}
        onClose={() => setShowSaveDialog(false)}
      />
    </ContentArea>
  );
}
