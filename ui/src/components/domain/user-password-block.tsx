import { useEffect, useRef } from "react";
import { Input } from "@/components/ui/input";
import type { BlockStatus } from "@/hooks/use-stack";

interface UserPasswordBlockProps {
  data: Record<string, unknown>;
  onChange: (data: Record<string, unknown>) => void;
  onStatusChange: (status: BlockStatus) => void;
}

export function UserPasswordBlock({
  data,
  onChange,
  onStatusChange,
}: UserPasswordBlockProps) {
  const name = (data.name as string) || "";
  const password = (data.password as string) || "";
  const prevStatus = useRef<BlockStatus | null>(null);

  useEffect(() => {
    const next: BlockStatus =
      name.trim() !== "" && password.trim() !== "" ? "ready" : "pending";
    if (next !== prevStatus.current) {
      prevStatus.current = next;
      onStatusChange(next);
    }
  }, [name, password, onStatusChange]);

  const update = (field: string, value: string) => {
    onChange({ ...data, [field]: value });
  };

  return (
    <div className="grid grid-cols-2 gap-3">
      <Input
        id="user-password-name"
        label="Username"
        placeholder="deploy"
        value={name}
        onChange={(e) => update("name", e.target.value)}
      />
      <Input
        id="user-password-password"
        label="New Password"
        placeholder="••••••••"
        type="password"
        value={password}
        onChange={(e) => update("password", e.target.value)}
      />
    </div>
  );
}
