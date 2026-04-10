import { useEffect, useRef } from "react";
import { Input } from "@/components/ui/input";
import type { BlockStatus } from "@/hooks/use-stack";

interface PowerBlockProps {
  data: Record<string, unknown>;
  onChange: (data: Record<string, unknown>) => void;
  onStatusChange: (status: BlockStatus) => void;
}

export function PowerBlock({
  data,
  onChange,
  onStatusChange,
}: PowerBlockProps) {
  const delay = (data.delay as string) || "";
  const prevStatus = useRef<BlockStatus | null>(null);

  useEffect(() => {
    const next: BlockStatus = "ready";
    if (next !== prevStatus.current) {
      prevStatus.current = next;
      onStatusChange(next);
    }
  }, [onStatusChange]);

  return (
    <Input
      id="power-delay"
      label="Delay (seconds, optional)"
      placeholder="0"
      value={delay}
      onChange={(e) => onChange({ ...data, delay: e.target.value })}
    />
  );
}
