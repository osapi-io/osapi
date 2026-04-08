import { useEffect, useRef } from "react";
import { Input } from "@/components/ui/input";
import type { BlockStatus } from "@/hooks/use-stack";

interface CommandBlockProps {
  data: Record<string, unknown>;
  onChange: (data: Record<string, unknown>) => void;
  onStatusChange: (status: BlockStatus) => void;
}

export function CommandBlock({
  data,
  onChange,
  onStatusChange,
}: CommandBlockProps) {
  const command = (data.command as string) || "";
  const args = (data.args as string) || "";
  const cwd = (data.cwd as string) || "";
  const prevStatus = useRef<BlockStatus | null>(null);

  useEffect(() => {
    const next: BlockStatus = command.trim() !== "" ? "ready" : "pending";
    if (next !== prevStatus.current) {
      prevStatus.current = next;
      onStatusChange(next);
    }
  }, [command, onStatusChange]);

  const update = (field: string, value: string) => {
    onChange({ ...data, [field]: value });
  };

  return (
    <div className="space-y-3">
      <div className="grid grid-cols-2 gap-3">
        <Input
          id="cmd-command"
          label="Command"
          placeholder="ls"
          value={command}
          onChange={(e) => update("command", e.target.value)}
        />
        <Input
          id="cmd-args"
          label="Arguments (space-separated)"
          placeholder="-la /tmp"
          value={args}
          onChange={(e) => update("args", e.target.value)}
        />
      </div>
      <Input
        id="cmd-cwd"
        label="Working Directory (optional)"
        placeholder="/tmp"
        value={cwd}
        onChange={(e) => update("cwd", e.target.value)}
      />
    </div>
  );
}
