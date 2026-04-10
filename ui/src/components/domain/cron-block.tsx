import { useEffect, useRef } from "react";
import { Input } from "@/components/ui/input";
import { Dropdown } from "@/components/ui/dropdown";
import { ObjectPicker } from "@/components/domain/object-picker";
import type { BlockStatus } from "@/hooks/use-stack";

interface CronBlockProps {
  data: Record<string, unknown>;
  onChange: (data: Record<string, unknown>) => void;
  onStatusChange: (status: BlockStatus) => void;
  upstreamObjects?: string[];
}

export function CronBlock({
  data,
  onChange,
  onStatusChange,
  upstreamObjects = [],
}: CronBlockProps) {
  const name = (data.name as string) || "";
  const schedule = (data.schedule as string) || "";
  const object = (data.object as string) || "";
  const contentType = (data.content_type as string) || "script";
  const prevStatus = useRef<BlockStatus | null>(null);

  useEffect(() => {
    const isReady =
      name.trim() !== "" && schedule.trim() !== "" && object.trim() !== "";
    const next: BlockStatus = isReady ? "ready" : "pending";
    if (next !== prevStatus.current) {
      prevStatus.current = next;
      onStatusChange(next);
    }
  }, [name, schedule, object, onStatusChange]);

  const update = (field: string, value: string) => {
    onChange({ ...data, [field]: value });
  };

  return (
    <div className="space-y-3">
      <div className="grid grid-cols-2 gap-3">
        <Input
          id="cron-name"
          label="Name"
          placeholder="backup-daily"
          value={name}
          onChange={(e) => update("name", e.target.value)}
        />
        <Input
          id="cron-schedule"
          label="Schedule"
          placeholder="0 3 * * *"
          value={schedule}
          onChange={(e) => update("schedule", e.target.value)}
        />
      </div>
      <div className="grid grid-cols-2 gap-3">
        <ObjectPicker
          id="cron-object"
          label="Object"
          value={object}
          onChange={(v) => update("object", v)}
          upstreamObjects={upstreamObjects}
        />
        <Dropdown
          id="cron-content-type"
          label="Content Type"
          value={contentType}
          onChange={(v) => update("content_type", v)}
          options={[
            { value: "script", label: "Script" },
            { value: "cron-drop-in", label: "Cron Drop-in" },
          ]}
        />
      </div>
    </div>
  );
}
