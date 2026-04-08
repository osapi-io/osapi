import { useEffect, useRef } from "react";
import { Input } from "@/components/ui/input";
import type { BlockStatus } from "@/hooks/use-stack";

interface SysctlBlockProps {
  data: Record<string, unknown>;
  onChange: (data: Record<string, unknown>) => void;
  onStatusChange: (status: BlockStatus) => void;
}

export function SysctlBlock({
  data,
  onChange,
  onStatusChange,
}: SysctlBlockProps) {
  const key = (data.key as string) || "";
  const value = (data.value as string) || "";
  const prevStatus = useRef<BlockStatus | null>(null);

  useEffect(() => {
    const next: BlockStatus =
      key.trim() !== "" && value.trim() !== "" ? "ready" : "pending";
    if (next !== prevStatus.current) {
      prevStatus.current = next;
      onStatusChange(next);
    }
  }, [key, value, onStatusChange]);

  const update = (field: string, val: string) => {
    onChange({ ...data, [field]: val });
  };

  return (
    <div className="grid grid-cols-2 gap-3">
      <Input
        id="sysctl-key"
        label="Sysctl Key"
        placeholder="net.ipv4.ip_forward"
        value={key}
        onChange={(e) => update("key", e.target.value)}
      />
      <Input
        id="sysctl-value"
        label="Value"
        placeholder="1"
        value={value}
        onChange={(e) => update("value", e.target.value)}
      />
    </div>
  );
}
