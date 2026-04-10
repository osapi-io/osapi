import { useEffect, useRef } from "react";
import { CronPicker } from "@/components/domain/cron-picker";
import type { BlockStatus } from "@/hooks/use-stack";

interface CronDeleteBlockProps {
  data: Record<string, unknown>;
  onChange: (data: Record<string, unknown>) => void;
  onStatusChange: (status: BlockStatus) => void;
}

export function CronDeleteBlock({
  data,
  onChange,
  onStatusChange,
}: CronDeleteBlockProps) {
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
    <CronPicker
      id="cron-delete-name"
      label="Cron Entry to Delete"
      value={name}
      onChange={(v) => onChange({ ...data, name: v })}
    />
  );
}
