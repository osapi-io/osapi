import { useEffect, useRef } from "react";
import { Input } from "@/components/ui/input";
import type { BlockStatus } from "@/hooks/use-stack";

interface GroupBlockProps {
  data: Record<string, unknown>;
  onChange: (data: Record<string, unknown>) => void;
  onStatusChange: (status: BlockStatus) => void;
  /** When true, shows member management fields for update operations */
  mode?: "create" | "update";
}

export function GroupBlock({
  data,
  onChange,
  onStatusChange,
  mode = "create",
}: GroupBlockProps) {
  const name = (data.name as string) || "";
  const gid = (data.gid as string) || "";
  const members = (data.members as string) || "";
  const prevStatus = useRef<BlockStatus | null>(null);

  useEffect(() => {
    const next: BlockStatus = name.trim() !== "" ? "ready" : "pending";
    if (next !== prevStatus.current) {
      prevStatus.current = next;
      onStatusChange(next);
    }
  }, [name, onStatusChange]);

  const update = (field: string, value: string) => {
    onChange({ ...data, [field]: value });
  };

  if (mode === "update") {
    return (
      <div className="grid grid-cols-2 gap-3">
        <Input
          id="group-name"
          label="Group Name"
          placeholder="docker"
          value={name}
          onChange={(e) => update("name", e.target.value)}
        />
        <Input
          id="group-members"
          label="Members (comma-separated)"
          placeholder="alice,bob"
          value={members}
          onChange={(e) => update("members", e.target.value)}
        />
      </div>
    );
  }

  return (
    <div className="grid grid-cols-2 gap-3">
      <Input
        id="group-name"
        label="Group Name"
        placeholder="docker"
        value={name}
        onChange={(e) => update("name", e.target.value)}
      />
      <Input
        id="group-gid"
        label="GID (optional)"
        placeholder="1001"
        value={gid}
        onChange={(e) => update("gid", e.target.value)}
      />
    </div>
  );
}
