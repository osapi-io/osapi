import { useEffect, useRef } from "react";
import { Input } from "@/components/ui/input";
import type { BlockStatus } from "@/hooks/use-stack";

interface UserBlockProps {
  data: Record<string, unknown>;
  onChange: (data: Record<string, unknown>) => void;
  onStatusChange: (status: BlockStatus) => void;
}

export function UserBlock({ data, onChange, onStatusChange }: UserBlockProps) {
  const name = (data.name as string) || "";
  const shell = (data.shell as string) || "";
  const home = (data.home as string) || "";
  const groups = (data.groups as string) || "";
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

  return (
    <div className="space-y-3">
      <div className="grid grid-cols-2 gap-3">
        <Input
          id="user-name"
          label="Username"
          placeholder="deploy"
          value={name}
          onChange={(e) => update("name", e.target.value)}
        />
        <Input
          id="user-shell"
          label="Shell (optional)"
          placeholder="/bin/bash"
          value={shell}
          onChange={(e) => update("shell", e.target.value)}
        />
      </div>
      <div className="grid grid-cols-2 gap-3">
        <Input
          id="user-home"
          label="Home Directory (optional)"
          placeholder="/home/deploy"
          value={home}
          onChange={(e) => update("home", e.target.value)}
        />
        <Input
          id="user-groups"
          label="Supplementary Groups (comma-separated)"
          placeholder="sudo,docker"
          value={groups}
          onChange={(e) => update("groups", e.target.value)}
        />
      </div>
    </div>
  );
}
