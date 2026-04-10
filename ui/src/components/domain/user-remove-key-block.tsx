import { useEffect, useRef } from "react";
import { Input } from "@/components/ui/input";
import type { BlockStatus } from "@/hooks/use-stack";

interface UserRemoveKeyBlockProps {
  data: Record<string, unknown>;
  onChange: (data: Record<string, unknown>) => void;
  onStatusChange: (status: BlockStatus) => void;
}

export function UserRemoveKeyBlock({
  data,
  onChange,
  onStatusChange,
}: UserRemoveKeyBlockProps) {
  const name = (data.name as string) || "";
  const fingerprint = (data.fingerprint as string) || "";
  const prevStatus = useRef<BlockStatus | null>(null);

  useEffect(() => {
    const next: BlockStatus =
      name.trim() !== "" && fingerprint.trim() !== "" ? "ready" : "pending";
    if (next !== prevStatus.current) {
      prevStatus.current = next;
      onStatusChange(next);
    }
  }, [name, fingerprint, onStatusChange]);

  const update = (field: string, value: string) => {
    onChange({ ...data, [field]: value });
  };

  return (
    <div className="grid grid-cols-2 gap-3">
      <Input
        id="remove-key-name"
        label="Username"
        placeholder="deploy"
        value={name}
        onChange={(e) => update("name", e.target.value)}
      />
      <Input
        id="remove-key-fingerprint"
        label="Key Fingerprint"
        placeholder="SHA256:abc123..."
        value={fingerprint}
        onChange={(e) => update("fingerprint", e.target.value)}
      />
    </div>
  );
}
