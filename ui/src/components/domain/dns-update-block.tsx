import { useEffect, useRef } from "react";
import { Input } from "@/components/ui/input";
import { FactInput } from "@/components/ui/fact-input";
import type { BlockStatus } from "@/hooks/use-stack";

interface DnsUpdateBlockProps {
  data: Record<string, unknown>;
  onChange: (data: Record<string, unknown>) => void;
  onStatusChange: (status: BlockStatus) => void;
}

export function DnsUpdateBlock({
  data,
  onChange,
  onStatusChange,
}: DnsUpdateBlockProps) {
  const interfaceName = (data.interface_name as string) || "";
  const servers = (data.servers as string) || "";
  const prevStatus = useRef<BlockStatus | null>(null);

  useEffect(() => {
    const next: BlockStatus =
      interfaceName.trim() !== "" && servers.trim() !== ""
        ? "ready"
        : "pending";
    if (next !== prevStatus.current) {
      prevStatus.current = next;
      onStatusChange(next);
    }
  }, [interfaceName, servers, onStatusChange]);

  const update = (field: string, value: string) => {
    onChange({ ...data, [field]: value });
  };

  return (
    <div className="space-y-3">
      <FactInput
        id="dns-update-interface"
        label="Interface Name"
        placeholder="eth0 or @fact.interface.primary"
        value={interfaceName}
        onChange={(v) => update("interface_name", v)}
      />
      <Input
        id="dns-update-servers"
        label="DNS Servers (comma-separated)"
        placeholder="1.1.1.1, 8.8.8.8"
        value={servers}
        onChange={(e) => update("servers", e.target.value)}
      />
    </div>
  );
}
