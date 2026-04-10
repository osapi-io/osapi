import { useEffect, useRef } from "react";
import { Input } from "@/components/ui/input";
import type { BlockStatus } from "@/hooks/use-stack";

interface PackageBlockProps {
  data: Record<string, unknown>;
  onChange: (data: Record<string, unknown>) => void;
  onStatusChange: (status: BlockStatus) => void;
}

export function PackageBlock({
  data,
  onChange,
  onStatusChange,
}: PackageBlockProps) {
  const name = (data.name as string) || "";
  const version = (data.version as string) || "";
  const prevStatus = useRef<BlockStatus | null>(null);

  useEffect(() => {
    const next: BlockStatus = name.trim() !== "" ? "ready" : "pending";
    if (next !== prevStatus.current) {
      prevStatus.current = next;
      onStatusChange(next);
    }
  }, [name, onStatusChange]);

  const update = (field: string, value: string) => {
    onChange({ ...data, [field]: value });
  };

  return (
    <div className="grid grid-cols-2 gap-3">
      <Input
        id="pkg-name"
        label="Package Name"
        placeholder="nginx"
        value={name}
        onChange={(e) => update("name", e.target.value)}
      />
      <Input
        id="pkg-version"
        label="Version (optional)"
        placeholder="1.24.0"
        value={version}
        onChange={(e) => update("version", e.target.value)}
      />
    </div>
  );
}
