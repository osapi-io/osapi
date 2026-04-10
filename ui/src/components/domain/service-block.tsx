import { useEffect, useRef } from "react";
import { Input } from "@/components/ui/input";
import { ObjectPicker } from "@/components/domain/object-picker";
import type { BlockStatus } from "@/hooks/use-stack";

interface ServiceBlockProps {
  data: Record<string, unknown>;
  onChange: (data: Record<string, unknown>) => void;
  onStatusChange: (status: BlockStatus) => void;
  upstreamObjects?: string[];
}

export function ServiceBlock({
  data,
  onChange,
  onStatusChange,
  upstreamObjects = [],
}: ServiceBlockProps) {
  const name = (data.name as string) || "";
  const object = (data.object as string) || "";
  const prevStatus = useRef<BlockStatus | null>(null);

  useEffect(() => {
    const isReady = name.trim() !== "" && object.trim() !== "";
    const next: BlockStatus = isReady ? "ready" : "pending";
    if (next !== prevStatus.current) {
      prevStatus.current = next;
      onStatusChange(next);
    }
  }, [name, object, onStatusChange]);

  const update = (field: string, value: string) => {
    onChange({ ...data, [field]: value });
  };

  return (
    <div className="space-y-3">
      <div className="grid grid-cols-2 gap-3">
        <Input
          id="service-name"
          label="Unit Name"
          placeholder="myapp.service"
          value={name}
          onChange={(e) => update("name", e.target.value)}
        />
        <ObjectPicker
          id="service-object"
          label="Unit File Object"
          value={object}
          onChange={(v) => update("object", v)}
          upstreamObjects={upstreamObjects}
        />
      </div>
    </div>
  );
}
