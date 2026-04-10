import { useEffect, useRef } from "react";
import { Input } from "@/components/ui/input";
import { Text } from "@/components/ui/text";
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
      <div className="space-y-1">
        <Text variant="label" size="xs">
          Public Key
        </Text>
        <textarea
          id="ssh-key-key"
          placeholder="ssh-ed25519 AAAA... user@host"
          value={key}
          onChange={(e) => update("key", e.target.value)}
          rows={3}
          className="w-full rounded-md border border-border bg-card px-3 py-2 text-sm text-text placeholder:text-text-muted/40 focus:border-primary/40 focus:outline-none"
        />
      </div>
    </div>
  );
}
