import { useEffect, useRef } from "react";
import { Input } from "@/components/ui/input";
import type { BlockStatus } from "@/hooks/use-stack";

interface LogQueryBlockProps {
  data: Record<string, unknown>;
  onChange: (data: Record<string, unknown>) => void;
  onStatusChange: (status: BlockStatus) => void;
}

export function LogQueryBlock({
  data,
  onChange,
  onStatusChange,
}: LogQueryBlockProps) {
  const lines = (data.lines as string) || "";
  const since = (data.since as string) || "";
  const priority = (data.priority as string) || "";
  const prevStatus = useRef<BlockStatus | null>(null);

  useEffect(() => {
    const next: BlockStatus = "ready";
    if (next !== prevStatus.current) {
      prevStatus.current = next;
      onStatusChange(next);
    }
  }, [onStatusChange]);

  const update = (field: string, value: string) => {
    onChange({ ...data, [field]: value });
  };

  return (
    <div className="grid grid-cols-3 gap-3">
      <Input
        id="log-lines"
        label="Lines"
        placeholder="100"
        value={lines}
        onChange={(e) => update("lines", e.target.value)}
      />
      <Input
        id="log-since"
        label="Since"
        placeholder="1h"
        value={since}
        onChange={(e) => update("since", e.target.value)}
      />
      <Input
        id="log-priority"
        label="Priority"
        placeholder="err"
        value={priority}
        onChange={(e) => update("priority", e.target.value)}
      />
    </div>
  );
}
