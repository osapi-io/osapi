import { useEffect, useRef } from "react";
import { Input } from "@/components/ui/input";
import { FactInput } from "@/components/ui/fact-input";
import type { BlockStatus } from "@/hooks/use-stack";

interface RouteBlockProps {
  data: Record<string, unknown>;
  onChange: (data: Record<string, unknown>) => void;
  onStatusChange: (status: BlockStatus) => void;
}

export function RouteBlock({
  data,
  onChange,
  onStatusChange,
}: RouteBlockProps) {
  const interfaceName = (data.interface_name as string) || "";
  const to = (data.to as string) || "";
  const via = (data.via as string) || "";
  const metric = (data.metric as string) || "";
  const prevStatus = useRef<BlockStatus | null>(null);

  useEffect(() => {
    const next: BlockStatus = interfaceName.trim() !== "" ? "ready" : "pending";
    if (next !== prevStatus.current) {
      prevStatus.current = next;
      onStatusChange(next);
    }
  }, [interfaceName, onStatusChange]);

  const update = (field: string, value: string) => {
    onChange({ ...data, [field]: value });
  };

  return (
    <div className="space-y-3">
      <FactInput
        id="route-interface"
        label="Interface Name"
        placeholder="eth0 or @fact.interface.primary"
        value={interfaceName}
        onChange={(v) => update("interface_name", v)}
      />
      <div className="grid grid-cols-3 gap-3">
        <Input
          id="route-to"
          label="Destination (to)"
          placeholder="10.0.0.0/8"
          value={to}
          onChange={(e) => update("to", e.target.value)}
        />
        <Input
          id="route-via"
          label="Gateway (via)"
          placeholder="192.168.1.1"
          value={via}
          onChange={(e) => update("via", e.target.value)}
        />
        <Input
          id="route-metric"
          label="Metric (optional)"
          placeholder="100"
          value={metric}
          onChange={(e) => update("metric", e.target.value)}
        />
      </div>
    </div>
  );
}
