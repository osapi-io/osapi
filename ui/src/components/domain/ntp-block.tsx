import { useEffect, useRef } from "react";
import { Input } from "@/components/ui/input";
import type { BlockStatus } from "@/hooks/use-stack";

interface NtpBlockProps {
  data: Record<string, unknown>;
  onChange: (data: Record<string, unknown>) => void;
  onStatusChange: (status: BlockStatus) => void;
}

export function NtpBlock({ data, onChange, onStatusChange }: NtpBlockProps) {
  const servers = (data.servers as string) || "";
  const prevStatus = useRef<BlockStatus | null>(null);

  useEffect(() => {
    const next: BlockStatus = servers.trim() !== "" ? "ready" : "pending";
    if (next !== prevStatus.current) {
      prevStatus.current = next;
      onStatusChange(next);
    }
  }, [servers, onStatusChange]);

  return (
    <Input
      id="ntp-servers"
      label="NTP Servers"
      placeholder="0.pool.ntp.org, 1.pool.ntp.org"
      value={servers}
      onChange={(e) => onChange({ ...data, servers: e.target.value })}
    />
  );
}
