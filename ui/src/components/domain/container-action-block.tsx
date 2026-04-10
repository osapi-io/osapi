import { useEffect, useRef } from "react";
import { ContainerPicker } from "@/components/domain/container-picker";
import type { BlockStatus } from "@/hooks/use-stack";

interface ContainerActionBlockProps {
  data: Record<string, unknown>;
  onChange: (data: Record<string, unknown>) => void;
  onStatusChange: (status: BlockStatus) => void;
  action: string;
}

export function ContainerActionBlock({
  data,
  onChange,
  onStatusChange,
  action,
}: ContainerActionBlockProps) {
  const containerId = (data.container_id as string) || "";
  const prevStatus = useRef<BlockStatus | null>(null);

  useEffect(() => {
    const next: BlockStatus = containerId.trim() !== "" ? "ready" : "pending";
    if (next !== prevStatus.current) {
      prevStatus.current = next;
      onStatusChange(next);
    }
  }, [containerId, onStatusChange]);

  return (
    <ContainerPicker
      id={`docker-${action}-container`}
      label={`Container to ${action}`}
      value={containerId}
      onChange={(v) => onChange({ ...data, container_id: v })}
    />
  );
}
