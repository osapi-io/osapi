import { useEffect, useRef } from "react";
import { SchedulePicker } from "@/components/domain/schedule-picker";
import type { BlockStatus } from "@/hooks/use-stack";

interface ScheduleDeleteBlockProps {
  data: Record<string, unknown>;
  onChange: (data: Record<string, unknown>) => void;
  onStatusChange: (status: BlockStatus) => void;
}

export function ScheduleDeleteBlock({
  data,
  onChange,
  onStatusChange,
}: ScheduleDeleteBlockProps) {
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
    <SchedulePicker
      id="cron-delete-name"
      label="Cron Entry to Delete"
      value={name}
      onChange={(v) => onChange({ ...data, name: v })}
    />
  );
}
