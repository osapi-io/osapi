import { useEffect, useRef } from "react";
import { ContainerPicker } from "@/components/domain/container-picker";
import { Input } from "@/components/ui/input";
import type { BlockStatus } from "@/hooks/use-stack";

interface DockerExecBlockProps {
  data: Record<string, unknown>;
  onChange: (data: Record<string, unknown>) => void;
  onStatusChange: (status: BlockStatus) => void;
}

export function DockerExecBlock({
  data,
  onChange,
  onStatusChange,
}: DockerExecBlockProps) {
  const containerId = (data.container_id as string) || "";
  const command = (data.command as string) || "";
  const prevStatus = useRef<BlockStatus | null>(null);

  useEffect(() => {
    const next: BlockStatus =
      containerId.trim() !== "" && command.trim() !== "" ? "ready" : "pending";
    if (next !== prevStatus.current) {
      prevStatus.current = next;
      onStatusChange(next);
    }
  }, [containerId, command, onStatusChange]);

  return (
    <div className="space-y-3">
      <ContainerPicker
        id="docker-exec-container"
        label="Container to exec in"
        value={containerId}
        onChange={(v) => onChange({ ...data, container_id: v })}
      />
      <Input
        id="docker-exec-command"
        label="Command (space-separated argv)"
        placeholder="sh -c 'echo hello'"
        value={command}
        onChange={(e) => onChange({ ...data, command: e.target.value })}
      />
    </div>
  );
}
