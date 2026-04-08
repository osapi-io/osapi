import { useEffect, useRef } from "react";
import { Input } from "@/components/ui/input";
import type { BlockStatus } from "@/hooks/use-stack";

interface ProcessSignalBlockProps {
  data: Record<string, unknown>;
  onChange: (data: Record<string, unknown>) => void;
  onStatusChange: (status: BlockStatus) => void;
}

export function ProcessSignalBlock({
  data,
  onChange,
  onStatusChange,
}: ProcessSignalBlockProps) {
  const pid = (data.pid as string) || "";
  const signal = (data.signal as string) || "";
  const prevStatus = useRef<BlockStatus | null>(null);

  useEffect(() => {
    const next: BlockStatus =
      pid.trim() !== "" && signal.trim() !== "" ? "ready" : "pending";
    if (next !== prevStatus.current) {
      prevStatus.current = next;
      onStatusChange(next);
    }
  }, [pid, signal, onStatusChange]);

  const update = (field: string, value: string) => {
    onChange({ ...data, [field]: value });
  };

  return (
    <div className="grid grid-cols-2 gap-3">
      <Input
        id="process-pid"
        label="PID"
        placeholder="1234"
        value={pid}
        onChange={(e) => update("pid", e.target.value)}
      />
      <Input
        id="process-signal"
        label="Signal"
        placeholder="SIGTERM"
        value={signal}
        onChange={(e) => update("signal", e.target.value)}
      />
    </div>
  );
}
