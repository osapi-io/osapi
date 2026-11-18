import { useEffect, useRef } from "react";
import { Input, Textarea } from "@/components/ui/input";
import type { BlockStatus } from "@/hooks/use-stack";

interface UserSSHKeyBlockProps {
  data: Record<string, unknown>;
  onChange: (data: Record<string, unknown>) => void;
  onStatusChange: (status: BlockStatus) => void;
}

export function UserSSHKeyBlock({
  data,
  onChange,
  onStatusChange,
}: UserSSHKeyBlockProps) {
  const name = (data.name as string) || "";
  const key = (data.key as string) || "";
  const prevStatus = useRef<BlockStatus | null>(null);

  useEffect(() => {
    const next: BlockStatus =
      name.trim() !== "" && key.trim() !== "" ? "ready" : "pending";
    if (next !== prevStatus.current) {
      prevStatus.current = next;
      onStatusChange(next);
    }
  }, [name, key, onStatusChange]);

  const update = (field: string, value: string) => {
    onChange({ ...data, [field]: value });
  };

  return (
    <div className="space-y-3">
      <Input
        id="ssh-key-name"
        label="Username"
        placeholder="deploy"
        value={name}
        onChange={(e) => update("name", e.target.value)}
      />
      <Textarea
        id="ssh-key-key"
        label="Public Key"
        placeholder="ssh-ed25519 AAAA... user@host"
        value={key}
        onChange={(e) => update("key", e.target.value)}
        rows={3}
      />
    </div>
  );
}
