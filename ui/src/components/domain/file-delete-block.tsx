import { useEffect, useRef } from "react";
import { ObjectPicker } from "@/components/domain/object-picker";
import type { BlockStatus } from "@/hooks/use-stack";

interface FileDeleteBlockProps {
  data: Record<string, unknown>;
  onChange: (data: Record<string, unknown>) => void;
  onStatusChange: (status: BlockStatus) => void;
}

export function FileDeleteBlock({
  data,
  onChange,
  onStatusChange,
}: FileDeleteBlockProps) {
  const name = (data.name as string) || "";
  const prevStatus = useRef<BlockStatus | null>(null);

  useEffect(() => {
    const next: BlockStatus = name.trim() !== "" ? "ready" : "pending";
    if (next !== prevStatus.current) {
      prevStatus.current = next;
      onStatusChange(next);
    }
  }, [name, onStatusChange]);

  return (
    <ObjectPicker
      id="file-delete-name"
      label="Object to Delete"
      value={name}
      onChange={(v) => onChange({ ...data, name: v })}
    />
  );
}
