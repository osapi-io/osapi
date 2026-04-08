import { useState, useCallback } from "react";

export interface StackBlock {
  type: string;
  label: string;
  description: string;
  target: string;
  data: Record<string, unknown>;
}

export interface Stack {
  id: string;
  name: string;
  description: string;
  blocks: StackBlock[];
  permissions: string[];
  createdAt: string;
  updatedAt: string;
}

// ---------------------------------------------------------------------------
// Mock data — will be replaced by API calls to /stacks
// ---------------------------------------------------------------------------

const MOCK_STACKS: Stack[] = [
  {
    id: "stack-1",
    name: "Setup Log Rotate Cron",
    description: "Upload logrotate script then create a cron to run it nightly",
    blocks: [
      {
        type: "file-upload",
        label: "Upload",
        description: "Upload to object store",
        target: "_all",
        data: { name: "logrotate.sh", content_type: "raw" },
      },
      {
        type: "cron-create",
        label: "Create",
        description: "Create a cron job",
        target: "_all",
        data: {
          name: "log-rotate",
          schedule: "0 2 * * *",
          object: "logrotate.sh",
          content_type: "script",
        },
      },
    ],
    permissions: ["file:write", "cron:write"],
    createdAt: "2026-03-20T14:30:00Z",
    updatedAt: "2026-03-24T09:15:00Z",
  },
  {
    id: "stack-2",
    name: "Deploy Config File",
    description: "Upload a config template and deploy it to all hosts",
    blocks: [
      {
        type: "file-upload",
        label: "Upload",
        description: "Upload to object store",
        target: "_all",
        data: { name: "app.conf", content_type: "template" },
      },
      {
        type: "file-deploy",
        label: "Deploy",
        description: "Deploy file to hosts",
        target: "_all",
        data: {
          object_name: "app.conf",
          path: "/etc/myapp/app.conf",
          mode: "0644",
          owner: "root",
          group: "root",
          content_type: "template",
        },
      },
      {
        type: "file-status",
        label: "Status",
        description: "Check deploy status",
        target: "_all",
        data: { path: "/etc/myapp/app.conf" },
      },
    ],
    permissions: ["file:write", "node:read"],
    createdAt: "2026-03-18T10:00:00Z",
    updatedAt: "2026-03-25T11:30:00Z",
  },
  {
    id: "stack-3",
    name: "Health Check",
    description: "Check uptime, disk, and memory across all hosts",
    blocks: [
      {
        type: "command",
        label: "Execute",
        description: "Run a command on hosts",
        target: "_all",
        data: { command: "uptime" },
      },
      {
        type: "command",
        label: "Execute",
        description: "Run a command on hosts",
        target: "_all",
        data: { command: "df", args: "-h /" },
      },
      {
        type: "command",
        label: "Execute",
        description: "Run a command on hosts",
        target: "_all",
        data: { command: "free", args: "-m" },
      },
    ],
    permissions: ["command:execute"],
    createdAt: "2026-03-15T08:00:00Z",
    updatedAt: "2026-03-25T11:30:00Z",
  },
  {
    id: "stack-4",
    name: "List All Resources",
    description: "List files, crons, and containers on all hosts",
    blocks: [
      {
        type: "file-list",
        label: "List",
        description: "List files in object store",
        target: "_all",
        data: {},
      },
      {
        type: "cron-list",
        label: "List",
        description: "List cron jobs",
        target: "_all",
        data: {},
      },
      {
        type: "docker-list",
        label: "List",
        description: "List containers",
        target: "_all",
        data: {},
      },
    ],
    permissions: ["file:read", "cron:read", "docker:read"],
    createdAt: "2026-03-12T09:00:00Z",
    updatedAt: "2026-03-12T09:00:00Z",
  },
  {
    id: "stack-5",
    name: "Setup Backup Cron",
    description: "Upload backup script and schedule it to run daily",
    blocks: [
      {
        type: "file-upload",
        label: "Upload",
        description: "Upload to object store",
        target: "_all",
        data: { name: "backup.sh", content_type: "raw" },
      },
      {
        type: "cron-create",
        label: "Create",
        description: "Create a cron job",
        target: "_all",
        data: {
          name: "daily-backup",
          schedule: "0 3 * * *",
          object: "backup.sh",
          content_type: "script",
        },
      },
      {
        type: "cron-list",
        label: "List",
        description: "List cron jobs",
        target: "_all",
        data: {},
      },
    ],
    permissions: ["file:write", "cron:write", "cron:read"],
    createdAt: "2026-03-22T16:45:00Z",
    updatedAt: "2026-03-22T16:45:00Z",
  },
  {
    id: "stack-6",
    name: "Deploy and Verify File",
    description: "Upload, deploy, then verify the file landed correctly",
    blocks: [
      {
        type: "file-upload",
        label: "Upload",
        description: "Upload to object store",
        target: "_all",
        data: { name: "motd", content_type: "raw" },
      },
      {
        type: "file-deploy",
        label: "Deploy",
        description: "Deploy file to hosts",
        target: "_all",
        data: {
          object_name: "motd",
          path: "/etc/motd",
          mode: "0644",
          content_type: "raw",
        },
      },
      {
        type: "command",
        label: "Execute",
        description: "Run a command on hosts",
        target: "_all",
        data: { command: "cat", args: "/etc/motd" },
      },
    ],
    permissions: ["file:write", "command:execute"],
    createdAt: "2026-03-19T15:30:00Z",
    updatedAt: "2026-03-25T08:00:00Z",
  },
  {
    id: "stack-7",
    name: "Remove Cron and Script",
    description:
      "Delete a cron job then clean up its script from the object store",
    blocks: [
      {
        type: "cron-delete",
        label: "Delete",
        description: "Delete a cron job",
        target: "_all",
        data: { name: "log-rotate" },
      },
      {
        type: "file-delete",
        label: "Delete",
        description: "Delete file from object store",
        target: "_all",
        data: { name: "logrotate.sh" },
      },
    ],
    permissions: ["cron:write", "file:write"],
    createdAt: "2026-03-24T10:00:00Z",
    updatedAt: "2026-03-26T07:45:00Z",
  },
  {
    id: "stack-8",
    name: "Undeploy and Clean Up",
    description:
      "Remove a deployed file from hosts then delete from object store",
    blocks: [
      {
        type: "file-undeploy",
        label: "Undeploy",
        description: "Remove deployed file",
        target: "_all",
        data: { path: "/etc/myapp/app.conf" },
      },
      {
        type: "file-delete",
        label: "Delete",
        description: "Delete file from object store",
        target: "_all",
        data: { name: "app.conf" },
      },
    ],
    permissions: ["file:write"],
    createdAt: "2026-03-10T12:00:00Z",
    updatedAt: "2026-03-23T14:20:00Z",
  },
];

export function useStacks() {
  const [stacks] = useState<Stack[]>(MOCK_STACKS);
  const [activeStackId, setActiveStackId] = useState<string | null>(null);

  const activeStack = stacks.find((s) => s.id === activeStackId) ?? null;

  const loadStack = useCallback((id: string) => {
    setActiveStackId(id);
  }, []);

  const clearActiveStack = useCallback(() => {
    setActiveStackId(null);
  }, []);

  const saveStack = useCallback(
    (_name: string, _description: string, _blocks: StackBlock[]) => {
      // TODO: POST /stacks — mock no-op for now
    },
    [],
  );

  return {
    stacks,
    activeStack,
    activeStackId,
    loadStack,
    clearActiveStack,
    saveStack,
  };
}
